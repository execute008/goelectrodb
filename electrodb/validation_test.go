package electrodb

import (
	"errors"
	"strings"
	"testing"
)

// Test validation functions
func TestValidationFunction(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"email": {
				Type:     AttributeTypeString,
				Required: true,
				Validate: func(value interface{}) error {
					email, ok := value.(string)
					if !ok {
						return errors.New("email must be a string")
					}
					if !strings.Contains(email, "@") {
						return errors.New("email must contain @")
					}
					return nil
				},
			},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// Valid email should pass
	validItem := Item{
		"id":    "123",
		"email": "test@example.com",
	}

	_, err = validator.ValidateAndTransformForWrite(validItem, false)
	if err != nil {
		t.Errorf("Valid email should pass validation, got error: %v", err)
	}

	// Invalid email should fail
	invalidItem := Item{
		"id":    "123",
		"email": "invalid-email",
	}

	_, err = validator.ValidateAndTransformForWrite(invalidItem, false)
	if err == nil {
		t.Error("Invalid email should fail validation")
	}
	if err != nil && !strings.Contains(err.Error(), "email must contain @") {
		t.Errorf("Expected validation error about @, got: %v", err)
	}
}

// Test enum validation
func TestEnumValidation(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
			"status": {
				Type:       AttributeTypeEnum,
				Required:   true,
				EnumValues: []interface{}{"active", "inactive", "pending"},
			},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// Valid enum value should pass
	validItem := Item{
		"id":     "123",
		"status": "active",
	}

	_, err = validator.ValidateAndTransformForWrite(validItem, false)
	if err != nil {
		t.Errorf("Valid enum should pass validation, got error: %v", err)
	}

	// Invalid enum value should fail
	invalidItem := Item{
		"id":     "123",
		"status": "invalid",
	}

	_, err = validator.ValidateAndTransformForWrite(invalidItem, false)
	if err == nil {
		t.Error("Invalid enum should fail validation")
	}
	if err != nil {
		electroErr, ok := err.(*ElectroError)
		if !ok || electroErr.Code != "InvalidEnumValue" {
			t.Errorf("Expected InvalidEnumValue error, got: %v", err)
		}
	}
}

// Test Get/Set transformations
func TestGetSetTransformations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
			"price": {
				Type:     AttributeTypeNumber,
				Required: true,
				// Set: multiply by 100 to store as cents
				Set: func(value interface{}) interface{} {
					if price, ok := value.(float64); ok {
						return int(price * 100)
					}
					return value
				},
				// Get: divide by 100 to return as dollars
				Get: func(value interface{}) interface{} {
					if cents, ok := value.(int); ok {
						return float64(cents) / 100.0
					}
					if cents, ok := value.(float64); ok {
						return cents / 100.0
					}
					return value
				},
			},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// Test Set transformation (before writing)
	item := Item{
		"id":    "123",
		"price": 25.00, // dollars
	}

	transformedItem, err := validator.ValidateAndTransformForWrite(item, false)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Should be stored as 2500 cents
	if transformedItem["price"] != 2500 {
		t.Errorf("Expected price to be transformed to 2500 cents, got %v", transformedItem["price"])
	}

	// Test Get transformation (after reading)
	storedItem := Item{
		"id":    "123",
		"price": 2500, // stored as cents
	}

	readItem := validator.TransformForRead(storedItem)

	// Should be returned as 25.00 dollars
	if readItem["price"] != 25.00 {
		t.Errorf("Expected price to be transformed to 25.00 dollars, got %v", readItem["price"])
	}
}

// Test ReadOnly enforcement
func TestReadOnlyEnforcement(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":        {Type: AttributeTypeString, Required: true},
			"createdAt": {Type: AttributeTypeNumber, ReadOnly: true},
			"name":      {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// ReadOnly should be allowed on create (isUpdate = false)
	createItem := Item{
		"id":        "123",
		"name":      "Test",
		"createdAt": 1234567890,
	}

	_, err = validator.ValidateAndTransformForWrite(createItem, false)
	if err != nil {
		t.Errorf("ReadOnly should be allowed on create, got error: %v", err)
	}

	// ReadOnly should be blocked on update (isUpdate = true)
	updateItem := Item{
		"id":        "123",
		"createdAt": 9999999999,
	}

	_, err = validator.ValidateAndTransformForWrite(updateItem, true)
	if err == nil {
		t.Error("ReadOnly should be blocked on update")
	}
	if err != nil {
		electroErr, ok := err.(*ElectroError)
		if !ok || electroErr.Code != "ReadOnlyViolation" {
			t.Errorf("Expected ReadOnlyViolation error, got: %v", err)
		}
	}
}

// Test ReadOnly enforcement in update operations
func TestReadOnlyInUpdateOperations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":        {Type: AttributeTypeString, Required: true},
			"createdAt": {Type: AttributeTypeNumber, ReadOnly: true},
			"counter":   {Type: AttributeTypeNumber, ReadOnly: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// Test SET operation on readonly
	setOps := map[string]interface{}{
		"createdAt": 1234567890,
	}
	err = validator.ValidateUpdateOperations(setOps, nil, nil, nil)
	if err == nil {
		t.Error("SET on readonly should fail")
	}

	// Test ADD operation on readonly
	addOps := map[string]interface{}{
		"counter": 1,
	}
	err = validator.ValidateUpdateOperations(nil, addOps, nil, nil)
	if err == nil {
		t.Error("ADD on readonly should fail")
	}

	// Test REMOVE operation on readonly
	remOps := []string{"createdAt"}
	err = validator.ValidateUpdateOperations(nil, nil, nil, remOps)
	if err == nil {
		t.Error("REMOVE on readonly should fail")
	}
}

// Test Hidden attribute filtering
func TestHiddenAttributes(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":       {Type: AttributeTypeString, Required: true},
			"name":     {Type: AttributeTypeString, Required: true},
			"password": {Type: AttributeTypeString, Hidden: true},
			"apiKey":   {Type: AttributeTypeString, Hidden: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// Item with hidden attributes
	item := Item{
		"id":       "123",
		"name":     "John Doe",
		"password": "secret123",
		"apiKey":   "key-123",
	}

	// Transform for read should filter hidden attributes
	readItem := validator.TransformForRead(item)

	if _, exists := readItem["password"]; exists {
		t.Error("Hidden attribute 'password' should be filtered out")
	}

	if _, exists := readItem["apiKey"]; exists {
		t.Error("Hidden attribute 'apiKey' should be filtered out")
	}

	if readItem["id"] != "123" {
		t.Error("Non-hidden attribute 'id' should be present")
	}

	if readItem["name"] != "John Doe" {
		t.Error("Non-hidden attribute 'name' should be present")
	}
}

// Test combined transformations and validations
func TestCombinedFeatures(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "User",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
			"email": {
				Type:     AttributeTypeString,
				Required: true,
				Validate: func(value interface{}) error {
					email, ok := value.(string)
					if !ok {
						return errors.New("email must be a string")
					}
					if !strings.Contains(email, "@") {
						return errors.New("email must contain @")
					}
					return nil
				},
				// Normalize email to lowercase
				Set: func(value interface{}) interface{} {
					if email, ok := value.(string); ok {
						return strings.ToLower(email)
					}
					return value
				},
			},
			"status": {
				Type:       AttributeTypeEnum,
				EnumValues: []interface{}{"active", "inactive"},
				Default: func() interface{} {
					return "active"
				},
			},
			"password": {
				Type:   AttributeTypeString,
				Hidden: true,
			},
			"createdAt": {
				Type:     AttributeTypeNumber,
				ReadOnly: true,
			},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// Test write transformation with all features
	item := Item{
		"id":        "123",
		"email":     "John@Example.COM", // should be normalized to lowercase
		"password":  "secret123",
		"createdAt": 1234567890,
		// status not provided, should get default value
	}

	transformedItem, err := validator.ValidateAndTransformForWrite(item, false)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Email should be normalized
	if transformedItem["email"] != "john@example.com" {
		t.Errorf("Expected email to be normalized to lowercase, got %v", transformedItem["email"])
	}

	// Now simulate reading the item back
	storedItem := Item{
		"id":        "123",
		"email":     "john@example.com",
		"password":  "secret123",
		"status":    "active",
		"createdAt": 1234567890,
	}

	readItem := validator.TransformForRead(storedItem)

	// Password should be filtered (hidden)
	if _, exists := readItem["password"]; exists {
		t.Error("Hidden attribute 'password' should be filtered out")
	}

	// Other attributes should be present
	if readItem["email"] != "john@example.com" {
		t.Error("Email should be present")
	}
	if readItem["status"] != "active" {
		t.Error("Status should be present")
	}
}

// Test default values application
func TestDefaultValues(t *testing.T) {
	defaultCalled := false
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
			"status": {
				Type: AttributeTypeString,
				Default: func() interface{} {
					defaultCalled = true
					return "pending"
				},
			},
			"priority": {
				Type: AttributeTypeNumber,
				Default: func() interface{} {
					return 0
				},
			},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	builder := NewParamsBuilder(entity)

	// Item without default attributes
	item := Item{
		"id": "123",
	}

	// Apply defaults
	enrichedItem := builder.applyDefaults(item)

	if !defaultCalled {
		t.Error("Default function should have been called")
	}

	if enrichedItem["status"] != "pending" {
		t.Errorf("Expected status to have default value 'pending', got %v", enrichedItem["status"])
	}

	if enrichedItem["priority"] != 0 {
		t.Errorf("Expected priority to have default value 0, got %v", enrichedItem["priority"])
	}

	// Item with explicitly set values should not use defaults
	item2 := Item{
		"id":       "456",
		"status":   "active",
		"priority": 5,
	}

	enrichedItem2 := builder.applyDefaults(item2)

	if enrichedItem2["status"] != "active" {
		t.Error("Explicit value should not be overridden by default")
	}

	if enrichedItem2["priority"] != 5 {
		t.Error("Explicit value should not be overridden by default")
	}
}

// Test validation with Set transformations in update operations
func TestUpdateOperationsWithTransformations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
			"name": {
				Type: AttributeTypeString,
				Set: func(value interface{}) interface{} {
					if str, ok := value.(string); ok {
						return strings.ToUpper(str)
					}
					return value
				},
			},
			"status": {
				Type:       AttributeTypeEnum,
				EnumValues: []interface{}{"active", "inactive"},
			},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	validator := NewValidator(entity)

	// Test SET operations with transformations
	setOps := map[string]interface{}{
		"name":   "john doe",
		"status": "active",
	}

	transformedSet, _, _ := validator.ApplySetTransformations(setOps, nil, nil)

	// Name should be transformed to uppercase
	if transformedSet["name"] != "JOHN DOE" {
		t.Errorf("Expected name to be transformed to uppercase, got %v", transformedSet["name"])
	}

	// Status should pass through (valid enum)
	if transformedSet["status"] != "active" {
		t.Errorf("Expected status to be 'active', got %v", transformedSet["status"])
	}
}
