package overflow

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
)

// a type to define a function used to compose Transaction interactions
type TransactionFunction func(filename string, opts ...InteractionOption) *OverflowResult

// a type to define a function used to compose Transaction interactions
type TransactionOptsFunction func(opts ...InteractionOption) *OverflowResult

//OverflowResult represents the state after running an transaction
type OverflowResult struct {
	StopOnError bool
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
	Fee    map[string]interface{}
	FeeGas int

	//The name of the Transaction
	Name string
}

// Get a uint64 field with the given fieldname(most often an id) from an event with a given suffix
func (o OverflowResult) GetIdFromEvent(eventName string, fieldName string) (uint64, error) {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			return event[0][fieldName].(uint64), nil
		}
	}
	err := fmt.Errorf("Could not find id field %s in event with suffix %s", fieldName, eventName)
	if o.StopOnError {
		panic(err)

	}
	return 0, err
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

// Get all events that end with the given suffix
func (o OverflowResult) GetEventsWithName(eventName string) []OverflowEvent {
	for name, event := range o.Events {
		if strings.HasSuffix(name, eventName) {
			return event
		}
	}
	return []OverflowEvent{}
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

// Assert that this transation was an success
func (o OverflowResult) AssertSuccess(t *testing.T) OverflowResult {
	t.Helper()
	assert.NoError(t, o.Err)
	return o
}

// Assert that the event with the given name suffix and fields are present
func (o OverflowResult) AssertEvent(t *testing.T, name string, fields OverflowEvent) OverflowResult {
	t.Helper()
	for eventName, events := range o.Events {
		if strings.HasSuffix(eventName, name) {

			newEvents := []OverflowEvent{}
			for _, event := range events {
				oe := OverflowEvent{}
				valid := false
				for key, value := range event {
					_, exist := fields[key]
					if exist {
						oe[key] = value
						valid = true
					}
				}
				if valid {
					newEvents = append(newEvents, oe)
				}
			}

			if !fields.ExistIn(newEvents) {
				assert.Fail(t, fmt.Sprintf("event not found %s", litter.Sdump(fields)))
				o.logEventsFailure(t, false)
			}
		}
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

	o.logEventsFailure(t, false)
	return o
}

// Assert that this transaction emitted no events
func (o OverflowResult) AssertNoEvents(t *testing.T) OverflowResult {
	t.Helper()
	res := assert.Empty(t, o.Events)
	o.logEventsFailure(t, res)
	return o
}

// internal method to log output to give good events output when an event asertion fails
func (o OverflowResult) logEventsFailure(t *testing.T, res bool) {
	t.Helper()
	if !res {
		t.Log("EXISTING EVENTS")
		t.Log("===============")
		events := o.Events
		for name, eventList := range events {
			for _, event := range eventList {
				t.Log(name)
				for key, value := range event {
					t.Log(fmt.Sprintf("   %s:%v", key, value))
				}
			}
		}
	}
}

// Assert that events with the given suffixes are present
func (o OverflowResult) AssertEmitEventName(t *testing.T, event ...string) OverflowResult {

	eventNames := []string{}
	for name, _ := range o.Events {
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
	assert.LessOrEqual(t, o.ComputationUsed, computation)
	if o.FeeGas != 0 {
		assert.Equal(t, o.ComputationUsed, o.FeeGas)
	}
	return o
}

// Assert that the transaction uses exactly the given computation amount
func (o OverflowResult) AssertComputationUsed(t *testing.T, computation int) OverflowResult {
	assert.Equal(t, computation, o.ComputationUsed)
	if o.FeeGas != 0 {
		assert.Equal(t, o.ComputationUsed, o.FeeGas)
	}

	return o
}

// Assert that a Debug.Log event was emitted that contains the given messages
func (o OverflowResult) AssertDebugLog(t *testing.T, message ...string) OverflowResult {
	var logMessages []interface{}
	for name, fe := range o.Events {
		if strings.HasSuffix(name, "Debug.Log") {
			for _, ev := range fe {
				logMessages = append(logMessages, ev["msg"])
			}
		}
	}
	for _, ev := range message {
		assert.Contains(t, logMessages, ev)
	}
	return o
}
