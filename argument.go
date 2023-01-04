package overflow

import (
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

// OverflowArgumentsBuilder
//
// # The old way of specifying arguments to interactions
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
type OverflowArgumentsBuilder struct {
	Overflow  *OverflowState
	Arguments []cadence.Value
	Error     error
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Build() []cadence.Value {
	if a.Error != nil {
		panic(a.Error)
	}
	return a.Arguments
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) RawAddress(address string) *OverflowArgumentsBuilder {
	account := flow.HexToAddress(address)
	accountArg := cadence.BytesToAddress(account.Bytes())
	return a.Argument(accountArg)
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Address(key string) *OverflowArgumentsBuilder {
	f := a.Overflow

	account, err := f.AccountE(key)
	if err != nil {
		a.Error = err
		return a
	}
	return a.Argument(cadence.BytesToAddress(account.Address().Bytes()))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) RawAccount(address string) *OverflowArgumentsBuilder {
	account := flow.HexToAddress(address)
	accountArg := cadence.BytesToAddress(account.Bytes())
	return a.Argument(accountArg)
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Account(key string) *OverflowArgumentsBuilder {
	f := a.Overflow

	account, err := f.AccountE(key)
	if err != nil {
		a.Error = err
		return a
	}
	return a.Argument(cadence.BytesToAddress(account.Address().Bytes()))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) String(value string) *OverflowArgumentsBuilder {
	return a.Argument(cadence.String(value))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Boolean(value bool) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewBool(value))
}

// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Bytes(value []byte) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewBytes(value))
}

// Int add an Int Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Int(value int) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewInt(value))
}

// Int8 add an Int8 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Int8(value int8) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewInt8(value))
}

// Int16 add an Int16 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Int16(value int16) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewInt16(value))
}

// Int32 add an Int32 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Int32(value int32) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewInt32(value))
}

// Int64 add an Int64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Int64(value int64) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewInt64(value))
}

// Int128 add an Int128 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Int128(value int) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewInt128(value))
}

// Int256 add an Int256 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Int256(value int) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewInt256(value))
}

// UInt add an UInt Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) UInt(value uint) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewUInt(value))
}

// UInt8 add an UInt8 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) UInt8(value uint8) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewUInt8(value))
}

// UInt16 add an UInt16 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) UInt16(value uint16) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewUInt16(value))
}

// UInt32 add an UInt32 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) UInt32(value uint32) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewUInt32(value))
}

// UInt64 add an UInt64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) UInt64(value uint64) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewUInt64(value))
}

// UInt128 add an UInt128 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) UInt128(value uint) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewUInt128(value))
}

// UInt256 add an UInt256 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) UInt256(value uint) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewUInt256(value))
}

// Word8 add a Word8 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Word8(value uint8) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewWord8(value))
}

// Word16 add a Word16 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Word16(value uint16) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewWord16(value))
}

// Word32 add a Word32 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Word32(value uint32) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewWord32(value))
}

// Word64 add a Word64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Word64(value uint64) *OverflowArgumentsBuilder {
	return a.Argument(cadence.NewWord64(value))
}

// Fix64 add a Fix64 Argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Fix64(value string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) DateStringAsUnixTimestamp(dateString string, timezone string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) UFix64(input float64) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) PublicPath(input string) *OverflowArgumentsBuilder {
	path := cadence.Path{Domain: "public", Identifier: input}
	return a.Argument(path)
}

// StoragePath argument
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) StoragePath(input string) *OverflowArgumentsBuilder {
	path := cadence.Path{Domain: "storage", Identifier: input}
	return a.Argument(path)
}

// PrivatePath argument
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) PrivatePath(input string) *OverflowArgumentsBuilder {
	path := cadence.Path{Domain: "private", Identifier: input}
	return a.Argument(path)
}

// Argument add an argument to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) Argument(value cadence.Value) *OverflowArgumentsBuilder {
	a.Arguments = append(a.Arguments, value)
	return a
}

// Add an {String:String} to the transaction
//
// Deprecated: This builder and all its methods are deprecated. Use the new Tx/Script methods and its argument method
func (a *OverflowArgumentsBuilder) StringMap(input map[string]string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) ScalarMap(input map[string]string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) StringArray(value ...string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) StringMapArray(value ...map[string]string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) ScalarMapArray(value ...map[string]string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) RawAddressArray(value ...string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) AccountArray(value ...string) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) UInt64Array(value ...uint64) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) UInt8Array(value ...uint8) *OverflowArgumentsBuilder {
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
func (a *OverflowArgumentsBuilder) UFix64Array(value ...float64) *OverflowArgumentsBuilder {
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
