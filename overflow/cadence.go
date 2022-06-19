package overflow

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/onflow/cadence"
)

// CadenceValueToJsonString converts a cadence.Value into a json pretty printed string
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
		result := map[string]interface{}{}
		for _, item := range field.Pairs {
			value := CadenceValueToInterfaceCompact(item.Value)
			key := getAndUnquoteStringAsPointer(item.Key)
			if value != nil && key != nil {
				result[*key] = value
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case cadence.Struct:
		result := map[string]interface{}{}
		subStructNames := field.StructType.Fields

		for j, subField := range field.Fields {
			value := CadenceValueToInterfaceCompact(subField)
			if value != nil {
				result[subStructNames[j].Identifier] = value
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case cadence.Array:
		var result []interface{}
		for _, item := range field.Values {
			value := CadenceValueToInterfaceCompact(item)
			if value != nil {
				result = append(result, value)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result

	case cadence.String:
		value := getAndUnquoteStringAsPointer(field)
		if value == nil {
			return nil
		}
		return *value

	default:
		return field.ToGoValue()
	}
}

func getAndUnquoteStringAsPointer(value cadence.Value) *string {
	result, err := strconv.Unquote(value.String())
	if err != nil {
		result = value.String()
	}

	if result == "" {
		return nil
	}
	return &result
}
