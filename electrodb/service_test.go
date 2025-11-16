package electrodb

import (
	"testing"
)

func TestNewService(t *testing.T) {
	service := NewService("TestService", nil)

	if service == nil {
		t.Fatal("Expected service to be non-nil")
	}

	if service.name != "TestService" {
		t.Errorf("Expected service name 'TestService', got '%s'", service.name)
	}

	if service.entities == nil {
		t.Error("Expected entities map to be initialized")
	}

	if len(service.entities) != 0 {
		t.Errorf("Expected 0 entities, got %d", len(service.entities))
	}
}

func TestServiceJoin(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	// Create first entity
	schema1 := &Schema{
		Service: "TestService",
		Entity:  "User",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"name": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	entity1, err := NewEntity(schema1, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Join first entity
	err = service.Join(entity1)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	if len(service.entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(service.entities))
	}

	// Try to join nil entity
	err = service.Join(nil)
	if err == nil {
		t.Error("Expected error when joining nil entity")
	}

	// Try to join duplicate entity
	err = service.Join(entity1)
	if err == nil {
		t.Error("Expected error when joining duplicate entity")
	}

	electroErr, ok := err.(*ElectroError)
	if !ok {
		t.Fatal("Expected ElectroError type")
	}

	if electroErr.Code != "DuplicateEntity" {
		t.Errorf("Expected error code 'DuplicateEntity', got '%s'", electroErr.Code)
	}
}

func TestServiceEntity(t *testing.T) {
	service := NewService("TestService", nil)

	schema := &Schema{
		Service: "TestService",
		Entity:  "User",
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

	err = service.Join(entity)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	// Get existing entity
	retrievedEntity, err := service.Entity("User")
	if err != nil {
		t.Fatalf("Failed to get entity: %v", err)
	}

	if retrievedEntity != entity {
		t.Error("Retrieved entity does not match original")
	}

	// Try to get non-existent entity
	_, err = service.Entity("NonExistent")
	if err == nil {
		t.Error("Expected error when getting non-existent entity")
	}

	electroErr, ok := err.(*ElectroError)
	if !ok {
		t.Fatal("Expected ElectroError type")
	}

	if electroErr.Code != "EntityNotFound" {
		t.Errorf("Expected error code 'EntityNotFound', got '%s'", electroErr.Code)
	}
}

func TestServiceCollections(t *testing.T) {
	service := NewService("StoreService", &ServiceConfig{
		Table: stringPtr("StoreTable"),
	})

	// Create Store entity
	storeSchema := &Schema{
		Service: "StoreService",
		Entity:  "Store",
		Table:   "StoreTable",
		Attributes: map[string]*AttributeDefinition{
			"id":       {Type: AttributeTypeString, Required: true},
			"mall":     {Type: AttributeTypeString, Required: true},
			"building": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
			"byMall": {
				Index:      stringPtr("gsi1pk-gsi1sk-index"),
				Collection: stringPtr("mall"),
				PK:         FacetDefinition{Field: "gsi1pk", Facets: []string{"mall"}},
				SK:         &FacetDefinition{Field: "gsi1sk", Facets: []string{"building"}},
			},
		},
	}

	storeEntity, err := NewEntity(storeSchema, nil)
	if err != nil {
		t.Fatalf("Failed to create store entity: %v", err)
	}

	// Create Employee entity with same collection
	employeeSchema := &Schema{
		Service: "StoreService",
		Entity:  "Employee",
		Table:   "StoreTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"mall": {Type: AttributeTypeString, Required: true},
			"name": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
			"byMall": {
				Index:      stringPtr("gsi1pk-gsi1sk-index"),
				Collection: stringPtr("mall"),
				PK:         FacetDefinition{Field: "gsi1pk", Facets: []string{"mall"}},
				SK:         &FacetDefinition{Field: "gsi1sk", Facets: []string{"name"}},
			},
		},
	}

	employeeEntity, err := NewEntity(employeeSchema, nil)
	if err != nil {
		t.Fatalf("Failed to create employee entity: %v", err)
	}

	// Join both entities
	err = service.Join(storeEntity)
	if err != nil {
		t.Fatalf("Failed to join store entity: %v", err)
	}

	err = service.Join(employeeEntity)
	if err != nil {
		t.Fatalf("Failed to join employee entity: %v", err)
	}

	// Check collections
	collections := service.Collections()
	if len(collections) == 0 {
		t.Fatal("Expected collections to be created")
	}

	// Get mall collection
	mallCollection, err := service.Collection("mall")
	if err != nil {
		t.Fatalf("Failed to get mall collection: %v", err)
	}

	if mallCollection.name != "mall" {
		t.Errorf("Expected collection name 'mall', got '%s'", mallCollection.name)
	}

	// Check that both entities are in the collection
	if len(mallCollection.entities) != 2 {
		t.Errorf("Expected 2 entities in collection, got %d", len(mallCollection.entities))
	}

	// Try to get non-existent collection
	_, err = service.Collection("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent collection")
	}
}

func TestCollectionQuery(t *testing.T) {
	service := NewService("StoreService", &ServiceConfig{
		Table: stringPtr("StoreTable"),
	})

	// Create Store entity
	storeSchema := &Schema{
		Service: "StoreService",
		Entity:  "Store",
		Table:   "StoreTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"mall": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
			"byMall": {
				Index:      stringPtr("gsi1pk-gsi1sk-index"),
				Collection: stringPtr("mall"),
				PK:         FacetDefinition{Field: "gsi1pk", Facets: []string{"mall"}},
			},
		},
	}

	storeEntity, err := NewEntity(storeSchema, nil)
	if err != nil {
		t.Fatalf("Failed to create store entity: %v", err)
	}

	err = service.Join(storeEntity)
	if err != nil {
		t.Fatalf("Failed to join store entity: %v", err)
	}

	// Get collection
	mallCollection, err := service.Collection("mall")
	if err != nil {
		t.Fatalf("Failed to get mall collection: %v", err)
	}

	// Test collection query params
	params, err := mallCollection.Query("EastPointe").Params()
	if err != nil {
		t.Fatalf("Failed to generate collection query params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}

	entitiesParams, ok := params["entities"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected entities in params")
	}

	if _, exists := entitiesParams["Store"]; !exists {
		t.Error("Expected Store entity params")
	}
}
