package overflow

import (
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

// FlowArgumentsBuilder
//
// The old way of specifing arguments to interactions
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
type FlowArgumentsBuilder struct {
	Overflow  *OverflowState
	Arguments []cadence.Value
	Error     error
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Build() []cadence.Value {
	if a.Error != nil {
		panic(a.Error)
	}
	return a.Arguments
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) RawAddress(address string) *FlowArgumentsBuilder {
	account := flow.HexToAddress(address)
	accountArg := cadence.BytesToAddress(account.Bytes())
	return a.Argument(accountArg)
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Address(key string) *FlowArgumentsBuilder {
	f := a.Overflow

	account, err := f.AccountE(key)
	if err != nil {
		a.Error = err
		return a
	}
	return a.Argument(cadence.BytesToAddress(account.Address().Bytes()))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) RawAccount(address string) *FlowArgumentsBuilder {
	account := flow.HexToAddress(address)
	accountArg := cadence.BytesToAddress(account.Bytes())
	return a.Argument(accountArg)
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Account(key string) *FlowArgumentsBuilder {
	f := a.Overflow

	account, err := f.AccountE(key)
	if err != nil {
		a.Error = err
		return a
	}
	return a.Argument(cadence.BytesToAddress(account.Address().Bytes()))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) String(value string) *FlowArgumentsBuilder {
	return a.Argument(cadence.String(value))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Boolean(value bool) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewBool(value))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Bytes(value []byte) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewBytes(value))
}

// Int add an Int Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Int(value int) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewInt(value))
}

// Int8 add an Int8 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Int8(value int8) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewInt8(value))
}

// Int16 add an Int16 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Int16(value int16) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewInt16(value))
}

// Int32 add an Int32 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Int32(value int32) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewInt32(value))
}

// Int64 add an Int64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Int64(value int64) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewInt64(value))
}

// Int128 add an Int128 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Int128(value int) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewInt128(value))
}

// Int256 add an Int256 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Int256(value int) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewInt256(value))
}

// UInt add an UInt Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt(value uint) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewUInt(value))
}

// UInt8 add an UInt8 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt8(value uint8) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewUInt8(value))
}

// UInt16 add an UInt16 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt16(value uint16) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewUInt16(value))
}

// UInt32 add an UInt32 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt32(value uint32) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewUInt32(value))
}

// UInt64 add an UInt64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt64(value uint64) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewUInt64(value))
}

// UInt128 add an UInt128 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt128(value uint) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewUInt128(value))
}

// UInt256 add an UInt256 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt256(value uint) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewUInt256(value))
}

// Word8 add a Word8 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Word8(value uint8) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewWord8(value))
}

// Word16 add a Word16 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Word16(value uint16) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewWord16(value))
}

// Word32 add a Word32 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Word32(value uint32) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewWord32(value))
}

// Word64 add a Word64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Word64(value uint64) *FlowArgumentsBuilder {
	return a.Argument(cadence.NewWord64(value))
}

// Fix64 add a Fix64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Fix64(value string) *FlowArgumentsBuilder {
	amount, err := cadence.NewFix64(value)
	if err != nil {
		a.Error = err
		return a
	}
	return a.Argument(amount)
}

// DateStringAsUnixTimestamp sends a dateString parsed in the timezone as a unix timeszone ufix
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) DateStringAsUnixTimestamp(dateString string, timezone string) *FlowArgumentsBuilder {
	value, err := parseTime(dateString, timezone)
	if err != nil {
		a.Error = err
		return a
	}

	//swallow the error since it will never happen here, we control the input
	amount, _ := cadence.NewUFix64(value)
	return a.Argument(amount)
}

// UFix64 add a UFix64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UFix64(input float64) *FlowArgumentsBuilder {
	value := fmt.Sprintf("%.8f", input)
	amount, err := cadence.NewUFix64(value)
	if err != nil {
		a.Error = err
		return a
	}
	return a.Argument(amount)
}

// PublicPath argument
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) PublicPath(input string) *FlowArgumentsBuilder {
	path := cadence.Path{Domain: "public", Identifier: input}
	return a.Argument(path)
}

// StoragePath argument
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) StoragePath(input string) *FlowArgumentsBuilder {
	path := cadence.Path{Domain: "storage", Identifier: input}
	return a.Argument(path)
}

// PrivatePath argument
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) PrivatePath(input string) *FlowArgumentsBuilder {
	path := cadence.Path{Domain: "private", Identifier: input}
	return a.Argument(path)
}

// Argument add an argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) Argument(value cadence.Value) *FlowArgumentsBuilder {
	a.Arguments = append(a.Arguments, value)
	return a
}

// Add an {String:String} to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) StringMap(input map[string]string) *FlowArgumentsBuilder {
	array := []cadence.KeyValuePair{}
	for key, val := range input {
		array = append(array, cadence.KeyValuePair{Key: cadenceString(key), Value: cadenceString(val)})
	}
	a.Arguments = append(a.Arguments, cadence.NewDictionary(array))
	return a
}

// Add an {String:UFix64} to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) ScalarMap(input map[string]string) *FlowArgumentsBuilder {
	array := []cadence.KeyValuePair{}
	for key, val := range input {
		UFix64Val, err := cadence.NewUFix64(val)
		if err != nil {
			a.Error = err
			return a
		}
		array = append(array, cadence.KeyValuePair{Key: cadenceString(key), Value: UFix64Val})
	}
	a.Arguments = append(a.Arguments, cadence.NewDictionary(array))
	return a
}

// Argument add an StringArray to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) StringArray(value ...string) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, val := range value {
		array = append(array, cadenceString(val))
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}

// Argument add an StringMapArray to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) StringMapArray(value ...map[string]string) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, vals := range value {
		dict := []cadence.KeyValuePair{}
		for key, val := range vals {
			dict = append(dict, cadence.KeyValuePair{Key: cadenceString(key), Value: cadenceString(val)})
		}
		array = append(array, cadence.NewDictionary(dict))
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}

// Argument add an StringArray to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) ScalarMapArray(value ...map[string]string) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, vals := range value {
		dict := []cadence.KeyValuePair{}
		for key, val := range vals {
			UFix64Val, err := cadence.NewUFix64(val)
			if err != nil {
				a.Error = err
				return a
			}
			dict = append(dict, cadence.KeyValuePair{Key: cadenceString(key), Value: UFix64Val})
		}
		array = append(array, cadence.NewDictionary(dict))
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}

// Argument add a RawAddressArray to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) RawAddressArray(value ...string) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, val := range value {
		address := flow.HexToAddress(val)
		cadenceAddress := cadence.BytesToAddress(address.Bytes())
		array = append(array, cadenceAddress)
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}

// Argument add a RawAddressArray to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) AccountArray(value ...string) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, val := range value {
		address, err := a.Overflow.AccountE(val)
		if err != nil {
			a.Error = err
			return a
		}
		cadenceAddress := cadence.BytesToAddress(address.Address().Bytes())
		array = append(array, cadenceAddress)
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}

// Argument add an argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt64Array(value ...uint64) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, val := range value {
		cadenceVal := cadence.NewUInt64(val)
		array = append(array, cadenceVal)
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}

// Argument add an argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UInt8Array(value ...uint8) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, val := range value {
		cadenceVal := cadence.NewUInt8(val)
		array = append(array, cadenceVal)
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}

// Argument add an argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *FlowArgumentsBuilder) UFix64Array(value ...float64) *FlowArgumentsBuilder {
	array := []cadence.Value{}
	for _, val := range value {
		stringValue := fmt.Sprintf("%f", val)
		amount, err := cadence.NewUFix64(stringValue)
		if err != nil {
			a.Error = err
			return a
		}
		array = append(array, amount)
	}
	a.Arguments = append(a.Arguments, cadence.NewArray(array))
	return a
}
