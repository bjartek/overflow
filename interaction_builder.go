package overflow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

// Flow Interaction Builder
//
// An interaction in overflow is either a script or a transaction

// OverflowInteractionBuilder used to create a builder pattern for an interaction
type OverflowInteractionBuilder struct {

	//the name of the integration, for inline variants
	Name string

	//force that this interaction will not print log, even if overflow state has specified it
	NoLog bool

	//The underlying state of overflow used to fetch some global settings
	Overflow *OverflowState

	//The file name of the interaction
	FileName string

	//The content of the interaction
	Content string

	//The list of raw arguments
	Arguments []cadence.Value

	NamedCadenceArguments CadenceArguments

	//The main signer used to sign the transaction
	// Payer: the account paying for the transaction fees.
	Payer *flowkit.Account

	//The propser account
	//    Proposer: the account that specifies a proposal key.
	Proposer *flowkit.Account

	//The payload signers that will sign the payload
	//Authorizers: zero or more accounts authorizing the transaction to mutate their state.
	PayloadSigners []*flowkit.Account

	//The gas limit to set for this given interaction
	GasLimit uint64

	//The basepath on where to look for interactions
	BasePath string

	//An error object to store errors that arrive as you configure an interaction
	Error error

	//The code of the tranasction in bytes
	TransactionCode []byte

	//The named arguments
	NamedArgs map[string]interface{}

	//Event filters to apply to the interaction
	EventFilter OverflowEventFilter

	//Wheter to ignore global event filters from OverflowState or not
	IgnoreGlobalEventFilters bool

	//Options to use when printing results
	PrintOptions *[]OverflowPrinterOption
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (f OverflowInteractionBuilder) Test(t *testing.T) OverflowTransactionResult {
	locale, _ := time.LoadLocation("UTC")
	time.Local = locale
	result := f.Send()
	var formattedEvents []*OverflowFormatedEvent
	for _, event := range result.RawEvents {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		formattedEvents = append(formattedEvents, ev)
	}
	return OverflowTransactionResult{
		Err:     result.Err,
		Events:  formattedEvents,
		Result:  result,
		Testing: t,
	}
}

// get the contract code
func (oib OverflowInteractionBuilder) getContractCode(codeFileName string) ([]byte, error) {
	code := []byte(oib.Content)
	var err error
	if oib.Content == "" {
		code, err = oib.Overflow.State.ReaderWriter().ReadFile(codeFileName)
		if err != nil {
			return nil, fmt.Errorf("%v Could not read interaction file from path=%s", emoji.PileOfPoo, codeFileName)
		}
	}
	return code, nil
}

// A function to customize the transaction builder
type OverflowInteractionOption func(*OverflowInteractionBuilder)

// force no printing for this interaction
func WithoutLog() OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.NoLog = true
	}
}

// set a list of args as key, value in an interaction, see Arg for options you can pass in
func WithArgs(args ...interface{}) OverflowInteractionOption {

	return func(oib *OverflowInteractionBuilder) {
		if len(args)%2 != 0 {
			oib.Error = fmt.Errorf("Please send in an even number of string : interface{} pairs")
			return
		}
		var i = 0
		for i < len(args) {
			key := args[0]
			value, labelOk := key.(string)
			if !labelOk {
				oib.Error = fmt.Errorf("even parameters in Args needs to be strings")
			}
			oib.NamedArgs[value] = args[1]
			i = i + 2
		}
	}
}

// set arguments to the interaction from a map. See Arg for options on what you can pass in
func WithArgsMap(args map[string]interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		for key, value := range args {
			oib.NamedArgs[key] = value
		}
	}
}

// set the name of this interaction, for inline interactions this will be the entire name for file interactions they will be combined
func WithName(name string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Name = name
	}
}

// Send an argument into a transaction
//
// The value is treated in the given way depending on type
// - cadence.Value is sent as straight argument
// - string argument are resolved into cadence.Value using flowkit
// - ofther values are converted to string with %v and resolved into cadence.Value using flowkit
// - if the type of the parameter is Address and the string you send in is a valid account in flow.json it will resolve
func WithArg(name string, value interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.NamedArgs[name] = value
	}
}

// Send an list of structs into a transaction

// use the `cadence` struct tag to name a field or it will be given the lowercase name of the field
func WithStructArgsCustomQualifier(name string, resolver InputResolver, values ...interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {

		array := []cadence.Value{}
		for _, value := range values {
			structValue, err := InputToCadence(value, resolver)
			if err != nil {
				oib.Error = err
				return
			}
			array = append(array, structValue)
		}
		oib.NamedArgs[name] = cadence.NewArray(array)
	}
}

// Send an struct as argument into a transaction

// use the `cadence` struct tag to name a field or it will be given the lowercase name of the field
func WithStructArgCustomResolver(name string, resolver InputResolver, value interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		structValue, err := InputToCadence(value, resolver)
		if err != nil {
			oib.Error = err
			return
		}
		oib.NamedArgs[name] = structValue
	}
}

// sending in a timestamp as an arg is quite complicated, use this method with the name of the arg, the datestring and the given timezone to parse it at
func WithArgDateTime(name string, dateString string, timezone string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		value, err := parseTime(dateString, timezone)
		if err != nil {
			oib.Error = err
			return
		}

		//swallow the error since it will never happen here, we control the input
		amount, _ := cadence.NewUFix64(value)

		oib.NamedArgs[name] = amount
	}
}

// Send in an array of addresses as an argument
func WithAddresses(name string, value ...string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		array := []cadence.Value{}

		for _, val := range value {
			account, err := oib.Overflow.AccountE(val)
			if err != nil {
				address, err := hexToAddress(val)
				if err != nil {
					oib.Error = errors.Wrap(err, fmt.Sprintf("%s is not an valid account name or an address", val))
					return
				}
				cadenceAddress := cadence.BytesToAddress(address.Bytes())
				array = append(array, cadenceAddress)
			} else {
				cadenceAddress := cadence.BytesToAddress(account.Address().Bytes())
				array = append(array, cadenceAddress)
			}
		}
		oib.NamedArgs[name] = cadence.NewArray(array)
	}
}

// print interactions using the following options
func WithPrintOptions(opts ...OverflowPrinterOption) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		if oib.PrintOptions == nil {
			oib.PrintOptions = &opts
		} else {
			allOpts := *oib.PrintOptions
			allOpts = append(allOpts, opts...)
			oib.PrintOptions = &allOpts
		}
	}
}

// set the proposer
func WithProposer(proposer string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		account, err := oib.Overflow.AccountE(proposer)
		if err != nil {
			oib.Error = err
			return
		}
		oib.Proposer = account
	}
}

// set the propser to be the service account
func WithProposerServiceAccount() OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		key := oib.Overflow.ServiceAccountName()
		account, _ := oib.Overflow.State.Accounts().ByName(key)
		oib.Proposer = account
	}
}

// set payer, proposer authorizer as the signer
func WithSigner(signer string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		account, err := oib.Overflow.AccountE(signer)
		if err != nil {
			oib.Error = err
			return
		}
		oib.Payer = account
		oib.Proposer = account
	}
}

// set service account as payer, proposer, authorizer
func WithSignerServiceAccount() OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		key := oib.Overflow.ServiceAccountName()
		account, _ := oib.Overflow.State.Accounts().ByName(key)
		oib.Payer = account
		oib.Proposer = account
	}
}

// set the gas limit
func WithMaxGas(gas uint64) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.GasLimit = gas
	}
}

// set a filter for events
func WithEventsFilter(filter OverflowEventFilter) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.EventFilter = filter
	}
}

// ignore global events filters defined on OverflowState
func WithoutGlobalEventFilter() OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.IgnoreGlobalEventFilters = true
	}
}

// set an aditional authorizer that will sign the payload
func WithPayloadSigner(signer ...string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		for _, signer := range signer {
			account, err := oib.Overflow.AccountE(signer)
			if err != nil {
				oib.Error = err
				return
			}
			oib.PayloadSigners = append(oib.PayloadSigners, account)
		}
	}
}

// Send a interaction builder as a Transaction returning an overflow result
func (oib OverflowInteractionBuilder) Send() *OverflowResult {
	result := &OverflowResult{
		StopOnError:     oib.Overflow.StopOnError,
		Err:             nil,
		Id:              [32]byte{},
		Meter:           &OverflowMeter{},
		RawLog:          []OverflowEmulatorLogMessage{},
		EmulatorLog:     []string{},
		ComputationUsed: 0,
		RawEvents:       []flow.Event{},
		Events:          map[string]OverflowEventList{},
		Transaction:     &flow.Transaction{},
		Fee:             map[string]interface{}{},
		FeeGas:          0,
		Name:            "",
		Arguments:       oib.NamedCadenceArguments,
	}
	if oib.Error != nil {
		result.Err = oib.Error
		return result
	}

	if oib.Proposer == nil {
		result.Err = fmt.Errorf("%v You need to set the proposer signer", emoji.PileOfPoo)
		return result
	}

	codeFileName := fmt.Sprintf("%s/%s.cdc", oib.BasePath, oib.FileName)

	if len(oib.TransactionCode) == 0 {
		code, err := oib.getContractCode(codeFileName)
		if err != nil {
			result.Err = err
			return result
		}
		oib.TransactionCode = code
	}

	oib.Overflow.Log.Reset()
	oib.Overflow.EmulatorLog.Reset()
	/*
		❗ Special case: if an account is both the payer and either a proposer or authorizer, it is only required to sign the envelope.
	*/
	// we append the payer at the end here so that it signs last
	signers := oib.PayloadSigners
	payer := oib.Payer
	if oib.Payer != nil {
		signers = append(signers, oib.Payer)
	}

	var authorizers []flow.Address
	for _, signer := range signers {
		authorizers = append(authorizers, signer.Address())
	}

	if oib.Payer == nil {
		payer = oib.Proposer
		signers = append(signers, oib.Proposer)
	}

	script := &services.Script{
		Code:     oib.TransactionCode,
		Args:     oib.Arguments,
		Filename: codeFileName,
	}
	addresses := services.NewTransactionAddresses(oib.Proposer.Address(), payer.Address(), authorizers)

	tx, err := oib.Overflow.Services.Transactions.Build(
		addresses,
		oib.Proposer.Key().Index(),
		script,
		oib.GasLimit,
		oib.Overflow.Network,
	)
	if err != nil {
		result.Err = err
		return result
	}

	for _, signer := range signers {
		err = tx.SetSigner(signer)
		if err != nil {
			result.Err = err
			return result
		}

		tx, err = tx.Sign()
		if err != nil {
			result.Err = err
			return result
		}
	}
	txId := tx.FlowTransaction().ID()
	result.Id = txId

	ftx, res, err := oib.Overflow.Services.Transactions.SendSigned(tx)
	result.Transaction = ftx

	if err != nil {
		result.Err = err
		return result
	}

	logMessage, err := oib.Overflow.readLog()
	if err != nil {
		result.Err = err
	}
	result.RawLog = logMessage

	result.Meter = &OverflowMeter{}
	var meter OverflowMeter
	scanner := bufio.NewScanner(oib.Overflow.EmulatorLog)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, "transaction execution data") {
			err = json.Unmarshal([]byte(txt), &meter)
			if err == nil {
				result.Meter = &meter
			}
		}
	}
	messages := []string{}
	for _, msg := range logMessage {
		if msg.ComputationUsed != 0 {
			result.ComputationUsed = msg.ComputationUsed
		}
		messages = append(messages, msg.Msg)
	}

	result.EmulatorLog = messages

	result.RawEvents = res.Events

	overflowEvents, fee := parseEvents(result.RawEvents)
	result.Fee = fee.Fields
	if len(result.Fee) != 0 {
		executionEffort, ok := result.Fee["executionEffort"].(float64)
		if !ok {
			result.Err = fmt.Errorf("Type conversion failed on execution effort of fee")
		}
		factor := 100000000
		gas := int(math.Round(executionEffort * float64(factor)))
		result.FeeGas = gas
	}

	if !oib.IgnoreGlobalEventFilters {

		fee := result.Fee["amount"]
		if oib.Overflow.FilterOutFeeEvents && fee != nil {
			overflowEvents = overflowEvents.FilterFees(fee.(float64))
		}

		if oib.Overflow.FilterOutEmptyWithDrawDepositEvents {
			overflowEvents = overflowEvents.FilterTempWithdrawDeposit()
		}

		if len(oib.Overflow.GlobalEventFilter) != 0 {
			overflowEvents = overflowEvents.FilterEvents(oib.Overflow.GlobalEventFilter)
		}
	}

	if len(oib.EventFilter) != 0 {
		overflowEvents = overflowEvents.FilterEvents(oib.EventFilter)
	}

	result.Events = overflowEvents

	result.Name = oib.Name
	oib.Overflow.Log.Reset()
	oib.Overflow.EmulatorLog.Reset()
	result.Err = res.Error
	return result
}
