package overflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/cmd"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/contracts"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// OverflowState

// OverflowState contains information about how to Overflow is confitured and the current runnig state
type OverflowState struct {

	//State is the current state of the configured overflow instance
	State *flowkit.State

	//the services from flowkit to performed operations on
	Services *services.Services

	//Configured variables that are taken from the builder since we need them in the execution of overflow later on
	Network                      string
	PrependNetworkToAccountNames bool
	ServiceAccountSuffix         string
	Gas                          int

	//flowkit, emulator and emulator debug log uses three different logging technologies so we have them all stored here
	//this flowkit Logger can go away when we can remove deprecations!
	Logger   output.Logger
	Log      *bytes.Buffer
	LogLevel int

	//https://github.com/bjartek/overflow/issues/45
	//This is not populated with anything yet since the emulator version that has this change is not in mainline yet
	EmulatorLog *bytes.Buffer

	//If there was an error starting overflow it is stored here
	Error error

	//Paths that points to where .cdc files are stored and the posibilty to specify something besides the standard `transactions`/`scripts`subdirectories
	BasePath            string
	TransactionBasePath string
	ScriptBasePath      string

	//Filters to events to remove uneeded noise
	FilterOutFeeEvents                  bool
	FilterOutEmptyWithDrawDepositEvents bool
	GlobalEventFilter                   OverflowEventFilter

	//Signal to overflow that if there is an error after running a single interaction we should panic
	StopOnError bool

	//Signal to overflow that if this is not nil we should print events on interaction completion
	PrintOptions *[]OverflowPrinterOption

	//Mint this amount of flow to new accounts
	NewUserFlowAmount float64
}

func (o *OverflowState) parseArguments(fileName string, code []byte, inputArgs map[string]interface{}) ([]cadence.Value, error) {
	var resultArgs []cadence.Value = make([]cadence.Value, 0)

	codes := map[common.Location]string{}
	location := common.StringLocation(fileName)
	program, must := cmd.PrepareProgram(string(code), location, codes)
	checker, _ := cmd.PrepareChecker(program, location, codes, nil, must)

	var parameterList []*ast.Parameter

	functionDeclaration := sema.FunctionEntryPointDeclaration(program)
	if functionDeclaration != nil {
		if functionDeclaration.ParameterList != nil {
			parameterList = functionDeclaration.ParameterList.Parameters
		}
	}

	transactionDeclaration := program.TransactionDeclarations()
	if len(transactionDeclaration) == 1 {
		if transactionDeclaration[0].ParameterList != nil {
			parameterList = transactionDeclaration[0].ParameterList.Parameters
		}
	}

	if parameterList == nil {
		return resultArgs, nil
	}

	argumentNotPresent := []string{}
	argumentNames := []string{}
	args := []interface{}{}
	for _, parameter := range parameterList {
		parameterName := parameter.Identifier.Identifier
		value, ok := inputArgs[parameterName]
		if !ok {
			argumentNotPresent = append(argumentNotPresent, parameterName)
		} else {
			argumentNames = append(argumentNames, parameterName)
			args = append(args, value)
		}
	}

	if len(argumentNotPresent) > 0 {
		err := fmt.Errorf("the transaction '%s' is missing %v", fileName, argumentNotPresent)
		return nil, err
	}

	redundantArgument := []string{}
	for inputKey, _ := range inputArgs {
		//If your IDE complains about this it is wrong, this is 1.18 generics not suported anywhere
		if !slices.Contains(argumentNames, inputKey) {
			redundantArgument = append(redundantArgument, inputKey)
		}
	}

	if len(redundantArgument) > 0 {
		err := fmt.Errorf("the transaction '%s' has the following extra arguments %v", fileName, redundantArgument)
		return nil, err
	}

	for index, argument := range args {

		cadenceVal, isCadenceValue := argument.(cadence.Value)
		if isCadenceValue {
			resultArgs = append(resultArgs, cadenceVal)
			continue
		}

		var argumentString string
		switch a := argument.(type) {
		case string:
			argumentString = a
		case float64:
			argumentString = fmt.Sprintf("%f", a)
		default:
			argumentString = fmt.Sprintf("%v", argument)

		}
		astType := parameterList[index].TypeAnnotation.Type
		semaType := checker.ConvertType(astType)

		switch semaType {
		case sema.StringType:
			if len(argumentString) > 0 && !strings.HasPrefix(argumentString, "\"") {
				argumentString = "\"" + argumentString + "\""
			}
		}

		switch semaType.(type) {
		case *sema.AddressType:

			account, _ := o.AccountE(argumentString)

			if account != nil {
				argumentString = account.Address().String()
			}

			if !strings.Contains(argumentString, "0x") {
				argumentString = fmt.Sprintf("0x%s", argumentString)
			}
		}

		var value, err = runtime.ParseLiteral(argumentString, semaType, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "argument `%s` with value `%s` is not expected type `%s`", parameterList[index].Identifier, argumentString, semaType)
		}
		resultArgs = append(resultArgs, value)
	}
	return resultArgs, nil
}

// AccountE fetch an account from State
// Note that if `PrependNetworkToAccountNames` is specified it is prefixed with the network so that you can use the same logical name accross networks
func (o *OverflowState) AccountE(key string) (*flowkit.Account, error) {
	if o.PrependNetworkToAccountNames {
		key = fmt.Sprintf("%s-%s", o.Network, key)
	}

	account, err := o.State.Accounts().ByName(key)
	if err != nil {
		return nil, err
	}

	return account, nil

}

// ServiceAccountName return the name of the current service account
// Note that if `PrependNetworkToAccountNames` is specified it is prefixed with the network so that you can use the same logical name accross networks
func (o *OverflowState) ServiceAccountName() string {
	if o.PrependNetworkToAccountNames {
		return fmt.Sprintf("%s-%s", o.Network, o.ServiceAccountSuffix)
	}
	return o.ServiceAccountSuffix
}

// CreateAccountsE ensures that all accounts present in the deployment block for the given network is present
func (o *OverflowState) CreateAccountsE() (*OverflowState, error) {
	p := o.State
	signerAccount, err := p.Accounts().ByName(o.ServiceAccountName())
	if err != nil {
		return nil, err
	}

	accounts := p.AccountNamesForNetwork(o.Network)
	sort.Strings(accounts)

	for _, accountName := range accounts {

		// this error can never happen here, there is a test for it.
		account, _ := p.Accounts().ByName(accountName)

		if _, err := o.Services.Accounts.Get(account.Address()); err == nil {
			continue
		}

		o.Logger.Info(fmt.Sprintf("Creating account %s", account.Name()))
		_, err := o.Services.Accounts.Create(
			signerAccount,
			[]crypto.PublicKey{account.Key().ToConfig().PrivateKey.PublicKey()},
			[]int{1000},
			[]crypto.SignatureAlgorithm{account.Key().SigAlgo()},
			[]crypto.HashAlgorithm{account.Key().HashAlgo()},
			[]string{})
		if err != nil {
			return nil, err
		}

		messages := []string{
			fmt.Sprintf("%v", emoji.Person),
			"Created account:",
			account.Name(),
			"with address:",
			account.Address().String(),
		}

		if o.Network == "emulator" && o.NewUserFlowAmount != 0.0 {
			o.MintFlowTokens(account.Address().String(), o.NewUserFlowAmount)
			if o.Error != nil {
				return nil, errors.Wrap(err, "could not mint flow tokens")
			}
			messages = append(messages, "with flow:", fmt.Sprintf("%.2f", o.NewUserFlowAmount))
		}

		if o.PrintOptions != nil && o.LogLevel == output.NoneLog {
			fmt.Println(strings.Join(messages, " "))
		}
	}
	return o, nil
}

// InitializeContracts installs all contracts in the deployment block for the configured network
func (o *OverflowState) InitializeContracts() *OverflowState {
	o.Log.Reset()
	contracts, err := o.Services.Project.Deploy(o.Network, false)
	if err != nil {
		log, _ := o.readLog()
		if len(log) != 0 {
			messages := []string{}
			for _, msg := range log {
				if msg.Level == "warning" || msg.Level == "error" {
					messages = append(messages, msg.Msg)
				}
			}
			o.Error = errors.Wrapf(err, "errors : %v", messages)
		} else {
			o.Error = err
		}
	} else {
		//we do not have log output from emulator but we want to print results
		if o.LogLevel == output.NoneLog && o.PrintOptions != nil {
			names := []string{}
			for _, c := range contracts {
				names = append(names, c.Name())
			}
			fmt.Printf("%v deploy contracts %s\n", emoji.Scroll, strings.Join(names, ", "))
		}
	}
	o.Log.Reset()
	return o
}

// GetAccount takes the account name  and returns the state of that account on the given network.
func (o *OverflowState) GetAccount(key string) (*flow.Account, error) {
	account, err := o.AccountE(key)
	if err != nil {
		return nil, err
	}
	rawAddress := account.Address()
	return o.Services.Accounts.Get(rawAddress)
}

// Deprecated: use the new Tx/Script method and the argument functions
func (o *OverflowState) ParseArgumentsWithoutType(fileName string, code []byte, inputArgs map[string]string) ([]cadence.Value, error) {
	var resultArgs []cadence.Value = make([]cadence.Value, 0)

	codes := map[common.Location]string{}
	location := common.StringLocation(fileName)
	program, must := cmd.PrepareProgram(string(code), location, codes)
	checker, _ := cmd.PrepareChecker(program, location, codes, nil, must)

	var parameterList []*ast.Parameter

	functionDeclaration := sema.FunctionEntryPointDeclaration(program)
	if functionDeclaration != nil {
		if functionDeclaration.ParameterList != nil {
			parameterList = functionDeclaration.ParameterList.Parameters
		}
	}

	transactionDeclaration := program.TransactionDeclarations()
	if len(transactionDeclaration) == 1 {
		if transactionDeclaration[0].ParameterList != nil {
			parameterList = transactionDeclaration[0].ParameterList.Parameters
		}
	}

	if parameterList == nil {
		return resultArgs, nil
	}

	argumentNotPresent := []string{}
	args := []string{}
	for _, parameter := range parameterList {
		parameterName := parameter.Identifier.Identifier
		value, ok := inputArgs[parameterName]
		if !ok {
			argumentNotPresent = append(argumentNotPresent, parameterName)
		} else {
			args = append(args, value)
		}
	}

	if len(argumentNotPresent) > 0 {
		err := fmt.Errorf("the following arguments where not present %v", argumentNotPresent)
		return nil, err
	}

	for index, argumentString := range args {
		astType := parameterList[index].TypeAnnotation.Type
		semaType := checker.ConvertType(astType)

		switch semaType {
		case sema.StringType:
			if len(argumentString) > 0 && !strings.HasPrefix(argumentString, "\"") {
				argumentString = "\"" + argumentString + "\""
			}
		}

		switch semaType.(type) {
		case *sema.AddressType:

			account, _ := o.AccountE(argumentString)

			if account != nil {
				argumentString = account.Address().String()
			}

			if !strings.Contains(argumentString, "0x") {
				argumentString = fmt.Sprintf("0x%s", argumentString)
			}
		}

		var value, err = runtime.ParseLiteral(argumentString, semaType, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "argument `%s` is not expected type `%s`", parameterList[index].Identifier, semaType)
		}
		resultArgs = append(resultArgs, value)
	}
	return resultArgs, nil
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (o *OverflowState) Arguments() *FlowArgumentsBuilder {
	return &FlowArgumentsBuilder{
		Overflow:  o,
		Arguments: []cadence.Value{},
	}
}

func (o OverflowState) readLog() ([]LogrusMessage, error) {

	var logMessage []LogrusMessage
	dec := json.NewDecoder(o.Log)
	for {
		var doc LogrusMessage

		err := dec.Decode(&doc)
		if err == io.EOF {
			// all done
			break
		}
		if err != nil {
			return []LogrusMessage{}, err
		}

		logMessage = append(logMessage, doc)
	}

	o.Log.Reset()
	return logMessage, nil

}

// If you store this in a struct and add arguments to it it will not reset between calls
func (o *OverflowState) TxFN(outerOpts ...OverflowInteractionOption) OverflowTransactionFunction {

	return func(filename string, opts ...OverflowInteractionOption) *OverflowResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Tx(filename, outerOpts...)

	}
}

func (o *OverflowState) TxFileNameFN(filename string, outerOpts ...OverflowInteractionOption) OverflowTransactionOptsFunction {

	return func(opts ...OverflowInteractionOption) *OverflowResult {
		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Tx(filename, outerOpts...)
	}
}

//The main function for running an transasction in overflow
func (o *OverflowState) Tx(filename string, opts ...OverflowInteractionOption) *OverflowResult {
	ftb := o.BuildInteraction(filename, "transaction", opts...)
	result := ftb.Send()

	if ftb.PrintOptions != nil && !ftb.NoLog {
		po := *ftb.PrintOptions
		result.Print(po...)
	}
	if o.StopOnError && result.Err != nil {
		panic(result.Err)
	}

	return result
}

// get the latest block
func (o *OverflowState) GetLatestBlock() (*flow.Block, error) {
	block, _, _, err := o.Services.Blocks.GetBlock("latest", "", false)
	return block, err
}

// get block at a given height
func (o *OverflowState) GetBlockAtHeight(height uint64) (*flow.Block, error) {
	block, _, _, err := o.Services.Blocks.GetBlock(fmt.Sprintf("%d", height), "", false)
	return block, err
}

// blockId should be a hexadecimal string
func (o *OverflowState) GetBlockById(blockId string) (*flow.Block, error) {
	block, _, _, err := o.Services.Blocks.GetBlock(blockId, "", false)
	return block, err
}

// create a flowInteractionBuilder from the sent in options
func (o *OverflowState) BuildInteraction(filename string, interactionType string, opts ...OverflowInteractionOption) *FlowInteractionBuilder {

	path := o.TransactionBasePath
	if interactionType == "script" {
		path = o.ScriptBasePath
	}
	ftb := &FlowInteractionBuilder{
		Overflow:       o,
		Payer:          nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(o.Gas),
		BasePath:       path,
		NamedArgs:      map[string]interface{}{},
		NoLog:          false,
		PrintOptions:   o.PrintOptions,
	}

	for _, opt := range opts {
		opt(ftb)
	}

	if strings.Contains(filename, "transaction (") ||
		strings.Contains(filename, "transaction {") ||
		strings.Contains(filename, "transaction{") ||
		strings.Contains(filename, "transaction(") ||
		strings.Contains(filename, "transaction ") ||
		strings.Contains(filename, "pub fun main(") {
		ftb.TransactionCode = []byte(filename)
		ftb.FileName = "inline"
	} else {
		filePath := fmt.Sprintf("%s/%s.cdc", ftb.BasePath, filename)
		code, err := ftb.getContractCode(filePath)
		ftb.TransactionCode = code
		ftb.FileName = filename
		if ftb.Name == "" {
			ftb.Name = filename
		} else {
			ftb.Name = fmt.Sprintf("%s (%s)", ftb.Name, filename)
		}
		if err != nil {
			ftb.Error = err
			return ftb
		}
	}

	if ftb.Error != nil {
		return ftb
	}

	parseArgs, err := o.parseArguments(ftb.FileName, ftb.TransactionCode, ftb.NamedArgs)
	if err != nil {
		ftb.Error = err
		return ftb
	}
	ftb.Arguments = parseArgs
	return ftb
}

// Parse the given overflow state into a solution/npm-module
func (o *OverflowState) ParseAll() (*OverflowSolution, error) {
	return o.ParseAllWithConfig(false, []string{}, []string{})
}

// Parse the gieven overflow state with filters
func (o *OverflowState) ParseAllWithConfig(skipContracts bool, txSkip []string, scriptSkip []string) (*OverflowSolution, error) {

	warnings := []string{}
	transactions := map[string]string{}
	err := filepath.Walk(fmt.Sprintf("%s/transactions/", o.BasePath), func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".cdc") {
			name := strings.TrimSuffix(info.Name(), ".cdc")
			for _, txSkip := range txSkip {
				match, err := regexp.MatchString(txSkip, name)
				if err != nil {
					return err
				}
				if match {
					return nil
				}
			}

			transactions[path] = name
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	scripts := map[string]string{}
	err = filepath.Walk(fmt.Sprintf("%s/scripts/", o.BasePath), func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".cdc") {
			name := strings.TrimSuffix(info.Name(), ".cdc")
			for _, scriptSkip := range txSkip {
				match, err := regexp.MatchString(scriptSkip, name)
				if err != nil {
					return err
				}
				if match {
					return nil
				}
			}
			scripts[path] = name
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	transactionDeclarations := map[string]*OverflowDeclarationInfo{}
	for path, name := range transactions {
		code, err := o.State.ReaderWriter().ReadFile(path)
		if err != nil {
			return nil, err
		}
		info := declarationInfo(path, code)
		if info != nil {
			transactionDeclarations[name] = info
		}
	}

	scriptDeclarations := map[string]*OverflowDeclarationInfo{}
	for path, name := range scripts {
		code, err := o.State.ReaderWriter().ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read file at path %s", path)
		}
		info := declarationInfo(path, code)
		if info != nil {
			scriptDeclarations[name] = info
		}
	}

	networks := o.State.Networks()
	solutionNetworks := map[string]*OverflowSolutionNetwork{}
	for _, nw := range *networks {

		contracts, err := o.contracts(nw.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot find contracts for network %s", nw.Name)
		}

		contractResult := map[string]string{}
		for _, contract := range contracts {
			contractResult[contract.Name()] = contract.TranspiledCode()
		}

		scriptResult := map[string]string{}
		for path, name := range scripts {
			code, err := o.State.ReaderWriter().ReadFile(path)
			if err != nil {
				return nil, err
			}
			result, err := o.Parse(path, code, nw.Name)
			if err == nil {
				scriptResult[name] = result
			} else {
				warnings = append(warnings, fmt.Sprintf("Could not create script %s for network %s", path, nw.Name))
			}
		}

		txResult := map[string]string{}
		for path, name := range transactions {
			code, err := o.State.ReaderWriter().ReadFile(path)
			if err != nil {
				return nil, err
			}
			result, err := o.Parse(path, code, nw.Name)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Could not create transaction %s for network %s", path, nw.Name))
			} else {
				txResult[name] = result
			}
		}

		contract := &contractResult
		if skipContracts {
			contract = nil
		}
		solutionNetworks[nw.Name] = &OverflowSolutionNetwork{
			Contracts:    contract,
			Transactions: txResult,
			Scripts:      scriptResult,
		}
	}

	return &OverflowSolution{
		Transactions: transactionDeclarations,
		Scripts:      scriptDeclarations,
		Networks:     solutionNetworks,
		Warnings:     warnings,
	}, nil
}

func (o *OverflowState) contracts(network string) ([]*contracts.Contract, error) {
	// check there are not multiple accounts with same contract
	if o.State.ContractConflictExists(network) {
		return nil, fmt.Errorf(
			"the same contract cannot be deployed to multiple accounts on the same network",
		)
	}

	// create new processor for contract
	processor := contracts.NewPreprocessor(
		contracts.FilesystemLoader{
			Reader: o.State.ReaderWriter(),
		},
		o.State.AliasesForNetwork(network),
	)

	// add all contracts needed to deploy to processor
	contractsNetwork, err := o.State.DeploymentContractsByNetwork(network)
	if err != nil {
		return nil, err
	}

	for _, contract := range contractsNetwork {
		err2 := processor.AddContractSource(
			contract.Name,
			contract.Source,
			contract.AccountAddress,
			contract.AccountName,
			contract.Args,
		)
		if err2 != nil {
			return nil, err2
		}
	}

	// resolve imports assigns accounts to imports
	err = processor.ResolveImports()
	if err != nil {
		return nil, err
	}

	// sort correct deployment order of contracts so we don't have import that is not yet deployed
	orderedContracts, err := processor.ContractDeploymentOrder()
	if err != nil {
		return nil, err
	}
	return orderedContracts, nil
}

// Parse a given file into a resolved version
func (o *OverflowState) Parse(codeFileName string, code []byte, network string) (string, error) {
	resolver, err := contracts.NewResolver(code)
	if err != nil {
		return "", err
	}

	if !resolver.HasFileImports() {
		return strings.TrimSpace(string(code)), nil
	}

	contractsNetwork, err := o.State.DeploymentContractsByNetwork(network)
	if err != nil {
		return "", err
	}

	aliases := o.State.AliasesForNetwork(network)

	resolvedCode, err := resolver.ResolveImports(
		codeFileName,
		contractsNetwork,
		aliases,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(resolvedCode)), nil
}
