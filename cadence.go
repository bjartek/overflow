package overflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"github.com/onflow/cadence"
)

// CadenceValueToJsonString converts a cadence.Value into a json pretty printed string
func CadenceValueToJsonString(value cadence.Value) (string, error) {
	result := CadenceValueToInterface(value)
	if result == nil {
		return "", nil
	}
	j, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		return "", err
	}

	return string(j), nil
}

// CadenceValueToInterface convert a candence.Value into interface{}
func CadenceValueToInterface(field cadence.Value) interface{} {
	if field == nil {
		return nil
	}

	switch field := field.(type) {
	case cadence.Optional:
		return CadenceValueToInterface(field.Value)
	case cadence.Dictionary:
		//fmt.Println("is dict ", field.ToGoValue(), " ", field.String())
		result := map[string]interface{}{}
		for _, item := range field.Pairs {
			value := CadenceValueToInterface(item.Value)
			key := getAndUnquoteString(item.Key)

			if value != nil && key != "" {
				result[key] = value
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case cadence.Struct:
		//fmt.Println("is struct ", field.ToGoValue(), " ", field.String())
		result := map[string]interface{}{}
		subStructNames := field.StructType.Fields

		for j, subField := range field.Fields {
			value := CadenceValueToInterface(subField)
			key := subStructNames[j].Identifier

			//	fmt.Println("struct ", key, "value", value)
			if value != nil {
				result[key] = value
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case cadence.Array:
		//fmt.Println("is array ", field.ToGoValue(), " ", field.String())
		var result []interface{}
		for _, item := range field.Values {
			value := CadenceValueToInterface(item)
			//	fmt.Printf("%+v\n", value)
			if value != nil {
				result = append(result, value)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result

	case cadence.Int:
		return field.Int()
	case cadence.Address:
		return field.String()
	case cadence.TypeValue:
		//fmt.Println("is type ", field.ToGoValue(), " ", field.String())
		return field.StaticType.ID()
	case cadence.String:
		//fmt.Println("is string ", field.ToGoValue(), " ", field.String())
		value := getAndUnquoteString(field)
		if value == "" {
			return nil
		}
		return value

	case cadence.UFix64:
		//fmt.Println("is ufix64 ", field.ToGoValue(), " ", field.String())

		float, _ := strconv.ParseFloat(field.String(), 64)
		return float
	case cadence.Fix64:
		float, _ := strconv.ParseFloat(field.String(), 64)
		return float

	default:
		//fmt.Println("is fallthrough ", field.ToGoValue(), " ", field.String())

		goValue := field.ToGoValue()
		if goValue != nil {
			return goValue
		}
		return field.String()
	}
}

// a resolver to resolve a input type into a name, can be used to resolve struct names for instance
type InputResolver func(string) (string, error)

func InputToCadence(v interface{}, resolver InputResolver) (cadence.Type, cadence.Value, error) {
	f := reflect.ValueOf(v)
	return ReflectToCadence(f, resolver)
}

func primitiveReflectTypeToCadenceype(typ reflect.Type) cadence.Type {
	kind := typ.Kind()
	switch kind {
	case reflect.Pointer:
		return cadence.OptionalType{Type: primitiveReflectTypeToCadenceype(typ.Elem())}
	case reflect.Int:
		return cadence.IntType{}
	case reflect.Int8:
		return cadence.Int8Type{}
	case reflect.Int16:
		return cadence.Int16Type{}
	case reflect.Int32:
		return cadence.Int32Type{}
	case reflect.Int64:
		return cadence.Int64Type{}
	case reflect.Bool:
		return cadence.BoolType{}
	case reflect.Uint:
		return cadence.UIntType{}
	case reflect.Uint8:
		return cadence.UInt8Type{}
	case reflect.Uint16:
		return cadence.UInt16Type{}
	case reflect.Uint32:
		return cadence.UInt32Type{}
	case reflect.Uint64:
		return cadence.UInt64Type{}
	case reflect.String:
		return cadence.StringType{}
	case reflect.Float64:
		return cadence.UFix64Type{}
	case reflect.Map:
		return cadence.DictionaryType{
			KeyType:     primitiveReflectTypeToCadenceype(typ.Key()),
			ElementType: primitiveReflectTypeToCadenceype(typ.Elem()),
		}
	case reflect.Slice, reflect.Array:
		return cadence.VariableSizedArrayType{ElementType: primitiveReflectTypeToCadenceype(typ.Elem())}
	}
	return nil
}

func ReflectToCadence(value reflect.Value, resolver InputResolver) (cadence.Type, cadence.Value, error) {
	inputType := value.Type()

	kind := inputType.Kind()
	switch kind {
	case reflect.Interface:
		value, err := cadence.NewValue(value.Interface())
		return value.Type(), value, err
	case reflect.Struct:
		var val []cadence.Value
		fields := []cadence.Field{}
		for i := 0; i < value.NumField(); i++ {
			fieldValue := value.Field(i)
			cadenceType, cadenceVal, err := ReflectToCadence(fieldValue, resolver)
			if err != nil {
				return nil, nil, err
			}

			field := inputType.Field(i)

			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				return nil, nil, err
			}

			name := ""
			tag, err := tags.Get("cadence")
			if err != nil {
				tag, _ = tags.Get("json")
			}
			if tag != nil {
				name = tag.Name
			}

			if name == "-" {
				continue
			}

			if name == "" {
				name = strings.ToLower(field.Name)
			}

			fields = append(fields, cadence.Field{
				Identifier: name,
				Type:       cadenceType,
			})

			val = append(val, cadenceVal)
		}

		resolvedIdentifier, err := resolver(inputType.Name())
		if err != nil {
			return nil, nil, err
		}
		structType := cadence.StructType{
			QualifiedIdentifier: resolvedIdentifier,
			Fields:              fields,
		}

		structValue := cadence.NewStruct(val).WithType(&structType)
		return structValue.Type(), structValue, nil

	case reflect.Pointer:
		if value.IsNil() {
			return cadence.OptionalType{}, cadence.NewOptional(nil), nil
		}

		ptrType, ptrValue, err := ReflectToCadence(value.Elem(), resolver)
		if err != nil {
			return nil, nil, err
		}
		return cadence.OptionalType{Type: ptrType}, cadence.NewOptional(ptrValue), nil

	case reflect.Int:
		value := cadence.NewInt(value.Interface().(int))
		return value.Type(), value, nil
	case reflect.Int8:
		value := cadence.NewInt8(value.Interface().(int8))
		return value.Type(), value, nil
	case reflect.Int16:
		value := cadence.NewInt16(value.Interface().(int16))
		return value.Type(), value, nil
	case reflect.Int32:
		value := cadence.NewInt32(value.Interface().(int32))
		return value.Type(), value, nil
	case reflect.Int64:
		value := cadence.NewInt64(value.Interface().(int64))
		return value.Type(), value, nil
	case reflect.Bool:
		value := cadence.NewBool(value.Interface().(bool))
		return value.Type(), value, nil
	case reflect.Uint:
		value := cadence.NewUInt(value.Interface().(uint))
		return value.Type(), value, nil
	case reflect.Uint8:
		value := cadence.NewUInt8(value.Interface().(uint8))
		return value.Type(), value, nil
	case reflect.Uint16:
		value := cadence.NewUInt16(value.Interface().(uint16))
		return value.Type(), value, nil
	case reflect.Uint32:
		value := cadence.NewUInt32(value.Interface().(uint32))
		return value.Type(), value, nil
	case reflect.Uint64:
		value := cadence.NewUInt64(value.Interface().(uint64))
		return value.Type(), value, nil
	case reflect.String:
		result, err := cadence.NewString(value.Interface().(string))
		return result.Type(), result, err
	case reflect.Float64:
		result, err := cadence.NewUFix64(fmt.Sprintf("%f", value.Interface().(float64)))
		return result.Type(), result, err

	case reflect.Map:
		array := []cadence.KeyValuePair{}
		iter := value.MapRange()

		for iter.Next() {
			key := iter.Key()
			val := iter.Value()
			_, cadenceKey, err := ReflectToCadence(key, resolver)
			if err != nil {
				return nil, nil, err
			}
			_, cadenceVal, err := ReflectToCadence(val, resolver)
			if err != nil {
				return nil, nil, err
			}
			array = append(array, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceVal})
		}
		value := cadence.NewDictionary(array)
		return primitiveReflectTypeToCadenceype(inputType), value, nil
	case reflect.Slice, reflect.Array:
		array := []cadence.Value{}
		for i := 0; i < value.Len(); i++ {
			arrValue := value.Index(i)
			_, cadenceVal, err := ReflectToCadence(arrValue, resolver)
			if err != nil {
				return nil, nil, err
			}
			array = append(array, cadenceVal)
		}
		value := cadence.NewArray(array)
		return primitiveReflectTypeToCadenceype(inputType), value, nil

	}

	return nil, nil, fmt.Errorf("Not supported type for now. Type : %s", inputType.Kind())
}
