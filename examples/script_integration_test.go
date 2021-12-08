package main

import (
	"testing"

	"github.com/bjartek/overflow/overflow"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	g := overflow.NewTestingEmulator().Start()
	t.Parallel()

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").RawAccountArgument("0x1cf0e2f2f715450").RunReturnsInterface()
		assert.Equal(t, "0x1cf0e2f2f715450", value)
	})

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").AccountArgument("first").RunReturnsInterface()
		assert.Equal(t, "0x1cf0e2f2f715450", value)
	})

	t.Run("Script in different folder", func(t *testing.T) {
		_, err := g.ScriptFromFile("block").ScriptPath("./cadence/scripts").RunReturns()
		assert.NoError(t, err)
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
		assert.Contains(t, err.Error(), "json: cannot unmarshal array into Go value of type main.TestReturn")
	})

}

type TestReturn struct {
	Name string
	Test string
}
