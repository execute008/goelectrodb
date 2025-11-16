package electrodb

import (
	"fmt"
	"strconv"
	"strings"
)

// ApplyPadding applies padding to attributes that have PaddingConfig
// This is used to zero-pad numbers for proper string sorting
func ApplyPadding(item Item, schema *Schema) Item {
	result := make(Item)
	for k, v := range item {
		result[k] = v
	}

	// Apply padding to each attribute that has padding config
	for attrName, attrDef := range schema.Attributes {
		if attrDef.Padding == nil {
			continue
		}

		value, exists := result[attrName]
		if !exists {
			continue
		}

		// Apply padding
		padded := padValue(value, attrDef.Padding)
		if padded != nil {
			result[attrName] = padded
		}
	}

	return result
}

// RemovePadding removes padding from attributes that have PaddingConfig
// This is used when reading from DynamoDB
func RemovePadding(item Item, schema *Schema) Item {
	if item == nil {
		return nil
	}

	result := make(Item)
	for k, v := range item {
		result[k] = v
	}

	// Remove padding from each attribute that has padding config
	for attrName, attrDef := range schema.Attributes {
		if attrDef.Padding == nil {
			continue
		}

		value, exists := result[attrName]
		if !exists {
			continue
		}

		// Remove padding
		unpadded := unpadValue(value, attrDef.Padding)
		if unpadded != nil {
			result[attrName] = unpadded
		}
	}

	return result
}

// padValue pads a single value according to padding config
func padValue(value interface{}, padding *PaddingConfig) interface{} {
	if padding == nil || padding.Length == 0 {
		return value
	}

	// Default padding character is "0"
	padChar := padding.Char
	if padChar == "" {
		padChar = "0"
	}

	// Convert value to string
	var strValue string
	switch v := value.(type) {
	case int:
		strValue = strconv.Itoa(v)
	case int64:
		strValue = strconv.FormatInt(v, 10)
	case float64:
		// For floats, convert to int first
		strValue = strconv.FormatInt(int64(v), 10)
	case string:
		strValue = v
	default:
		// For other types, convert to string using fmt
		strValue = fmt.Sprintf("%v", v)
	}

	// Pad the string
	if len(strValue) < padding.Length {
		padCount := padding.Length - len(strValue)
		padded := strings.Repeat(padChar, padCount) + strValue
		return padded
	}

	return strValue
}

// unpadValue removes padding from a value
func unpadValue(value interface{}, padding *PaddingConfig) interface{} {
	if padding == nil {
		return value
	}

	// Default padding character is "0"
	padChar := padding.Char
	if padChar == "" {
		padChar = "0"
	}

	strValue, ok := value.(string)
	if !ok {
		// If it's not a string, return as-is
		return value
	}

	// Remove leading padding characters
	unpadded := strings.TrimLeft(strValue, padChar)

	// If the entire string was padding, return "0"
	if unpadded == "" {
		unpadded = "0"
	}

	// Try to convert back to number if it looks like one
	if intVal, err := strconv.ParseInt(unpadded, 10, 64); err == nil {
		// Return as int64 to match common numeric types
		return intVal
	}

	// Return as string if not a number
	return unpadded
}
