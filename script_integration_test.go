package overflow

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AwesomeStruct struct {
	First struct {
		Nested string `json:"nested"`
	} `json:"first"`
}

func TestScriptIntegrationNew(t *testing.T) {
	o, err := OverflowTesting()
	require.NoError(t, err)
	require.NotNil(t, o)

	mapScript := o.ScriptFileNameFN(`access(all) fun main() : {String: {String: String}} {
return { "first" : {  "nested" : "nestedvalue"}}

}`)

	t.Run("Get length from pointer", func(t *testing.T) {
		mapScript().AssertLengthWithPointer(t, "/first/nested", 11)
	})

	t.Run("Get length from pointer map", func(t *testing.T) {
		mapScript().AssertLengthWithPointer(t, "/first", 23)
	})

	t.Run("Get value from pointer", func(t *testing.T) {
		mapScript().AssertWithPointer(t, "/first/nested", "nestedvalue")
	})

	t.Run("Get error from pointer", func(t *testing.T) {
		mapScript().AssertWithPointerError(t, "/first/nested2", "Object has no key 'nested2'")
	})

	t.Run("Get value using want", func(t *testing.T) {
		mapScript().AssertWithPointerWant(t, "/first/nested", autogold.Want("nested", "nestedvalue"))
	})

	t.Run("Get value using want map", func(t *testing.T) {
		//Note that wants must have a unique name
		mapScript().AssertWithPointerWant(t, "/first", autogold.Want("nestedMap", map[string]interface{}{"nested": "nestedvalue"}))
	})

	t.Run("Marhal result using pointer", func(t *testing.T) {

		var result map[string]string
		//Note that wants must have a unique name
		err := mapScript().MarshalPointerAs("/first", &result)
		assert.NoError(t, err)

		assert.Equal(t, "nestedvalue", result["nested"])
	})

	t.Run("Marhal result", func(t *testing.T) {

		var result AwesomeStruct
		err := mapScript().MarshalAs(&result)
		assert.NoError(t, err)

		assert.Equal(t, "nestedvalue", result.First.Nested)
	})

	t.Run("Get assert with want", func(t *testing.T) {
		mapScript().AssertWant(t, autogold.Want("assertWant", map[string]interface{}{"first": map[string]interface{}{"nested": "nestedvalue"}}))
	})

	t.Run("Use relative import", func(t *testing.T) {
		res := o.Script(`
import Debug from "../contracts/Debug.cdc"

access(all) fun main() : AnyStruct {
return "foo"
}

`)
		require.NoError(t, res.Err)
		assert.Equal(t, "foo", res.Output)

	})

	t.Run("Use new import syntax", func(t *testing.T) {
		res := o.Script(`
import "Debug"

access(all) fun main() : AnyStruct {
return "foo"
}

`)
		require.NoError(t, res.Err)
		assert.Equal(t, "foo", res.Output)

	})
}
