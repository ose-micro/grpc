package grpc

import (
	"fmt"
	"log"

	"google.golang.org/protobuf/types/known/structpb"
)

func ConvertToStruct(data map[string]interface{}) (*structpb.Struct, error) {
	fields, err := ConvertToValue(data)
	if err != nil {
		return nil, err
	}
	return &structpb.Struct{Fields: fields}, nil
}

func ConvertToValue(data map[string]interface{}) (map[string]*structpb.Value, error) {
	result := make(map[string]*structpb.Value)
	for key, v := range data {
		val, err := ToProtoValue(v)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", key, err)
		}
		result[key] = val
	}
	return result, nil
}
func ToProtoValue(v interface{}) (*structpb.Value, error) {
	switch val := v.(type) {
	case string:
		return structpb.NewStringValue(val), nil
	case int:
		return structpb.NewNumberValue(float64(val)), nil
	case int32:
		return structpb.NewNumberValue(float64(val)), nil
	case int64:
		return structpb.NewNumberValue(float64(val)), nil
	case float32:
		return structpb.NewNumberValue(float64(val)), nil
	case float64:
		return structpb.NewNumberValue(val), nil
	case bool:
		return structpb.NewBoolValue(val), nil
	case []interface{}:
		list := make([]*structpb.Value, len(val))
		for i, elem := range val {
			pv, err := ToProtoValue(elem)
			if err != nil {
				return nil, err
			}
			list[i] = pv
		}
		return structpb.NewListValue(&structpb.ListValue{Values: list}), nil

	case []map[string]interface{}:
		// Convert slice of maps to list of struct values
		list := make([]*structpb.Value, len(val))
		for i, elem := range val {
			structVal, err := ConvertToStruct(elem)
			if err != nil {
				return nil, err
			}
			list[i] = structpb.NewStructValue(structVal)
		}
		return structpb.NewListValue(&structpb.ListValue{Values: list}), nil

	case map[string]interface{}:
		fields, err := ConvertToValue(val)
		if err != nil {
			return nil, err
		}
		return structpb.NewStructValue(&structpb.Struct{Fields: fields}), nil

	case nil:
		return structpb.NewNullValue(), nil

	default:
		log.Printf("Unsupported type %T with value %v\n", val, val)
		return nil, fmt.Errorf("unsupported type %T", val)
	}
}
