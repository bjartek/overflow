package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsInAccountCreation(t *testing.T) {

	t.Run("Should give error on wrong account name", func(t *testing.T) {
		_, err := NewTestingEmulator().Config("fixtures/invalid_account_in_deployment.json").StartE()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deployment contains nonexisting account emulator-firs")
	})
}

func TestGetAccount(t *testing.T) {

	t.Run("Should return the account", func(t *testing.T) {
		g, _ := NewTestingEmulator().StartE()
		account, err := g.GetAccount("account")

		assert.Nil(t, err)
		assert.Equal(t, "f8d6e0586b0a20c7", account.Address.String())
	})

	t.Run("Should return an error if account doesn't exist", func(t *testing.T) {
		g, _ := NewTestingEmulator().StartE()
		_, err := g.GetAccount("doesnotexist")
		assert.ErrorContains(t, err, "could not find account with name emulator-doesnotexist in the configuration")

	})

	t.Run("Should return an error if sa does not exist", func(t *testing.T) {
		_, err := NewTestingEmulator().SetServiceSuffix("dummy").StartE()

		assert.ErrorContains(t, err, "could not find account with name emulator-dummy in the configuration")

	})

}
