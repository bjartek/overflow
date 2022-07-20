package overflow

import (
	"reflect"
	"strings"

	"github.com/onflow/flow-go-sdk"
)

// Event parsing
//
// a type alias to an OverflowEventFilter to filter out all events with a given suffix and the fields with given suffixes
type OverflowEventFilter map[string][]string

// a type holding all events that are emitted from a Transaction
type OverflowEvents map[string][]OverflowEvent

// a type representing the terse output of an raw Flow Event
type OverflowEvent map[string]interface{}

// Check if an event exist in the other events
func (o OverflowEvent) ExistIn(events []OverflowEvent) bool {
	for _, ev := range events {
		result := reflect.DeepEqual(o, ev)
		if result {
			return true
		}
	}
	return false
}

//Parse raw flow events into a list of events and a fee event
func ParseEvents(events []flow.Event) (OverflowEvents, OverflowEvent) {
	overflowEvents := OverflowEvents{}
	fee := OverflowEvent{}
	for _, event := range events {

		var fieldNames []string

		for _, eventTypeFields := range event.Value.EventType.Fields {
			fieldNames = append(fieldNames, eventTypeFields.Identifier)
		}

		finalFields := map[string]interface{}{}

		for id, field := range event.Value.Fields {
			name := fieldNames[id]
			value := CadenceValueToInterface(field)
			if value != nil {
				finalFields[name] = value
			}
		}

		events, ok := overflowEvents[event.Type]
		if !ok {
			events = []OverflowEvent{}
		}
		events = append(events, finalFields)
		overflowEvents[event.Type] = events

		if strings.HasSuffix(event.Type, "FlowFees.FeesDeducted") {
			fee = finalFields
		}

	}
	return overflowEvents, fee
}

// Filter out temp withdraw deposit events
func (overflowEvents OverflowEvents) FilterTempWithdrawDeposit() OverflowEvents {
	filteredEvents := overflowEvents
	for name, events := range overflowEvents {
		if strings.HasSuffix(name, "TokensWithdrawn") {

			withDrawnEvents := []OverflowEvent{}
			for _, value := range events {
				if value["from"] != nil {
					withDrawnEvents = append(withDrawnEvents, value)
				}
			}
			if len(withDrawnEvents) != 0 {
				filteredEvents[name] = withDrawnEvents
			} else {
				delete(filteredEvents, name)
			}
		}

		if strings.HasSuffix(name, "TokensDeposited") {
			despoitEvents := []OverflowEvent{}
			for _, value := range events {
				if value["to"] != nil {
					despoitEvents = append(despoitEvents, value)
				}
			}
			if len(despoitEvents) != 0 {
				filteredEvents[name] = despoitEvents
			} else {
				delete(filteredEvents, name)
			}
		}
	}
	return filteredEvents
}

// Filtter out fee events
func (overflowEvents OverflowEvents) FilterFees(fee float64) OverflowEvents {

	filteredEvents := overflowEvents
	for name, events := range overflowEvents {
		if strings.HasSuffix(name, "FlowFees.FeesDeducted") {
			delete(filteredEvents, name)
		}

		if strings.HasSuffix(name, "FlowToken.TokensWithdrawn") {

			withDrawnEvents := []OverflowEvent{}
			for _, value := range events {
				if value["amount"].(float64) != fee {
					withDrawnEvents = append(withDrawnEvents, value)
				}
			}
			if len(withDrawnEvents) != 0 {
				filteredEvents[name] = withDrawnEvents
			} else {
				delete(filteredEvents, name)
			}
		}

		if strings.HasSuffix(name, "FlowToken.TokensDeposited") {
			despoitEvents := []OverflowEvent{}
			for _, value := range events {
				if value["amount"].(float64) != fee {
					despoitEvents = append(despoitEvents, value)
				}
			}
			if len(despoitEvents) != 0 {
				filteredEvents[name] = despoitEvents
			} else {
				delete(filteredEvents, name)
			}

		}
	}
	return filteredEvents
}

// Filter out events given the sent in filter
func (overflowEvents OverflowEvents) FilterEvents(ignoreFields OverflowEventFilter) OverflowEvents {
	filteredEvents := OverflowEvents{}
	for name, events := range overflowEvents {

		//find if we should ignore fields
		ignoreFieldNames := []string{}
		for ignoreEvent, fields := range ignoreFields {
			if strings.HasSuffix(name, ignoreEvent) {
				ignoreFieldNames = fields
			}
		}

		eventList := []OverflowEvent{}
		for _, ev := range events {
			event := OverflowEvent{}
			for key, value := range ev {
				valid := true
				for _, ig := range ignoreFieldNames {
					if strings.HasSuffix(key, ig) {
						valid = false
					}
				}
				if valid {
					event[key] = value
				}
			}
			if len(event) != 0 {
				eventList = append(eventList, event)
			}
		}
		if len(eventList) != 0 {
			filteredEvents[name] = eventList
		}
	}
	return filteredEvents
}
