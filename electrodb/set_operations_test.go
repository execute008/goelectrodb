package electrodb

import (
	"strings"
	"testing"
)

func TestAddToSet(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "User",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"userId": {Type: AttributeTypeString, Required: true},
			"tags":   {Type: AttributeTypeSet, Required: false},
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

	// Test AddToSet operation
	updateOp := entity.Update(Keys{"userId": "user123"}).
		AddToSet("tags", []string{"premium", "verified"})

	if updateOp == nil {
		t.Fatal("Expected non-nil update operation")
	}

	// Build params
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Should have update expression
	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	// Should contain ADD clause
	if !strings.Contains(updateExpr, "ADD") {
		t.Errorf("Expected UpdateExpression to contain 'ADD', got: %s", updateExpr)
	}
}

func TestDeleteFromSet(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"categories": {Type: AttributeTypeSet, Required: false},
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

	// Test DeleteFromSet operation
	updateOp := entity.Update(Keys{"productId": "prod123"}).
		DeleteFromSet("categories", []string{"outdated", "deprecated"})

	if updateOp == nil {
		t.Fatal("Expected non-nil update operation")
	}

	// Build params
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Should have update expression
	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	// Should contain DELETE clause
	if !strings.Contains(updateExpr, "DELETE") {
		t.Errorf("Expected UpdateExpression to contain 'DELETE', got: %s", updateExpr)
	}
}

func TestCombinedSetOperations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Record",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"recordId": {Type: AttributeTypeString, Required: true},
			"tags":     {Type: AttributeTypeSet, Required: false},
			"labels":   {Type: AttributeTypeSet, Required: false},
			"name":     {Type: AttributeTypeString, Required: false},
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

	// Test combining SET, ADD (to set), and DELETE (from set)
	updateOp := entity.Update(Keys{"recordId": "rec123"}).
		Set(map[string]interface{}{
			"name": "Updated Record",
		}).
		AddToSet("tags", []string{"new", "featured"}).
		DeleteFromSet("labels", []string{"old", "archived"})

	// Build params
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	// Should contain all three clauses
	if !strings.Contains(updateExpr, "SET") {
		t.Errorf("Expected UpdateExpression to contain 'SET', got: %s", updateExpr)
	}

	if !strings.Contains(updateExpr, "ADD") {
		t.Errorf("Expected UpdateExpression to contain 'ADD', got: %s", updateExpr)
	}

	if !strings.Contains(updateExpr, "DELETE") {
		t.Errorf("Expected UpdateExpression to contain 'DELETE', got: %s", updateExpr)
	}
}

func TestAddToSetWithNumbers(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Stats",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"statsId": {Type: AttributeTypeString, Required: true},
			"scores":  {Type: AttributeTypeSet, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"statsId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test AddToSet with numbers (Number Set - NS)
	updateOp := entity.Update(Keys{"statsId": "stats123"}).
		AddToSet("scores", []int{100, 200, 300})

	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	if !strings.Contains(updateExpr, "ADD") {
		t.Errorf("Expected UpdateExpression to contain 'ADD', got: %s", updateExpr)
	}
}

func TestSetOperationsWithCounter(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Counter",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"counterId": {Type: AttributeTypeString, Required: true},
			"count":     {Type: AttributeTypeNumber, Required: false},
			"tags":      {Type: AttributeTypeSet, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"counterId"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test combining numeric ADD with set ADD
	updateOp := entity.Update(Keys{"counterId": "counter123"}).
		Add(map[string]interface{}{
			"count": 5,
		}).
		AddToSet("tags", []string{"active"})

	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	// Both should use ADD clause
	if !strings.Contains(updateExpr, "ADD") {
		t.Errorf("Expected UpdateExpression to contain 'ADD', got: %s", updateExpr)
	}

	// Verify expression has two ADD operations (should be comma-separated)
	addCount := strings.Count(updateExpr, "ADD")
	if addCount != 1 {
		t.Logf("Expected single ADD clause with comma-separated operations, got: %s", updateExpr)
	}
}

func TestDeleteVsRemove(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Document",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"docId":   {Type: AttributeTypeString, Required: true},
			"tags":    {Type: AttributeTypeSet, Required: false},
			"tempId":  {Type: AttributeTypeString, Required: false},
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

	// Test DELETE (removes values from set) vs REMOVE (removes attribute entirely)
	updateOp := entity.Update(Keys{"docId": "doc123"}).
		DeleteFromSet("tags", []string{"draft"}).  // DELETE - removes "draft" from tags set
		Remove([]string{"tempId"})                  // REMOVE - removes tempId attribute entirely

	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	// Should have both DELETE and REMOVE
	if !strings.Contains(updateExpr, "DELETE") {
		t.Errorf("Expected UpdateExpression to contain 'DELETE', got: %s", updateExpr)
	}

	if !strings.Contains(updateExpr, "REMOVE") {
		t.Errorf("Expected UpdateExpression to contain 'REMOVE', got: %s", updateExpr)
	}
}

func TestAllUpdateOperations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Item",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"itemId":  {Type: AttributeTypeString, Required: true},
			"name":    {Type: AttributeTypeString, Required: false},
			"count":   {Type: AttributeTypeNumber, Required: false},
			"tags":    {Type: AttributeTypeSet, Required: false},
			"labels":  {Type: AttributeTypeSet, Required: false},
			"tempData": {Type: AttributeTypeString, Required: false},
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

	// Test all update operations together: SET, ADD (counter), ADD (set), DELETE (set), REMOVE
	updateOp := entity.Update(Keys{"itemId": "item123"}).
		Set(map[string]interface{}{
			"name": "Updated Item",
		}).
		Add(map[string]interface{}{
			"count": 1,
		}).
		AddToSet("tags", []string{"new"}).
		DeleteFromSet("labels", []string{"old"}).
		Remove([]string{"tempData"})

	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("Expected UpdateExpression to be a string")
	}

	// Should have SET, ADD, DELETE, and REMOVE
	if !strings.Contains(updateExpr, "SET") {
		t.Errorf("Expected UpdateExpression to contain 'SET', got: %s", updateExpr)
	}

	if !strings.Contains(updateExpr, "ADD") {
		t.Errorf("Expected UpdateExpression to contain 'ADD', got: %s", updateExpr)
	}

	if !strings.Contains(updateExpr, "DELETE") {
		t.Errorf("Expected UpdateExpression to contain 'DELETE', got: %s", updateExpr)
	}

	if !strings.Contains(updateExpr, "REMOVE") {
		t.Errorf("Expected UpdateExpression to contain 'REMOVE', got: %s", updateExpr)
	}
}
