package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
 Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/

func TestTransactionIntegration(t *testing.T) {
	o, err := OverflowTesting(WithLogFull())
	o.Tx("mint_tokens", WithSignerServiceAccount(), WithArg("recipient", "first"), WithArg("amount", 1.0)).AssertSuccess(t)

	assert.NoError(t, err)
	t.Parallel()

	t.Run("fail on missing signer", func(t *testing.T) {
		o.Tx("create_nft_collection").AssertFailure(t, "💩 You need to set the proposer signer")
	})

	t.Run("fail on wrong transaction name", func(t *testing.T) {
		o.Tx("create_nft_collectio", WithSigner("first")).
			AssertFailure(t, "💩 Could not read interaction file from path=./transactions/create_nft_collectio.cdc")
	})

	t.Run("Create NFT collection with different base path", func(t *testing.T) {
		o.Tx("../tx/create_nft_collection",
			WithSigner("first")).
			AssertSuccess(t).
			AssertNoEvents(t)
	})

	t.Run("Mint tokens assert events", func(t *testing.T) {
		result := o.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.1)).
			AssertSuccess(t).
			AssertEventCount(t, 3).
			AssertEmitEventName(t, "FlowToken.TokensDeposited").
			AssertEvent(t, "FlowToken.TokensDeposited", map[string]interface{}{
				"amount": 100.1,
			},
			)
		assert.Equal(t, 1, len(result.GetEventsWithName("TokensDeposited")))

	})

	t.Run("Assert get id", func(t *testing.T) {
		result := o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`,
			WithSigner("first"),
			WithArg("id", 1),
		).AssertSuccess(t)

		res, err := result.GetIdFromEvent("LogNum", "id")
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), res)
		assert.Equal(t, []uint64{1}, result.GetIdsFromEvent("LogNum", "id"))

	})

	t.Run("Inline transaction with debug log", func(t *testing.T) {
		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(message:String) {
		  prepare(acct: AuthAccount) {
			Debug.log(message) } }`,
			WithSigner("first"),
			WithArg("message", "foobar"),
		).
			AssertDebugLog(t, "foobar").
			AssertComputationUsed(t, 7).
			AssertComputationLessThenOrEqual(t, 40).
			AssertEmulatorLog(t, "Transaction submitted")
	})

	t.Run("Mint tokens and marshal event", func(t *testing.T) {
		result := o.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.1)).
			AssertSuccess(t)

		var events []interface{}
		err := result.MarshalEventsWithName("TokensDeposited", &events)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))

		var singleEvent interface{}
		err = result.GetEventsWithName("TokensDeposited")[0].MarshalAs(&singleEvent)
		assert.NoError(t, err)
		assert.NotNil(t, singleEvent)
	})

}
func TestTransactionEventFiltering(t *testing.T) {

	filter := OverflowEventFilter{
		"Log": []string{"msg"},
	}

	filterLocal := OverflowEventFilter{
		"LogNum": []string{"id"},
	}

	o, err := OverflowTesting(WithGlobalEventFilter(filter))
	assert.NoError(t, err)
	o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(message:String) {
		  prepare(acct: AuthAccount) {
			Debug.log(message) 
			Debug.id(1)
		} }`,
		WithEventsFilter(filterLocal),
		WithSigner("first"),
		WithArg("message", "foobar"),
	).AssertSuccess(t).AssertEventCount(t, 0)
}
