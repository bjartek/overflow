package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifierIntegration(t *testing.T) {
	o, err := OverflowTesting()
	assert.NoError(t, err)

	result, err := o.QualifiedIdentifier("MetadataViews", "Display")
	assert.NoError(t, err)
	assert.Equal(t, "A.f8d6e0586b0a20c7.MetadataViews.Display", result)
}

func TestIdentifierTestnet(t *testing.T) {
	o := Overflow(WithNetwork("testnet"))
	assert.NoError(t, o.Error)

	result, err := o.QualifiedIdentifier("MetadataViews", "Display")
	assert.NoError(t, err)
	assert.Equal(t, "A.631e88ae7f1d7c20.MetadataViews.Display", result)
}
