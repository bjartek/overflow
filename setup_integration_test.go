package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupIntegration(t *testing.T) {
	o := Overflow()
	t.Run("Should create inmemory emulator client", func(t *testing.T) {
		assert.Equal(t, "emulator", o.Network.Name)
	})

	t.Run("should get account", func(t *testing.T) {
		account := o.Account("first")
		assert.Equal(t, "179b6b1cb6755e31", account.Address.String())
	})

	t.Run("should get address", func(t *testing.T) {
		account := o.Address("first")
		assert.Equal(t, "0x179b6b1cb6755e31", account)
	})

	t.Run("panic on wrong account name", func(t *testing.T) {
		assert.Panics(t, func() { o.Address("foobar") })
	})
}
