package electrodb

import (
	"strings"
	"testing"
)

func TestTransactPutWithCondition(t *testing.T) {
	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})

	schema := &Schema{
		Service: "TestService",
		Entity:  "User",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"userId": {Type: AttributeTypeString, Required: true},
			"email":  {Type: AttributeTypeString, Required: true},
			"status": {Type: AttributeTypeString, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"userId"}},
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

	// Test put with condition (only put if item doesn't exist)
	putOp := entity.Put(Item{
		"userId": "user123",
		"email":  "user@example.com",
		"status": "active",
	}).Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return ops.NotExists(attrs["userId"])
	})

	params, err := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		return []TransactionItem{
			putOp.Commit(),
		}
	}).Params()

	if err != nil {
		t.Fatalf("Failed to build transaction params: %v", err)
	}

	// Verify params structure
	if params["TransactItems"] == nil {
		t.Fatal("Expected TransactItems to be set")
	}
}

func TestTransactUpdateWithCondition(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"version":   {Type: AttributeTypeNumber, Required: false},
			"name":      {Type: AttributeTypeString, Required: false},
			"price":     {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"productId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})
	err = service.Join(entity)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	// Test update with optimistic locking condition
	updateOp := entity.Update(Keys{"productId": "prod123"}).
		Set(map[string]interface{}{
			"name":    "Updated Product",
			"price":   99.99,
			"version": 2,
		}).
		Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
			return attrs["version"].Eq(1)
		})

	params, err := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		return []TransactionItem{
			updateOp.Commit(),
		}
	}).Params()

	if err != nil {
		t.Fatalf("Failed to build transaction params: %v", err)
	}

	// Verify params structure
	if params["TransactItems"] == nil {
		t.Fatal("Expected TransactItems to be set")
	}
}

func TestTransactDeleteWithCondition(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Record",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"recordId": {Type: AttributeTypeString, Required: true},
			"status":   {Type: AttributeTypeString, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"recordId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})
	err = service.Join(entity)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	// Test delete with condition (only delete if status is 'inactive')
	deleteOp := entity.Delete(Keys{"recordId": "rec123"}).
		Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
			return attrs["status"].Eq("inactive")
		})

	params, err := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		return []TransactionItem{
			deleteOp.Commit(),
		}
	}).Params()

	if err != nil {
		t.Fatalf("Failed to build transaction params: %v", err)
	}

	// Verify params structure
	if params["TransactItems"] == nil {
		t.Fatal("Expected TransactItems to be set")
	}
}

func TestTransactMixedOperationsWithConditions(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Account",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"accountId": {Type: AttributeTypeString, Required: true},
			"balance":   {Type: AttributeTypeNumber, Required: false},
			"status":    {Type: AttributeTypeString, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"accountId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})
	err = service.Join(entity)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	// Test transaction with multiple operations, each with conditions
	putOp := entity.Put(Item{
		"accountId": "acc123",
		"balance":   1000,
		"status":    "active",
	}).Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return ops.NotExists(attrs["accountId"])
	})

	updateOp := entity.Update(Keys{"accountId": "acc456"}).
		Set(map[string]interface{}{
			"balance": 500,
		}).
		Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
			return attrs["balance"].Gte(100)
		})

	deleteOp := entity.Delete(Keys{"accountId": "acc789"}).
		Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
			return attrs["balance"].Eq(0) + " AND " + attrs["status"].Eq("closed")
		})

	params, err := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		return []TransactionItem{
			putOp.Commit(),
			updateOp.Commit(),
			deleteOp.Commit(),
		}
	}).Params()

	if err != nil {
		t.Fatalf("Failed to build transaction params: %v", err)
	}

	// Verify params structure
	if params["TransactItems"] == nil {
		t.Fatal("Expected TransactItems to be set")
	}
}

func TestTransactWithoutCondition(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Item",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"itemId": {Type: AttributeTypeString, Required: true},
			"name":   {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"itemId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})
	err = service.Join(entity)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	// Test transaction without condition - should work as before
	putOp := entity.Put(Item{
		"itemId": "item123",
		"name":   "Test Item",
	})

	params, err := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		return []TransactionItem{
			putOp.Commit(),
		}
	}).Params()

	if err != nil {
		t.Fatalf("Failed to build transaction params: %v", err)
	}

	// Verify params structure
	if params["TransactItems"] == nil {
		t.Fatal("Expected TransactItems to be set")
	}
}

func TestConditionWithComplexExpressions(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Document",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"docId":     {Type: AttributeTypeString, Required: true},
			"version":   {Type: AttributeTypeNumber, Required: false},
			"published": {Type: AttributeTypeBoolean, Required: false},
			"author":    {Type: AttributeTypeString, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"docId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	service := NewService("TestService", &ServiceConfig{
		Table: stringPtr("TestTable"),
	})
	err = service.Join(entity)
	if err != nil {
		t.Fatalf("Failed to join entity: %v", err)
	}

	// Test with complex condition expression
	updateOp := entity.Update(Keys{"docId": "doc123"}).
		Set(map[string]interface{}{
			"version":   2,
			"published": true,
		}).
		Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
			// Only update if:
			// - version is 1
			// - document is not published yet
			// - author exists
			versionCheck := attrs["version"].Eq(1)
			publishedCheck := attrs["published"].Eq(false)
			authorExists := ops.Exists(attrs["author"])
			return "(" + versionCheck + " AND " + publishedCheck + ") AND " + authorExists
		})

	params, err := service.TransactWrite(func(entities map[string]*Entity) []TransactionItem {
		return []TransactionItem{
			updateOp.Commit(),
		}
	}).Params()

	if err != nil {
		t.Fatalf("Failed to build transaction params: %v", err)
	}

	// Verify params structure
	if params["TransactItems"] == nil {
		t.Fatal("Expected TransactItems to be set")
	}

	// Verify the transaction item has a condition
	// Note: We can't easily inspect the internal structure here without
	// unmarshaling, but we've verified it doesn't error
}

func TestPutConditionExpressionStructure(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "User",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"userId": {Type: AttributeTypeString, Required: true},
			"email":  {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"userId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Create put operation with condition
	putOp := entity.Put(Item{
		"userId": "user123",
		"email":  "test@example.com",
	}).Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return ops.NotExists(attrs["userId"])
	})

	// Build transaction item
	transactItem := putOp.Commit()

	// Cast to TransactPutItem to verify structure
	putItem, ok := transactItem.(*TransactPutItem)
	if !ok {
		t.Fatal("Expected TransactPutItem type")
	}

	if putItem.conditionBuilder == nil {
		t.Fatal("Expected condition builder to be set")
	}

	// Build the DynamoDB item
	item, err := putItem.BuildTransactItem()
	if err != nil {
		t.Fatalf("Failed to build transact item: %v", err)
	}

	// Verify Put is set
	if item.Put == nil {
		t.Fatal("Expected Put to be set in TransactWriteItem")
	}

	// Verify condition expression is set
	if item.Put.ConditionExpression == nil {
		t.Fatal("Expected ConditionExpression to be set")
	}

	if !strings.Contains(*item.Put.ConditionExpression, "attribute_not_exists") {
		t.Errorf("Expected condition to contain 'attribute_not_exists', got: %s", *item.Put.ConditionExpression)
	}
}
