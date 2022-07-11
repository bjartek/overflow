package overflow

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/assert"
)

func TestFilterOverflowEvents(t *testing.T) {

	events := OverflowEvents{
		"A.123.Test.Deposit": []OverflowEvent{{
			"id":     1,
			"string": "string",
		}},
	}

	t.Run("Filter out all events should yield empty", func(t *testing.T) {

		filter := OverflowEventFilter{
			"Deposit": []string{"id", "string"},
		}
		filtered := events.FilterEvents(filter)
		assert.Empty(t, filtered)

	})
	t.Run("Filter out single field", func(t *testing.T) {

		filter := OverflowEventFilter{
			"Deposit": []string{"id"},
		}
		filtered := events.FilterEvents(filter)
		want := autogold.Want("string", OverflowEvents{"A.123.Test.Deposit": []OverflowEvent{
			OverflowEvent{"string": "string"},
		}})
		want.Equal(t, filtered)
	})
	t.Run("Filter fees", func(t *testing.T) {

		eventsWithFees := OverflowEvents{
			"A.f919ee77447b7497.FlowFees.FeesDeducted": []OverflowEvent{{
				"amount":          0.00000918,
				"inclusionEffort": 1.00000000,
				"executionEffort": 0.00000164,
			}},
			"A.1654653399040a61.FlowToken.TokensWithdrawn": []OverflowEvent{{
				"amount": 0.00000918,
				"from":   "0x55ad22f01ef568a1",
			}},
			"A.1654653399040a61.FlowToken.TokensDeposited": []OverflowEvent{{
				"amount": 0.00000918,
				"to":     "0xf919ee77447b7497",
			}, {
				"amount": 1.00000000,
				"to":     "0xf919ee77447b7497",
			}},
		}
		filtered := eventsWithFees.FilterFees(0.00000918)
		want := autogold.Want("fees filtered", OverflowEvents{"A.1654653399040a61.FlowToken.TokensDeposited": []OverflowEvent{
			OverflowEvent{
				"amount": 1,
				"to":     "0xf919ee77447b7497",
			},
		}})
		want.Equal(t, filtered)
	})

	t.Run("Filter empty deposit withdraw", func(t *testing.T) {

		eventsWithFees := OverflowEvents{
			"A.1654653399040a61.FlowToken.TokensWithdrawn": []OverflowEvent{{
				"amount": 0.00000918,
				"from":   nil,
			}},
			"A.1654653399040a61.FlowToken.TokensDeposited": []OverflowEvent{{
				"amount": 0.00000918,
				"to":     nil,
			}, {
				"amount": 1.00000000,
				"to":     "0xf919ee77447b7497",
			}},
		}
		filtered := eventsWithFees.FilterTempWithdrawDeposit()
		want := autogold.Want("fees empty deposit withdraw", OverflowEvents{"A.1654653399040a61.FlowToken.TokensDeposited": []OverflowEvent{
			OverflowEvent{
				"amount": 1,
				"to":     "0xf919ee77447b7497",
			},
		}})
		want.Equal(t, filtered)
	})

}
