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
	Testing *testing.T
}

func (f FlowTransactionBuilder) Test(t *testing.T) TransactionResult {
	locale, _ := time.LoadLocation("UTC")
	time.Local = locale
	events, err := f.RunE()
	var formattedEvents []*FormatedEvent
	for _, event := range events {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		formattedEvents = append(formattedEvents, ev)
	}
	return TransactionResult{
		Err:     err,
		Events:  formattedEvents,
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

	if !res {
		for _, ev := range t.Events {
			t.Testing.Log(ev.String())
		}
	}

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

	if !res {
		for _, ev := range t.Events {
			t.Testing.Log(ev.String())
		}
	}

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

	if !res {
		for _, ev := range t.Events {
			t.Testing.Log(ev.String())
		}
	}

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

	if !result {
		for _, ev := range events {
			t.Testing.Log(ev.String())
		}
	}

	return t
}
func (t TransactionResult) AssertEmitEvent(event ...*FormatedEvent) TransactionResult {
	res := false
	for _, ev := range event {
		if assert.Contains(t.Testing, t.Events, ev) {
			res = true
		}
	}

	if !res {
		for _, ev := range t.Events {
			t.Testing.Log(ev.String())
		}
	}

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
