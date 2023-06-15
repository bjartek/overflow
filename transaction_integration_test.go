package overflow

import (
	"testing"

	"github.com/onflow/flow-go/utils/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
 Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/

func TestTransactionIntegration(t *testing.T) {

	customResolver := func(input string) (string, error) {
		return "A.f8d6e0586b0a20c7.Debug.Foo", nil
	}
	o, err := OverflowTesting(WithCoverageReport())
	require.NoError(t, err)
	require.NotNil(t, o)
	o.Tx("mint_tokens", WithSignerServiceAccount(), WithArg("recipient", "first"), WithArg("amount", 1.0)).AssertSuccess(t)

	t.Run("fail on missing signer", func(t *testing.T) {
		o.Tx("create_nft_collection").AssertFailure(t, "💩 You need to set the proposer signer")
	})

	t.Run("fail on wrong transaction name", func(t *testing.T) {
		o.Tx("create_nft_collectio", WithSigner("first")).
			AssertFailure(t, "💩 Could not read interaction file from path=./transactions/create_nft_collectio.cdc")
	})

	t.Run("mint tokens with different base path", func(t *testing.T) {
		o.Tx("../tx/mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.1)).
			AssertSuccess(t)
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

		report := o.GetCoverageReport()
		assert.Equal(t, "17.6%", report.Summary().Coverage)
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

	/*
		* This is not a big deal but transaction submitted is not here anymore
			t.Run("emulator log", func(t *testing.T) {
				res := o.Tx(`
				import "Debug"
				transaction(message:String) {
				  prepare(acct: AuthAccount) {
					Debug.log(message) } }`,
					WithSigner("first"),
					WithArg("message", "foobar"),
				)

				res.
					AssertSuccess(t).
					AssertEmulatorLog(t, "Transaction submitted")
			})
	*/

	t.Run("Inline transaction with debug log", func(t *testing.T) {
		res := o.Tx(`
		import "Debug"
		transaction(message:String) {
		  prepare(acct: AuthAccount) {
			Debug.log(message) } }`,
			WithSigner("first"),
			WithArg("message", "foobar"),
		)

		res.
			AssertSuccess(t).
			AssertDebugLog(t, "foobar").
			AssertComputationUsed(t, 7).
			AssertComputationLessThenOrEqual(t, 40)
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

	t.Run("Send struct to transaction", func(t *testing.T) {

		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(foo: Debug.Foo) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArg("foo", Debug_Foo{Bar: "baz"}),
		).AssertSuccess(t)

	})

	t.Run("Send struct to transaction With Skip field", func(t *testing.T) {

		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(foo: Debug.Foo) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArg("foo", Debug_Foo_Skip{Bar: "baz", Skip: "skip"}),
		).AssertSuccess(t)

	})

	t.Run("Send list of struct to transaction custom qualifier", func(t *testing.T) {

		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(foo: [Debug.Foo]) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithStructArgsCustomQualifier("foo", customResolver, Foo{Bar: "baz"}, Foo{Bar: "baz2"}),
		).AssertSuccess(t)

	})

	t.Run("Send struct to transaction custom qualifier", func(t *testing.T) {

		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(foo: Debug.Foo) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithStructArgCustomResolver("foo", customResolver, Foo{Bar: "baz"}),
		).AssertSuccess(t)

	})

	t.Run("Send list of struct to transaction", func(t *testing.T) {

		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(foo: [Debug.Foo]) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArgs("foo", []Debug_Foo{{Bar: "baz"}, {Bar: "baz2"}}),
		).AssertSuccess(t)

	})

	t.Run("Send nestedstruct to transaction", func(t *testing.T) {

		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(foo: Debug.FooBar) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArg("foo", Debug_FooBar{Bar: "bar", Foo: Debug_Foo{Bar: "baz"}}),
		).AssertSuccess(t)

	})

	t.Run("Send nestedstruct with array to transaction", func(t *testing.T) {

		o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(foo: Debug.FooListBar) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArg("foo", Debug_FooListBar{Bar: "bar", Foo: []Debug_Foo2{{Bar: "0xf8d6e0586b0a20c7"}}}),
		).AssertSuccess(t)

	})

	t.Run("Send HttpFile to transaction", func(t *testing.T) {

		o.Tx(`
		import MetadataViews from "../contracts/MetadataViews.cdc"
		transaction(foo: AnyStruct{MetadataViews.File}) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArg("foo", MetadataViews_HTTPFile{Url: "foo"}),
		).AssertSuccess(t)

	})

	t.Run("Send IpfsFile to transaction", func(t *testing.T) {

		o.Tx(`
		import MetadataViews from "../contracts/MetadataViews.cdc"
		transaction(foo: AnyStruct{MetadataViews.File}) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArg("foo", MetadataViews_IPFSFile{Cid: "foo"}),
		).AssertSuccess(t)

	})

	t.Run("Send IpfsFile with path to transaction", func(t *testing.T) {

		path := "/Foo"
		o.Tx(`
		import MetadataViews from "../contracts/MetadataViews.cdc"
		transaction(foo: AnyStruct{MetadataViews.File}) {
		  prepare(acct: AuthAccount) {
		 } 
	 }`,
			WithSigner("first"),
			WithArg("foo", MetadataViews_IPFSFile{Cid: "foo", Path: &path}),
		).AssertSuccess(t)

	})

	t.Run("Send IpfsDisplay to transaction", func(t *testing.T) {

		o.Tx(`
				import MetadataViews from "../contracts/MetadataViews.cdc"
				transaction(display: MetadataViews.Display) {
				  prepare(acct: AuthAccount) {
				 }
			 }`,
			WithSigner("first"),
			WithArg("display", MetadataViews_Display_IPFS{Name: "foo", Description: "desc", Thumbnail: MetadataViews_IPFSFile{Cid: "foo"}}),
		).AssertSuccess(t)

	})

	t.Run("Send HttpDisplay to transaction", func(t *testing.T) {

		o.Tx(`
			import MetadataViews from "../contracts/MetadataViews.cdc"
			transaction(display: MetadataViews.Display) {
			  prepare(acct: AuthAccount) {
			 }
		 }`,
			WithSigner("first"),
			WithArg("display", MetadataViews_Display_Http{Name: "foo", Description: "desc", Thumbnail: MetadataViews_HTTPFile{Url: "foo"}}),
		).AssertSuccess(t)

	})

	t.Run("Send Trait to transaction", func(t *testing.T) {

		o.Tx(`
			import MetadataViews from "../contracts/MetadataViews.cdc"
			transaction(trait: MetadataViews.Trait) {
			  prepare(acct: AuthAccount) {
			 }
		 }`,
			WithSigner("first"),
			WithArg("trait", MetadataViews_Trait{Name: "foo", Value: "bar"}),
		).AssertSuccess(t)

	})

	bytes, err := o.GetCoverageReport().MarshalJSON()
	require.NoError(t, err)
	err = io.WriteFile("coverage-report.json", bytes)
	require.NoError(t, err)

}

func TestTransactionEventFiltering(t *testing.T) {

	filter := OverflowEventFilter{
		"Log": []string{"msg"},
	}

	filterLocal := OverflowEventFilter{
		"LogNum": []string{"id"},
	}

	o, err := OverflowTesting(WithGlobalEventFilter(filter))
	require.NotNil(t, o)
	require.NoError(t, err)
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

func TestFillUpSpace(t *testing.T) {
	o, err := OverflowTesting(WithFlowForNewUsers(0.001))
	assert.NoError(t, err)

	result := o.GetFreeCapacity("first")
	assert.Equal(t, 199205, result)
	o.FillUpStorage("first")
	assert.NoError(t, o.Error)

	result2 := o.GetFreeCapacity("first")
	assert.LessOrEqual(t, result2, 42000)

}
