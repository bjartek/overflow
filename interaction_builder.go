package overflow

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/bjartek/underflow"
	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/transactions"
	"github.com/pkg/errors"
)

// Flow Interaction Builder
//
// An interaction in overflow is either a script or a transaction

// OverflowInteractionBuilder used to create a builder pattern for an interaction
type OverflowInteractionBuilder struct {
	Ctx context.Context
	// the name of the integration, for inline variants
	Name string

	// force that this interaction will not print log, even if overflow state has specified it
	NoLog bool

	// The underlying state of overflow used to fetch some global settings
	Overflow *OverflowState

	// The file name of the interaction
	FileName string

	// The content of the interaction
	Content string

	// The list of raw arguments
	Arguments []cadence.Value

	NamedCadenceArguments CadenceArguments

	// The main signer used to sign the transaction
	// Payer: the account paying for the transaction fees.
	Payer *accounts.Account

	// The propser account
	//    Proposer: the account that specifies a proposal key.
	Proposer *accounts.Account

	// The payload signers that will sign the payload
	// Authorizers: zero or more accounts authorizing the transaction to mutate their state.
	PayloadSigners []*accounts.Account

	// The gas limit to set for this given interaction
	GasLimit uint64

	// The basepath on where to look for interactions
	BasePath string

	// An error object to store errors that arrive as you configure an interaction
	Error error

	// The code of the tranasction in bytes
	TransactionCode []byte

	// The named arguments
	NamedArgs map[string]interface{}

	// Event filters to apply to the interaction
	EventFilter OverflowEventFilter

	// Whether to ignore global event filters from OverflowState or not
	IgnoreGlobalEventFilters bool

	// Options to use when printing results
	PrintOptions *[]OverflowPrinterOption

	// Query to use for running scripts
	ScriptQuery *flowkit.ScriptQuery

	//
	StopOnError *bool

	Testing OverflowTestingAsssertions

	AutoSigner bool
}

type OverflowTestingAsssertions struct {
	T       *testing.T
	Failure *string
	Events  []EventAssertion
	Require bool
}

type EventAssertion struct {
	Fields  map[string]interface{}
	Suffix  string
	Require bool
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

func WithContext(ctx context.Context) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Ctx = ctx
	}
}

// set a list of args as key, value in an interaction, see Arg for options you can pass in
func WithArgs(args ...interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		if len(args)%2 != 0 {
			oib.Error = fmt.Errorf("please send in an even number of string : interface{} pairs")
			return
		}
		i := 0
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
func WithStructArgsCustomQualifier(name string, resolver underflow.InputResolver, values ...interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		array := []cadence.Value{}
		for _, value := range values {
			structValue, err := underflow.InputToCadence(value, resolver)
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
func WithStructArgCustomResolver(name string, resolver underflow.InputResolver, value interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		structValue, err := underflow.InputToCadence(value, resolver)
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

		// swallow the error since it will never happen here, we control the input
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
				cadenceAddress := cadence.BytesToAddress(account.Address.Bytes())
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

// set payer, proposer authorizer as the signer
func WithManualProposer(account *accounts.Account) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
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
func WithManualSigner(account *accounts.Account) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Payer = account
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

// set payer, proposer authorizer as the signer
func WithAutoSigner() OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.AutoSigner = true
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

func WithAuthorizer(signer ...string) OverflowInteractionOption {
	return WithPayloadSigner(signer...)
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

// alias for adding manual payload signers
func WithManualAuthorizer(signer ...*accounts.Account) OverflowInteractionOption {
	return WithManualPayloadSigner(signer...)
}

// set an aditional authorizer that will sign the payload
func WithManualPayloadSigner(signer ...*accounts.Account) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.PayloadSigners = append(oib.PayloadSigners, signer...)
	}
}

// set what block height to execute a script at! NB! if very old will not work on normal AN
func WithExecuteScriptAtBlockHeight(height uint64) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.ScriptQuery = &flowkit.ScriptQuery{Height: height}
	}
}

// set what block height to execute a script at! NB! if very old will not work on normal AN
func WithExecuteScriptAtBlockIdHex(blockId string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.ScriptQuery = &flowkit.ScriptQuery{ID: flow.HexToID(blockId)}
	}
}

// set what block height to execute a script at! NB! if very old will not work on normal AN
func WithExecuteScriptAtBlockIdentifier(blockId flow.Identifier) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.ScriptQuery = &flowkit.ScriptQuery{ID: blockId}
	}
}

func WithPanicInteractionOnError(stop bool) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.StopOnError = &stop
	}
}

func WithAssertFailure(t *testing.T, message string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Testing.T = t
		oib.Testing.Failure = &message
		oib.Testing.Require = false
	}
}

func WithRequireFailure(t *testing.T, message string) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Testing.T = t
		oib.Testing.Failure = &message
		oib.Testing.Require = true
	}
}

// a helper to modify an event assertion if you have a sigle one and you want to change the value
func WithAssertEventReplaceField(suffix string, field string, value interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		for i, ev := range oib.Testing.Events {
			if ev.Suffix == suffix {
				oib.Testing.Events[i].Fields[field] = value
				return
			}
		}
	}
}

func WithAssertEvent(t *testing.T, suffix string, fields map[string]interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Testing.T = t

		oib.Testing.Events = append(oib.Testing.Events, EventAssertion{
			Fields:  fields,
			Suffix:  suffix,
			Require: false,
		})
		oib.Testing.Require = false
	}
}

func WithRequireEvent(t *testing.T, suffix string, fields map[string]interface{}) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Testing.T = t

		oib.Testing.Events = append(oib.Testing.Events, EventAssertion{
			Fields:  fields,
			Suffix:  suffix,
			Require: true,
		})
		oib.Testing.Require = false
	}
}

func WithEventAssertions(t *testing.T, ea ...EventAssertion) OverflowInteractionOption {
	return func(oib *OverflowInteractionBuilder) {
		oib.Testing.T = t

		oib.Testing.Events = append(oib.Testing.Events, ea...)
	}
}

// Send a interaction builder as a Transaction returning an overflow result
func (oib OverflowInteractionBuilder) Send() *OverflowResult {
	result := &OverflowResult{
		StopOnError:      oib.Overflow.StopOnError,
		Err:              nil,
		Id:               [32]byte{},
		Meter:            &OverflowMeter{},
		RawLog:           []OverflowEmulatorLogMessage{},
		EmulatorLog:      []string{},
		ComputationUsed:  0,
		RawEvents:        []flow.Event{},
		Events:           map[string]OverflowEventList{},
		Transaction:      &flow.Transaction{},
		Fee:              map[string]interface{}{},
		FeeGas:           0,
		Name:             "",
		Arguments:        oib.NamedCadenceArguments,
		UnderflowOptions: oib.Overflow.UnderflowOptions,
	}
	if oib.StopOnError != nil {
		result.StopOnError = *oib.StopOnError
	}
	if oib.Error != nil {
		result.Err = oib.Error
		return result
	}

	codeFileName := fmt.Sprintf("%s/%s.cdc", oib.BasePath, oib.FileName)

	result.DeclarationInfo = *declarationInfo(oib.TransactionCode)

	// if we have more then one should the following be payload signers?
	if oib.AutoSigner {
		if len(result.DeclarationInfo.Authorizers) != 1 {
			result.Err = errors.New("currently do not support more then 1 signer when using authSigner")
			return result
		}

		account, err := oib.Overflow.AccountE(result.DeclarationInfo.Authorizers[0].Name)
		if err != nil {
			result.Err = err
			return result
		}
		oib.Payer = account
		oib.Proposer = account
	}

	if oib.Proposer == nil {
		result.Err = fmt.Errorf("%v You need to set the proposer signer", emoji.PileOfPoo)
		return result
	}
	oib.Overflow.Log.Reset()
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
		authorizers = append(authorizers, signer.Address)
	}

	if oib.Payer == nil {
		payer = oib.Proposer
		signers = append(signers, oib.Proposer)
	}

	script := flowkit.Script{
		Code:     oib.TransactionCode,
		Args:     oib.Arguments,
		Location: codeFileName,
	}

	addresses := transactions.AddressesRoles{
		Proposer:    oib.Proposer.Address,
		Authorizers: authorizers,
		Payer:       payer.Address,
	}

	tx, err := oib.Overflow.Flowkit.BuildTransaction(
		oib.Ctx,
		addresses,
		oib.Proposer.Key.Index(),
		script,
		oib.GasLimit,
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

	ftx, res, err := oib.Overflow.Flowkit.SendSignedTransaction(oib.Ctx, tx)
	result.Transaction = ftx
	result.TransactionResult = res

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
	messages := []string{}
	for _, msg := range logMessage {

		if msg.ComputationUsed != 0 {
			result.ComputationUsed = msg.ComputationUsed
		}
		if strings.Contains(msg.Msg, "transaction execution data") {
			var meter OverflowMeter
			bytes, _ := json.Marshal(msg.Fields)
			err = json.Unmarshal(bytes, &meter)
			if err == nil {
				result.Meter = &meter
			}
			continue
		}

		messages = append(messages, msg.String())
	}
	result.EmulatorLog = messages

	result.RawEvents = res.Events

	overflowEvents, fee := oib.Overflow.ParseEvents(result.RawEvents)
	result.Fee = fee.Fields
	if len(result.Fee) != 0 {
		executionEffort, ok := result.Fee["executionEffort"].(float64)
		if !ok {
			result.Err = fmt.Errorf("type conversion failed on execution effort of fee")
		}
		factor := 100000000
		gas := int(math.Round(executionEffort * float64(factor)))
		result.FeeGas = gas
	}

	feeAmount := result.Fee["amount"]
	eventsWithoutFees, feeFromEvents := overflowEvents.FilterFees(feeAmount.(float64), fmt.Sprintf("0x%s", result.Transaction.Payer.Hex()))
	result.Balance = feeFromEvents

	if !oib.IgnoreGlobalEventFilters {
		if oib.Overflow.FilterOutFeeEvents && feeAmount != nil {
			overflowEvents = eventsWithoutFees
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
	result.Err = errors.Wrapf(res.Error, "transaction=%s", codeFileName)

	if result.Err != nil && result.StopOnError {
		panic(result.Err)
	}

	return result
}
