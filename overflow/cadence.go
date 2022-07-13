package overflow

import (
	"encoding/json"
	"github.com/onflow/cadence"
	"log"
	"strconv"
	"strings"
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
			if strings.Contains(field.String(), "\\u") || strings.Contains(field.String(), "\\U") {
				return field.ToGoValue()
			}

			return CadenceValueToGoValue(field)
		}
		return result
	}
}
