package overflow

/*
func TestPanicIfStopOnFailure(t *testing.T) {
	o, err := OverflowTesting(WithPanicOnError())
	assert.NoError(t, err)

	t.Run("transaction", func(t *testing.T) {
		assert.PanicsWithError(t, "💩 You need to set the proposer signer", func() {
			o.Tx("create_nft_collection")
		})
	})

	t.Run("script", func(t *testing.T) {

		assert.PanicsWithError(t, "💩 Could not read interaction file from path=./scripts/asdf.cdc", func() {
			o.Script("asdf")
		})
	})
}
*/
