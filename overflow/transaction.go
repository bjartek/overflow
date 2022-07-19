package overflow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/onflow/flow-go-sdk"
)

//The main function for running an transasction in overflow
func (o *OverflowState) Tx(filename string, opts ...InteractionOption) *OverflowResult {
	result := o.BuildInteraction(filename, "transaction", opts...).Send()

	if o.PrintOptions != nil {
		po := *o.PrintOptions
		result.Print(po...)
	}
	if o.StopOnError && result.Err != nil {
		panic(result.Err)
	}

	return result
}

//Composition functions for Transactions
type TransactionFunction func(filename string, opts ...InteractionOption) *OverflowResult
type TransactionOptsFunction func(opts ...InteractionOption) *OverflowResult

// If you store this in a struct and add arguments to it it will not reset between calls
func (o *OverflowState) TxFN(outerOpts ...InteractionOption) TransactionFunction {

	return func(filename string, opts ...InteractionOption) *OverflowResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Tx(filename, outerOpts...)

	}
}

func (o *OverflowState) TxFileNameFN(filename string, outerOpts ...InteractionOption) TransactionOptsFunction {

	return func(opts ...InteractionOption) *OverflowResult {
		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Tx(filename, outerOpts...)
	}
}

//OverflowResult represents the state after running an transaction
type OverflowResult struct {
	//The error if any
	Err error

	//The id of the transaction
	Id flow.Identifier

	//If running on an emulator
	//the meter that contains useful debug information on memory and interactions
	Meter *Meter
	//The Raw log from the emulator
	RawLog []LogrusMessage
	// The log from the emulator
	EmulatorLog []string

	//The computation used
	ComputationUsed int

	//The raw unfiltered events
	RawEvents []flow.Event

	//Events that are filtered and parsed into a terse format
	Events OverflowEvents

	//The underlying transaction if we need to look into that
	Transaction *flow.Transaction

	//TODO: consider marshalling this as a struct for convenience
	//The fee event if any
	Fee map[string]interface{}

	//The name of the Transaction
	Name string
}

func (o OverflowResult) GetIdFromEvent(eventName string, fieldName string) (uint64, error) {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			return event[0][fieldName].(uint64), nil
		}
	}
	return 0, fmt.Errorf("Could not find id field %s in event with suffix %s", fieldName, eventName)
}

func (o OverflowResult) GetIdsFromEvent(eventName string, fieldName string) []uint64 {
	var ids []uint64
	for name, events := range o.Events {
		if strings.HasSuffix(name, eventName) {
			for _, event := range events {
				ids = append(ids, event[fieldName].(uint64))
			}
		}
	}
	return ids
}

func (o OverflowResult) GetEventsWithName(eventName string) []OverflowEvent {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			return event
		}
	}
	return []OverflowEvent{}
}

func (o OverflowState) readLog() ([]LogrusMessage, error) {

	var logMessage []LogrusMessage
	dec := json.NewDecoder(o.Log)
	for {
		var doc LogrusMessage

		err := dec.Decode(&doc)
		if err == io.EOF {
			// all done
			break
		}
		if err != nil {
			return []LogrusMessage{}, err
		}

		logMessage = append(logMessage, doc)
	}

	o.Log.Reset()
	return logMessage, nil

}

// Send a intereaction builder as a Transaction returning an overflow result
func (t FlowInteractionBuilder) Send() *OverflowResult {
	result := &OverflowResult{}
	if t.Error != nil {
		result.Err = t.Error
		return result
	}

	if t.Proposer == nil {
		result.Err = fmt.Errorf("%v You need to set the main signer", emoji.PileOfPoo)
		return result
	}

	codeFileName := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)

	if len(t.TransactionCode) == 0 {
		code, err := t.getContractCode(codeFileName)
		if err != nil {
			result.Err = err
			return result
		}
		t.TransactionCode = code
	}

	t.Overflow.Log.Reset()
	t.Overflow.EmulatorLog.Reset()
	// we append the mainSigners at the end here so that it signs last
	signers := t.PayloadSigners
	if t.MainSigner != nil {
		signers = append(signers, t.MainSigner)
	}

	var authorizers []flow.Address
	for _, signer := range signers {
		authorizers = append(authorizers, signer.Address())
	}
	if t.MainSigner == nil {
		signers = append(signers, t.Proposer)
	}

	tx, err := t.Overflow.Services.Transactions.Build(
		t.Proposer.Address(),
		authorizers,
		t.Proposer.Address(),
		t.Proposer.Key().Index(),
		t.TransactionCode,
		codeFileName,
		t.GasLimit,
		t.Arguments,
		t.Overflow.Network,
		true,
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

	txBytes := []byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode()))
	ftx, res, err := t.Overflow.Services.Transactions.SendSigned(txBytes, true)
	result.Transaction = ftx

	if err != nil {
		result.Err = err
		return result
	}

	logMessage, err := t.Overflow.readLog()
	if err != nil {
		result.Err = err
	}
	result.RawLog = logMessage

	result.Meter = &Meter{}
	var meter Meter
	scanner := bufio.NewScanner(t.Overflow.EmulatorLog)
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

	overflowEvents, fee := ParseEvents(result.RawEvents)
	result.Fee = fee
	if !t.IgnoreGlobalEventFilters {

		fee := result.Fee["amount"]
		if t.Overflow.FilterOutFeeEvents && fee != nil {
			overflowEvents = overflowEvents.FilterFees(fee.(float64))
		}

		if t.Overflow.FilterOutEmptyWithDrawDepositEvents {
			overflowEvents = overflowEvents.FilterTempWithdrawDeposit()
		}

		if len(t.Overflow.GlobalEventFilter) != 0 {
			overflowEvents = overflowEvents.FilterEvents(t.Overflow.GlobalEventFilter)
		}
	}

	if len(t.EventFilter) != 0 {
		overflowEvents = overflowEvents.FilterEvents(t.EventFilter)
	}

	result.Events = overflowEvents

	result.Name = t.FileName
	t.Overflow.Log.Reset()
	t.Overflow.EmulatorLog.Reset()
	result.Err = res.Error
	return result
}
