package overflow

import (
	"fmt"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/pkg/errors"
)

// FlowInteractionBuilder used to create a builder pattern for a transaction
type FlowInteractionBuilder struct {
	Overflow       *OverflowState
	FileName       string
	Content        string
	Arguments      []cadence.Value
	MainSigner     *flowkit.Account
	PayloadSigners []*flowkit.Account
	GasLimit       uint64
	BasePath       string
	Error          error

	//these are used for v3, but can still be here for v2
	TransactionCode []byte
	NamedArgs       map[string]interface{}
	Proposer        *flowkit.Account

	EventFilter              OverflowEventFilter
	IgnoreGlobalEventFilters bool
}

//A function to customize the transaction builder
type InteractionOption func(*FlowInteractionBuilder)

func (o *OverflowState) BuildInteraction(filename string, interactionType string, opts ...InteractionOption) *FlowInteractionBuilder {

	path := o.TransactionBasePath
	if interactionType == "script" {
		path = o.ScriptBasePath
	}
	ftb := &FlowInteractionBuilder{
		Overflow:       o,
		MainSigner:     nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(o.Gas),
		BasePath:       path,
		NamedArgs:      map[string]interface{}{},
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
		if err != nil {
			ftb.Error = err
			return ftb
		}
	}
	for _, opt := range opts {
		opt(ftb)
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

func (t FlowInteractionBuilder) getContractCode(codeFileName string) ([]byte, error) {
	code := []byte(t.Content)
	var err error
	if t.Content == "" {
		code, err = t.Overflow.State.ReaderWriter().ReadFile(codeFileName)
		if err != nil {
			return nil, fmt.Errorf("%v Could not read interaction file from path=%s", emoji.PileOfPoo, codeFileName)
		}
	}
	return code, nil
}

func Args(args ...interface{}) func(ftb *FlowInteractionBuilder) {

	return func(ftb *FlowInteractionBuilder) {
		if len(args)%2 != 0 {
			ftb.Error = fmt.Errorf("Please send in an even number of string : interface{} pairs")
			return
		}
		var i = 0
		for i < len(args) {
			key := args[0]
			value, labelOk := key.(string)
			if !labelOk {
				ftb.Error = fmt.Errorf("even parameters in Args needs to be strings")
			}
			ftb.NamedArgs[value] = args[1]
			i = i + 2
		}
	}
}

func ArgsM(args map[string]interface{}) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		for key, value := range args {
			ftb.NamedArgs[key] = value
		}
	}
}

/// Send an argument into a transaction
/// @param name: string the name of the parameter
/// @param value: the value of the argument, se below
///
/// The value is treated in the given way depending on type
///  - cadence.Value is sent as straight argument
///  - string argument are resolved into cadence.Value using flowkit
///  - ofther values are converted to string with %v and resolved into cadence.Value using flowkit
///  - if the type of the paramter is Address and the string you send in is a valid account in flow.json it will resolve
///
/// Examples:
///  If you want to send the UFix64 number "42.0" into a transaciton you have to use it as a string since %v of fmt.Sprintf will make it 42
func Arg(name string, value interface{}) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		ftb.NamedArgs[name] = value
	}
}

func DateTimeArg(name string, dateString string, timezone string) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		value, err := parseTime(dateString, timezone)
		if err != nil {
			ftb.Error = err
			return
		}

		//swallow the error since it will never happen here, we control the input
		amount, _ := cadence.NewUFix64(value)

		ftb.NamedArgs[name] = amount
	}
}

func Addresses(name string, value ...string) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		array := []cadence.Value{}

		for _, val := range value {
			account, err := ftb.Overflow.AccountE(val)
			if err != nil {
				address, err := HexToAddress(val)
				if err != nil {
					ftb.Error = errors.Wrap(err, fmt.Sprintf("%s is not an valid account name or an address", val))
					return
				}
				cadenceAddress := cadence.BytesToAddress(address.Bytes())
				array = append(array, cadenceAddress)
			} else {
				cadenceAddress := cadence.BytesToAddress(account.Address().Bytes())
				array = append(array, cadenceAddress)
			}
		}
		ftb.NamedArgs[name] = cadence.NewArray(array)
	}
}

func ProposeAs(proposer string) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		account, err := ftb.Overflow.AccountE(proposer)
		if err != nil {
			ftb.Error = err
			return
		}
		ftb.Proposer = account
	}
}

func ProposeAsServiceAccount() func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		key := ftb.Overflow.ServiceAccountName()
		account, _ := ftb.Overflow.State.Accounts().ByName(key)
		ftb.Proposer = account
	}
}

func SignProposeAndPayAs(signer string) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		account, err := ftb.Overflow.AccountE(signer)
		if err != nil {
			ftb.Error = err
			return
		}
		ftb.MainSigner = account
		ftb.Proposer = account
	}
}

func SignProposeAndPayAsServiceAccount() func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		key := ftb.Overflow.ServiceAccountName()
		account, _ := ftb.Overflow.State.Accounts().ByName(key)
		ftb.MainSigner = account
		ftb.Proposer = account
	}
}

func Gas(gas uint64) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		ftb.GasLimit = gas
	}
}

func EventFilter(filter OverflowEventFilter) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		ftb.EventFilter = filter
	}
}

func IgnoreGlobalEventFilters() func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		ftb.IgnoreGlobalEventFilters = true
	}
}

func PayloadSigner(signer ...string) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		for _, signer := range signer {
			account, err := ftb.Overflow.AccountE(signer)
			if err != nil {
				ftb.Error = err
				return
			}
			ftb.PayloadSigners = append(ftb.PayloadSigners, account)
		}
	}
}
