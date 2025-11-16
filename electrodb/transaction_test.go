package electrodb

import (
	"testing"
)

func TestTransactWrite(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	// Create entities
	userSchema := &Schema{
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

	productSchema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"name":  {Type: AttributeTypeString, Required: true},
			"price": {Type: AttributeTypeNumber, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	userEntity, err := NewEntity(userSchema, nil)
	if err != nil {
		t.Fatalf("Failed to create user entity: %v", err)
	}

	productEntity, err := NewEntity(productSchema, nil)
	if err != nil {
		t.Fatalf("Failed to create product entity: %v", err)
	}

	err = service.Join(userEntity)
	if err != nil {
		t.Fatalf("Failed to join user entity: %v", err)
	}

	err = service.Join(productEntity)
	if err != nil {
		t.Fatalf("Failed to join product entity: %v", err)
	}

	// Build transaction
	txBuilder := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		user := entities["User"]
		product := entities["Product"]

		return []TransactionItem{
			user.Put(Item{"id": "user1", "name": "John Doe"}).Commit(),
			product.Put(Item{"id": "prod1", "name": "Widget", "price": 99}).Commit(),
			user.Update(Keys{"id": "user2"}).Set(map[string]interface{}{"name": "Jane Doe"}).Commit(),
			product.Delete(Keys{"id": "prod2"}).Commit(),
		}
	})

	if txBuilder == nil {
		t.Fatal("Expected transaction builder to be non-nil")
	}

	// Test params generation
	params, err := txBuilder.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}

	transactItems, ok := params["TransactItems"]
	if !ok {
		t.Fatal("Expected TransactItems in params")
	}

	if transactItems == nil {
		t.Fatal("Expected TransactItems to be non-nil")
	}
}

func TestTransactWriteEmpty(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	// Build empty transaction
	txBuilder := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		return []TransactionItem{}
	})

	// Test params generation with empty items
	params, err := txBuilder.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}
}

func TestTransactGet(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	// Create entity
	userSchema := &Schema{
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

	userEntity, err := NewEntity(userSchema, nil)
	if err != nil {
		t.Fatalf("Failed to create user entity: %v", err)
	}

	err = service.Join(userEntity)
	if err != nil {
		t.Fatalf("Failed to join user entity: %v", err)
	}

	// Build transaction get
	txBuilder := service.TransactGet(func(entities map[string]*Entity) []TransactionItem {
		user := entities["User"]

		return []TransactionItem{
			user.Get(Keys{"id": "user1"}).Commit(),
			user.Get(Keys{"id": "user2"}).Commit(),
			user.Get(Keys{"id": "user3"}).Commit(),
		}
	})

	if txBuilder == nil {
		t.Fatal("Expected transaction builder to be non-nil")
	}

	// Test params generation
	params, err := txBuilder.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}

	transactItems, ok := params["TransactItems"]
	if !ok {
		t.Fatal("Expected TransactItems in params")
	}

	if transactItems == nil {
		t.Fatal("Expected TransactItems to be non-nil")
	}
}

func TestCommitOperations(t *testing.T) {
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

	// Test Put.Commit()
	putItem := entity.Put(Item{"id": "1", "name": "Test"}).Commit()
	if putItem == nil {
		t.Error("Expected put commit to return TransactionItem")
	}

	// Build transact item
	transactItem, err := putItem.BuildTransactItem()
	if err != nil {
		t.Errorf("Failed to build transact item: %v", err)
	}

	if transactItem.Put == nil {
		t.Error("Expected Put to be set in transact item")
	}

	// Test Update.Commit()
	updateItem := entity.Update(Keys{"id": "1"}).Set(map[string]interface{}{"name": "Updated"}).Commit()
	if updateItem == nil {
		t.Error("Expected update commit to return TransactionItem")
	}

	transactItem, err = updateItem.BuildTransactItem()
	if err != nil {
		t.Errorf("Failed to build transact item: %v", err)
	}

	if transactItem.Update == nil {
		t.Error("Expected Update to be set in transact item")
	}

	// Test Delete.Commit()
	deleteItem := entity.Delete(Keys{"id": "1"}).Commit()
	if deleteItem == nil {
		t.Error("Expected delete commit to return TransactionItem")
	}

	transactItem, err = deleteItem.BuildTransactItem()
	if err != nil {
		t.Errorf("Failed to build transact item: %v", err)
	}

	if transactItem.Delete == nil {
		t.Error("Expected Delete to be set in transact item")
	}

	// Test Get.Commit()
	getItem := entity.Get(Keys{"id": "1"}).Commit()
	if getItem == nil {
		t.Error("Expected get commit to return TransactionItem")
	}

	// Get should work for TransactGet but not TransactWrite
	_, err = getItem.BuildTransactItem()
	if err == nil {
		t.Error("Expected error when using Get in TransactWrite")
	}

	transactGetItem, err := getItem.BuildTransactGetItem()
	if err != nil {
		t.Errorf("Failed to build transact get item: %v", err)
	}

	if transactGetItem.Get == nil {
		t.Error("Expected Get to be set in transact get item")
	}

	// Test that Put cannot be used in TransactGet
	_, err = putItem.BuildTransactGetItem()
	if err == nil {
		t.Error("Expected error when using Put in TransactGet")
	}
}

func TestTransactWriteMixedOperations(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"name":  {Type: AttributeTypeString, Required: true},
			"count": {Type: AttributeTypeNumber, Required: false},
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

	// Build transaction with multiple operation types
	txBuilder := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		e := entities["TestEntity"]

		return []TransactionItem{
			// Put new item
			e.Put(Item{"id": "1", "name": "Item 1", "count": 0}).Commit(),
			// Update existing item
			e.Update(Keys{"id": "2"}).
				Set(map[string]interface{}{"name": "Updated"}).
				Add(map[string]interface{}{"count": 1}).
				Commit(),
			// Delete item
			e.Delete(Keys{"id": "3"}).Commit(),
			// Put another item
			e.Put(Item{"id": "4", "name": "Item 4", "count": 10}).Commit(),
		}
	})

	params, err := txBuilder.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	if params == nil {
		t.Fatal("Expected params to be non-nil")
	}

	// Verify all operations are included
	// In a real scenario, we would check the structure more thoroughly
}
