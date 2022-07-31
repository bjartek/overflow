package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupIntegration(t *testing.T) {

	o := Overflow()
	t.Run("Should create inmemory emulator client", func(t *testing.T) {
		assert.Equal(t, "emulator", o.Network)
	})

	t.Run("should get account", func(t *testing.T) {
		account := o.Account("first")
		assert.Equal(t, "01cf0e2f2f715450", account.Address().String())
	})

	t.Run("should get address", func(t *testing.T) {
		account := o.Address("first")
		assert.Equal(t, "0x01cf0e2f2f715450", account)
	})

	t.Run("panic on wrong account name", func(t *testing.T) {
		assert.Panics(t, func() { o.Address("foobar") })
	})

}
