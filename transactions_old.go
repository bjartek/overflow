package overflow

import (
	"fmt"
	"time"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
)

// Deprecated: Use Tx()
func (o *OverflowState) SimpleTxArgs(filename string, signer string, args *OverflowArgumentsBuilder) {
	o.TransactionFromFile(filename).SignProposeAndPayAs(signer).Args(args).RunPrintEventsFull()
}

// Deprecated: Use Tx()
//
// TransactionFromFile will start a flow transaction builder
func (o *OverflowState) TransactionFromFile(filename string) OverflowInteractionBuilder {
	return OverflowInteractionBuilder{
		Overflow:       o,
		FileName:       filename,
		Payer:          nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(o.Gas),
		BasePath:       fmt.Sprintf("%s/transactions", o.BasePath),
	}
}

// Deprecated: Use Tx()
//
// Transaction will start a flow transaction builder using the inline transaction
func (o *OverflowState) Transaction(content string) OverflowInteractionBuilder {
	return OverflowInteractionBuilder{
		Overflow:       o,
		FileName:       "inline",
		Content:        content,
		Payer:          nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(o.Gas),
		BasePath:       fmt.Sprintf("%s/transactions", o.BasePath),
	}
}

// Deprecated: Use ArgM
func (t OverflowInteractionBuilder) NamedArguments(args map[string]string) OverflowInteractionBuilder {

	codeFileName := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)
	code, err := t.getContractCode(codeFileName)
	if err != nil {
		t.Error = err
	}
	parseArgs, err := t.Overflow.ParseArgumentsWithoutType(t.FileName, code, args)
	if err != nil {
		t.Error = err
	}
	t.Arguments = parseArgs
	return t
}

// Deprecated: Use Args
//
// Specify arguments to send to transaction using a raw list of values
func (t OverflowInteractionBuilder) ArgsV(args []cadence.Value) OverflowInteractionBuilder {
	t.Arguments = args
	return t
}

// Deprecated: Use Arg
//
// Specify arguments to send to transaction using a builder you send in
func (t OverflowInteractionBuilder) Args(args *OverflowArgumentsBuilder) OverflowInteractionBuilder {
	t.Arguments = args.Build()
	return t
}

// Deprecated: Use Arg
//
// Specify arguments to send to transaction using a function that takes a builder where you call the builder
func (t OverflowInteractionBuilder) ArgsFn(fn func(*OverflowArgumentsBuilder)) OverflowInteractionBuilder {
	args := t.Overflow.Arguments()
	fn(args)
	t.Arguments = args.Build()
	return t
}

// Deprecated: Use Tx function
func (t OverflowInteractionBuilder) TransactionPath(path string) OverflowInteractionBuilder {
	t.BasePath = path
	return t
}

// Deprecated: Use Tx function
//
// Gas sets the gas limit for this transaction
func (t OverflowInteractionBuilder) Gas(limit uint64) OverflowInteractionBuilder {
	t.GasLimit = limit
	return t
}

// Deprecated: Use Tx function
//
// SignProposeAndPayAs set the payer, proposer and envelope signer
func (t OverflowInteractionBuilder) SignProposeAndPayAs(signer string) OverflowInteractionBuilder {
	account, err := t.Overflow.AccountE(signer)
	if err != nil {
		t.Error = err
		return t
	}
	t.Proposer = account
	t.Payer = account
	return t
}

// Deprecated: Use Tx function
//
// SignProposeAndPayAsService set the payer, proposer and envelope signer
func (t OverflowInteractionBuilder) SignProposeAndPayAsService() OverflowInteractionBuilder {
	key := t.Overflow.ServiceAccountName()
	//swallow error as you cannot start a overflow without a valid sa
	account, _ := t.Overflow.State.Accounts().ByName(key)
	t.Payer = account
	t.Proposer = account
	return t
}

// Deprecated: Use Tx function
//
// PayloadSigner set a signer for the payload
func (t OverflowInteractionBuilder) PayloadSigner(value string) OverflowInteractionBuilder {
	signer, err := t.Overflow.AccountE(value)
	if err != nil {
		t.Error = err
		return t
	}
	t.PayloadSigners = append(t.PayloadSigners, signer)
	return t
}

// Deprecated: use Send().PrintEvents()
//
// RunPrintEventsFull will run a transaction and print all events
func (t OverflowInteractionBuilder) RunPrintEventsFull() {
	PrintEvents(t.Run(), map[string][]string{})
}

// Deprecated: use Send().PrintEventsFiltered()
//
// RunPrintEvents will run a transaction and print all events ignoring some fields
func (t OverflowInteractionBuilder) RunPrintEvents(ignoreFields map[string][]string) {
	PrintEvents(t.Run(), ignoreFields)
}

// Deprecated: use Send and get entire result
//
// Run run the transaction
func (t OverflowInteractionBuilder) Run() []flow.Event {
	result := t.Send()
	if result.Err != nil {
		t.Overflow.Logger.Error(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, t.FileName, result.Err))
		panic(result.Err)
	}
	return result.RawEvents
}

// Deprecated: Use Tx().Print().GetIdFromEvent
func (t OverflowInteractionBuilder) RunGetIdFromEventPrintAll(eventName string, fieldName string) uint64 {
	result := t.Send()
	if result.Err != nil {
		panic(result.Err)
	}

	PrintEvents(result.RawEvents, map[string][]string{})

	res, err := result.GetIdFromEvent(eventName, fieldName)
	if err != nil {
		panic(err)
	}
	return res
}

// Deprecated, use Send().GetIdFromEvent
func (t OverflowInteractionBuilder) RunGetIdFromEvent(eventName string, fieldName string) uint64 {

	result := t.Send()
	if result.Err != nil {
		panic(result.Err)
	}
	res, err := result.GetIdFromEvent(eventName, fieldName)
	if err != nil {
		panic(err)
	}
	return res
}

// Deprecated: Use Tx().Print().GetIdsFromEvent
func (t OverflowInteractionBuilder) RunGetIds(eventName string, fieldName string) ([]uint64, error) {

	result := t.Send()
	if result.Err != nil {
		return nil, result.Err
	}
	return result.GetIdsFromEvent(eventName, fieldName), nil
}

// Deprecated: use Tx().GetEventsWithName
func (t OverflowInteractionBuilder) RunGetEventsWithNameOrError(eventName string) ([]OverflowFormatedEvent, error) {

	result := t.Send()
	if result.Err != nil {
		return nil, result.Err
	}
	var events []OverflowFormatedEvent
	for _, event := range result.RawEvents {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			events = append(events, *ev)
		}
	}
	return events, nil

}

// Deprecated: Use Send().GetEventsWithName()
func (t OverflowInteractionBuilder) RunGetEventsWithName(eventName string) []OverflowFormatedEvent {

	result := t.Send()
	if result.Err != nil {
		panic(result.Err)
	}
	var events []OverflowFormatedEvent
	for _, event := range result.RawEvents {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			events = append(events, *ev)
		}
	}
	return events
}

// Deprecated: use Send()
//
// RunE runs returns events and error
func (t OverflowInteractionBuilder) RunE() ([]flow.Event, error) {
	result := t.Send()
	return result.RawEvents, result.Err
}
