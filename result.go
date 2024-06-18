package overflow

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bjartek/underflow"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CadenceArguments map[string]cadence.Value

// a type to define a function used to compose Transaction interactions
type OverflowTransactionFunction func(filename string, opts ...OverflowInteractionOption) *OverflowResult

// a type to define a function used to compose Transaction interactions
type OverflowTransactionOptsFunction func(opts ...OverflowInteractionOption) *OverflowResult

// OverflowResult represents the state after running an transaction
type OverflowResult struct {
	StopOnError bool
	// The error if any
	Err error

	// The id of the transaction
	Id flow.Identifier

	// If running on an emulator
	// the meter that contains useful debug information on memory and interactions
	Meter *OverflowMeter
	// The Raw log from the emulator
	RawLog []OverflowEmulatorLogMessage
	// The log from the emulator
	EmulatorLog []string

	// The computation used
	ComputationUsed int

	// The raw unfiltered events
	RawEvents []flow.Event

	// Events that are filtered and parsed into a terse format
	Events OverflowEvents

	// The underlying transaction if we need to look into that
	Transaction *flow.Transaction

	// The transaction result if we need to look into that
	TransactionResult *flow.TransactionResult

	// TODO: consider marshalling this as a struct for convenience
	// The fee event if any
	Fee    map[string]interface{}
	FeeGas int

	// The name of the Transaction
	Name string

	Arguments        CadenceArguments
	UnderflowOptions underflow.Options
	DeclarationInfo  OverflowDeclarationInfo
}

func (o OverflowResult) PrintArguments(t *testing.T) {
	printOrLog(t, "=== Arguments ===")
	maxLength := 0
	for name := range o.Arguments {
		if len(name) > maxLength {
			maxLength = len(name)
		}
	}

	format := fmt.Sprintf("%%%ds -> %%v", maxLength)

	for name, arg := range o.Arguments {
		value, err := underflow.CadenceValueToJsonStringWithOption(arg, o.UnderflowOptions)
		if err != nil {
			panic(err)
		}
		printOrLog(t, fmt.Sprintf(format, name, value))
	}
}

// Get a uint64 field with the given fieldname(most often an id) from an event with a given suffix
func (o OverflowResult) GetIdFromEvent(eventName string, fieldName string) (uint64, error) {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			return event[0].Fields[fieldName].(uint64), nil
		}
	}
	err := fmt.Errorf("could not find id field %s in event with suffix %s", fieldName, eventName)
	if o.StopOnError {
		panic(err)
	}
	return 0, err
}

// Get a byteArray field with the given fieldname from an event with a given suffix
func (o OverflowResult) GetByteArrayFromEvent(eventName string, fieldName string) ([]byte, error) {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			return getByteArray(event[0].Fields[fieldName])
		}
	}
	err := fmt.Errorf("could not find field %s in event with suffix %s", fieldName, eventName)
	if o.StopOnError {
		panic(err)
	}
	return nil, err
}

func (o OverflowResult) GetIdsFromEvent(eventName string, fieldName string) []uint64 {
	var ids []uint64
	for name, events := range o.Events {
		if strings.HasSuffix(name, eventName) {
			for _, event := range events {
				ids = append(ids, event.Fields[fieldName].(uint64))
			}
		}
	}
	return ids
}

// Get all events that end with the given suffix
func (o OverflowResult) GetEventsWithName(eventName string) []OverflowEvent {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			return event
		}
	}
	return []OverflowEvent{}
}

// Get all events that end with the given suffix
func (o OverflowResult) MarshalEventsWithName(eventName string, result interface{}) error {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			err := event.MarshalAs(result)
			return err
		}
	}
	return nil
}

// Assert that this particular transaction was a failure that has a message that contains the sendt in assertion
func (o OverflowResult) RequireFailure(t *testing.T, msg string) OverflowResult {
	t.Helper()

	require.Error(t, o.Err)
	if o.Err != nil {
		require.Contains(t, o.Err.Error(), msg)
	}
	return o
}

// Assert that this particular transaction was a failure that has a message that contains the sendt in assertion
func (o OverflowResult) AssertFailure(t *testing.T, msg string) OverflowResult {
	t.Helper()

	assert.Error(t, o.Err)
	if o.Err != nil {
		assert.Contains(t, o.Err.Error(), msg)
	}
	return o
}

// Require that this transaction was an success
func (o OverflowResult) RequireSuccess(t *testing.T) OverflowResult {
	t.Helper()
	require.NoError(t, o.Err)
	return o
}

// Assert that this transaction was an success
func (o OverflowResult) AssertSuccess(t *testing.T) OverflowResult {
	t.Helper()
	assert.NoError(t, o.Err)
	return o
}

// Assert that the event with the given name suffix and fields are present
func (o OverflowResult) AssertEvent(t *testing.T, name string, fields map[string]interface{}) OverflowResult {
	t.Helper()
	newFields := OverflowEvent{Fields: map[string]interface{}{}}
	for key, value := range fields {
		if value != nil {
			newFields.Fields[key] = value
		}
	}
	hit := false
	for eventName, events := range o.Events {
		if strings.HasSuffix(eventName, name) {
			hit = true
			newEvents := []OverflowEvent{}
			for _, event := range events {
				oe := OverflowEvent{Fields: map[string]interface{}{}}
				valid := false
				for key, value := range event.Fields {
					_, exist := newFields.Fields[key]
					if exist {
						oe.Fields[key] = value
						valid = true
					}
				}
				if valid {
					newEvents = append(newEvents, oe)
				}
			}

			if !newFields.ExistIn(newEvents) {
				assert.Fail(t, fmt.Sprintf("transaction %s missing event %s with fields %s", o.Name, name, litter.Sdump(newFields.Fields)))
				newEventsMap := OverflowEvents{eventName: newEvents}
				newEventsMap.Print(t)
			}
		}
	}
	if !hit {
		assert.Fail(t, fmt.Sprintf("event not found %s, %s", name, litter.Sdump(newFields)))
		o.Events.Print(t)
	}
	return o
}

// Require that the event with the given name suffix and fields are present
func (o OverflowResult) RequireEvent(t *testing.T, name string, fields map[string]interface{}) OverflowResult {
	t.Helper()
	newFields := OverflowEvent{Fields: map[string]interface{}{}}
	for key, value := range fields {
		if value != nil {
			newFields.Fields[key] = value
		}
	}
	hit := false
	for eventName, events := range o.Events {
		if strings.HasSuffix(eventName, name) {
			hit = true
			newEvents := []OverflowEvent{}
			for _, event := range events {
				oe := OverflowEvent{Fields: map[string]interface{}{}}
				valid := false
				for key, value := range event.Fields {
					_, exist := newFields.Fields[key]
					if exist {
						oe.Fields[key] = value
						valid = true
					}
				}
				if valid {
					newEvents = append(newEvents, oe)
				}
			}

			if !newFields.ExistIn(newEvents) {
				require.Fail(t, fmt.Sprintf("transaction %s missing event %s with fields %s", o.Name, name, litter.Sdump(newFields.Fields)))
				newEventsMap := OverflowEvents{eventName: newEvents}
				newEventsMap.Print(t)
			}
		}
	}
	if !hit {
		require.Fail(t, fmt.Sprintf("event not found %s, %s", name, litter.Sdump(newFields)))
		o.Events.Print(t)
	}
	return o
}

// Assert that the transaction result contains the amount of events
func (o OverflowResult) AssertEventCount(t *testing.T, number int) OverflowResult {
	t.Helper()
	num := 0
	for _, ev := range o.Events {
		num = num + len(ev)
	}
	assert.Equal(t, number, num)

	o.Events.Print(t)
	return o
}

// Assert that this transaction emitted no events
func (o OverflowResult) AssertNoEvents(t *testing.T) OverflowResult {
	t.Helper()
	res := assert.Empty(t, o.Events)
	if !res {
		o.Events.Print(t)
	}
	return o
}

// Assert that events with the given suffixes are present
func (o OverflowResult) AssertEmitEventName(t *testing.T, event ...string) OverflowResult {
	t.Helper()

	eventNames := []string{}
	for name := range o.Events {
		eventNames = append(eventNames, name)
	}

	for _, ev := range event {
		valid := false
		for _, name := range eventNames {
			if strings.HasSuffix(name, ev) {
				valid = true
			}
		}

		if !valid {
			assert.Fail(t, fmt.Sprintf("event with suffix %s not present in %v", ev, eventNames))
		}
	}

	return o
}

// Assert that the internal log of the emulator contains the given message
func (o OverflowResult) AssertEmulatorLog(t *testing.T, message string) OverflowResult {
	t.Helper()

	for _, log := range o.EmulatorLog {
		if strings.Contains(log, message) {
			return o
		}
	}

	assert.Fail(t, "No emulator log contain message "+message, o.EmulatorLog)

	return o
}

// Assert that this transaction did not use more then the given amount of computation
func (o OverflowResult) AssertComputationLessThenOrEqual(t *testing.T, computation int) OverflowResult {
	t.Helper()

	assert.LessOrEqual(t, o.ComputationUsed, computation)
	if o.FeeGas != 0 {
		// TODO: add back in again once fixed
		assert.Equal(t, o.ComputationUsed, o.FeeGas)
	}

	return o
}

// Assert that the transaction uses exactly the given computation amount
func (o OverflowResult) AssertComputationUsed(t *testing.T, computation int) OverflowResult {
	t.Helper()
	assert.Equal(t, computation, o.ComputationUsed)
	if o.FeeGas != 0 {
		// TODO: add back in again once fixed
		assert.Equal(t, o.ComputationUsed, o.FeeGas)
	}
	return o
}

// Assert that a Debug.Log event was emitted that contains the given messages
func (o OverflowResult) AssertDebugLog(t *testing.T, message ...string) OverflowResult {
	t.Helper()
	var logMessages []interface{}
	for name, fe := range o.Events {
		if strings.HasSuffix(name, "Debug.Log") {
			for _, ev := range fe {
				logMessages = append(logMessages, ev.Fields["msg"])
			}
		}
	}
	for _, ev := range message {
		assert.Contains(t, logMessages, ev)
	}
	return o
}

func getByteArray(data interface{}) ([]byte, error) {
	slice, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected a slice of interfaces")
	}
	byteSlice := make([]byte, len(slice))
	for i, val := range slice {
		b, ok := val.(uint8)
		if !ok {
			return nil, fmt.Errorf("unexpected type at index %d", i)
		}
		byteSlice[i] = b
	}
	return byteSlice, nil
}
