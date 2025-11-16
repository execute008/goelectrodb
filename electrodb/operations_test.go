package electrodb

import (
	"testing"
)

// Test Append operation
func TestAppendOperation(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"tags":  {Type: AttributeTypeList},
			"items": {Type: AttributeTypeList},
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

	// Test appending to a list
	updateOp := entity.Update(Keys{"id": "123"}).
		Append(map[string]interface{}{
			"tags": []string{"new", "append"},
		})

	if updateOp == nil {
		t.Fatal("Update operation should not be nil")
	}

	if len(updateOp.appendOps) != 1 {
		t.Errorf("Expected 1 append operation, got %d", len(updateOp.appendOps))
	}

	// Test params generation
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("UpdateExpression should be present")
	}

	// Should contain list_append function
	if !contains(updateExpr, "list_append") {
		t.Errorf("Update expression should contain list_append, got: %s", updateExpr)
	}
}

// Test Prepend operation
func TestPrependOperation(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"items": {Type: AttributeTypeList},
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

	// Test prepending to a list
	updateOp := entity.Update(Keys{"id": "123"}).
		Prepend(map[string]interface{}{
			"items": []string{"first", "second"},
		})

	if updateOp == nil {
		t.Fatal("Update operation should not be nil")
	}

	if len(updateOp.prependOps) != 1 {
		t.Errorf("Expected 1 prepend operation, got %d", len(updateOp.prependOps))
	}

	// Test params generation
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("UpdateExpression should be present")
	}

	// Should contain list_append function
	if !contains(updateExpr, "list_append") {
		t.Errorf("Update expression should contain list_append for prepend, got: %s", updateExpr)
	}
}

// Test Subtract operation
func TestSubtractOperation(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":      {Type: AttributeTypeString, Required: true},
			"balance": {Type: AttributeTypeNumber},
			"stock":   {Type: AttributeTypeNumber},
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

	// Test subtracting from a number
	updateOp := entity.Update(Keys{"id": "123"}).
		Subtract(map[string]interface{}{
			"balance": 50,
			"stock":   5,
		})

	if updateOp == nil {
		t.Fatal("Update operation should not be nil")
	}

	if len(updateOp.subtractOps) != 2 {
		t.Errorf("Expected 2 subtract operations, got %d", len(updateOp.subtractOps))
	}

	// Test params generation
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("UpdateExpression should be present")
	}

	// Should contain subtraction operator
	if !contains(updateExpr, " - ") {
		t.Errorf("Update expression should contain subtraction operator, got: %s", updateExpr)
	}
}

// Test Data operation
func TestDataOperation(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"items": {Type: AttributeTypeList},
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

	// Test removing specific list indices
	updateOp := entity.Update(Keys{"id": "123"}).
		Data(map[string]interface{}{
			"items": []int{0, 2}, // Remove items at indices 0 and 2
		})

	if updateOp == nil {
		t.Fatal("Update operation should not be nil")
	}

	if len(updateOp.dataOps) != 1 {
		t.Errorf("Expected 1 data operation, got %d", len(updateOp.dataOps))
	}

	// Test params generation
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("UpdateExpression should be present")
	}

	// Should contain REMOVE clause with indices
	if !contains(updateExpr, "REMOVE") {
		t.Errorf("Update expression should contain REMOVE for data operation, got: %s", updateExpr)
	}
	if !contains(updateExpr, "[0]") {
		t.Errorf("Update expression should contain list index [0], got: %s", updateExpr)
	}
}

// Test combined operations
func TestCombinedOperations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":      {Type: AttributeTypeString, Required: true},
			"name":    {Type: AttributeTypeString},
			"balance": {Type: AttributeTypeNumber},
			"tags":    {Type: AttributeTypeList},
			"notes":   {Type: AttributeTypeList},
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

	// Test using multiple operation types together
	updateOp := entity.Update(Keys{"id": "123"}).
		Set(map[string]interface{}{
			"name": "Updated Name",
		}).
		Add(map[string]interface{}{
			"balance": 100,
		}).
		Subtract(map[string]interface{}{
			"balance": 25,
		}).
		Append(map[string]interface{}{
			"tags": []string{"new-tag"},
		}).
		Prepend(map[string]interface{}{
			"notes": []string{"important"},
		}).
		Remove([]string{"oldAttribute"})

	if updateOp == nil {
		t.Fatal("Update operation should not be nil")
	}

	// Verify all operations were recorded
	if len(updateOp.setOps) != 1 {
		t.Errorf("Expected 1 set operation, got %d", len(updateOp.setOps))
	}
	if len(updateOp.addOps) != 1 {
		t.Errorf("Expected 1 add operation, got %d", len(updateOp.addOps))
	}
	if len(updateOp.subtractOps) != 1 {
		t.Errorf("Expected 1 subtract operation, got %d", len(updateOp.subtractOps))
	}
	if len(updateOp.appendOps) != 1 {
		t.Errorf("Expected 1 append operation, got %d", len(updateOp.appendOps))
	}
	if len(updateOp.prependOps) != 1 {
		t.Errorf("Expected 1 prepend operation, got %d", len(updateOp.prependOps))
	}
	if len(updateOp.remOps) != 1 {
		t.Errorf("Expected 1 remove operation, got %d", len(updateOp.remOps))
	}

	// Test params generation with all operations
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("UpdateExpression should be present")
	}

	// Verify all clauses are present
	if !contains(updateExpr, "SET") {
		t.Error("Update expression should contain SET clause")
	}
	if !contains(updateExpr, "ADD") {
		t.Error("Update expression should contain ADD clause")
	}
	if !contains(updateExpr, "REMOVE") {
		t.Error("Update expression should contain REMOVE clause")
	}
	if !contains(updateExpr, "list_append") {
		t.Error("Update expression should contain list_append for append/prepend")
	}
	if !contains(updateExpr, " - ") {
		t.Error("Update expression should contain subtraction")
	}
}

// Test Patch operation (alias for Update)
func TestPatchOperation(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"name": {Type: AttributeTypeString},
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

	// Patch should work exactly like Update
	patchOp := entity.Patch(Keys{"id": "123"}).
		Set(map[string]interface{}{
			"name": "Patched Name",
		})

	if patchOp == nil {
		t.Fatal("Patch operation should not be nil")
	}

	updateOp := entity.Update(Keys{"id": "123"}).
		Set(map[string]interface{}{
			"name": "Updated Name",
		})

	// Both should have the same structure
	if len(patchOp.setOps) != len(updateOp.setOps) {
		t.Error("Patch and Update should have same number of set operations")
	}
}

// Test transaction support for new operations
func TestTransactionWithNewOperations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":      {Type: AttributeTypeString, Required: true},
			"balance": {Type: AttributeTypeNumber},
			"tags":    {Type: AttributeTypeList},
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

	// Create update with new operations
	updateOp := entity.Update(Keys{"id": "123"}).
		Subtract(map[string]interface{}{
			"balance": 50,
		}).
		Append(map[string]interface{}{
			"tags": []string{"processed"},
		})

	// Commit to transaction
	transactItem := updateOp.Commit()
	if transactItem == nil {
		t.Fatal("Transaction item should not be nil")
	}

	// Should be a TransactUpdateItem
	tui, ok := transactItem.(*TransactUpdateItem)
	if !ok {
		t.Fatal("Transaction item should be TransactUpdateItem")
	}

	// Verify operations were copied
	if len(tui.subtractOps) != 1 {
		t.Errorf("Expected 1 subtract operation in transaction, got %d", len(tui.subtractOps))
	}
	if len(tui.appendOps) != 1 {
		t.Errorf("Expected 1 append operation in transaction, got %d", len(tui.appendOps))
	}

	// Test building transaction item
	_, err = tui.BuildTransactItem()
	if err != nil {
		t.Errorf("Failed to build transaction item: %v", err)
	}
}

// Test multiple appends and prepends
func TestMultipleListOperations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"list1": {Type: AttributeTypeList},
			"list2": {Type: AttributeTypeList},
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

	// Test appending to multiple lists
	updateOp := entity.Update(Keys{"id": "123"}).
		Append(map[string]interface{}{
			"list1": []string{"a", "b"},
			"list2": []int{1, 2},
		})

	if len(updateOp.appendOps) != 2 {
		t.Errorf("Expected 2 append operations, got %d", len(updateOp.appendOps))
	}

	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	updateExpr := params["UpdateExpression"].(string)

	// Both lists should be in the expression
	if !contains(updateExpr, "list_append") {
		t.Error("Update expression should contain list_append")
	}
}

// Test subtract with different types
func TestSubtractWithDifferentTypes(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":     {Type: AttributeTypeString, Required: true},
			"intVal": {Type: AttributeTypeNumber},
			"floatVal": {Type: AttributeTypeNumber},
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

	// Test subtracting different number types
	updateOp := entity.Update(Keys{"id": "123"}).
		Subtract(map[string]interface{}{
			"intVal":   10,
			"floatVal": 5.5,
		})

	if len(updateOp.subtractOps) != 2 {
		t.Errorf("Expected 2 subtract operations, got %d", len(updateOp.subtractOps))
	}

	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to generate params: %v", err)
	}

	// Verify expression was built
	if params["UpdateExpression"] == nil {
		t.Error("UpdateExpression should be present")
	}
}
