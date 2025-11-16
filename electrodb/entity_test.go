package electrodb

import (
	"testing"
)

// TestNewEntity tests basic entity creation
func TestNewEntity(t *testing.T) {
	schema := &Schema{
		Service: "MallStoreDirectory",
		Entity:  "MallStores",
		Table:   "StoreDirectory",
		Version: "1",
		Attributes: map[string]*AttributeDefinition{
			"id": {
				Type:     AttributeTypeString,
				Required: false,
				Field:    "storeLocationId",
			},
			"mall": {
				Type:     AttributeTypeString,
				Required: true,
				Field:    "mall",
			},
			"store": {
				Type:     AttributeTypeString,
				Required: true,
				Field:    "storeId",
			},
			"building": {
				Type:     AttributeTypeString,
				Required: true,
				Field:    "buildingId",
			},
			"unit": {
				Type:     AttributeTypeString,
				Required: true,
				Field:    "unitId",
			},
		},
		Indexes: map[string]*IndexDefinition{
			"store": {
				PK: FacetDefinition{
					Field:  "pk",
					Facets: []string{"id"},
				},
			},
			"units": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK: FacetDefinition{
					Field:  "gsi1pk",
					Facets: []string{"mall"},
				},
				SK: &FacetDefinition{
					Field:  "gsi1sk",
					Facets: []string{"building", "unit", "store"},
				},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	if entity == nil {
		t.Fatal("Entity is nil")
	}

	if entity.schema.Service != "MallStoreDirectory" {
		t.Errorf("Expected service 'MallStoreDirectory', got '%s'", entity.schema.Service)
	}

	if entity.schema.Entity != "MallStores" {
		t.Errorf("Expected entity 'MallStores', got '%s'", entity.schema.Entity)
	}
}

// TestNewEntityWithNilSchema tests entity creation with nil schema
func TestNewEntityWithNilSchema(t *testing.T) {
	entity, err := NewEntity(nil, nil)
	if err == nil {
		t.Fatal("Expected error when creating entity with nil schema")
	}

	if entity != nil {
		t.Fatal("Expected entity to be nil")
	}

	electroErr, ok := err.(*ElectroError)
	if !ok {
		t.Fatal("Expected ElectroError type")
	}

	if electroErr.Code != "InvalidSchema" {
		t.Errorf("Expected error code 'InvalidSchema', got '%s'", electroErr.Code)
	}
}

// TestSchemaValidation tests schema validation
func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name        string
		schema      *Schema
		expectError bool
		errorCode   string
	}{
		{
			name: "missing service name",
			schema: &Schema{
				Entity: "TestEntity",
				Table:  "TestTable",
				Attributes: map[string]*AttributeDefinition{
					"id": {Type: AttributeTypeString},
				},
				Indexes: map[string]*IndexDefinition{
					"primary": {
						PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
					},
				},
			},
			expectError: true,
			errorCode:   "InvalidSchema",
		},
		{
			name: "missing entity name",
			schema: &Schema{
				Service: "TestService",
				Table:   "TestTable",
				Attributes: map[string]*AttributeDefinition{
					"id": {Type: AttributeTypeString},
				},
				Indexes: map[string]*IndexDefinition{
					"primary": {
						PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
					},
				},
			},
			expectError: true,
			errorCode:   "InvalidSchema",
		},
		{
			name: "missing table name",
			schema: &Schema{
				Service: "TestService",
				Entity:  "TestEntity",
				Attributes: map[string]*AttributeDefinition{
					"id": {Type: AttributeTypeString},
				},
				Indexes: map[string]*IndexDefinition{
					"primary": {
						PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
					},
				},
			},
			expectError: true,
			errorCode:   "InvalidSchema",
		},
		{
			name: "no attributes",
			schema: &Schema{
				Service:    "TestService",
				Entity:     "TestEntity",
				Table:      "TestTable",
				Attributes: map[string]*AttributeDefinition{},
				Indexes: map[string]*IndexDefinition{
					"primary": {
						PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
					},
				},
			},
			expectError: true,
			errorCode:   "InvalidSchema",
		},
		{
			name: "no indexes",
			schema: &Schema{
				Service: "TestService",
				Entity:  "TestEntity",
				Table:   "TestTable",
				Attributes: map[string]*AttributeDefinition{
					"id": {Type: AttributeTypeString},
				},
				Indexes: map[string]*IndexDefinition{},
			},
			expectError: true,
			errorCode:   "InvalidSchema",
		},
		{
			name: "invalid facet reference",
			schema: &Schema{
				Service: "TestService",
				Entity:  "TestEntity",
				Table:   "TestTable",
				Attributes: map[string]*AttributeDefinition{
					"id": {Type: AttributeTypeString},
				},
				Indexes: map[string]*IndexDefinition{
					"primary": {
						PK: FacetDefinition{Field: "pk", Facets: []string{"nonexistent"}},
					},
				},
			},
			expectError: true,
			errorCode:   "InvalidSchema",
		},
		{
			name: "valid schema",
			schema: &Schema{
				Service: "TestService",
				Entity:  "TestEntity",
				Table:   "TestTable",
				Attributes: map[string]*AttributeDefinition{
					"id":   {Type: AttributeTypeString},
					"name": {Type: AttributeTypeString},
				},
				Indexes: map[string]*IndexDefinition{
					"primary": {
						PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
						SK: &FacetDefinition{Field: "sk", Facets: []string{"name"}},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := NewEntity(tt.schema, nil)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error for test '%s', but got none", tt.name)
				}

				electroErr, ok := err.(*ElectroError)
				if !ok {
					t.Fatalf("Expected ElectroError type for test '%s'", tt.name)
				}

				if electroErr.Code != tt.errorCode {
					t.Errorf("Expected error code '%s', got '%s'", tt.errorCode, electroErr.Code)
				}

				if entity != nil {
					t.Errorf("Expected entity to be nil for test '%s'", tt.name)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for test '%s': %v", tt.name, err)
				}

				if entity == nil {
					t.Fatalf("Expected entity to be non-nil for test '%s'", tt.name)
				}
			}
		})
	}
}

// TestEntityPutWithoutClient tests that put operations fail when no client is provided
func TestEntityPutWithoutClient(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"name":  {Type: AttributeTypeString, Required: true},
			"email": {Type: AttributeTypeString, Required: false},
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

	// Attempt to put without a client
	item := Item{
		"id":    "123",
		"name":  "Test User",
		"email": "test@example.com",
	}

	_, err = entity.Put(item).Go()
	if err == nil {
		t.Fatal("Expected error when executing put without client")
	}

	electroErr, ok := err.(*ElectroError)
	if !ok {
		t.Fatal("Expected ElectroError type")
	}

	if electroErr.Code != "NoClientProvided" {
		t.Errorf("Expected error code 'NoClientProvided', got '%s'", electroErr.Code)
	}
}

// TestEntityGet tests get operation creation
func TestEntityGet(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
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

	keys := Keys{"id": "123"}
	getOp := entity.Get(keys)

	if getOp == nil {
		t.Fatal("Expected get operation to be non-nil")
	}

	// Test that params can be generated without error
	params, err := getOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}
}

// TestEntityUpdate tests update operation creation
func TestEntityUpdate(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"count": {Type: AttributeTypeNumber, Required: false},
			"name":  {Type: AttributeTypeString, Required: false},
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

	keys := Keys{"id": "123"}
	updateOp := entity.Update(keys).Set(map[string]interface{}{
		"name": "New Name",
	}).Add(map[string]interface{}{
		"count": 1,
	})

	if updateOp == nil {
		t.Fatal("Expected update operation to be non-nil")
	}

	// Test that params can be generated without error
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}
}

// TestEntityDelete tests delete operation creation
func TestEntityDelete(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
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

	keys := Keys{"id": "123"}
	deleteOp := entity.Delete(keys)

	if deleteOp == nil {
		t.Fatal("Expected delete operation to be non-nil")
	}

	// Test that params can be generated without error
	params, err := deleteOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}
}
