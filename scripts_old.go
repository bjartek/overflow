package overflow

import (
	"encoding/json"
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/pkg/errors"
)

// OverflowScriptBuilder is a struct to hold information for running a script
//
// Deprecated: use FlowInteractionBuilder and the Script method
type OverflowScriptBuilder struct {
	Overflow       *OverflowState
	FileName       string
	Arguments      []cadence.Value
	ScriptAsString string
	BasePath       string
	Error          error
}

// Script start a script builder with the inline script as body
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (o *OverflowState) InlineScript(content string) OverflowScriptBuilder {
	return OverflowScriptBuilder{
		Overflow:       o,
		FileName:       "inline",
		Arguments:      []cadence.Value{},
		ScriptAsString: content,
		BasePath:       fmt.Sprintf("%s/scripts", o.BasePath),
	}
}

// ScriptFromFile will start a flow script builder
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (o *OverflowState) ScriptFromFile(filename string) OverflowScriptBuilder {
	return OverflowScriptBuilder{
		Overflow:       o,
		FileName:       filename,
		Arguments:      []cadence.Value{},
		ScriptAsString: "",
		BasePath:       fmt.Sprintf("%s/scripts", o.BasePath),
	}
}

//Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) ScriptPath(path string) OverflowScriptBuilder {
	t.BasePath = path
	return t
}

//Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) NamedArguments(args map[string]string) OverflowScriptBuilder {

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
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) ArgsV(args []cadence.Value) OverflowScriptBuilder {
	t.Arguments = args
	return t
}

// Specify arguments to send to transaction using a builder you send in
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) Args(args *OverflowArgumentsBuilder) OverflowScriptBuilder {
	t.Arguments = args.Build()
	return t
}

// Specify arguments to send to transaction using a function that takes a builder where you call the builder
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) ArgsFn(fn func(*OverflowArgumentsBuilder)) OverflowScriptBuilder {
	args := t.Overflow.Arguments()
	fn(args)
	t.Arguments = args.Build()
	return t
}

// Run executes a read only script
// Deprecated:  use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) Run() {
	result := t.RunFailOnError()
	res, err := CadenceValueToJsonString(result)
	if err != nil {
		panic(err)
	}
	t.Overflow.Logger.Info(fmt.Sprintf("%v Script run from result: %v\n", emoji.Star, res))
}

// Deprecated: use FlowInteractionBuilder
func (t OverflowScriptBuilder) getScriptCode(scriptFilePath string) ([]byte, error) {

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
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) RunReturns() (cadence.Value, error) {

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

	flowScript := &services.Script{
		Code:     script,
		Args:     t.Arguments,
		Filename: scriptFilePath,
	}
	result, err := f.Services.Scripts.Execute(flowScript, f.Network)
	if err != nil {
		return nil, err
	}

	t.Overflow.EmulatorLog.Reset()
	f.Logger.Info(fmt.Sprintf("%v Script run from path %s\n", emoji.Star, scriptFilePath))
	return result, nil
}

// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) RunFailOnError() cadence.Value {
	result, err := t.RunReturns()
	if err != nil {
		panic(errors.Wrapf(err, "scriptName:%s", t.FileName))
	}
	return result

}

// RunMarshalAs runs the script and marshals the result into the provided value, returning an error if any
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) RunMarshalAs(value interface{}) error {
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
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) RunReturnsJsonString() string {
	res, err := CadenceValueToJsonString(t.RunFailOnError())
	if err != nil {
		panic(err)
	}
	return res
}

// RunReturnsInterface runs the script and returns interface{}
//
// Deprecated: use FlowInteractionBuilder and the Script method
func (t OverflowScriptBuilder) RunReturnsInterface() interface{} {
	return CadenceValueToInterface(t.RunFailOnError())
}
