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
	Transactions map[string]*SolutionDeclarationInfo `json:"transactions"`

	//all scripts with name and paramters they have
	Scripts map[string]*SolutionDeclarationInfo `json:"scripts"`

	//all networks with associated scripts/tranasctions/contracts preresolved
	Networks map[string]*SolutionNetwork `json:"networks"`

	//warnings accumulated during parsing
	Warnings []string `json:"warnings"`
}

// a type containing information about paramter types and orders
type SolutionDeclarationInfo struct {
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
type SolutionMerged struct {
	Networks map[string]SolutionMergedNetwork `json:"networks"`
}

//a network in the merged solution
type SolutionMergedNetwork struct {
	Scripts      map[string]SolutionCodeWithSpec `json:"scripts"`
	Transactions map[string]SolutionCodeWithSpec `json:"transactions,omitempty"`
	Contracts    *map[string]string              `json:"contracts,omitempty"`
}

// representing code with specification if parameters
type SolutionCodeWithSpec struct {
	Code string                   `json:"code"`
	Spec *SolutionDeclarationInfo `json:"spec"`
}

// merge the given Solution into a MergedSolution that is suited for exposing as an NPM module
func (s *Solution) MergeSpecAndCode() *SolutionMerged {

	networks := map[string]SolutionMergedNetwork{}

	networkNames := []string{}
	for name, _ := range s.Networks {
		networkNames = append(networkNames, name)
	}

	for name, network := range s.Networks {

		scripts := map[string]SolutionCodeWithSpec{}
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
				scripts[scriptName] = SolutionCodeWithSpec{
					Code: formatCode(code),
					Spec: s.Scripts[rawScriptName],
				}
			}
		}

		transactions := map[string]SolutionCodeWithSpec{}
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
				transactions[txName] = SolutionCodeWithSpec{
					Code: formatCode(code),
					Spec: s.Transactions[rawTxName],
				}
			}
		}

		networks[name] = SolutionMergedNetwork{
			Contracts:    network.Contracts,
			Scripts:      scripts,
			Transactions: transactions,
		}

	}
	return &SolutionMerged{Networks: networks}
}

func declarationInfo(codeFileName string, code []byte) *SolutionDeclarationInfo {
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
	return &SolutionDeclarationInfo{
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
