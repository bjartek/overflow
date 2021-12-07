package overflow

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
)

//FlowScriptBuilder is a struct to hold information for running a script
type FlowScriptBuilder struct {
	Overflow       *Overflow
	FileName       string
	Arguments      []cadence.Value
	ScriptAsString string
	BasePath       string
}

//Script start a script builder with the inline script as body
func (f *Overflow) Script(content string) FlowScriptBuilder {
	return FlowScriptBuilder{
		Overflow:       f,
		FileName:       "inline",
		Arguments:      []cadence.Value{},
		ScriptAsString: content,
		BasePath:       "./scripts",
	}
}

//ScriptFromFile will start a flow script builder
func (f *Overflow) ScriptFromFile(filename string) FlowScriptBuilder {
	return FlowScriptBuilder{
		Overflow:       f,
		FileName:       filename,
		Arguments:      []cadence.Value{},
		ScriptAsString: "",
		BasePath:       "./scripts",
	}
}

func (t FlowScriptBuilder) ScriptPath(path string) FlowScriptBuilder {
	t.BasePath = path
	return t
}

// Specify arguments to send to transaction using a raw list of values
func (t FlowScriptBuilder) ArgsV(args []cadence.Value) FlowScriptBuilder {
	t.Arguments = args
	return t
}

// Specify arguments to send to transaction using a builder you send in
func (t FlowScriptBuilder) Args(args *FlowArgumentsBuilder) FlowScriptBuilder {
	t.Arguments = args.Build()
	return t
}

// Specify arguments to send to transaction using a function that takes a builder where you call the builder
func (t FlowScriptBuilder) ArgsFn(fn func(*FlowArgumentsBuilder)) FlowScriptBuilder {
	args := t.Overflow.Arguments()
	fn(args)
	t.Arguments = args.Build()
	return t
}

// Run executes a read only script
func (t FlowScriptBuilder) Run() {
	result := t.RunFailOnError()
	t.Overflow.Logger.Info(fmt.Sprintf("%v Script run from result: %v\n", emoji.Star, CadenceValueToJsonString(result)))
}

// RunReturns executes a read only script
func (t FlowScriptBuilder) RunReturns() (cadence.Value, error) {

	f := t.Overflow
	scriptFilePath := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)

	var err error
	script := []byte(t.ScriptAsString)
	if t.ScriptAsString == "" {
		script, err = f.State.ReaderWriter().ReadFile(scriptFilePath)
		if err != nil {
			return nil, err
		}
	}

	result, err := f.Services.Scripts.Execute(
		script,
		t.Arguments,
		scriptFilePath,
		f.Network)
	if err != nil {
		return nil, err
	}

	f.Logger.Info(fmt.Sprintf("%v Script run from path %s\n", emoji.Star, scriptFilePath))
	return result, nil
}
func (t FlowScriptBuilder) RunFailOnError() cadence.Value {
	result, err := t.RunReturns()
	if err != nil {
		t.Overflow.Logger.Error(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, t.FileName, err))
		os.Exit(1)
	}
	return result

}

//RunMarshalAs runs the script and marshals the result into the provided value, returning an error if any
func (t FlowScriptBuilder) RunMarshalAs(value interface{}) error {
	result, err := t.RunReturns()
	if err != nil {
		return err
	}
	jsonResult := CadenceValueToJsonString(result)
	err = json.Unmarshal([]byte(jsonResult), &value)
	return err
}

//RunReturnsJsonString runs the script and returns pretty printed json string
func (t FlowScriptBuilder) RunReturnsJsonString() string {
	return CadenceValueToJsonString(t.RunFailOnError())
}

//RunReturnsInterface runs the script and returns interface{}
func (t FlowScriptBuilder) RunReturnsInterface() interface{} {
	return CadenceValueToInterface(t.RunFailOnError())
}
