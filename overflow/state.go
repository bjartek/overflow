package overflow

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/cmd"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

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
	Logger output.Logger
	Log    *bytes.Buffer

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

	StopOnError bool
}

func (f *OverflowState) parseArguments(fileName string, code []byte, inputArgs map[string]interface{}) ([]cadence.Value, error) {
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

			account, _ := f.AccountE(argumentString)

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

//AccountE fetch an account from State
//Note that if `PrependNetworkToAccountNames` is specified it is prefixed with the network so that you can use the same logical name accross networks
func (f *OverflowState) AccountE(key string) (*flowkit.Account, error) {
	if f.PrependNetworkToAccountNames {
		key = fmt.Sprintf("%s-%s", f.Network, key)
	}

	account, err := f.State.Accounts().ByName(key)
	if err != nil {
		return nil, err
	}

	return account, nil

}

// ServiceAccountName return the name of the current service account
//Note that if `PrependNetworkToAccountNames` is specified it is prefixed with the network so that you can use the same logical name accross networks
func (o *OverflowState) ServiceAccountName() string {
	if o.PrependNetworkToAccountNames {
		return fmt.Sprintf("%s-%s", o.Network, o.ServiceAccountSuffix)
	}
	return o.ServiceAccountSuffix
}
