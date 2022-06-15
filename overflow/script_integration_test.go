package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScriptIntegration(t *testing.T) {
	g := NewTestingEmulator().Start()
	t.Parallel()

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").Args(g.Arguments().RawAccount("0x1cf0e2f2f715450")).RunReturnsInterface()
		assert.Equal(t, "0x01cf0e2f2f715450", value)
	})

	t.Run("run and output result", func(t *testing.T) {
		g.ScriptFromFile("test").Args(g.Arguments().RawAccount("0x1cf0e2f2f715450")).Run()
	})

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").Args(g.Arguments().Account("first")).RunReturnsInterface()
		assert.Equal(t, "0x01cf0e2f2f715450", value)
	})

	t.Run("Script in different folder", func(t *testing.T) {
		_, err := g.ScriptFromFile("block").ScriptPath("./cadence/scripts").RunReturns()
		assert.NoError(t, err)
	})

	t.Run("Script should report failure", func(t *testing.T) {
		value, err := g.ScriptFromFile("asdf").RunReturns()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "open ./scripts/asdf.cdc: no such file or directory")
		assert.Nil(t, value)

	})

	t.Run("Script should report failure", func(t *testing.T) {
		value, err := g.Script("asdf").RunReturns()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Parsing failed")
		assert.Nil(t, value)

	})

	t.Run("marshal test result into struct", func(t *testing.T) {
		var result []TestReturn
		err := g.Script(`
pub struct Report{
				 pub let name: String
				 pub let test: String

				 init(name: String, test:String) {
					 self.name=name
					 self.test=test
				 }
			 }

			 pub fun main() : [Report] {
				 return [Report(name:"name1", test: "test1"), Report(name:"name2", test: "test2")]
			 }
		`).RunMarshalAs(&result)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, TestReturn{Name: "name1", Test: "test1"}, result[0])
	})

	t.Run("marshal test result into struct should fail if wrong type", func(t *testing.T) {
		var result TestReturn
		err := g.Script(`
pub struct Report{
				 pub let name: String
				 pub let test: String

				 init(name: String, test:String) {
					 self.name=name
					 self.test=test
				 }
			 }

			 pub fun main() : [Report] {
				 return [Report(name:"name1", test: "test1"), Report(name:"name2", test: "test2")]
			 }
		`).RunMarshalAs(&result)
		assert.Error(t, err, "should return error")
		assert.Contains(t, err.Error(), "json: cannot unmarshal array into Go value of type overflow.TestReturn")
	})

	t.Run("Named arguments", func(t *testing.T) {
		value := g.ScriptFromFile("test").
			NamedArguments(map[string]string{
				"account": "first",
			}).RunReturnsInterface()

		assert.Equal(t, "0x01cf0e2f2f715450", value)
	})

	t.Run("Named arguments error if not all arguments", func(t *testing.T) {
		_, err := g.ScriptFromFile("test").
			NamedArguments(map[string]string{
				"aaccount": "first",
			}).RunReturns()
		assert.ErrorContains(t, err, "the following arguments where not present [account]")
	})

	t.Run("Named arguments error if wrong file", func(t *testing.T) {
		_, err := g.ScriptFromFile("foo/test").
			NamedArguments(map[string]string{
				"aaccount": "first",
			}).RunReturns()
		assert.ErrorContains(t, err, "open ./scripts/foo/test.cdc: no such file or directory")
	})

	t.Run("Named arguments blank", func(t *testing.T) {
		value := g.ScriptFromFile("block").
			NamedArguments(map[string]string{}).
			RunReturnsInterface()

		assert.Equal(t, "4", value)
	})

}

type TestReturn struct {
	Name string
	Test string
}
