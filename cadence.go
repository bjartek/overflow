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

func InputToCadence(v interface{}, resolver InputResolver) (cadence.Value, error) {
	f := reflect.ValueOf(v)
	return ReflectToCadence(f, resolver)
}

func ReflectToCadence(value reflect.Value, resolver InputResolver) (cadence.Value, error) {
	inputType := value.Type()

	kind := inputType.Kind()
	switch kind {
	case reflect.Interface:
		return cadence.NewValue(value.Interface())
	case reflect.Struct:
		var val []cadence.Value
		fields := []cadence.Field{}
		for i := 0; i < value.NumField(); i++ {
			fieldValue := value.Field(i)
			cadenceVal, err := ReflectToCadence(fieldValue, resolver)
			if err != nil {
				return nil, err
			}
			cadenceType := cadenceVal.Type()

			field := inputType.Field(i)

			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				return nil, err
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
			return nil, err
		}
		structType := cadence.StructType{
			QualifiedIdentifier: resolvedIdentifier,
			Fields:              fields,
		}

		structValue := cadence.NewStruct(val).WithType(&structType)
		return structValue, nil

	case reflect.Pointer:
		if value.IsNil() {
			return cadence.NewOptional(nil), nil
		}

		ptrValue, err := ReflectToCadence(value.Elem(), resolver)
		if err != nil {
			return nil, err
		}
		return cadence.NewOptional(ptrValue), nil

	case reflect.Int:
		return cadence.NewInt(value.Interface().(int)), nil
	case reflect.Int8:
		return cadence.NewInt8(value.Interface().(int8)), nil
	case reflect.Int16:
		return cadence.NewInt16(value.Interface().(int16)), nil
	case reflect.Int32:
		return cadence.NewInt32(value.Interface().(int32)), nil
	case reflect.Int64:
		return cadence.NewInt64(value.Interface().(int64)), nil
	case reflect.Bool:
		return cadence.NewBool(value.Interface().(bool)), nil
	case reflect.Uint:
		return cadence.NewUInt(value.Interface().(uint)), nil
	case reflect.Uint8:
		return cadence.NewUInt8(value.Interface().(uint8)), nil
	case reflect.Uint16:
		return cadence.NewUInt16(value.Interface().(uint16)), nil
	case reflect.Uint32:
		return cadence.NewUInt32(value.Interface().(uint32)), nil
	case reflect.Uint64:
		return cadence.NewUInt64(value.Interface().(uint64)), nil
	case reflect.String:
		result, err := cadence.NewString(value.Interface().(string))
		return result, err
	case reflect.Float64:
		result, err := cadence.NewUFix64(fmt.Sprintf("%f", value.Interface().(float64)))
		return result, err

	case reflect.Map:
		array := []cadence.KeyValuePair{}
		iter := value.MapRange()

		for iter.Next() {
			key := iter.Key()
			val := iter.Value()
			cadenceKey, err := ReflectToCadence(key, resolver)
			if err != nil {
				return nil, err
			}
			cadenceVal, err := ReflectToCadence(val, resolver)
			if err != nil {
				return nil, err
			}
			array = append(array, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceVal})
		}
		return cadence.NewDictionary(array), nil
	case reflect.Slice, reflect.Array:
		array := []cadence.Value{}
		for i := 0; i < value.Len(); i++ {
			arrValue := value.Index(i)
			cadenceVal, err := ReflectToCadence(arrValue, resolver)
			if err != nil {
				return nil, err
			}
			array = append(array, cadenceVal)
		}
		return cadence.NewArray(array), nil

	}

	return nil, fmt.Errorf("Not supported type for now. Type : %s", inputType.Kind())
}
