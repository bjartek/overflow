package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorsInAccountCreation(t *testing.T) {

	t.Run("Should deploy contracts to multiple accounts", func(t *testing.T) {
		_, err := OverflowTesting(WithFlowConfig("testdata/flow-with-multiple-deployments.json"), WithLogFull(), WithFlowForNewUsers(100.0))
		assert.NoError(t, err)
	})

	t.Run("Should give error on wrong contract name", func(t *testing.T) {
		assert.Panics(t, func() {
			NewTestingEmulator().Config("testdata/non_existing_contract.json").Start()
		})
	})

	t.Run("Should give error on invalid env var in flow.json", func(t *testing.T) {
		_, err := NewTestingEmulator().Config("testdata/invalid_env_flow.json").StartE()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid private key for account: emulator-5")
	})

	t.Run("Should give error on wrong account name", func(t *testing.T) {
		_, err := NewTestingEmulator().Config("testdata/invalid_account_in_deployment.json").StartE()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deployment contains nonexisting account emulator-firs")
	})

}

func TestGetAccount(t *testing.T) {

	t.Run("Should return the account", func(t *testing.T) {
		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		account, err := g.GetAccount("account")

		assert.Nil(t, err)
		assert.Equal(t, "f8d6e0586b0a20c7", account.Address.String())
	})

	t.Run("Should return an error if account doesn't exist", func(t *testing.T) {
		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		_, err = g.GetAccount("doesnotexist")
		assert.ErrorContains(t, err, "could not find account with name emulator-doesnotexist in the configuration")

	})

	t.Run("Should return an error if sa does not exist", func(t *testing.T) {
		_, err := NewTestingEmulator().SetServiceSuffix("dummy").StartE()

		assert.ErrorContains(t, err, "could not find account with name emulator-dummy in the configuration")

	})

}
