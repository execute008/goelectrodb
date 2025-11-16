package electrodb

import (
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "User",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"userId": {Type: AttributeTypeString, Required: true},
			"email":  {Type: AttributeTypeString, Required: true},
			"name":   {Type: AttributeTypeString, Required: false},
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

	// Test Create operation
	createOp := entity.Create(Item{
		"userId": "user123",
		"email":  "test@example.com",
		"name":   "Test User",
	})

	if createOp == nil {
		t.Fatal("Expected non-nil create operation")
	}

	// Verify that Create has a condition builder (to prevent overwrite)
	if createOp.conditionBuilder == nil {
		t.Error("Expected Create to have a condition builder to prevent overwrite")
	}

	// Build params to verify the condition expression
	params, err := createOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Should have the item
	if params["Item"] == nil {
		t.Error("Expected Item to be set")
	}
}

func TestCreateConditionExpression(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"name":      {Type: AttributeTypeString, Required: true},
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

	// Create operation should have condition to prevent overwrite
	createOp := entity.Create(Item{
		"productId": "prod123",
		"name":      "Test Product",
	})

	// Verify condition builder exists
	if createOp.conditionBuilder == nil {
		t.Fatal("Expected condition builder on Create operation")
	}

	// Build the condition expression
	condExpr, condNames, condValues := createOp.conditionBuilder.Build()

	// Should have attribute_not_exists condition
	if !strings.Contains(condExpr, "attribute_not_exists") {
		t.Errorf("Expected condition to contain 'attribute_not_exists', got: %s", condExpr)
	}

	// Should have attribute names
	if len(condNames) == 0 {
		t.Error("Expected condition to have attribute names")
	}

	// Should not have values (attribute_not_exists doesn't need values)
	if len(condValues) != 0 {
		t.Error("Expected no attribute values for attribute_not_exists")
	}
}

func TestUpsert(t *testing.T) {
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

	// Test Upsert operation (should be same as Put)
	upsertOp := entity.Upsert(Item{
		"userId": "user123",
		"email":  "test@example.com",
	})

	if upsertOp == nil {
		t.Fatal("Expected non-nil upsert operation")
	}

	// Upsert should NOT have a condition builder (allows overwrite)
	if upsertOp.conditionBuilder != nil {
		t.Error("Expected Upsert to not have a condition builder (should allow overwrite)")
	}

	// Build params
	params, err := upsertOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Should have the item
	if params["Item"] == nil {
		t.Error("Expected Item to be set")
	}
}

func TestUpsertUpdate(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
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

	// Test UpsertUpdate operation
	upsertUpdateOp := entity.UpsertUpdate(Keys{"productId": "prod123"}).
		Set(map[string]interface{}{
			"name":  "Updated Product",
			"price": 99.99,
		})

	if upsertUpdateOp == nil {
		t.Fatal("Expected non-nil upsert update operation")
	}

	// Build params
	params, err := upsertUpdateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Should have update expression
	if params["UpdateExpression"] == nil {
		t.Error("Expected UpdateExpression to be set")
	}

	// Should have SET clause for the attributes
	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	if !strings.Contains(updateExpr, "SET") {
		t.Errorf("Expected UpdateExpression to contain 'SET', got: %s", updateExpr)
	}
}

func TestCreateVsUpsert(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Item",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"itemId": {Type: AttributeTypeString, Required: true},
			"value":  {Type: AttributeTypeString, Required: false},
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

	item := Item{
		"itemId": "item123",
		"value":  "test",
	}

	// Create should have condition builder
	createOp := entity.Create(item)
	if createOp.conditionBuilder == nil {
		t.Error("Create should have condition builder")
	}

	// Upsert should NOT have condition builder
	upsertOp := entity.Upsert(item)
	if upsertOp.conditionBuilder != nil {
		t.Error("Upsert should not have condition builder")
	}

	// Put should NOT have condition builder (same as Upsert)
	putOp := entity.Put(item)
	if putOp.conditionBuilder != nil {
		t.Error("Put should not have condition builder")
	}
}

func TestCreateWithManualCondition(t *testing.T) {
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

	// Test that manual conditions can be added to Create (overriding default)
	createOp := entity.Create(Item{
		"userId": "user123",
		"email":  "test@example.com",
		"status": "active",
	}).Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		// Custom condition - only create if userId doesn't exist AND status is active
		return ops.NotExists(attrs["userId"]) + " AND " + attrs["status"].Eq("active")
	})

	// Should have condition builder
	if createOp.conditionBuilder == nil {
		t.Error("Expected condition builder on Create operation with manual condition")
	}

	// Build the condition
	condExpr, _, _ := createOp.conditionBuilder.Build()

	// Should have both conditions
	if !strings.Contains(condExpr, "attribute_not_exists") {
		t.Errorf("Expected condition to contain 'attribute_not_exists', got: %s", condExpr)
	}

	if !strings.Contains(condExpr, "AND") {
		t.Errorf("Expected condition to contain 'AND', got: %s", condExpr)
	}
}

func TestUpdateVsUpsertUpdate(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Record",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"recordId": {Type: AttributeTypeString, Required: true},
			"data":     {Type: AttributeTypeString, Required: false},
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

	keys := Keys{"recordId": "rec123"}

	// Both Update and UpsertUpdate should produce the same operation
	// (UpdateItem in DynamoDB naturally upserts)
	updateOp := entity.Update(keys).Set(map[string]interface{}{"data": "value1"})
	upsertUpdateOp := entity.UpsertUpdate(keys).Set(map[string]interface{}{"data": "value2"})

	updateParams, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build update params: %v", err)
	}

	upsertParams, err := upsertUpdateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build upsert update params: %v", err)
	}

	// Both should have UpdateExpression
	if updateParams["UpdateExpression"] == nil || upsertParams["UpdateExpression"] == nil {
		t.Error("Both operations should have UpdateExpression")
	}

	// Both should have the same structure (UpsertUpdate is just an alias for clarity)
	if updateParams["TableName"] != upsertParams["TableName"] {
		t.Error("Both operations should target the same table")
	}
}
