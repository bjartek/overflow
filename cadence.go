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
func MarshalAsCadenceStruct(qualifiedIdentifier string, t interface{}) (cadence.Value, error) {
	var val []cadence.Value

	reflect.TypeOf(t)
	s := reflect.ValueOf(t).Elem()
	if s.Kind() != reflect.Struct {
		panic(s.Kind())
	}
	typeOfT := s.Type()
	fields := []cadence.Field{}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		cadenceType, cadenceVal, err := ParseInput(f)
		if err != nil {
			return nil, err
		}

		fields = append(fields, cadence.Field{
			Identifier: strings.ToLower(typeOfT.Field(i).Name),
			Type:       cadenceType,
		})

		val = append(val, cadenceVal)
	}

	structType := cadence.StructType{
		QualifiedIdentifier: qualifiedIdentifier,
		Fields:              fields,
	}

	value :=
		cadence.NewStruct(val).WithType(&structType)

	return value, nil
}

func ParseInputValue(v interface{}) (cadence.Type, cadence.Value, error) {
	f := reflect.ValueOf(v)
	return ParseInput(f)
}

func ParseInput(f reflect.Value) (cadence.Type, cadence.Value, error) {
	inputType := f.Type()

	fmt.Printf("%v\n", f)
	fmt.Printf("%v\n", inputType)
	fmt.Printf("%s\n", inputType.Name())
	fmt.Printf("%s\n", inputType.Kind())

	switch inputType {

	case reflect.TypeOf(uint64(0)):
		return cadence.UInt64Type{}, cadence.NewUInt64(f.Interface().(uint64)), nil

	case reflect.TypeOf("string"):
		result, err := cadence.NewString(f.Interface().(string))
		fmt.Println("string")
		return cadence.StringType{}, result, err

	case reflect.TypeOf(float64(0.0)):
		result, err := cadence.NewUFix64(fmt.Sprintf("%.2f", f.Interface().(float64)))
		return cadence.UFix64Type{}, result, err

	case reflect.TypeOf(map[interface{}]interface{}{}):
		array := []cadence.KeyValuePair{}
		for key, val := range f.Interface().(map[interface{}]interface{}) {
			//how can key here be string or int or uint64
			reflectKey := reflect.ValueOf(key)
			_, cadenceKey, err := ParseInput(reflectKey)
			if err != nil {
				return nil, nil, err
			}
			reflectVal := reflect.ValueOf(val)
			_, cadenceVal, err := ParseInput(reflectVal)
			if err != nil {
				return nil, nil, err
			}
			array = append(array, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceVal})
		}
		return cadence.DictionaryType{}, cadence.NewDictionary(array), nil

	case reflect.TypeOf(map[string]interface{}{}):
		array := []cadence.KeyValuePair{}
		for key, val := range f.Interface().(map[string]interface{}) {
			reflectKey := reflect.ValueOf(key)
			_, cadenceKey, err := ParseInput(reflectKey)
			if err != nil {
				return nil, nil, err
			}
			reflectVal := reflect.ValueOf(val)
			_, cadenceVal, err := ParseInput(reflectVal)
			if err != nil {
				return nil, nil, err
			}
			array = append(array, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceVal})
		}
		return cadence.DictionaryType{}, cadence.NewDictionary(array), nil

	case reflect.TypeOf([]interface{}{}):
		array := []cadence.Value{}
		for _, val := range f.Interface().([]interface{}) {
			reflectVal := reflect.ValueOf(val)
			_, cadenceVal, err := ParseInput(reflectVal)
			if err != nil {
				return nil, cadenceVal, err
			}
			array = append(array, cadenceVal)
		}
		return cadence.VariableSizedArrayType{}, cadence.NewArray(array), nil

	case reflect.TypeOf([]uint64{}):
		array := []cadence.Value{}
		for _, val := range f.Interface().([]uint64) {
			reflectVal := reflect.ValueOf(val)
			_, cadenceVal, err := ParseInput(reflectVal)
			if err != nil {
				return nil, cadenceVal, err
			}
			array = append(array, cadenceVal)
		}
		return cadence.VariableSizedArrayType{}, cadence.NewArray(array), nil
	}

	panic(fmt.Sprintf("Not supported type for now. Type : %s", f.Type()))
}
