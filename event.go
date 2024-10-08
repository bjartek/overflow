package overflow

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/bjartek/underflow"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/sanity-io/litter"
	"golang.org/x/exp/slices"
)

// Event parsing
//
// a type alias to an OverflowEventFilter to filter out all events with a given suffix and the fields with given suffixes
type OverflowEventFilter map[string][]string

type OverflowEventList []OverflowEvent

// a type holding all events that are emitted from a Transaction
type OverflowEvents map[string]OverflowEventList

func (me OverflowEvents) GetStakeholders(stakeholders map[string][]string) map[string][]string {
	for _, events := range me {
		for _, event := range events {
			eventStakeholders := event.GetStakeholders()
			for stakeholder, roles := range eventStakeholders {

				allRoles, ok := stakeholders[stakeholder]
				if !ok {
					allRoles = []string{}
				}
				allRoles = append(allRoles, roles...)
				stakeholders[stakeholder] = allRoles
			}
		}
	}

	return stakeholders
}

type OverflowEvent struct {
	Fields        map[string]interface{} `json:"fields"`
	Addresses     map[string][]string    `json:"addresses"`
	Id            string                 `json:"id"`
	TransactionId string                 `json:"transactionID"`
	Name          string                 `json:"name"`
	RawEvent      cadence.Event          `json:"rawEvent"`
	EventIndex    uint32                 `json:"eventIndex"`
}

type FeeBalance struct {
	PayerBalance    float64 `json:"payerBalance"`
	TotalFeeBalance float64 `json:"totalFeeBalance"`
}

// Check if an event exist in the other events
func (o OverflowEvent) ExistIn(events []OverflowEvent) bool {
	for _, ev := range events {
		if litter.Sdump(o.Fields) == litter.Sdump(ev.Fields) {
			return true
		}
	}
	return false
}

// list of address to a list of roles for that address
func (me OverflowEvent) GetStakeholders() map[string][]string {
	stakeholder := map[string][]string{}
	for name, value := range me.Addresses {
		for _, address := range value {

			existing, ok := stakeholder[address]
			if !ok {
				existing = []string{}
			}
			eventName := me.Name
			if strings.Contains(eventName, "FungibleToken.Withdrawn") ||
				strings.Contains(eventName, "FungibleToken.Deposited") ||
				strings.Contains(eventName, "NonFungibleToken.Deposited") ||
				strings.Contains(eventName, "NonFungibleToken.Withdrawn") {
				vaultType, _ := me.Fields["type"].(string)
				existing = append(existing, fmt.Sprintf("%s/%s", vaultType, name))
			} else {
				existing = append(existing, fmt.Sprintf("%s/%s", me.Name, name))
			}
			stakeholder[address] = existing
		}
	}
	return stakeholder
}

func (e OverflowEventList) MarshalAs(marshalTo interface{}) error {
	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, marshalTo)
	if err != nil {
		return err
	}
	return nil
}

func (e OverflowEvent) MarshalAs(marshalTo interface{}) error {
	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, marshalTo)
	if err != nil {
		return err
	}
	return nil
}

func (o *OverflowState) ParseEvents(events []flow.Event) (OverflowEvents, OverflowEvent) {
	return o.ParseEventsWithIdPrefix(events, "")
}

func (o *OverflowState) ParseEventsWithIdPrefix(events []flow.Event, idPrefix string) (OverflowEvents, OverflowEvent) {
	overflowEvents := OverflowEvents{}
	fee := OverflowEvent{}
	for i, event := range events {
		finalFields := map[string]interface{}{}
		addresses := map[string][]string{}

		for name, field := range cadence.FieldsMappedByName(event.Value) {

			adr := underflow.ExtractAddresses(field)
			if len(adr) > 0 {
				addresses[name] = adr
			}
			value := underflow.CadenceValueToInterfaceWithOption(field, o.UnderflowOptions)
			if value != nil {
				finalFields[name] = value
			}
		}

		events, ok := overflowEvents[event.Type]
		if !ok {
			events = []OverflowEvent{}
		}
		events = append(events, OverflowEvent{
			Id:            fmt.Sprintf("%s%s-%d", idPrefix, event.TransactionID.Hex(), i),
			Fields:        finalFields,
			Name:          event.Type,
			TransactionId: event.TransactionID.String(),
			EventIndex:    uint32(event.EventIndex),
			Addresses:     addresses,
			RawEvent:      event.Value,
		})
		overflowEvents[event.Type] = events
		if strings.HasSuffix(event.Type, "FlowFees.FeesDeducted") {
			fee = OverflowEvent{
				Fields:        finalFields,
				Name:          event.Type,
				TransactionId: event.TransactionID.String(),
			}
		}

	}
	return overflowEvents, fee
}

func (overflowEvents OverflowEvents) FilterTempWithdrawDeposit() OverflowEvents {
	filteredEvents := overflowEvents
	for name, events := range overflowEvents {

		if strings.HasSuffix(name, "TokensWithdrawn") {

			withDrawnEvents := []OverflowEvent{}
			for _, value := range events {
				if value.Fields["from"] != nil {
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
			depositEvents := []OverflowEvent{}
			for _, value := range events {
				if value.Fields["to"] != nil {
					depositEvents = append(depositEvents, value)
				}
			}
			if len(depositEvents) != 0 {
				filteredEvents[name] = depositEvents
			} else {
				delete(filteredEvents, name)
			}
		}

		if strings.HasSuffix(name, ".FungibleToken.Withdrawn") {
			withDrawnEvents := []OverflowEvent{}
			for _, value := range events {
				if value.Fields["from"] != nil {
					withDrawnEvents = append(withDrawnEvents, value)
				}
			}
			if len(withDrawnEvents) != 0 {
				filteredEvents[name] = withDrawnEvents
			} else {
				delete(filteredEvents, name)
			}
		}

		if strings.HasSuffix(name, ".FungibleToken.Deposited") {
			depositEvents := []OverflowEvent{}
			for _, value := range events {
				if value.Fields["to"] != nil {
					depositEvents = append(depositEvents, value)
				}
			}
			if len(depositEvents) != 0 {
				filteredEvents[name] = depositEvents
			} else {
				delete(filteredEvents, name)
			}

		}
	}
	return filteredEvents
}

var feeReceipients = []string{"0xf919ee77447b7497", "0x912d5440f7e3769e", "0xe5a8b7f23e8b548f", "0xab086ce9cc29fc80"}

// Filtter out fee events
func (overflowEvents OverflowEvents) FilterFees(fee float64, payer string) (OverflowEvents, FeeBalance) {
	filteredEvents := overflowEvents

	fees := FeeBalance{}
	for name, events := range overflowEvents {
		if strings.HasSuffix(name, "FlowFees.FeesDeducted") {
			delete(filteredEvents, name)
		}

		if strings.HasSuffix(name, ".FungibleToken.Withdrawn") {
			withDrawnEvents := []OverflowEvent{}
			for _, value := range events {
				ftType, _ := value.Fields["type"].(string)
				if !strings.HasSuffix(ftType, "FlowToken.Vault") {
					withDrawnEvents = append(withDrawnEvents, value)
					continue
				}
				amount, _ := value.Fields["amount"].(float64)
				from, ok := value.Fields["from"].(string)

				if ok && amount == fee && from == payer {
					balance, _ := value.Fields["balanceAfter"].(float64)
					fees.PayerBalance = balance
					continue
				}

				withDrawnEvents = append(withDrawnEvents, value)
			}
			if len(withDrawnEvents) != 0 {
				filteredEvents[name] = withDrawnEvents
			} else {
				delete(filteredEvents, name)
			}
		}

		if strings.HasSuffix(name, ".FungibleToken.Deposited") {
			depositEvents := []OverflowEvent{}
			for _, value := range events {
				ftType, _ := value.Fields["type"].(string)
				if !strings.HasSuffix(ftType, "FlowToken.Vault") {
					depositEvents = append(depositEvents, value)
					continue
				}
				amount, _ := value.Fields["amount"].(float64)
				to, ok := value.Fields["to"].(string)

				if ok && amount == fee && slices.Contains(feeReceipients, to) {
					balance, _ := value.Fields["balanceAfter"].(float64)
					fees.TotalFeeBalance = balance
					continue
				}
				depositEvents = append(depositEvents, value)
			}
			if len(depositEvents) != 0 {
				filteredEvents[name] = depositEvents
			} else {
				delete(filteredEvents, name)
			}
		}

		if strings.HasSuffix(name, "FlowToken.TokensWithdrawn") {

			withDrawnEvents := []OverflowEvent{}
			for _, value := range events {

				amount, _ := value.Fields["amount"].(float64)
				from, ok := value.Fields["from"].(string)

				if ok && amount == fee && from == payer {
					continue
				}

				withDrawnEvents = append(withDrawnEvents, value)
			}
			if len(withDrawnEvents) != 0 {
				filteredEvents[name] = withDrawnEvents
			} else {
				delete(filteredEvents, name)
			}
		}

		if strings.HasSuffix(name, "FlowToken.TokensDeposited") {
			depositEvents := []OverflowEvent{}
			for _, value := range events {

				amount, _ := value.Fields["amount"].(float64)
				to, ok := value.Fields["to"].(string)

				if ok && amount == fee && slices.Contains(feeReceipients, to) {
					continue
				}
				depositEvents = append(depositEvents, value)
			}
			if len(depositEvents) != 0 {
				filteredEvents[name] = depositEvents
			} else {
				delete(filteredEvents, name)
			}

		}
	}
	return filteredEvents, fees
}

func printOrLog(t *testing.T, s string) {
	if t == nil {
		fmt.Println(s)
	} else {
		t.Log(s)
		t.Helper()
	}
}

func (overflowEvents OverflowEvents) Print(t *testing.T) {
	if t != nil {
		t.Helper()
	}

	events := []OverflowEvent{}
	printOrLog(t, "=== Events ===")
	for _, eventList := range overflowEvents {
		for _, event := range eventList {
			events = append(events, event)
		}
	}

	slices.SortStableFunc(events, func(a OverflowEvent, b OverflowEvent) int {
		return int(a.EventIndex) - int(b.EventIndex)
	})

	for _, event := range events {
		printOrLog(t, event.Name)
		length := 0
		for key := range event.Fields {
			keyLength := len(key)
			if keyLength > length {
				length = keyLength
			}
		}

		format := fmt.Sprintf("%%%ds -> %%v", length+2)
		for key, value := range event.Fields {
			printOrLog(t, fmt.Sprintf(format, key, value))
		}
	}
}

// Filter out events given the sent in filter
func (overflowEvents OverflowEvents) FilterEvents(ignoreFields OverflowEventFilter) OverflowEvents {
	filteredEvents := OverflowEvents{}
	for name, events := range overflowEvents {

		// find if we should ignore fields
		ignoreFieldNames := []string{}
		for ignoreEvent, fields := range ignoreFields {
			if strings.HasSuffix(name, ignoreEvent) {
				ignoreFieldNames = fields
			}
		}

		eventList := []OverflowEvent{}
		for _, ev := range events {
			event := OverflowEvent{Fields: map[string]interface{}{}}
			for key, value := range ev.Fields {
				valid := true
				for _, ig := range ignoreFieldNames {
					if strings.HasSuffix(key, ig) {
						valid = false
					}
				}
				if valid {
					event.Fields[key] = value
				}
			}
			if len(event.Fields) != 0 {
				eventList = append(eventList, event)
			}
		}
		if len(eventList) != 0 {
			filteredEvents[name] = eventList
		}
	}
	return filteredEvents
}
