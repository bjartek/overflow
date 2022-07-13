package overflow

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/onflow/cadence"
)

// TODO: consider submitting this to cadence
// CadenceValueToJsonString converts a cadence.Value into a json pretty printed string
//Deprecated use CadenceValueToJsonStringCompact
func CadenceValueToJsonString(value cadence.Value) string {
	if value == nil {
		return "{}"
	}

	result := CadenceValueToInterface(value)
	j, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	return string(j)
}

// CadenceValueToJsonString converts a cadence.Value into a json pretty printed string
func CadenceValueToJsonStringCompact(value cadence.Value) string {
	result := CadenceValueToInterfaceCompact(value)
	if result == nil {
		return ""
	}
	j, err := json.MarshalIndent(result, "", "  ")

	if err != nil {
		panic(err)
	}

	return string(j)
}

// CadenceValueToInterface convert a candence.Value into interface{}
//Deprecated use CadenceValueToInterfaceCompact
func CadenceValueToInterface(field cadence.Value) interface{} {
	if field == nil {
		return ""
	}

	switch field := field.(type) {
	case cadence.Optional:
		return CadenceValueToInterface(field.Value)
	case cadence.Dictionary:
		result := map[string]interface{}{}
		for _, item := range field.Pairs {
			key, err := strconv.Unquote(item.Key.String())
			if err != nil {
				result[item.Key.String()] = CadenceValueToInterface(item.Value)
				continue
			}

			result[key] = CadenceValueToInterface(item.Value)
		}
		return result
	case cadence.Struct:
		result := map[string]interface{}{}
		subStructNames := field.StructType.Fields

		for j, subField := range field.Fields {
			result[subStructNames[j].Identifier] = CadenceValueToInterface(subField)
		}
		return result
	case cadence.Array:
		var result []interface{}
		for _, item := range field.Values {
			result = append(result, CadenceValueToInterface(item))
		}
		return result

	default:
		result, err := strconv.Unquote(field.String())
		if err != nil {
			return field.String()
		}
		return result
	}
}

// CadenceValueToInterface convert a candence.Value into interface{}
func CadenceValueToInterfaceCompact(field cadence.Value) interface{} {
	if field == nil {
		return nil
	}

	switch field := field.(type) {
	case cadence.Optional:
		return CadenceValueToInterfaceCompact(field.Value)
	case cadence.Dictionary:
		//fmt.Println("is dict ", field.ToGoValue(), " ", field.String())
		result := map[string]interface{}{}
		for _, item := range field.Pairs {
			value := CadenceValueToInterfaceCompact(item.Value)
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
			value := CadenceValueToInterfaceCompact(subField)
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
			value := CadenceValueToInterfaceCompact(item)
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
