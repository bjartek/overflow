package overflow

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

// TransactionResult
//
// The old result object from an transaction
//
// Deprecated: use the new Tx() method and OverflowResult
type TransactionResult struct {
	Err     error
	Events  []*FormatedEvent
	Result  *OverflowResult
	Testing *testing.T
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertFailure(msg string) TransactionResult {
	assert.Error(t.Testing, t.Err)
	if t.Err != nil {
		assert.Contains(t.Testing, t.Err.Error(), msg)
	}
	return t
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertSuccess() TransactionResult {
	t.Testing.Helper()

	if t.Err != nil {
		assert.Fail(t.Testing, fmt.Sprintf("Received unexpected error:\n%+v", t.Err), fmt.Sprintf("transactionName:%s", t.Result.Name))
	}
	return t
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEventCount(number int) TransactionResult {
	assert.Equal(t.Testing, len(t.Events), number)
	return t
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertNoEvents() TransactionResult {
	res := assert.Empty(t.Testing, t.Events)

	t.logFailure(res)
	return t
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) logFailure(res bool) {
	if !res {
		for _, ev := range t.Events {
			t.Testing.Log(litter.Sdump(ev))
		}
	}
}

// Deprecated: use the new Tx() method and Asserts on the result
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

// Deprecated: use the new Tx() method and Asserts on the result
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

// Deprecated: use the new Tx() method and Asserts on the result
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

// Deprecated: use the new Tx() method and Asserts on the result
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

// Deprecated: use the new Tx() method and Asserts on the result
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

// Deprecated: use the new Tx() method and Asserts on the result
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

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertEmulatorLog(message string) TransactionResult {

	for _, log := range t.Result.EmulatorLog {
		if strings.Contains(log, message) {
			return t
		}
	}

	assert.Fail(t.Testing, "No emulator log contain message "+message, t.Result.EmulatorLog)

	return t
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertComputationLessThenOrEqual(computation int) TransactionResult {
	assert.LessOrEqual(t.Testing, t.Result.ComputationUsed, computation)
	return t
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) AssertComputationUsed(computation int) TransactionResult {
	assert.Equal(t.Testing, computation, t.Result.ComputationUsed)
	return t
}

// Deprecated: use the new Tx() method and Asserts on the result
func (t TransactionResult) GetIdFromEvent(eventName string, fieldName string) uint64 {

	for _, ev := range t.Events {
		if ev.Name == eventName {
			return ev.GetFieldAsUInt64(fieldName)
		}
	}
	panic("from event name of fieldname")

}
