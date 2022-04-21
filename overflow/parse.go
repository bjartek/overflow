package overflow

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/cmd"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/flow-cli/pkg/flowkit/contracts"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Solution struct {
	Transactions map[string]*DeclarationInfo `json:"transactions"`
	Scripts      map[string]*DeclarationInfo `json:"scripts"`
	Networks     map[string]*SolutionNetwork `json:"networks"`
}

type SolutionNetwork struct {
	Scripts      map[string]string  `json:"scripts"`
	Transactions map[string]string  `json:"transactions,omitempty"`
	Contracts    *map[string]string `json:"contracts,omitempty"`
}

type DeclarationInfo struct {
	ParameterOrder []string          `json:"order"`
	Parameters     map[string]string `json:"parameters"`
}

func (o *Overflow) ParseAll() (*Solution, error) {
	return o.ParseAllWithConfig(false, []string{}, []string{})
}

type MergedSolution struct {
	Networks map[string]MergedSolutionNetwork `json:"networks"`
}

type MergedSolutionNetwork struct {
	Scripts      map[string]CodeWithSpec `json:"scripts"`
	Transactions map[string]CodeWithSpec `json:"transactions,omitempty"`
	Contracts    *map[string]string      `json:"contracts,omitempty"`
}

type CodeWithSpec struct {
	Code string           `json:"code"`
	Spec *DeclarationInfo `json:"spec"`
}

func (s *Solution) MergeSpecAndCode() *MergedSolution {

	networks := map[string]MergedSolutionNetwork{}
	for name, network := range s.Networks {

		scripts := map[string]CodeWithSpec{}
		for name, code := range network.Scripts {
			scripts[name] = CodeWithSpec{
				Code: FormatCode(code),
				Spec: s.Scripts[name],
			}
		}

		transactions := map[string]CodeWithSpec{}
		for name, code := range network.Transactions {
			transactions[name] = CodeWithSpec{
				Code: FormatCode(code),
				Spec: s.Transactions[name],
			}
		}

		networks[name] = MergedSolutionNetwork{
			Contracts:    network.Contracts,
			Scripts:      scripts,
			Transactions: transactions,
		}

	}
	return &MergedSolution{Networks: networks}
}

func (o *Overflow) ParseAllWithConfig(skipContracts bool, txSkip []string, scriptSkip []string) (*Solution, error) {

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

	transactionDeclarations := map[string]*DeclarationInfo{}
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

	scriptDeclarations := map[string]*DeclarationInfo{}
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
	solutionNetworks := map[string]*SolutionNetwork{}
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
				log.Printf("Could not create script %s for network %s", path, nw.Name)

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
				log.Printf("Could not create transaction %s for network %s", path, nw.Name)
			} else {
				txResult[name] = result
			}
		}

		contract := &contractResult
		if skipContracts {
			contract = nil
		}
		solutionNetworks[nw.Name] = &SolutionNetwork{
			Contracts:    contract,
			Transactions: txResult,
			Scripts:      scriptResult,
		}
	}

	return &Solution{
		Transactions: transactionDeclarations,
		Scripts:      scriptDeclarations,
		Networks:     solutionNetworks,
	}, nil
}

func (o *Overflow) contracts(network string) ([]*contracts.Contract, error) {
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
			contract.Target,
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

func declarationInfo(codeFileName string, code []byte) *DeclarationInfo {
	params := params(codeFileName, code)
	if params == nil {
		return nil
	}
	parametersMap := make(map[string]string, len(params.Parameters))
	var parameterList []string
	for _, parameter := range params.Parameters {
		parametersMap[parameter.Identifier.Identifier] = parameter.TypeAnnotation.Type.String()
		parameterList = append(parameterList, parameter.Identifier.Identifier)
	}
	if len(parameterList) == 0 {
		return nil
	}
	return &DeclarationInfo{
		ParameterOrder: parameterList,
		Parameters:     parametersMap,
	}
}

func (o *Overflow) Parse(codeFileName string, code []byte, network string) (string, error) {
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

func params(fileName string, code []byte) *ast.ParameterList {

	codes := map[common.LocationID]string{}
	location := common.StringLocation(fileName)
	program, _ := cmd.PrepareProgram(string(code), location, codes)

	transactionDeclaration := program.SoleTransactionDeclaration()
	if transactionDeclaration != nil {
		if transactionDeclaration.ParameterList != nil {
			return transactionDeclaration.ParameterList
		}
	}

	functionDeclaration := sema.FunctionEntryPointDeclaration(program)
	if functionDeclaration != nil {
		if functionDeclaration.ParameterList != nil {
			return functionDeclaration.ParameterList
		}
	}

	return nil
}

func FormatCode(input string) string {
	return strings.ReplaceAll(strings.TrimSpace(input), "\t", "    ")
}
