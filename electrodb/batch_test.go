package electrodb

import (
	"testing"
)

func TestBatchGetRequest(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"name": {Type: AttributeTypeString, Required: false},
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

	keys := []Keys{
		{"id": "1"},
		{"id": "2"},
		{"id": "3"},
	}

	batchGetRequest := entity.BatchGet(keys)

	if batchGetRequest == nil {
		t.Fatal("Expected batch get request to be non-nil")
	}

	if len(batchGetRequest.keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(batchGetRequest.keys))
	}
}

func TestBatchWriteRequest(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
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

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	items := []Item{
		{"id": "1", "name": "Item 1"},
		{"id": "2", "name": "Item 2"},
	}

	deleteKeys := []Keys{
		{"id": "3"},
	}

	batchWriteRequest := entity.BatchWrite().Put(items).Delete(deleteKeys)

	if batchWriteRequest == nil {
		t.Fatal("Expected batch write request to be non-nil")
	}

	if len(batchWriteRequest.puts) != 2 {
		t.Errorf("Expected 2 put items, got %d", len(batchWriteRequest.puts))
	}

	if len(batchWriteRequest.deletes) != 1 {
		t.Errorf("Expected 1 delete key, got %d", len(batchWriteRequest.deletes))
	}
}

func TestBatchWriteTooLarge(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
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

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Create more than MaxBatchWriteItems items
	items := make([]Item, MaxBatchWriteItems+1)
	for i := 0; i < MaxBatchWriteItems+1; i++ {
		items[i] = Item{
			"id":   string(rune(i)),
			"name": "Item",
		}
	}

	batchWriteRequest := entity.BatchWrite().Put(items)

	_, err = batchWriteRequest.Go()
	if err == nil {
		t.Fatal("Expected error for batch too large")
	}

	electroErr, ok := err.(*ElectroError)
	if !ok {
		t.Fatal("Expected ElectroError type")
	}

	if electroErr.Code != "BatchTooLarge" {
		t.Errorf("Expected error code 'BatchTooLarge', got '%s'", electroErr.Code)
	}
}

func TestServiceBatchGet(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	// Create two entities
	schema1 := &Schema{
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

	schema2 := &Schema{
		Service: "TestService",
		Entity:  "Product",
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

	entity1, err := NewEntity(schema1, nil)
	if err != nil {
		t.Fatalf("Failed to create entity1: %v", err)
	}

	entity2, err := NewEntity(schema2, nil)
	if err != nil {
		t.Fatalf("Failed to create entity2: %v", err)
	}

	err = service.Join(entity1)
	if err != nil {
		t.Fatalf("Failed to join entity1: %v", err)
	}

	err = service.Join(entity2)
	if err != nil {
		t.Fatalf("Failed to join entity2: %v", err)
	}

	// Create batch get request
	batchGetRequest := service.BatchGet().
		Get("User", []Keys{{"id": "user1"}, {"id": "user2"}}).
		Get("Product", []Keys{{"id": "prod1"}})

	if batchGetRequest == nil {
		t.Fatal("Expected batch get request to be non-nil")
	}

	if len(batchGetRequest.requests) != 2 {
		t.Errorf("Expected 2 entity requests, got %d", len(batchGetRequest.requests))
	}

	if len(batchGetRequest.requests["User"]) != 2 {
		t.Errorf("Expected 2 user keys, got %d", len(batchGetRequest.requests["User"]))
	}

	if len(batchGetRequest.requests["Product"]) != 1 {
		t.Errorf("Expected 1 product key, got %d", len(batchGetRequest.requests["Product"]))
	}
}

func TestServiceBatchWrite(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	schema := &Schema{
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

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	err = service.Join(entity)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	items := []Item{
		{"id": "1", "name": "User 1"},
		{"id": "2", "name": "User 2"},
	}

	deleteKeys := []Keys{
		{"id": "3"},
	}

	batchWriteRequest := service.BatchWrite().
		Put("User", items).
		Delete("User", deleteKeys)

	if batchWriteRequest == nil {
		t.Fatal("Expected batch write request to be non-nil")
	}

	if len(batchWriteRequest.puts["User"]) != 2 {
		t.Errorf("Expected 2 put items, got %d", len(batchWriteRequest.puts["User"]))
	}

	if len(batchWriteRequest.deletes["User"]) != 1 {
		t.Errorf("Expected 1 delete key, got %d", len(batchWriteRequest.deletes["User"]))
	}
}
