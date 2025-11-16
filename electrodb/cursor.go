package electrodb

import (
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// encodeCursor converts a DynamoDB LastEvaluatedKey to a base64-encoded cursor string
func encodeCursor(lastKey map[string]types.AttributeValue) (string, error) {
	if lastKey == nil || len(lastKey) == 0 {
		return "", nil
	}

	// Convert the LastEvaluatedKey to a JSON-serializable format
	cursorData := make(map[string]interface{})
	for key, value := range lastKey {
		cursorData[key] = attributeValueToInterface(value)
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(cursorData)
	if err != nil {
		return "", NewElectroError("CursorEncodingError", "Failed to encode cursor", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return encoded, nil
}

// decodeCursor converts a base64-encoded cursor string back to a DynamoDB ExclusiveStartKey
func decodeCursor(cursor string) (map[string]types.AttributeValue, error) {
	if cursor == "" {
		return nil, nil
	}

	// Decode from base64
	jsonBytes, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, NewElectroError("CursorDecodingError", "Failed to decode cursor", err)
	}

	// Unmarshal from JSON
	var cursorData map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &cursorData); err != nil {
		return nil, NewElectroError("CursorDecodingError", "Failed to unmarshal cursor", err)
	}

	// Convert back to DynamoDB attribute values
	exclusiveStartKey := make(map[string]types.AttributeValue)
	for key, value := range cursorData {
		attrValue, err := interfaceToAttributeValue(value)
		if err != nil {
			return nil, err
		}
		exclusiveStartKey[key] = attrValue
	}

	return exclusiveStartKey, nil
}

// attributeValueToInterface converts a DynamoDB AttributeValue to a Go interface{}
// for JSON serialization
func attributeValueToInterface(av types.AttributeValue) interface{} {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		return map[string]string{"S": v.Value}
	case *types.AttributeValueMemberN:
		return map[string]string{"N": v.Value}
	case *types.AttributeValueMemberB:
		return map[string][]byte{"B": v.Value}
	case *types.AttributeValueMemberBOOL:
		return map[string]bool{"BOOL": v.Value}
	case *types.AttributeValueMemberNULL:
		return map[string]bool{"NULL": v.Value}
	case *types.AttributeValueMemberSS:
		return map[string][]string{"SS": v.Value}
	case *types.AttributeValueMemberNS:
		return map[string][]string{"NS": v.Value}
	case *types.AttributeValueMemberBS:
		return map[string][][]byte{"BS": v.Value}
	case *types.AttributeValueMemberM:
		m := make(map[string]interface{})
		for k, val := range v.Value {
			m[k] = attributeValueToInterface(val)
		}
		return map[string]map[string]interface{}{"M": m}
	case *types.AttributeValueMemberL:
		l := make([]interface{}, len(v.Value))
		for i, val := range v.Value {
			l[i] = attributeValueToInterface(val)
		}
		return map[string][]interface{}{"L": l}
	default:
		return nil
	}
}

// interfaceToAttributeValue converts a Go interface{} back to a DynamoDB AttributeValue
func interfaceToAttributeValue(val interface{}) (types.AttributeValue, error) {
	// The value should be a map with a single key indicating the type
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, NewElectroError("CursorDecodingError", "Invalid cursor format", nil)
	}

	// Check for each possible type
	if s, ok := m["S"].(string); ok {
		return &types.AttributeValueMemberS{Value: s}, nil
	}
	if n, ok := m["N"].(string); ok {
		return &types.AttributeValueMemberN{Value: n}, nil
	}
	if b, ok := m["B"].([]byte); ok {
		return &types.AttributeValueMemberB{Value: b}, nil
	}
	if b, ok := m["BOOL"].(bool); ok {
		return &types.AttributeValueMemberBOOL{Value: b}, nil
	}
	if n, ok := m["NULL"].(bool); ok {
		return &types.AttributeValueMemberNULL{Value: n}, nil
	}
	if ss, ok := m["SS"].([]interface{}); ok {
		strSlice := make([]string, len(ss))
		for i, v := range ss {
			if str, ok := v.(string); ok {
				strSlice[i] = str
			}
		}
		return &types.AttributeValueMemberSS{Value: strSlice}, nil
	}
	if ns, ok := m["NS"].([]interface{}); ok {
		strSlice := make([]string, len(ns))
		for i, v := range ns {
			if str, ok := v.(string); ok {
				strSlice[i] = str
			}
		}
		return &types.AttributeValueMemberNS{Value: strSlice}, nil
	}
	if mapVal, ok := m["M"].(map[string]interface{}); ok {
		attrMap := make(map[string]types.AttributeValue)
		for k, v := range mapVal {
			attrValue, err := interfaceToAttributeValue(v)
			if err != nil {
				return nil, err
			}
			attrMap[k] = attrValue
		}
		return &types.AttributeValueMemberM{Value: attrMap}, nil
	}
	if list, ok := m["L"].([]interface{}); ok {
		attrList := make([]types.AttributeValue, len(list))
		for i, v := range list {
			attrValue, err := interfaceToAttributeValue(v)
			if err != nil {
				return nil, err
			}
			attrList[i] = attrValue
		}
		return &types.AttributeValueMemberL{Value: attrList}, nil
	}

	return nil, NewElectroError("CursorDecodingError", "Unknown attribute value type", nil)
}
