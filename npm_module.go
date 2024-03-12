package overflow

import (
	"fmt"
	"strings"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/sema"
)

// NPM Module
//
// Overflow has support for generating an NPM module from a set of interactions

// a type representing the raw solutions that contains all transactions, scripts, networks and warnings of any
type OverflowSolution struct {
	// all transactions with name and what paremters they have
	Transactions map[string]*OverflowDeclarationInfo `json:"transactions"`

	// all scripts with name and parameter they have
	Scripts map[string]*OverflowDeclarationInfo `json:"scripts"`

	// all networks with associated scripts/tranasctions/contracts preresolved
	Networks map[string]*OverflowSolutionNetwork `json:"networks"`

	// warnings accumulated during parsing
	Warnings []string `json:"warnings"`
}

type OverflowAuthorizers [][]string

// a type containing information about parameter types and orders
type OverflowDeclarationInfo struct {
	Parameters     map[string]string   `json:"parameters"`
	Authorizers    OverflowAuthorizers `json:"-"`
	ParameterOrder []string            `json:"order"`
}

// a type representing one network in a solution, so mainnet/testnet/emulator
type OverflowSolutionNetwork struct {
	Scripts      map[string]string  `json:"scripts"`
	Transactions map[string]string  `json:"transactions,omitempty"`
	Contracts    *map[string]string `json:"contracts,omitempty"`
}

// a type representing a merged solution that will be serialized as the json file for the npm module
type OverflowSolutionMerged struct {
	Networks map[string]OverflowSolutionMergedNetwork `json:"networks"`
}

// a network in the merged solution
type OverflowSolutionMergedNetwork struct {
	Scripts      map[string]OverflowCodeWithSpec `json:"scripts"`
	Transactions map[string]OverflowCodeWithSpec `json:"transactions,omitempty"`
	Contracts    *map[string]string              `json:"contracts,omitempty"`
}

// representing code with specification if parameters
type OverflowCodeWithSpec struct {
	Spec *OverflowDeclarationInfo `json:"spec"`
	Code string                   `json:"code"`
}

// merge the given Solution into a MergedSolution that is suited for exposing as an NPM module
func (s *OverflowSolution) MergeSpecAndCode() *OverflowSolutionMerged {
	networks := map[string]OverflowSolutionMergedNetwork{}

	networkNames := []string{}
	for name := range s.Networks {
		networkNames = append(networkNames, name)
	}

	for name, network := range s.Networks {

		scripts := map[string]OverflowCodeWithSpec{}
		for rawScriptName, code := range network.Scripts {

			scriptName := rawScriptName

			valid := true
			for _, networkName := range networkNames {
				overwriteNetworkScriptName := fmt.Sprintf("%s%s", networkName, scriptName)
				_, ok := s.Scripts[overwriteNetworkScriptName]
				if ok {
					if networkName == name {
						valid = false
						break
					}
				}
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
				scripts[scriptName] = OverflowCodeWithSpec{
					Code: formatCode(code),
					Spec: s.Scripts[rawScriptName],
				}
			}
		}

		transactions := map[string]OverflowCodeWithSpec{}
		for rawTxName, code := range network.Transactions {

			txName := rawTxName
			txValid := true
			for _, networkName := range networkNames {
				overwriteNetworkTxName := fmt.Sprintf("%s%s", networkName, txName)
				_, ok := s.Transactions[overwriteNetworkTxName]
				if ok {
					if networkName == name {
						txValid = false
						break
					}
				}
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
				transactions[txName] = OverflowCodeWithSpec{
					Code: formatCode(code),
					Spec: s.Transactions[rawTxName],
				}
			}
		}

		networks[name] = OverflowSolutionMergedNetwork{
			Contracts:    network.Contracts,
			Scripts:      scripts,
			Transactions: transactions,
		}

	}
	return &OverflowSolutionMerged{Networks: networks}
}

func declarationInfo(code []byte) *OverflowDeclarationInfo {
	params, authorizerTypes := paramsAndAuthorizers(code)
	if params == nil {
		return &OverflowDeclarationInfo{
			ParameterOrder: []string{},
			Parameters:     map[string]string{},
			Authorizers:    authorizerTypes,
		}
	}
	parametersMap := make(map[string]string, len(params.Parameters))
	var parameterList []string
	for _, parameter := range params.Parameters {
		parametersMap[parameter.Identifier.Identifier] = parameter.TypeAnnotation.Type.String()
		parameterList = append(parameterList, parameter.Identifier.Identifier)
	}
	if len(parameterList) == 0 {
		return &OverflowDeclarationInfo{
			ParameterOrder: []string{},
			Parameters:     map[string]string{},
			Authorizers:    authorizerTypes,
		}
	}
	return &OverflowDeclarationInfo{
		ParameterOrder: parameterList,
		Parameters:     parametersMap,
		Authorizers:    authorizerTypes,
	}
}

func paramsAndAuthorizers(code []byte) (*ast.ParameterList, OverflowAuthorizers) {
	program, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return nil, nil
	}

	authorizers := OverflowAuthorizers{}
	// if we have any transtion declaration then return it
	for _, txd := range program.TransactionDeclarations() {
		if txd.Prepare != nil {
			prepareParams := txd.Prepare.FunctionDeclaration.ParameterList
			if prepareParams != nil {
				for _, parg := range txd.Prepare.FunctionDeclaration.ParameterList.ParametersByIdentifier() {
					// name := parg.Identifier.Identifier
					ta := parg.TypeAnnotation
					if ta != nil {
						rt, ok := ta.Type.(*ast.ReferenceType)
						if ok {

							entitlements := []string{}
							switch authorization := rt.Authorization.(type) {
							case ast.EntitlementSet:
								for _, entitlement := range authorization.Entitlements() {
									entitlements = append(entitlements, entitlement.Identifier.Identifier)
								}
							}
							authorizers = append(authorizers, entitlements)
						} else {
							authorizers = append(authorizers, []string{})
						}
					}
				}
			}
		}
		return txd.ParameterList, authorizers
	}

	functionDeclaration := sema.FunctionEntryPointDeclaration(program)
	if functionDeclaration != nil {
		if functionDeclaration.ParameterList != nil {
			return functionDeclaration.ParameterList, nil
		}
	}

	return nil, nil
}

func formatCode(input string) string {
	return strings.ReplaceAll(strings.TrimSpace(input), "\t", "    ")
}
