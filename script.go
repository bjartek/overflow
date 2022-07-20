package overflow

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonpointer"
)

// Scripts
//
// A read only interaction against the flow blockcahin

// a type used for composing scripts
type ScriptFunction func(filename string, opts ...InteractionOption) *OverflowScriptResult

// a type used for composing scripts
type ScriptOptsFunction func(opts ...InteractionOption) *OverflowScriptResult

// compose interactionOptions into a new Script function
func (o *OverflowState) ScriptFN(outerOpts ...InteractionOption) ScriptFunction {

	return func(filename string, opts ...InteractionOption) *OverflowScriptResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Script(filename, outerOpts...)
	}
}

// compose fileName and interactionOptions into a new Script function
func (o *OverflowState) ScriptFileNameFN(filename string, outerOpts ...InteractionOption) ScriptOptsFunction {

	return func(opts ...InteractionOption) *OverflowScriptResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Script(filename, outerOpts...)
	}
}

// run a script with the given code/filanem an options
func (o *OverflowState) Script(filename string, opts ...InteractionOption) *OverflowScriptResult {
	interaction := o.BuildInteraction(filename, "script", opts...)

	result := interaction.runScript()

	if o.PrintOptions != nil {
		result.Print()
	}
	if o.StopOnError && result.Err != nil {
		panic(result.Err)
	}
	return result

}

func (fbi *FlowInteractionBuilder) runScript() *OverflowScriptResult {

	o := fbi.Overflow
	osc := &OverflowScriptResult{Input: fbi}
	if fbi.Error != nil {
		osc.Err = fbi.Error
		return osc
	}

	filePath := fmt.Sprintf("%s/%s.cdc", fbi.BasePath, fbi.FileName)

	o.EmulatorLog.Reset()
	o.Log.Reset()

	result, err := o.Services.Scripts.Execute(
		fbi.TransactionCode,
		fbi.Arguments,
		filePath,
		o.Network)

	osc.Result = result
	osc.Output = CadenceValueToInterface(result)
	if err != nil {
		osc.Err = errors.Wrapf(err, "scriptFileName:%s", fbi.FileName)
	}

	var logMessage []LogrusMessage
	dec := json.NewDecoder(o.Log)
	for {
		var doc LogrusMessage

		err := dec.Decode(&doc)
		if err == io.EOF {
			// all done
			break
		}
		if err != nil {
			osc.Err = err
		}

		logMessage = append(logMessage, doc)
	}

	o.EmulatorLog.Reset()
	o.Log.Reset()

	osc.Log = logMessage

	return osc
}

// result after running a script
type OverflowScriptResult struct {
	Err    error
	Result cadence.Value
	Input  *FlowInteractionBuilder
	Log    []LogrusMessage
	Output interface{}
}

// get the script as json
func (osr *OverflowScriptResult) GetAsJson() (string, error) {
	if osr.Err != nil {
		return "", errors.Wrapf(osr.Err, "script: %s", osr.Input.FileName)
	}
	j, err := json.MarshalIndent(osr.Output, "", "    ")

	if err != nil {
		return "", errors.Wrapf(err, "script: %s", osr.Input.FileName)
	}

	return string(j), nil
}

// get the script as interface{}
func (osr *OverflowScriptResult) GetAsInterface() (interface{}, error) {
	if osr.Err != nil {
		return nil, errors.Wrapf(osr.Err, "script: %s", osr.Input.FileName)
	}
	return osr.Output, nil
}

// Assert that a jsonPointer into the result is an error
func (osr *OverflowScriptResult) AssertWithPointerError(t *testing.T, pointer string, message string) *OverflowScriptResult {
	_, err := osr.GetWithPointer(pointer)
	assert.Error(t, err)
	assert.ErrorContains(t, err, message, "output", litter.Sdump(osr.Output))

	return osr
}

// Assert that a jsonPointer into the result is equal to the given value
func (osr *OverflowScriptResult) AssertWithPointer(t *testing.T, pointer string, value interface{}) *OverflowScriptResult {
	result, err := osr.GetWithPointer(pointer)
	assert.NoError(t, err)

	assert.Equal(t, value, result, "output", litter.Sdump(osr.Output))

	return osr
}

//Assert that a jsonPointer into the result is equal to the given autogold Want
func (osr *OverflowScriptResult) AssertWithPointerWant(t *testing.T, pointer string, want autogold.Value) *OverflowScriptResult {
	result, err := osr.GetWithPointer(pointer)
	assert.NoError(t, err)

	switch result.(type) {
	case []interface{}:
	case map[interface{}]interface{}:
		want.Equal(t, litter.Sdump(result))
	default:
		want.Equal(t, result)
	}
	return osr
}

// Assert that the length of a jsonPointer is equal to length
func (osr *OverflowScriptResult) AssertLengthWithPointer(t *testing.T, pointer string, length int) *OverflowScriptResult {
	result, err := osr.GetWithPointer(pointer)
	assert.NoError(t, err)
	switch res := result.(type) {
	case []interface{}:
	case map[interface{}]interface{}:
		assert.Equal(t, length, len(res), litter.Sdump(osr.Output))
	default:
		assert.Equal(t, length, len(fmt.Sprintf("%v", res)), litter.Sdump(osr.Output))
	}
	return osr
}

// Marshal the script output as the given sent in type
func (osr *OverflowScriptResult) MarshalAs(marshalTo interface{}) error {
	bytes, err := json.Marshal(osr.Output)
	if err != nil {
		return err
	}

	json.Unmarshal(bytes, marshalTo)
	return nil
}

// Marshal the given jsonPointer as the given type
func (osr *OverflowScriptResult) MarshalPointerAs(pointer string, marshalTo interface{}) error {
	ptr, err := gojsonpointer.NewJsonPointer(pointer)
	if err != nil {
		return err
	}
	result, _, err := ptr.Get(osr.Output)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	json.Unmarshal(bytes, marshalTo)
	return nil
}

// get the given jsonPointer as interface{}
func (osr *OverflowScriptResult) GetWithPointer(pointer string) (interface{}, error) {

	ptr, err := gojsonpointer.NewJsonPointer(pointer)
	if err != nil {
		return nil, err
	}
	result, _, err := ptr.Get(osr.Output)
	return result, err
}

// Assert that the result is equal to the given autogold.Want
func (osr *OverflowScriptResult) AssertWant(t *testing.T, want autogold.Value) *OverflowScriptResult {
	assert.NoError(t, osr.Err)

	switch osr.Output.(type) {
	case []interface{}:
	case map[interface{}]interface{}:
		want.Equal(t, litter.Sdump(osr.Output))
	default:
		want.Equal(t, osr.Output)
	}
	return osr
}

// Print the result
func (osr *OverflowScriptResult) Print() {
	json, err := osr.GetAsJson()
	if err != nil {
		color.Red(err.Error())
		return
	}
	fmt.Printf("%v Script %s run from result: %v\n", emoji.Star, osr.Input.FileName, json)
}
