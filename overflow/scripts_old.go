package overflow

import (
	"encoding/json"
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"
)

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
