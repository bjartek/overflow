package v3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	o, _ := OverflowTesting()

	t.Run("Run simple transacon", func(t *testing.T) {
		res := o.Tx("arguments", Arg("test", "foo"), SignProposeAndPayAsServiceAccount())
		assert.NoError(t, res.Err)
	})

	t.Run("compose a function", func(t *testing.T) {
		serviceAccountTx := o.TxFN(SignProposeAndPayAsServiceAccount())
		res := serviceAccountTx("arguments", Arg("test", "foo"))
		assert.NoError(t, res.Err)
	})

	t.Run("create function with name", func(t *testing.T) {
		argumentTx := o.TxFileNameFN("arguments", SignProposeAndPayAsServiceAccount())
		res := argumentTx(Arg("test", "foo"))
		assert.NoError(t, res.Err)
	})

}
