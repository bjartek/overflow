package overflow

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

//Deprecated use the new Tx() method and OverflowResult
type TransactionResult struct {
	Err     error
	Events  []*FormatedEvent
	Result  *OverflowResult
	Testing *testing.T
}

// Evertyhing from here down is deprecated

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (f FlowInteractionBuilder) Test(t *testing.T) TransactionResult {
	locale, _ := time.LoadLocation("UTC")
	time.Local = locale
	result := f.Send()
	var formattedEvents []*FormatedEvent
	for _, event := range result.RawEvents {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		formattedEvents = append(formattedEvents, ev)
	}
	return TransactionResult{
		Err:     result.Err,
		Events:  formattedEvents,
		Result:  result,
		Testing: t,
	}
}

func (o OverflowResult) AssertFailure(t *testing.T, msg string) OverflowResult {
	t.Helper()
	assert.Error(t, o.Err)
	if o.Err != nil {
		assert.Contains(t, o.Err.Error(), msg)
	}
	return o
}

func (o OverflowResult) AssertSuccess(t *testing.T) OverflowResult {
	t.Helper()
	assert.NoError(t, o.Err)
	return o
}

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

func (o OverflowResult) AssertNoEvents(t *testing.T) OverflowResult {
	t.Helper()
	res := assert.Empty(t, o.Events)
	o.logEventsFailure(t, res)
	return o
}

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

func (o OverflowResult) AssertEmulatorLog(t *testing.T, message string) OverflowResult {

	for _, log := range o.EmulatorLog {
		if strings.Contains(log, message) {
			return o
		}
	}

	assert.Fail(t, "No emulator log contain message "+message, o.EmulatorLog)

	return o
}

func (o OverflowResult) AssertComputationLessThenOrEqual(t *testing.T, computation int) OverflowResult {
	assert.LessOrEqual(t, o.ComputationUsed, computation)
	return o
}

func (o OverflowResult) AssertComputationUsed(t *testing.T, computation int) OverflowResult {
	assert.Equal(t, computation, o.ComputationUsed)
	return o
}

//Deprecated use the new Tx() method and Asserts on the result
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

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertFailure(msg string) TransactionResult {
	assert.Error(t.Testing, t.Err)
	if t.Err != nil {
		assert.Contains(t.Testing, t.Err.Error(), msg)
	}
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertSuccess() TransactionResult {
	t.Testing.Helper()

	if t.Err != nil {
		assert.Fail(t.Testing, fmt.Sprintf("Received unexpected error:\n%+v", t.Err), fmt.Sprintf("transactionName:%s", t.Result.Name))
	}
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEventCount(number int) TransactionResult {
	assert.Equal(t.Testing, len(t.Events), number)
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertNoEvents() TransactionResult {
	res := assert.Empty(t.Testing, t.Events)

	t.logFailure(res)
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) logFailure(res bool) {
	if !res {
		for _, ev := range t.Events {
			t.Testing.Log(litter.Sdump(ev))
		}
	}
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEmitEventNameShortForm(event ...string) TransactionResult {
	var eventNames []string
	for _, fe := range t.Events {
		eventNames = append(eventNames, fe.ShortName())
	}

	res := false
	for _, ev := range event {
		if assert.Contains(t.Testing, eventNames, ev) {
			res = true
		}
	}

	t.logFailure(res)

	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEmitEventName(event ...string) TransactionResult {
	var eventNames []string
	for _, fe := range t.Events {
		eventNames = append(eventNames, fe.Name)
	}

	res := false
	for _, ev := range event {
		if assert.Contains(t.Testing, eventNames, ev) {
			res = true
		}
	}

	t.logFailure(res)

	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEmitEventJson(event ...string) TransactionResult {

	var jsonEvents []string
	for _, fe := range t.Events {
		jsonEvents = append(jsonEvents, fe.String())
	}

	res := true
	for _, ev := range event {
		//TODO: keep as before if this fails
		if !slices.Contains(jsonEvents, ev) {
			assert.Fail(t.Testing, fmt.Sprintf("event not found %s", litter.Sdump(ev)))
			res = false
		}
	}

	t.logFailure(res)
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertPartialEvent(expected *FormatedEvent) TransactionResult {

	events := t.Events
	newEvents := []*FormatedEvent{}
	for _, ev := range events {
		//todo do we need more then just name here?
		if ev.Name == expected.Name {
			fields := map[string]interface{}{}
			for key, value := range ev.Fields {
				_, exist := expected.Fields[key]
				if exist {
					fields[key] = value
				}
			}

			if len(fields) > 0 {
				newEvents = append(newEvents, &FormatedEvent{
					Name:        ev.Name,
					Time:        ev.Time,
					BlockHeight: ev.BlockHeight,
					Fields:      fields,
				})
			}
		}
	}
	if !expected.ExistIn(newEvents) {
		assert.Fail(t.Testing, fmt.Sprintf("event not found %s", litter.Sdump(expected)))
		t.logFailure(false)
	}

	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEmitEvent(event ...*FormatedEvent) TransactionResult {
	res := true
	for _, ev := range event {
		//This is not a compile error

		if !ev.ExistIn(t.Events) {
			assert.Fail(t.Testing, fmt.Sprintf("event not found %s", litter.Sdump(ev)))
			res = false
		}
	}

	t.logFailure(res)

	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertDebugLog(message ...string) TransactionResult {
	var logMessages []interface{}
	for _, fe := range t.Events {
		if strings.HasSuffix(fe.Name, "Debug.Log") {
			logMessages = append(logMessages, fe.Fields["msg"])
		}
	}

	for _, ev := range message {
		assert.Contains(t.Testing, logMessages, ev)
	}
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEmulatorLog(message string) TransactionResult {

	for _, log := range t.Result.EmulatorLog {
		if strings.Contains(log, message) {
			return t
		}
	}

	assert.Fail(t.Testing, "No emulator log contain message "+message, t.Result.EmulatorLog)

	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertComputationLessThenOrEqual(computation int) TransactionResult {
	assert.LessOrEqual(t.Testing, t.Result.ComputationUsed, computation)
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertComputationUsed(computation int) TransactionResult {
	assert.Equal(t.Testing, computation, t.Result.ComputationUsed)
	return t
}

//Deprecated use the new Tx() method and Asserts on the result
func (t TransactionResult) GetIdFromEvent(eventName string, fieldName string) uint64 {

	for _, ev := range t.Events {
		if ev.Name == eventName {
			return ev.GetFieldAsUInt64(fieldName)
		}
	}
	panic("from event name of fieldname")

}
