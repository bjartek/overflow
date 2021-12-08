package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsInAccountCreation(t *testing.T) {

	t.Run("Should give erro on wrong account name", func(t *testing.T) {
		_, err := NewTestingEmulator().Config("fixtures/invalid_account_in_deployment.json").StartE()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deployment contains nonexisting account emulator-firs")
	})

}
