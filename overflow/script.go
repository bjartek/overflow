package overflow

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"
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
			panic(err)
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
}

func (osr *OverflowScriptResult) GetAsJson() string {
	if osr.Err != nil {
		panic(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, osr.Input.FileName, osr.Err))
	}
	return CadenceValueToJsonStringCompact(osr.Result)
}

func (osr *OverflowScriptResult) GetAsInterface() interface{} {
	if osr.Err != nil {
		panic(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, osr.Input.FileName, osr.Err))
	}
	return CadenceValueToInterfaceCompact(osr.Result)
}

func (osr *OverflowScriptResult) Print() {
	json := osr.GetAsJson()
	osr.Input.Overflow.Logger.Info(fmt.Sprintf("%v Script %s run from result: %v\n", emoji.Star, osr.Input.FileName, json))
}

func (osr *OverflowScriptResult) MarhalAs(value interface{}) error {
	if osr.Err != nil {
		return osr.Err
	}
	jsonResult := CadenceValueToJsonStringCompact(osr.Result)
	err := json.Unmarshal([]byte(jsonResult), &value)
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
	t.Overflow.Logger.Info(fmt.Sprintf("%v Script run from result: %v\n", emoji.Star, CadenceValueToJsonString(result)))
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
	jsonResult := CadenceValueToJsonString(result)
	err = json.Unmarshal([]byte(jsonResult), &value)
	return err
}

// RunReturnsJsonString runs the script and returns pretty printed json string
// Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) RunReturnsJsonString() string {
	return CadenceValueToJsonString(t.RunFailOnError())
}

//RunReturnsInterface runs the script and returns interface{}
// Deprecation use FlowInteractionBuilder and the Script method
func (t FlowScriptBuilder) RunReturnsInterface() interface{} {
	return CadenceValueToInterface(t.RunFailOnError())
}
