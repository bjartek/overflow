package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAddress(t *testing.T) {

	o, err := OverflowTesting()
	assert.NoError(t, err)

	testCases := map[string]string{
		"first":     "0x01cf0e2f2f715450",
		"FlowToken": "0x0ae53cb6e3f42a79",
		"Debug":     "0xf8d6e0586b0a20c7",
	}

	for name, result := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.EqualValues(t, result, o.Address(name))
		})
	}

}
