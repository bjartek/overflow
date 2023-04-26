package overflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/cmd"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/accounts"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/gateway"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/flowkit/project"
	"github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// Overflow client is an interface with the most used v1 api methods for overflow
type OverflowClient interface {
	ScriptFN(outerOpts ...OverflowInteractionOption) OverflowScriptFunction
	ScriptFileNameFN(filename string, outerOpts ...OverflowInteractionOption) OverflowScriptOptsFunction
	Script(filename string, opts ...OverflowInteractionOption) *OverflowScriptResult

	QualifiedIdentifierFromSnakeCase(typeName string) (string, error)
	QualifiedIdentifier(contract string, name string) (string, error)

	AddContract(name string, contract *flowkit.Script, update bool) error

	GetNetwork() string
	AccountE(key string) (*accounts.Account, error)
	Address(key string) string
	Account(key string) *accounts.Account

	//Note that this returns a flow account and not a flowkit account like the others, is this needed?
	GetAccount(key string) (*flow.Account, error)

	Tx(filename string, opts ...OverflowInteractionOption) *OverflowResult
	TxFN(outerOpts ...OverflowInteractionOption) OverflowTransactionFunction
	TxFileNameFN(filename string, outerOpts ...OverflowInteractionOption) OverflowTransactionOptsFunction

	GetLatestBlock() (*flow.Block, error)
	GetBlockAtHeight(height uint64) (*flow.Block, error)
	GetBlockById(blockId string) (*flow.Block, error)
	FetchEventsWithResult(opts ...OverflowEventFetcherOption) EventFetcherResult

	UploadFile(filename string, accountName string) error
	DownloadAndUploadFile(url string, accountName string) error
	DownloadImageAndUploadAsDataUrl(url, accountName string) error
	UploadImageAsDataUrl(filename string, accountName string) error
	UploadString(content string, accountName string) error
	GetFreeCapacity(accountName string) int
	MintFlowTokens(accountName string, amount float64) *OverflowState
	FillUpStorage(accountName string) *OverflowState

	SignUserMessage(account string, message string) (string, error)
}

// beta client with unstable features
type OverflowBetaClient interface {
	OverflowClient
	GetTransactionResultByBlockId(blockId flow.Identifier) ([]*flow.TransactionResult, error)
	GetTransactionByBlockId(blockId flow.Identifier) ([]*flow.Transaction, error)
	GetTransactions(ctx context.Context, id flow.Identifier) ([]OverflowTransaction, error)
	StreamTransactions(ctx context.Context, poll time.Duration, height uint64, channel chan<- BlockResult) error
}

// OverflowState contains information about how to Overflow is confitured and the current runnig state
type OverflowState struct {
	State *flowkit.State
	//the services from flowkit to performed operations on
	Flowkit *flowkit.Flowkit

	EmulatorGatway *gateway.EmulatorGateway

	ArchiveFlowkit *flowkit.Flowkit

	//Configured variables that are taken from the builder since we need them in the execution of overflow later on
	Network                      config.Network
	PrependNetworkToAccountNames bool
	ServiceAccountSuffix         string
	Gas                          int

	//flowkit, emulator and emulator debug log uses three different logging technologies so we have them all stored here
	//this flowkit Logger can go away when we can remove deprecations!
	Logger   output.Logger
	Log      *bytes.Buffer
	LogLevel int

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

	InputResolver InputResolver
}

type OverflowArgument struct {
	Name  string
	Value interface{}
	Type  ast.Type
}

type OverflowArguments map[string]OverflowArgument
type OverflowArgumentList []OverflowArgument

func (o *OverflowState) AddContract(ctx context.Context, name string, code []byte, args []cadence.Value, filename string, update bool) error {
	script := flowkit.Script{
		Code:     code,
		Args:     args,
		Location: filename,
	}
	account, err := o.AccountE(name)
	if err != nil {
		return err
	}
	_, _, err = o.Flowkit.AddContract(ctx, account, script, flowkit.UpdateExistingContract(update))
	return err

}
func (o *OverflowState) GetNetwork() string {
	return o.Network.Name
}

// Qualified identifier from a snakeCase string Account_Contract_Struct
func (o *OverflowState) QualifiedIdentifierFromSnakeCase(typeName string) (string, error) {

	words := strings.Split(typeName, "_")
	if len(words) < 2 {
		return "", fmt.Errorf("Invalid snake_case type string Contract_Name")
	}
	return o.QualifiedIdentifier(words[0], words[1])
}

// Create a qualified identifier from account, contract, name

// account can either be a name from  accounts or the raw value
func (o *OverflowState) QualifiedIdentifier(contract string, name string) (string, error) {

	flowContract, err := o.State.Contracts().ByName(contract)
	if err != nil {
		return "", err
	}

	//we found the contract specified in contracts section
	if flowContract != nil {
		alias := flowContract.Aliases.ByNetwork(o.Network.Name)
		if alias != nil {
			return fmt.Sprintf("A.%s.%s.%s", alias.Address.String(), contract, name), nil
		}
	}

	flowDeploymentContracts, err := o.State.DeploymentContractsByNetwork(o.Network)
	if err != nil {
		return "", err

	}

	for _, flowDeploymentContract := range flowDeploymentContracts {
		if flowDeploymentContract.Name == contract {
			return fmt.Sprintf("A.%s.%s.%s", flowDeploymentContract.AccountAddress, contract, name), nil
		}
	}

	return "", fmt.Errorf("You are trying to get the qualified identifier for something you are not creating or have mentioned in flow.json with name=%s", contract)

}

func (o *OverflowState) parseArguments(fileName string, code []byte, inputArgs map[string]interface{}) ([]cadence.Value, CadenceArguments, error) {
	var resultArgs []cadence.Value = make([]cadence.Value, 0)
	resultArgsMap := CadenceArguments{}

	codes := map[common.Location][]byte{}
	location := common.StringLocation(fileName)
	program, must := cmd.PrepareProgram(code, location, codes)
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
		return resultArgs, resultArgsMap, nil
	}
	argumentNotPresent := []string{}
	argumentNames := []string{}
	args := OverflowArgumentList{}
	for _, parameter := range parameterList {
		parameterName := parameter.Identifier.Identifier
		value, ok := inputArgs[parameterName]
		if !ok {
			argumentNotPresent = append(argumentNotPresent, parameterName)
		} else {
			argumentNames = append(argumentNames, parameterName)
			args = append(args, OverflowArgument{
				Name:  parameterName,
				Value: value,
				Type:  parameter.TypeAnnotation.Type,
			})
		}
	}

	if len(argumentNotPresent) > 0 {
		err := fmt.Errorf("the interaction '%s' is missing %v", fileName, argumentNotPresent)
		return nil, nil, err
	}

	redundantArgument := []string{}
	for inputKey := range inputArgs {
		//If your IDE complains about this it is wrong, this is 1.18 generics not suported anywhere
		if !slices.Contains(argumentNames, inputKey) {
			redundantArgument = append(redundantArgument, inputKey)
		}
	}

	if len(redundantArgument) > 0 {
		err := fmt.Errorf("the interaction '%s' has the following extra arguments %v", fileName, redundantArgument)
		return nil, nil, err
	}

	for _, oa := range args {

		name := oa.Name
		argument := oa.Value

		cadenceVal, isCadenceValue := argument.(cadence.Value)
		if isCadenceValue {
			resultArgs = append(resultArgs, cadenceVal)
			resultArgsMap[name] = cadenceVal
			continue
		}

		var argumentString string
		switch a := argument.(type) {
		case nil:
			argumentString = "nil"
		case string:
			argumentString = a
		case int:
			argumentString = fmt.Sprintf("%v", a)
		default:
			cadenceVal, err := InputToCadence(argument, o.InputResolver)
			if err != nil {
				return nil, nil, err
			}
			resultArgs = append(resultArgs, cadenceVal)
			resultArgsMap[name] = cadenceVal
			continue

		}
		semaType := checker.ConvertType(oa.Type)

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
				argumentString = account.Address.String()
			}

			if !strings.Contains(argumentString, "0x") {
				argumentString = fmt.Sprintf("0x%s", argumentString)
			}
		}

		inter, interErr := interpreter.NewInterpreter(nil, nil, &interpreter.Config{})
		if interErr != nil {
			return nil, nil, interErr
		}
		var value, err = runtime.ParseLiteral(argumentString, semaType, inter)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "argument `%s` with value `%s` is not expected type `%s`", name, argumentString, semaType)
		}
		resultArgs = append(resultArgs, value)
		resultArgsMap[name] = value
	}
	return resultArgs, resultArgsMap, nil
}

// AccountE fetch an account from State
// Note that if `PrependNetworkToAccountNames` is specified it is prefixed with the network so that you can use the same logical name across networks
func (o *OverflowState) AccountE(key string) (*accounts.Account, error) {
	if o.PrependNetworkToAccountNames {
		key = fmt.Sprintf("%s-%s", o.Network.Name, key)
	}

	account, err := o.State.Accounts().ByName(key)
	if err != nil {
		return nil, err
	}

	return account, nil

}

// return the address of an given account
func (o *OverflowState) Address(key string) string {
	return fmt.Sprintf("0x%s", o.FlowAddress(key))
}

// return the flow Address of the given name
func (o *OverflowState) FlowAddress(key string) flow.Address {
	account, err := o.AccountE(key)
	if err == nil {
		return account.Address
	}

	flowContract, err := o.State.Contracts().ByName(key)
	if err != nil {
		panic(err)
	}

	//we found the contract specified in contracts section
	if flowContract != nil {
		alias := flowContract.Aliases.ByNetwork(o.Network.Name)
		if alias != nil {
			return alias.Address
		}
	}

	flowDeploymentContracts, err := o.State.DeploymentContractsByNetwork(o.Network)
	if err != nil {
		panic(err)
	}

	for _, flowDeploymentContract := range flowDeploymentContracts {
		if flowDeploymentContract.Name == key {
			return flowDeploymentContract.AccountAddress
		}
	}
	panic("Not valid user account, contract or deployment contract")
}

// return the account of a given account
func (o *OverflowState) Account(key string) *accounts.Account {
	account, err := o.AccountE(key)
	if err != nil {
		panic(err)
	}

	return account
}

// ServiceAccountName return the name of the current service account
// Note that if `PrependNetworkToAccountNames` is specified it is prefixed with the network so that you can use the same logical name across networks
func (o *OverflowState) ServiceAccountName() string {
	if o.PrependNetworkToAccountNames {
		return fmt.Sprintf("%s-%s", o.Network.Name, o.ServiceAccountSuffix)
	}
	return o.ServiceAccountSuffix
}

// CreateAccountsE ensures that all accounts present in the deployment block for the given network is present
func (o *OverflowState) CreateAccountsE(ctx context.Context) (*OverflowState, error) {
	p := o.State
	signerAccount, err := p.Accounts().ByName(o.ServiceAccountName())
	if err != nil {
		return nil, err
	}

	acct := *p.AccountsForNetwork(o.Network)

	sort.SliceStable(acct, func(i, j int) bool {
		return strings.Compare(acct[i].Name, acct[j].Name) < 1
	})

	for _, account := range acct {
		if _, err := o.Flowkit.GetAccount(ctx, account.Address); err == nil {
			continue
		}

		keys := []accounts.PublicKey{{
			Public:   account.Key.ToConfig().PrivateKey.PublicKey(),
			Weight:   1000,
			SigAlgo:  account.Key.SigAlgo(),
			HashAlgo: account.Key.HashAlgo(),
		}}

		o.Logger.Info(fmt.Sprintf("Creating account %s", account.Name))
		_, _, err := o.Flowkit.CreateAccount(ctx, signerAccount, keys)
		if err != nil {
			return nil, err
		}

		messages := []string{
			fmt.Sprintf("%v", emoji.Person),
			"Created account:",
			account.Name,
			"with address:",
			account.Address.String(),
		}

		if o.Network.Name == "emulator" && o.NewUserFlowAmount != 0.0 {
			res := o.MintFlowTokens(account.Address.String(), o.NewUserFlowAmount)
			if res.Error != nil {
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
func (o *OverflowState) InitializeContracts(ctx context.Context) *OverflowState {
	o.Log.Reset()
	contracts, err := o.Flowkit.DeployProject(ctx, flowkit.UpdateExistingContract(true))
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
				names = append(names, c.Name)
			}
			fmt.Printf("%v deploy contracts %s\n", emoji.Scroll, strings.Join(names, ", "))
		}
	}
	o.Log.Reset()
	return o
}

// GetAccount takes the account name  and returns the state of that account on the given network.
func (o *OverflowState) GetAccount(ctx context.Context, key string) (*flow.Account, error) {
	account, err := o.AccountE(key)
	if err != nil {
		return nil, err
	}
	rawAddress := account.Address
	return o.Flowkit.GetAccount(ctx, rawAddress)
}

func (o OverflowState) readLog() ([]OverflowEmulatorLogMessage, error) {

	var logMessage []OverflowEmulatorLogMessage
	dec := json.NewDecoder(o.Log)
	for {
		var msg map[string]interface{}

		err := dec.Decode(&msg)
		if err == io.EOF {
			// all done
			break
		}

		if err != nil {
			return []OverflowEmulatorLogMessage{}, err
		}
		doc := OverflowEmulatorLogMessage{Msg: msg["message"].(string), Level: msg["level"].(string)}

		delete(msg, "message")
		delete(msg, "level")
		rawCom, ok := msg["computationUsed"]
		if ok {
			field := rawCom.(float64)
			doc.ComputationUsed = int(field)
			delete(msg, "computationUsed")
		}
		doc.Fields = msg
		logMessage = append(logMessage, doc)
	}

	o.Log.Reset()
	return logMessage, nil

}

// If you store this in a struct and add arguments to it it will not reset between calls
func (o *OverflowState) TxFN(outerOpts ...OverflowInteractionOption) OverflowTransactionFunction {

	return func(filename string, opts ...OverflowInteractionOption) *OverflowResult {
		outerOpts = append(outerOpts, opts...)
		return o.Tx(filename, outerOpts...)

	}
}

func (o *OverflowState) TxFileNameFN(filename string, outerOpts ...OverflowInteractionOption) OverflowTransactionOptsFunction {

	return func(opts ...OverflowInteractionOption) *OverflowResult {
		outerOpts = append(outerOpts, opts...)
		return o.Tx(filename, outerOpts...)
	}
}

// The main function for running an transasction in overflow
func (o *OverflowState) Tx(filename string, opts ...OverflowInteractionOption) *OverflowResult {
	ftb := o.BuildInteraction(filename, "transaction", opts...)
	result := ftb.Send()

	if ftb.PrintOptions != nil && !ftb.NoLog {
		po := *ftb.PrintOptions
		result.Print(po...)
	}
	if o.StopOnError && result.Err != nil {
		result.PrintArguments(nil)
		panic(result.Err)
	}

	return result
}

// get the latest block
func (o *OverflowState) GetLatestBlock(ctx context.Context) (*flow.Block, error) {

	bc, err := flowkit.NewBlockQuery("latest")
	if err != nil {
		return nil, err
	}
	return o.Flowkit.GetBlock(ctx, bc)
}

// get block at a given height
func (o *OverflowState) GetBlockAtHeight(ctx context.Context, height uint64) (*flow.Block, error) {
	bc := flowkit.BlockQuery{Height: height}
	return o.Flowkit.GetBlock(ctx, bc)
}

// blockId should be a hexadecimal string
func (o *OverflowState) GetBlockById(ctx context.Context, blockId string) (*flow.Block, error) {
	bid := flow.HexToID(blockId)
	bc := flowkit.BlockQuery{ID: &bid}
	return o.Flowkit.GetBlock(ctx, bc)
}

// create a flowInteractionBuilder from the sent in options
func (o *OverflowState) BuildInteraction(filename string, interactionType string, opts ...OverflowInteractionOption) *OverflowInteractionBuilder {

	path := o.TransactionBasePath
	if interactionType == "script" {
		path = o.ScriptBasePath
	}
	ftb := &OverflowInteractionBuilder{
		Overflow:       o,
		Payer:          nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*accounts.Account{},
		GasLimit:       uint64(o.Gas),
		BasePath:       path,
		NamedArgs:      map[string]interface{}{},
		NoLog:          false,
		PrintOptions:   o.PrintOptions,
		ScriptQuery:    &flowkit.ScriptQuery{},
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

	parseArgs, namedCadenceArguments, err := o.parseArguments(ftb.FileName, ftb.TransactionCode, ftb.NamedArgs)
	if err != nil {
		ftb.Error = err
		return ftb
	}
	ftb.Arguments = parseArgs
	ftb.NamedCadenceArguments = namedCadenceArguments
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

		contractResult, err := o.contracts(nw)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot find contracts for network %s", nw.Name)
		}

		scriptResult := map[string]string{}
		for path, name := range scripts {
			code, err := o.State.ReaderWriter().ReadFile(path)
			if err != nil {
				return nil, err
			}
			result, err := o.Parse(path, code, nw)
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
			result, err := o.Parse(path, code, nw)
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

func (o *OverflowState) contracts(network config.Network) (map[string]string, error) {

	contracts, err := o.State.DeploymentContractsByNetwork(network)
	if err != nil {
		return nil, err
	}

	deployment, err := project.NewDeployment(contracts, o.State.AliasesForNetwork(network))
	if err != nil {
		return nil, err
	}

	sorted, err := deployment.Sort()
	if err != nil {
		return nil, err
	}

	resolvedContracts := map[string]string{}
	for _, p := range sorted {
		code, err := o.Parse(p.Location(), p.Code(), network)
		if err != nil {
			return resolvedContracts, err
		}
		resolvedContracts[p.Name] = code
	}
	return resolvedContracts, nil
}

// Parse a given file into a resolved version
func (o *OverflowState) Parse(codeFileName string, code []byte, network config.Network) (string, error) {

	program, err := project.NewProgram(code, []cadence.Value{}, codeFileName)
	if err != nil {
		return "", err
	}

	if !program.HasImports() {
		return strings.TrimSpace(string(program.Code())), nil
	}

	contracts, err := o.State.DeploymentContractsByNetwork(network)
	if err != nil {
		return "", err
	}

	importReplacer := project.NewImportReplacer(
		contracts,
		o.State.AliasesForNetwork(network),
	)

	program2, err := importReplacer.Replace(program)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(program2.Code())), nil
}

func (o *OverflowState) GetCoverageReport() *runtime.CoverageReport {
	return o.EmulatorGatway.CoverageReport()
}

func (o *OverflowState) RollbackToBlockHeight(height uint64) error {
	return o.EmulatorGatway.RollbackToBlockHeight(height)
}
