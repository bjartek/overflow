package overflow

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TransactionResult struct {
	Err     error
	Events  []*FormatedEvent
	Result  *OverflowResult
	Testing *testing.T
}

func (o OverflowResult) Test(t *testing.T) TransactionResult {
	locale, _ := time.LoadLocation("UTC")
	time.Local = locale
	var formattedEvents []*FormatedEvent
	for _, event := range o.RawEvents {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		formattedEvents = append(formattedEvents, ev)
	}
	return TransactionResult{
		Err:     o.Err,
		Events:  formattedEvents,
		Result:  &o,
		Testing: t,
	}
}

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

func (t TransactionResult) AssertFailure(msg string) TransactionResult {
	assert.Error(t.Testing, t.Err)
	if t.Err != nil {
		assert.Contains(t.Testing, t.Err.Error(), msg)
	}
	return t
}
func (t TransactionResult) AssertSuccess() TransactionResult {
	assert.NoError(t.Testing, t.Err)
	return t
}

func (t TransactionResult) AssertEventCount(number int) TransactionResult {
	assert.Equal(t.Testing, len(t.Events), number)
	return t
}

func (t TransactionResult) AssertNoEvents() TransactionResult {
	res := assert.Empty(t.Testing, t.Events)

	t.logFailure(res)
	return t
}

func (t TransactionResult) logFailure(res bool) {
	if !res {
		for _, ev := range t.Events {
			t.Testing.Log(ev.String())
		}
	}
}

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

func (t TransactionResult) AssertEmitEventJson(event ...string) TransactionResult {

	var jsonEvents []string
	for _, fe := range t.Events {
		jsonEvents = append(jsonEvents, fe.String())
	}

	res := false
	for _, ev := range event {
		if assert.Contains(t.Testing, jsonEvents, ev) {
			res = true
		}
	}

	t.logFailure(res)
	return t
}

func (t TransactionResult) AssertPartialEvent(expected *FormatedEvent) TransactionResult {

	events := t.Events
	for index, ev := range events {
		//todo do we need more then just name here?
		if ev.Name == expected.Name {
			for key := range ev.Fields {
				_, exist := expected.Fields[key]
				if !exist {
					delete(events[index].Fields, key)
				}
			}
		}
	}

	result := assert.Contains(t.Testing, events, expected)

	t.logFailure(result)

	return t
}
func (t TransactionResult) AssertEmitEvent(event ...*FormatedEvent) TransactionResult {
	res := false
	for _, ev := range event {
		if assert.Contains(t.Testing, t.Events, ev) {
			res = true
		}
	}

	t.logFailure(res)

	return t
}

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

func (t TransactionResult) AssertEmulatorLog(message string) TransactionResult {

	for _, log := range t.Result.EmulatorLog {
		if strings.Contains(log, message) {
			return t
		}
	}

	assert.Fail(t.Testing, "No emulator log contain message "+message, t.Result.EmulatorLog)

	return t
}

func (t TransactionResult) AssertComputationLessThenOrEqual(computation int) TransactionResult {
	assert.LessOrEqual(t.Testing, t.Result.ComputationUsed, computation)
	return t
}

func (t TransactionResult) AssertComputationUsed(computation int) TransactionResult {
	assert.Equal(t.Testing, computation, t.Result.ComputationUsed)
	return t
}

func (t TransactionResult) GetIdFromEvent(eventName string, fieldName string) uint64 {

	for _, ev := range t.Events {
		if ev.Name == eventName {
			return ev.GetFieldAsUInt64(fieldName)
		}
	}
	panic("from event name of fieldname")

}
