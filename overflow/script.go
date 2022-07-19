package overflow

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/enescakir/emoji"
	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonpointer"
)

//Composition functions for Scripts
type ScriptFunction func(filename string, opts ...InteractionOption) *OverflowScriptResult
type ScriptOptsFunction func(opts ...InteractionOption) *OverflowScriptResult

func (o *OverflowState) ScriptFN(outerOpts ...InteractionOption) ScriptFunction {

	return func(filename string, opts ...InteractionOption) *OverflowScriptResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Script(filename, outerOpts...)
	}
}

func (o *OverflowState) ScriptFileNameFN(filename string, outerOpts ...InteractionOption) ScriptOptsFunction {

	return func(opts ...InteractionOption) *OverflowScriptResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Script(filename, outerOpts...)
	}
}

func (o *OverflowState) Script(filename string, opts ...InteractionOption) *OverflowScriptResult {
	interaction := o.BuildInteraction(filename, "script", opts...)

	osc := &OverflowScriptResult{Input: interaction}
	if interaction.Error != nil {
		osc.Err = interaction.Error
		return osc
	}

	filePath := fmt.Sprintf("%s/%s.cdc", interaction.BasePath, interaction.FileName)

	o.EmulatorLog.Reset()
	o.Log.Reset()

	result, err := o.Services.Scripts.Execute(
		interaction.TransactionCode,
		interaction.Arguments,
		filePath,
		o.Network)

	osc.Result = result
	osc.Output = CadenceValueToInterface(result)
	osc.Err = errors.Wrapf(err, "scriptFileName:%s", interaction.FileName)

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
	if osc.Err != nil {
		return osc
	}

	o.Logger.Info(fmt.Sprintf("%v Script run from path %s\n", emoji.Star, filePath))
	return osc
}

type OverflowScriptResult struct {
	Err    error
	Result cadence.Value
	Input  *FlowInteractionBuilder
	Log    []LogrusMessage
	Output interface{}
}

func (osr *OverflowScriptResult) GetAsJson() (string, error) {
	if osr.Err != nil {
		return "", errors.Wrapf(osr.Err, "script: %s", osr.Input.FileName)
	}
	j, err := json.MarshalIndent(osr.Output, "", "    ")

	if err != nil {
		return "", errors.Wrapf(err, "script %s", osr.Input.FileName)
	}

	return string(j), nil
}

func (osr *OverflowScriptResult) GetAsInterface() (interface{}, error) {
	if osr.Err != nil {
		return nil, errors.Wrapf(osr.Err, "script: %s", osr.Input.FileName)
	}
	return osr.Output, nil
}

func (osr *OverflowScriptResult) AssertWithPointerError(t *testing.T, pointer string, message string) *OverflowScriptResult {
	_, err := osr.GetWithPointer(pointer)
	assert.Error(t, err)
	assert.ErrorContains(t, err, message, "output", litter.Sdump(osr.Output))

	return osr
}

func (osr *OverflowScriptResult) AssertWithPointer(t *testing.T, pointer string, value interface{}) *OverflowScriptResult {
	result, err := osr.GetWithPointer(pointer)
	assert.NoError(t, err)

	assert.Equal(t, value, result, "output", litter.Sdump(osr.Output))

	return osr
}

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

func (osr *OverflowScriptResult) MarshalAs(marshalTo interface{}) error {
	bytes, err := json.Marshal(osr.Output)
	if err != nil {
		return err
	}

	json.Unmarshal(bytes, marshalTo)
	return nil
}

func (osr *OverflowScriptResult) MarshalPointerAs(pointer string, marshalTo interface{}) error {
	ptr, err := gojsonpointer.NewJsonPointer(pointer)
	if err != nil {
		return err
	}
	result, _, err := ptr.Get(osr.Output)

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	json.Unmarshal(bytes, marshalTo)
	return nil
}

func (osr *OverflowScriptResult) GetWithPointer(pointer string) (interface{}, error) {

	ptr, err := gojsonpointer.NewJsonPointer(pointer)
	if err != nil {
		return nil, err
	}
	result, _, err := ptr.Get(osr.Output)
	return result, err
}

func (osr *OverflowScriptResult) AssertWant(t *testing.T, want autogold.Value) *OverflowScriptResult {
	assert.NoError(t, osr.Err)

	switch osr.Output.(type) {
	case []interface{}:
	case map[interface{}]interface{}:
		want.Equal(t, litter.Sdump(osr.Output))
	default:
		want.Equal(t, osr.Output)
	}
	/*
		if osr.Output == nil {
			want.Equal(t, osr.Output)
		} else {
			want.Equal(t, litter.Sdump(osr.Output))
		}
	*/
	return osr
}

func (osr *OverflowScriptResult) Print() {
	json, err := osr.GetAsJson()
	if err != nil {
		osr.Input.Overflow.Logger.Error(err.Error())
		return
	}
	osr.Input.Overflow.Logger.Info(fmt.Sprintf("%v Script %s run from result: %v\n", emoji.Star, osr.Input.FileName, json))
}

func (osr *OverflowScriptResult) MarhalAs(value interface{}) error {
	if osr.Err != nil {
		return osr.Err
	}
	result, err := osr.GetAsJson()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(result), &value)
	return err
}

// everything below here is deprecated

//FlowScriptBuilder is a struct to hold information for running a script
//Deprecation use FlowInteractionBuilder and the Script method
type FlowScriptBuilder struct {
	Overflow       *OverflowState
	FileName       string
	Arguments      []cadence.Value
	ScriptAsString string
	BasePath       string
	Error          error
}

//Script start a script builder with the inline script as body
//Deprecation use FlowInteractionBuilder and the Script method
func (o *OverflowState) InlineScript(content string) FlowScriptBuilder {
	return FlowScriptBuilder{
		Overflow:       o,
		FileName:       "inline",
		Arguments:      []cadence.Value{},
		ScriptAsString: content,
		BasePath:       fmt.Sprintf("%s/scripts", o.BasePath),
	}
}

//ScriptFromFile will start a flow script builder
//Deprecation use FlowInteractionBuilder and the Script method
func (o *OverflowState) ScriptFromFile(filename string) FlowScriptBuilder {
	return FlowScriptBuilder{
		Overflow:       o,
		FileName:       filename,
		Arguments:      []cadence.Value{},
		ScriptAsString: "",
		BasePath:       fmt.Sprintf("%s/scripts", o.BasePath),
	}
}

//Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) ScriptPath(path string) FlowScriptBuilder {
	t.BasePath = path
	return t
}

//Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) NamedArguments(args map[string]string) FlowScriptBuilder {

	scriptFilePath := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)
	code, err := t.getScriptCode(scriptFilePath)
	if err != nil {
		t.Error = err
		return t
	}
	parseArgs, err := t.Overflow.ParseArgumentsWithoutType(t.FileName, code, args)
	if err != nil {
		t.Error = err
		return t
	}
	t.Arguments = parseArgs
	return t
}

// Specify arguments to send to transaction using a raw list of values
//Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) ArgsV(args []cadence.Value) FlowScriptBuilder {
	t.Arguments = args
	return t
}

//Specify arguments to send to transaction using a builder you send in
//Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) Args(args *FlowArgumentsBuilder) FlowScriptBuilder {
	t.Arguments = args.Build()
	return t
}

// Specify arguments to send to transaction using a function that takes a builder where you call the builder
//Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) ArgsFn(fn func(*FlowArgumentsBuilder)) FlowScriptBuilder {
	args := t.Overflow.Arguments()
	fn(args)
	t.Arguments = args.Build()
	return t
}

// Run executes a read only script
//Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) Run() {
	result := t.RunFailOnError()
	res, err := CadenceValueToJsonString(result)
	if err != nil {
		panic(err)
	}
	t.Overflow.Logger.Info(fmt.Sprintf("%v Script run from result: %v\n", emoji.Star, res))
}

// Deprecation use FlowInteractionBuilder
func (t FlowScriptBuilder) getScriptCode(scriptFilePath string) ([]byte, error) {

	var err error
	script := []byte(t.ScriptAsString)
	if t.ScriptAsString == "" {
		script, err = t.Overflow.State.ReaderWriter().ReadFile(scriptFilePath)
		if err != nil {
			return nil, err
		}
	}

	return script, nil
}

// RunReturns executes a read only script
// Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) RunReturns() (cadence.Value, error) {

	if t.Error != nil {
		return nil, t.Error
	}

	f := t.Overflow
	scriptFilePath := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)
	script, err := t.getScriptCode(scriptFilePath)
	if err != nil {
		return nil, err
	}

	t.Overflow.EmulatorLog.Reset()

	result, err := f.Services.Scripts.Execute(
		script,
		t.Arguments,
		scriptFilePath,
		f.Network)
	if err != nil {
		return nil, err
	}

	t.Overflow.EmulatorLog.Reset()
	f.Logger.Info(fmt.Sprintf("%v Script run from path %s\n", emoji.Star, scriptFilePath))
	return result, nil
}

// Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) RunFailOnError() cadence.Value {
	result, err := t.RunReturns()
	if err != nil {
		panic(errors.Wrapf(err, "scriptName:%s", t.FileName))
	}
	return result

}

// RunMarshalAs runs the script and marshals the result into the provided value, returning an error if any
// Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) RunMarshalAs(value interface{}) error {
	result, err := t.RunReturns()
	if err != nil {
		return err
	}
	jsonResult, err := CadenceValueToJsonString(result)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(jsonResult), &value)
	return err
}

// RunReturnsJsonString runs the script and returns pretty printed json string
// Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) RunReturnsJsonString() string {
	res, err := CadenceValueToJsonString(t.RunFailOnError())
	if err != nil {
		panic(err)
	}
	return res
}

//RunReturnsInterface runs the script and returns interface{}
// Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) RunReturnsInterface() interface{} {
	return CadenceValueToInterface(t.RunFailOnError())
}
