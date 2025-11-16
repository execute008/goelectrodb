package electrodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestEncodeCursor(t *testing.T) {
	// Test with string attribute value
	lastKey := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "$service#entity#id123"},
		"sk": &types.AttributeValueMemberS{Value: "$service#entity#name#john"},
	}

	encoded, err := encodeCursor(lastKey)
	if err != nil {
		t.Fatalf("Failed to encode cursor: %v", err)
	}

	if encoded == "" {
		t.Error("Expected non-empty cursor")
	}
}

func TestEncodeDecodeStringCursor(t *testing.T) {
	// Original last key with string values
	originalKey := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "$service#entity#user123"},
		"sk": &types.AttributeValueMemberS{Value: "$service#entity#email#test@example.com"},
	}

	// Encode
	encoded, err := encodeCursor(originalKey)
	if err != nil {
		t.Fatalf("Failed to encode cursor: %v", err)
	}

	// Decode
	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("Failed to decode cursor: %v", err)
	}

	// Verify both keys exist
	if len(decoded) != 2 {
		t.Errorf("Expected 2 keys in decoded cursor, got %d", len(decoded))
	}

	// Verify pk value
	pkValue, ok := decoded["pk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("Expected pk to be a string attribute")
	}
	if pkValue.Value != "$service#entity#user123" {
		t.Errorf("Expected pk value '$service#entity#user123', got '%s'", pkValue.Value)
	}

	// Verify sk value
	skValue, ok := decoded["sk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("Expected sk to be a string attribute")
	}
	if skValue.Value != "$service#entity#email#test@example.com" {
		t.Errorf("Expected sk value '$service#entity#email#test@example.com', got '%s'", skValue.Value)
	}
}

func TestEncodeDecodeNumberCursor(t *testing.T) {
	// Original last key with number values
	originalKey := map[string]types.AttributeValue{
		"pk":    &types.AttributeValueMemberS{Value: "$service#entity#user123"},
		"score": &types.AttributeValueMemberN{Value: "42.5"},
	}

	// Encode
	encoded, err := encodeCursor(originalKey)
	if err != nil {
		t.Fatalf("Failed to encode cursor: %v", err)
	}

	// Decode
	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("Failed to decode cursor: %v", err)
	}

	// Verify score value
	scoreValue, ok := decoded["score"].(*types.AttributeValueMemberN)
	if !ok {
		t.Fatal("Expected score to be a number attribute")
	}
	if scoreValue.Value != "42.5" {
		t.Errorf("Expected score value '42.5', got '%s'", scoreValue.Value)
	}
}

func TestEncodeDecodeBooleanCursor(t *testing.T) {
	// Original last key with boolean value
	originalKey := map[string]types.AttributeValue{
		"pk":     &types.AttributeValueMemberS{Value: "$service#entity#user123"},
		"active": &types.AttributeValueMemberBOOL{Value: true},
	}

	// Encode
	encoded, err := encodeCursor(originalKey)
	if err != nil {
		t.Fatalf("Failed to encode cursor: %v", err)
	}

	// Decode
	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("Failed to decode cursor: %v", err)
	}

	// Verify active value
	activeValue, ok := decoded["active"].(*types.AttributeValueMemberBOOL)
	if !ok {
		t.Fatal("Expected active to be a boolean attribute")
	}
	if activeValue.Value != true {
		t.Errorf("Expected active value 'true', got '%v'", activeValue.Value)
	}
}

func TestEncodeEmptyCursor(t *testing.T) {
	// Test with nil
	encoded, err := encodeCursor(nil)
	if err != nil {
		t.Fatalf("Failed to encode nil cursor: %v", err)
	}
	if encoded != "" {
		t.Error("Expected empty string for nil cursor")
	}

	// Test with empty map
	encoded, err = encodeCursor(map[string]types.AttributeValue{})
	if err != nil {
		t.Fatalf("Failed to encode empty cursor: %v", err)
	}
	if encoded != "" {
		t.Error("Expected empty string for empty cursor")
	}
}

func TestDecodeEmptyCursor(t *testing.T) {
	decoded, err := decodeCursor("")
	if err != nil {
		t.Fatalf("Failed to decode empty cursor: %v", err)
	}
	if decoded != nil {
		t.Error("Expected nil for empty cursor")
	}
}

func TestDecodeInvalidCursor(t *testing.T) {
	// Test with invalid base64
	_, err := decodeCursor("!!!invalid-base64!!!")
	if err == nil {
		t.Error("Expected error for invalid base64")
	}

	// Test with valid base64 but invalid JSON
	invalidJSON := "bm90IGpzb24=" // "not json" in base64
	_, err = decodeCursor(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestEncodeDecodeComplexCursor(t *testing.T) {
	// Test with multiple attribute types
	originalKey := map[string]types.AttributeValue{
		"pk":       &types.AttributeValueMemberS{Value: "$service#entity#user123"},
		"sk":       &types.AttributeValueMemberS{Value: "$service#entity#timestamp"},
		"score":    &types.AttributeValueMemberN{Value: "100"},
		"active":   &types.AttributeValueMemberBOOL{Value: true},
		"inactive": &types.AttributeValueMemberBOOL{Value: false},
	}

	// Encode
	encoded, err := encodeCursor(originalKey)
	if err != nil {
		t.Fatalf("Failed to encode cursor: %v", err)
	}

	// Decode
	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("Failed to decode cursor: %v", err)
	}

	// Verify all keys exist
	if len(decoded) != 5 {
		t.Errorf("Expected 5 keys in decoded cursor, got %d", len(decoded))
	}

	// Verify each value
	if pk, ok := decoded["pk"].(*types.AttributeValueMemberS); !ok || pk.Value != "$service#entity#user123" {
		t.Error("pk value mismatch")
	}
	if sk, ok := decoded["sk"].(*types.AttributeValueMemberS); !ok || sk.Value != "$service#entity#timestamp" {
		t.Error("sk value mismatch")
	}
	if score, ok := decoded["score"].(*types.AttributeValueMemberN); !ok || score.Value != "100" {
		t.Error("score value mismatch")
	}
	if active, ok := decoded["active"].(*types.AttributeValueMemberBOOL); !ok || active.Value != true {
		t.Error("active value mismatch")
	}
	if inactive, ok := decoded["inactive"].(*types.AttributeValueMemberBOOL); !ok || inactive.Value != false {
		t.Error("inactive value mismatch")
	}
}
