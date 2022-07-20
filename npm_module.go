package overflow

import (
	"strings"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/cmd"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
)

// NPM Module
//
// Overflow has support for generating an NPM module from a set of interactions

// a type representing the raw solutions that contains all transactions, scripts, networks and warnings of any
type Solution struct {

	//all transactions with name and what paremters they have
	Transactions map[string]*DeclarationInfo `json:"transactions"`

	//all scripts with name and paramters they have
	Scripts map[string]*DeclarationInfo `json:"scripts"`

	//all networks with associated scripts/tranasctions/contracts preresolved
	Networks map[string]*SolutionNetwork `json:"networks"`

	//warnings accumulated during parsing
	Warnings []string `json:"warnings"`
}

// a type containing information about paramter types and orders
type DeclarationInfo struct {
	ParameterOrder []string          `json:"order"`
	Parameters     map[string]string `json:"parameters"`
}

// a type representing one network in a solution, so mainnet/testnet/emulator
type SolutionNetwork struct {
	Scripts      map[string]string  `json:"scripts"`
	Transactions map[string]string  `json:"transactions,omitempty"`
	Contracts    *map[string]string `json:"contracts,omitempty"`
}

// a type representing a merged solution that will be serialized as the json file for the npm module
type MergedSolution struct {
	Networks map[string]MergedSolutionNetwork `json:"networks"`
}

//a network in the merged solution
type MergedSolutionNetwork struct {
	Scripts      map[string]CodeWithSpec `json:"scripts"`
	Transactions map[string]CodeWithSpec `json:"transactions,omitempty"`
	Contracts    *map[string]string      `json:"contracts,omitempty"`
}

// representing code with specification if parameters
type CodeWithSpec struct {
	Code string           `json:"code"`
	Spec *DeclarationInfo `json:"spec"`
}

// merge the given Solution into a MergedSolution that is suited for exposing as an NPM module
func (s *Solution) MergeSpecAndCode() *MergedSolution {

	networks := map[string]MergedSolutionNetwork{}

	networkNames := []string{}
	for name, _ := range s.Networks {
		networkNames = append(networkNames, name)
	}

	for name, network := range s.Networks {

		scripts := map[string]CodeWithSpec{}
		for rawScriptName, code := range network.Scripts {

			scriptName := rawScriptName

			valid := true
			for _, networkName := range networkNames {
				if strings.HasPrefix(scriptName, networkName) {
					if networkName == name {
						scriptName = strings.TrimPrefix(scriptName, networkName)
						valid = true
						break
					} else {
						valid = false
						break
					}
				}
			}
			if valid {
				scripts[scriptName] = CodeWithSpec{
					Code: formatCode(code),
					Spec: s.Scripts[rawScriptName],
				}
			}
		}

		transactions := map[string]CodeWithSpec{}
		for rawTxName, code := range network.Transactions {

			txName := rawTxName
			txValid := true
			for _, networkName := range networkNames {

				if strings.HasPrefix(txName, networkName) {
					if networkName == name {
						txName = strings.TrimPrefix(txName, networkName)
						txValid = true
						break
					} else {
						txValid = false
						break
					}
				}
			}
			if txValid {
				transactions[txName] = CodeWithSpec{
					Code: formatCode(code),
					Spec: s.Transactions[rawTxName],
				}
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

func params(fileName string, code []byte) *ast.ParameterList {

	codes := map[common.Location]string{}
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

func formatCode(input string) string {
	return strings.ReplaceAll(strings.TrimSpace(input), "\t", "    ")
}
