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

	// TODO: test case for non-existent account
	// For some reason, when fetching an account that doesnt exist,
	// flowkit's Services.Accounts.Get() method throws an error & exits, rather than just returning an error.

	// t.Run("Should return an error if account doesn't exist", func(t *testing.T) {
	// 	g, _ := NewTestingEmulator().StartE()
	// 	account, _ := g.GetAccount("doesnotexist")

	// 	assert.Nil(t, account)
	// })
}
