package overflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
func StructToCadence(qualifiedIdentifier string, t interface{}) (cadence.Value, error) {
	var val []cadence.Value

	s := reflect.ValueOf(t)
	if s.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct")
	}
	typeOfT := s.Type()
	fields := []cadence.Field{}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		cadenceType, cadenceVal, err := ReflectToCadence(f)
		if err != nil {
			return nil, err
		}

		field := typeOfT.Field(i)
		name := field.Tag.Get("cadence")
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		fields = append(fields, cadence.Field{
			Identifier: name,
			Type:       cadenceType,
		})

		val = append(val, cadenceVal)
	}

	structType := cadence.StructType{
		QualifiedIdentifier: qualifiedIdentifier,
		Fields:              fields,
	}

	value := cadence.NewStruct(val).WithType(&structType)

	return value, nil
}

func InputToCadence(v interface{}) (cadence.Type, cadence.Value, error) {
	f := reflect.ValueOf(v)
	return ReflectToCadence(f)
}

func ReflectToCadence(f reflect.Value) (cadence.Type, cadence.Value, error) {
	inputType := f.Type()

	fmt.Printf("value %v\n", f)
	fmt.Printf("type %v\n", inputType)
	fmt.Printf("name %s\n", inputType.Name())
	fmt.Printf("kind %s\n", inputType.Kind())

	kind := inputType.Kind()
	switch kind {

	/*
		Bool
		Int
		Int8
		Int16
		Int32
		Int64
		Uint
		Uint8
		Uint16
		Uint32
		Uintptr
		Float32
		Pointer
	*/
	case reflect.Uint64:
		return cadence.UInt64Type{}, cadence.NewUInt64(f.Interface().(uint64)), nil

	case reflect.String:
		result, err := cadence.NewString(f.Interface().(string))
		fmt.Println("string")
		return cadence.StringType{}, result, err

	case reflect.Float64:
		result, err := cadence.NewUFix64(fmt.Sprintf("%.2f", f.Interface().(float64)))
		return cadence.UFix64Type{}, result, err

	case reflect.Map:
		array := []cadence.KeyValuePair{}
		var typeKey cadence.Type
		var typeVal cadence.Type
		iter := f.MapRange()

		for iter.Next() {
			key := iter.Key()
			val := iter.Value()
			typ, cadenceKey, err := ReflectToCadence(key)
			typeKey = typ
			if err != nil {
				return nil, nil, err
			}
			typ, cadenceVal, err := ReflectToCadence(val)
			typeVal = typ
			if err != nil {
				return nil, nil, err
			}
			array = append(array, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceVal})
		}
		//we need to return a better type, with key and elements
		return cadence.DictionaryType{
			KeyType:     typeKey,
			ElementType: typeVal,
		}, cadence.NewDictionary(array), nil
	case reflect.Slice, reflect.Array:
		var sliceType cadence.Type
		array := []cadence.Value{}
		for i := 0; i < f.Len(); i++ {
			arrValue := f.Index(i)
			typ, cadenceVal, err := ReflectToCadence(arrValue)
			sliceType = typ
			if err != nil {
				return nil, cadenceVal, err
			}
			array = append(array, cadenceVal)
		}
		return cadence.VariableSizedArrayType{ElementType: sliceType}, cadence.NewArray(array), nil

	}

	return nil, nil, fmt.Errorf("Not supported type for now. Type : %s", inputType.Kind())
}
