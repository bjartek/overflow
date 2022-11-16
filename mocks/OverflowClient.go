// Code generated by mockery v2.14.1. DO NOT EDIT.

package mocks

import (
	flowkit "github.com/onflow/flow-cli/pkg/flowkit"
	flow "github.com/onflow/flow-go-sdk"

	mock "github.com/stretchr/testify/mock"

	overflow "github.com/bjartek/overflow"

	services "github.com/onflow/flow-cli/pkg/flowkit/services"
)

// OverflowClient is an autogenerated mock type for the OverflowClient type
type OverflowClient struct {
	mock.Mock
}

type OverflowClient_Expecter struct {
	mock *mock.Mock
}

func (_m *OverflowClient) EXPECT() *OverflowClient_Expecter {
	return &OverflowClient_Expecter{mock: &_m.Mock}
}

// Account provides a mock function with given fields: key
func (_m *OverflowClient) Account(key string) *flowkit.Account {
	ret := _m.Called(key)

	var r0 *flowkit.Account
	if rf, ok := ret.Get(0).(func(string) *flowkit.Account); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flowkit.Account)
		}
	}

	return r0
}

// OverflowClient_Account_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Account'
type OverflowClient_Account_Call struct {
	*mock.Call
}

// Account is a helper method to define mock.On call
//   - key string
func (_e *OverflowClient_Expecter) Account(key interface{}) *OverflowClient_Account_Call {
	return &OverflowClient_Account_Call{Call: _e.mock.On("Account", key)}
}

func (_c *OverflowClient_Account_Call) Run(run func(key string)) *OverflowClient_Account_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *OverflowClient_Account_Call) Return(_a0 *flowkit.Account) *OverflowClient_Account_Call {
	_c.Call.Return(_a0)
	return _c
}

// AccountE provides a mock function with given fields: key
func (_m *OverflowClient) AccountE(key string) (*flowkit.Account, error) {
	ret := _m.Called(key)

	var r0 *flowkit.Account
	if rf, ok := ret.Get(0).(func(string) *flowkit.Account); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flowkit.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverflowClient_AccountE_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AccountE'
type OverflowClient_AccountE_Call struct {
	*mock.Call
}

// AccountE is a helper method to define mock.On call
//   - key string
func (_e *OverflowClient_Expecter) AccountE(key interface{}) *OverflowClient_AccountE_Call {
	return &OverflowClient_AccountE_Call{Call: _e.mock.On("AccountE", key)}
}

func (_c *OverflowClient_AccountE_Call) Run(run func(key string)) *OverflowClient_AccountE_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *OverflowClient_AccountE_Call) Return(_a0 *flowkit.Account, _a1 error) *OverflowClient_AccountE_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// AddContract provides a mock function with given fields: name, contract, update
func (_m *OverflowClient) AddContract(name string, contract *services.Contract, update bool) error {
	ret := _m.Called(name, contract, update)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *services.Contract, bool) error); ok {
		r0 = rf(name, contract, update)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OverflowClient_AddContract_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddContract'
type OverflowClient_AddContract_Call struct {
	*mock.Call
}

// AddContract is a helper method to define mock.On call
//   - name string
//   - contract *services.Contract
//   - update bool
func (_e *OverflowClient_Expecter) AddContract(name interface{}, contract interface{}, update interface{}) *OverflowClient_AddContract_Call {
	return &OverflowClient_AddContract_Call{Call: _e.mock.On("AddContract", name, contract, update)}
}

func (_c *OverflowClient_AddContract_Call) Run(run func(name string, contract *services.Contract, update bool)) *OverflowClient_AddContract_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(*services.Contract), args[2].(bool))
	})
	return _c
}

func (_c *OverflowClient_AddContract_Call) Return(_a0 error) *OverflowClient_AddContract_Call {
	_c.Call.Return(_a0)
	return _c
}

// Address provides a mock function with given fields: key
func (_m *OverflowClient) Address(key string) string {
	ret := _m.Called(key)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// OverflowClient_Address_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Address'
type OverflowClient_Address_Call struct {
	*mock.Call
}

// Address is a helper method to define mock.On call
//   - key string
func (_e *OverflowClient_Expecter) Address(key interface{}) *OverflowClient_Address_Call {
	return &OverflowClient_Address_Call{Call: _e.mock.On("Address", key)}
}

func (_c *OverflowClient_Address_Call) Run(run func(key string)) *OverflowClient_Address_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *OverflowClient_Address_Call) Return(_a0 string) *OverflowClient_Address_Call {
	_c.Call.Return(_a0)
	return _c
}

// FetchEventsWithResult provides a mock function with given fields: opts
func (_m *OverflowClient) FetchEventsWithResult(opts ...overflow.OverflowEventFetcherOption) overflow.EventFetcherResult {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 overflow.EventFetcherResult
	if rf, ok := ret.Get(0).(func(...overflow.OverflowEventFetcherOption) overflow.EventFetcherResult); ok {
		r0 = rf(opts...)
	} else {
		r0 = ret.Get(0).(overflow.EventFetcherResult)
	}

	return r0
}

// OverflowClient_FetchEventsWithResult_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchEventsWithResult'
type OverflowClient_FetchEventsWithResult_Call struct {
	*mock.Call
}

// FetchEventsWithResult is a helper method to define mock.On call
//   - opts ...overflow.OverflowEventFetcherOption
func (_e *OverflowClient_Expecter) FetchEventsWithResult(opts ...interface{}) *OverflowClient_FetchEventsWithResult_Call {
	return &OverflowClient_FetchEventsWithResult_Call{Call: _e.mock.On("FetchEventsWithResult",
		append([]interface{}{}, opts...)...)}
}

func (_c *OverflowClient_FetchEventsWithResult_Call) Run(run func(opts ...overflow.OverflowEventFetcherOption)) *OverflowClient_FetchEventsWithResult_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]overflow.OverflowEventFetcherOption, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(overflow.OverflowEventFetcherOption)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *OverflowClient_FetchEventsWithResult_Call) Return(_a0 overflow.EventFetcherResult) *OverflowClient_FetchEventsWithResult_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetAccount provides a mock function with given fields: key
func (_m *OverflowClient) GetAccount(key string) (*flow.Account, error) {
	ret := _m.Called(key)

	var r0 *flow.Account
	if rf, ok := ret.Get(0).(func(string) *flow.Account); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverflowClient_GetAccount_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAccount'
type OverflowClient_GetAccount_Call struct {
	*mock.Call
}

// GetAccount is a helper method to define mock.On call
//   - key string
func (_e *OverflowClient_Expecter) GetAccount(key interface{}) *OverflowClient_GetAccount_Call {
	return &OverflowClient_GetAccount_Call{Call: _e.mock.On("GetAccount", key)}
}

func (_c *OverflowClient_GetAccount_Call) Run(run func(key string)) *OverflowClient_GetAccount_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *OverflowClient_GetAccount_Call) Return(_a0 *flow.Account, _a1 error) *OverflowClient_GetAccount_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetBlockAtHeight provides a mock function with given fields: height
func (_m *OverflowClient) GetBlockAtHeight(height uint64) (*flow.Block, error) {
	ret := _m.Called(height)

	var r0 *flow.Block
	if rf, ok := ret.Get(0).(func(uint64) *flow.Block); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverflowClient_GetBlockAtHeight_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBlockAtHeight'
type OverflowClient_GetBlockAtHeight_Call struct {
	*mock.Call
}

// GetBlockAtHeight is a helper method to define mock.On call
//   - height uint64
func (_e *OverflowClient_Expecter) GetBlockAtHeight(height interface{}) *OverflowClient_GetBlockAtHeight_Call {
	return &OverflowClient_GetBlockAtHeight_Call{Call: _e.mock.On("GetBlockAtHeight", height)}
}

func (_c *OverflowClient_GetBlockAtHeight_Call) Run(run func(height uint64)) *OverflowClient_GetBlockAtHeight_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uint64))
	})
	return _c
}

func (_c *OverflowClient_GetBlockAtHeight_Call) Return(_a0 *flow.Block, _a1 error) *OverflowClient_GetBlockAtHeight_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetBlockById provides a mock function with given fields: blockId
func (_m *OverflowClient) GetBlockById(blockId string) (*flow.Block, error) {
	ret := _m.Called(blockId)

	var r0 *flow.Block
	if rf, ok := ret.Get(0).(func(string) *flow.Block); ok {
		r0 = rf(blockId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(blockId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverflowClient_GetBlockById_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBlockById'
type OverflowClient_GetBlockById_Call struct {
	*mock.Call
}

// GetBlockById is a helper method to define mock.On call
//   - blockId string
func (_e *OverflowClient_Expecter) GetBlockById(blockId interface{}) *OverflowClient_GetBlockById_Call {
	return &OverflowClient_GetBlockById_Call{Call: _e.mock.On("GetBlockById", blockId)}
}

func (_c *OverflowClient_GetBlockById_Call) Run(run func(blockId string)) *OverflowClient_GetBlockById_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *OverflowClient_GetBlockById_Call) Return(_a0 *flow.Block, _a1 error) *OverflowClient_GetBlockById_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetLatestBlock provides a mock function with given fields:
func (_m *OverflowClient) GetLatestBlock() (*flow.Block, error) {
	ret := _m.Called()

	var r0 *flow.Block
	if rf, ok := ret.Get(0).(func() *flow.Block); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverflowClient_GetLatestBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLatestBlock'
type OverflowClient_GetLatestBlock_Call struct {
	*mock.Call
}

// GetLatestBlock is a helper method to define mock.On call
func (_e *OverflowClient_Expecter) GetLatestBlock() *OverflowClient_GetLatestBlock_Call {
	return &OverflowClient_GetLatestBlock_Call{Call: _e.mock.On("GetLatestBlock")}
}

func (_c *OverflowClient_GetLatestBlock_Call) Run(run func()) *OverflowClient_GetLatestBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *OverflowClient_GetLatestBlock_Call) Return(_a0 *flow.Block, _a1 error) *OverflowClient_GetLatestBlock_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetNetwork provides a mock function with given fields:
func (_m *OverflowClient) GetNetwork() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// OverflowClient_GetNetwork_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNetwork'
type OverflowClient_GetNetwork_Call struct {
	*mock.Call
}

// GetNetwork is a helper method to define mock.On call
func (_e *OverflowClient_Expecter) GetNetwork() *OverflowClient_GetNetwork_Call {
	return &OverflowClient_GetNetwork_Call{Call: _e.mock.On("GetNetwork")}
}

func (_c *OverflowClient_GetNetwork_Call) Run(run func()) *OverflowClient_GetNetwork_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *OverflowClient_GetNetwork_Call) Return(_a0 string) *OverflowClient_GetNetwork_Call {
	_c.Call.Return(_a0)
	return _c
}

// QualifiedIdentiferFromSnakeCase provides a mock function with given fields: typeName
func (_m *OverflowClient) QualifiedIdentiferFromSnakeCase(typeName string) (string, error) {
	ret := _m.Called(typeName)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(typeName)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(typeName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverflowClient_QualifiedIdentiferFromSnakeCase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'QualifiedIdentiferFromSnakeCase'
type OverflowClient_QualifiedIdentiferFromSnakeCase_Call struct {
	*mock.Call
}

// QualifiedIdentiferFromSnakeCase is a helper method to define mock.On call
//   - typeName string
func (_e *OverflowClient_Expecter) QualifiedIdentiferFromSnakeCase(typeName interface{}) *OverflowClient_QualifiedIdentiferFromSnakeCase_Call {
	return &OverflowClient_QualifiedIdentiferFromSnakeCase_Call{Call: _e.mock.On("QualifiedIdentiferFromSnakeCase", typeName)}
}

func (_c *OverflowClient_QualifiedIdentiferFromSnakeCase_Call) Run(run func(typeName string)) *OverflowClient_QualifiedIdentiferFromSnakeCase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *OverflowClient_QualifiedIdentiferFromSnakeCase_Call) Return(_a0 string, _a1 error) *OverflowClient_QualifiedIdentiferFromSnakeCase_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// QualifiedIdentifier provides a mock function with given fields: contract, name
func (_m *OverflowClient) QualifiedIdentifier(contract string, name string) (string, error) {
	ret := _m.Called(contract, name)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(contract, name)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(contract, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverflowClient_QualifiedIdentifier_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'QualifiedIdentifier'
type OverflowClient_QualifiedIdentifier_Call struct {
	*mock.Call
}

// QualifiedIdentifier is a helper method to define mock.On call
//   - contract string
//   - name string
func (_e *OverflowClient_Expecter) QualifiedIdentifier(contract interface{}, name interface{}) *OverflowClient_QualifiedIdentifier_Call {
	return &OverflowClient_QualifiedIdentifier_Call{Call: _e.mock.On("QualifiedIdentifier", contract, name)}
}

func (_c *OverflowClient_QualifiedIdentifier_Call) Run(run func(contract string, name string)) *OverflowClient_QualifiedIdentifier_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *OverflowClient_QualifiedIdentifier_Call) Return(_a0 string, _a1 error) *OverflowClient_QualifiedIdentifier_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Script provides a mock function with given fields: filename, opts
func (_m *OverflowClient) Script(filename string, opts ...overflow.OverflowInteractionOption) *overflow.OverflowScriptResult {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, filename)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *overflow.OverflowScriptResult
	if rf, ok := ret.Get(0).(func(string, ...overflow.OverflowInteractionOption) *overflow.OverflowScriptResult); ok {
		r0 = rf(filename, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*overflow.OverflowScriptResult)
		}
	}

	return r0
}

// OverflowClient_Script_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Script'
type OverflowClient_Script_Call struct {
	*mock.Call
}

// Script is a helper method to define mock.On call
//   - filename string
//   - opts ...overflow.OverflowInteractionOption
func (_e *OverflowClient_Expecter) Script(filename interface{}, opts ...interface{}) *OverflowClient_Script_Call {
	return &OverflowClient_Script_Call{Call: _e.mock.On("Script",
		append([]interface{}{filename}, opts...)...)}
}

func (_c *OverflowClient_Script_Call) Run(run func(filename string, opts ...overflow.OverflowInteractionOption)) *OverflowClient_Script_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]overflow.OverflowInteractionOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(overflow.OverflowInteractionOption)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *OverflowClient_Script_Call) Return(_a0 *overflow.OverflowScriptResult) *OverflowClient_Script_Call {
	_c.Call.Return(_a0)
	return _c
}

// ScriptFN provides a mock function with given fields: outerOpts
func (_m *OverflowClient) ScriptFN(outerOpts ...overflow.OverflowInteractionOption) overflow.OverflowScriptFunction {
	_va := make([]interface{}, len(outerOpts))
	for _i := range outerOpts {
		_va[_i] = outerOpts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 overflow.OverflowScriptFunction
	if rf, ok := ret.Get(0).(func(...overflow.OverflowInteractionOption) overflow.OverflowScriptFunction); ok {
		r0 = rf(outerOpts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(overflow.OverflowScriptFunction)
		}
	}

	return r0
}

// OverflowClient_ScriptFN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ScriptFN'
type OverflowClient_ScriptFN_Call struct {
	*mock.Call
}

// ScriptFN is a helper method to define mock.On call
//   - outerOpts ...overflow.OverflowInteractionOption
func (_e *OverflowClient_Expecter) ScriptFN(outerOpts ...interface{}) *OverflowClient_ScriptFN_Call {
	return &OverflowClient_ScriptFN_Call{Call: _e.mock.On("ScriptFN",
		append([]interface{}{}, outerOpts...)...)}
}

func (_c *OverflowClient_ScriptFN_Call) Run(run func(outerOpts ...overflow.OverflowInteractionOption)) *OverflowClient_ScriptFN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]overflow.OverflowInteractionOption, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(overflow.OverflowInteractionOption)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *OverflowClient_ScriptFN_Call) Return(_a0 overflow.OverflowScriptFunction) *OverflowClient_ScriptFN_Call {
	_c.Call.Return(_a0)
	return _c
}

// ScriptFileNameFN provides a mock function with given fields: filename, outerOpts
func (_m *OverflowClient) ScriptFileNameFN(filename string, outerOpts ...overflow.OverflowInteractionOption) overflow.OverflowScriptOptsFunction {
	_va := make([]interface{}, len(outerOpts))
	for _i := range outerOpts {
		_va[_i] = outerOpts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, filename)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 overflow.OverflowScriptOptsFunction
	if rf, ok := ret.Get(0).(func(string, ...overflow.OverflowInteractionOption) overflow.OverflowScriptOptsFunction); ok {
		r0 = rf(filename, outerOpts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(overflow.OverflowScriptOptsFunction)
		}
	}

	return r0
}

// OverflowClient_ScriptFileNameFN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ScriptFileNameFN'
type OverflowClient_ScriptFileNameFN_Call struct {
	*mock.Call
}

// ScriptFileNameFN is a helper method to define mock.On call
//   - filename string
//   - outerOpts ...overflow.OverflowInteractionOption
func (_e *OverflowClient_Expecter) ScriptFileNameFN(filename interface{}, outerOpts ...interface{}) *OverflowClient_ScriptFileNameFN_Call {
	return &OverflowClient_ScriptFileNameFN_Call{Call: _e.mock.On("ScriptFileNameFN",
		append([]interface{}{filename}, outerOpts...)...)}
}

func (_c *OverflowClient_ScriptFileNameFN_Call) Run(run func(filename string, outerOpts ...overflow.OverflowInteractionOption)) *OverflowClient_ScriptFileNameFN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]overflow.OverflowInteractionOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(overflow.OverflowInteractionOption)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *OverflowClient_ScriptFileNameFN_Call) Return(_a0 overflow.OverflowScriptOptsFunction) *OverflowClient_ScriptFileNameFN_Call {
	_c.Call.Return(_a0)
	return _c
}

// Tx provides a mock function with given fields: filename, opts
func (_m *OverflowClient) Tx(filename string, opts ...overflow.OverflowInteractionOption) *overflow.OverflowResult {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, filename)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *overflow.OverflowResult
	if rf, ok := ret.Get(0).(func(string, ...overflow.OverflowInteractionOption) *overflow.OverflowResult); ok {
		r0 = rf(filename, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*overflow.OverflowResult)
		}
	}

	return r0
}

// OverflowClient_Tx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Tx'
type OverflowClient_Tx_Call struct {
	*mock.Call
}

// Tx is a helper method to define mock.On call
//   - filename string
//   - opts ...overflow.OverflowInteractionOption
func (_e *OverflowClient_Expecter) Tx(filename interface{}, opts ...interface{}) *OverflowClient_Tx_Call {
	return &OverflowClient_Tx_Call{Call: _e.mock.On("Tx",
		append([]interface{}{filename}, opts...)...)}
}

func (_c *OverflowClient_Tx_Call) Run(run func(filename string, opts ...overflow.OverflowInteractionOption)) *OverflowClient_Tx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]overflow.OverflowInteractionOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(overflow.OverflowInteractionOption)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *OverflowClient_Tx_Call) Return(_a0 *overflow.OverflowResult) *OverflowClient_Tx_Call {
	_c.Call.Return(_a0)
	return _c
}

// TxFN provides a mock function with given fields: outerOpts
func (_m *OverflowClient) TxFN(outerOpts ...overflow.OverflowInteractionOption) overflow.OverflowTransactionFunction {
	_va := make([]interface{}, len(outerOpts))
	for _i := range outerOpts {
		_va[_i] = outerOpts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 overflow.OverflowTransactionFunction
	if rf, ok := ret.Get(0).(func(...overflow.OverflowInteractionOption) overflow.OverflowTransactionFunction); ok {
		r0 = rf(outerOpts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(overflow.OverflowTransactionFunction)
		}
	}

	return r0
}

// OverflowClient_TxFN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TxFN'
type OverflowClient_TxFN_Call struct {
	*mock.Call
}

// TxFN is a helper method to define mock.On call
//   - outerOpts ...overflow.OverflowInteractionOption
func (_e *OverflowClient_Expecter) TxFN(outerOpts ...interface{}) *OverflowClient_TxFN_Call {
	return &OverflowClient_TxFN_Call{Call: _e.mock.On("TxFN",
		append([]interface{}{}, outerOpts...)...)}
}

func (_c *OverflowClient_TxFN_Call) Run(run func(outerOpts ...overflow.OverflowInteractionOption)) *OverflowClient_TxFN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]overflow.OverflowInteractionOption, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(overflow.OverflowInteractionOption)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *OverflowClient_TxFN_Call) Return(_a0 overflow.OverflowTransactionFunction) *OverflowClient_TxFN_Call {
	_c.Call.Return(_a0)
	return _c
}

// TxFileNameFN provides a mock function with given fields: filename, outerOpts
func (_m *OverflowClient) TxFileNameFN(filename string, outerOpts ...overflow.OverflowInteractionOption) overflow.OverflowTransactionOptsFunction {
	_va := make([]interface{}, len(outerOpts))
	for _i := range outerOpts {
		_va[_i] = outerOpts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, filename)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 overflow.OverflowTransactionOptsFunction
	if rf, ok := ret.Get(0).(func(string, ...overflow.OverflowInteractionOption) overflow.OverflowTransactionOptsFunction); ok {
		r0 = rf(filename, outerOpts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(overflow.OverflowTransactionOptsFunction)
		}
	}

	return r0
}

// OverflowClient_TxFileNameFN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TxFileNameFN'
type OverflowClient_TxFileNameFN_Call struct {
	*mock.Call
}

// TxFileNameFN is a helper method to define mock.On call
//   - filename string
//   - outerOpts ...overflow.OverflowInteractionOption
func (_e *OverflowClient_Expecter) TxFileNameFN(filename interface{}, outerOpts ...interface{}) *OverflowClient_TxFileNameFN_Call {
	return &OverflowClient_TxFileNameFN_Call{Call: _e.mock.On("TxFileNameFN",
		append([]interface{}{filename}, outerOpts...)...)}
}

func (_c *OverflowClient_TxFileNameFN_Call) Run(run func(filename string, outerOpts ...overflow.OverflowInteractionOption)) *OverflowClient_TxFileNameFN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]overflow.OverflowInteractionOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(overflow.OverflowInteractionOption)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *OverflowClient_TxFileNameFN_Call) Return(_a0 overflow.OverflowTransactionOptsFunction) *OverflowClient_TxFileNameFN_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewOverflowClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewOverflowClient creates a new instance of OverflowClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOverflowClient(t mockConstructorTestingTNewOverflowClient) *OverflowClient {
	mock := &OverflowClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
