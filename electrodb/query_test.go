package electrodb

import (
	"strings"
	"testing"
)

func TestQueryWithWhereClause(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "User",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"userId":   {Type: AttributeTypeString, Required: true},
			"email":    {Type: AttributeTypeString, Required: true},
			"age":      {Type: AttributeTypeNumber, Required: false},
			"active":   {Type: AttributeTypeBoolean, Required: false},
			"tenantId": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"userId"}},
			},
			"byTenant": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"tenantId"}},
				SK:    &FacetDefinition{Field: "gsi1sk", Facets: []string{"email"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test query with filter expression
	query := entity.Query("byTenant").Query("tenant1").Where(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return attrs["age"].Gt(21) + " AND " + attrs["active"].Eq(true)
	})

	// Get params to verify the filter expression is built correctly
	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Verify filter expression exists
	filterExpr, ok := params["FilterExpression"].(string)
	if !ok || filterExpr == "" {
		t.Fatal("Expected FilterExpression to be set")
	}

	// Verify filter expression contains expected operators
	if !strings.Contains(filterExpr, ">") {
		t.Errorf("Expected filter expression to contain '>', got: %s", filterExpr)
	}

	if !strings.Contains(filterExpr, "AND") {
		t.Errorf("Expected filter expression to contain 'AND', got: %s", filterExpr)
	}

	// Verify expression attribute names exist
	exprAttrNames, ok := params["ExpressionAttributeNames"].(map[string]string)
	if !ok {
		t.Fatal("Expected ExpressionAttributeNames to be set")
	}

	// Should have attribute names for 'age' and 'active'
	if len(exprAttrNames) < 2 {
		t.Errorf("Expected at least 2 expression attribute names, got %d", len(exprAttrNames))
	}

	// Verify expression attribute values exist
	exprAttrValues := params["ExpressionAttributeValues"]
	if exprAttrValues == nil {
		t.Fatal("Expected ExpressionAttributeValues to be set")
	}
}

func TestQueryWithComplexWhereClause(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId":   {Type: AttributeTypeString, Required: true},
			"category":    {Type: AttributeTypeString, Required: true},
			"name":        {Type: AttributeTypeString, Required: true},
			"price":       {Type: AttributeTypeNumber, Required: false},
			"inStock":     {Type: AttributeTypeBoolean, Required: false},
			"description": {Type: AttributeTypeString, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"productId"}},
			},
			"byCategory": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"category"}},
				SK:    &FacetDefinition{Field: "gsi1sk", Facets: []string{"name"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test complex filter with multiple conditions
	query := entity.Query("byCategory").Query("electronics").Where(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		priceCondition := attrs["price"].Between(100, 500)
		stockCondition := attrs["inStock"].Eq(true)
		descCondition := attrs["description"].Contains("premium")
		return "(" + priceCondition + " AND " + stockCondition + ") AND " + descCondition
	})

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	filterExpr, ok := params["FilterExpression"].(string)
	if !ok || filterExpr == "" {
		t.Fatal("Expected FilterExpression to be set")
	}

	// Verify BETWEEN operator
	if !strings.Contains(filterExpr, "BETWEEN") {
		t.Errorf("Expected filter expression to contain 'BETWEEN', got: %s", filterExpr)
	}

	// Verify contains function
	if !strings.Contains(filterExpr, "contains") {
		t.Errorf("Expected filter expression to contain 'contains', got: %s", filterExpr)
	}

	// Verify parentheses for grouping
	if !strings.Contains(filterExpr, "(") || !strings.Contains(filterExpr, ")") {
		t.Errorf("Expected filter expression to contain parentheses, got: %s", filterExpr)
	}
}

func TestQueryWithFunctionBasedFilters(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Record",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"recordId": {Type: AttributeTypeString, Required: true},
			"email":    {Type: AttributeTypeString, Required: false},
			"metadata": {Type: AttributeTypeMap, Required: false},
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

	// Test with function-based filters (exists, not_exists, begins_with)
	query := entity.Query("primary").Query("record1").Where(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		hasMetadata := ops.Exists(attrs["metadata"])
		emailFilter := attrs["email"].Begins("admin@")
		return hasMetadata + " AND " + emailFilter
	})

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	filterExpr, ok := params["FilterExpression"].(string)
	if !ok || filterExpr == "" {
		t.Fatal("Expected FilterExpression to be set")
	}

	// Verify attribute_exists function
	if !strings.Contains(filterExpr, "attribute_exists") {
		t.Errorf("Expected filter expression to contain 'attribute_exists', got: %s", filterExpr)
	}

	// Verify begins_with function
	if !strings.Contains(filterExpr, "begins_with") {
		t.Errorf("Expected filter expression to contain 'begins_with', got: %s", filterExpr)
	}
}

func TestQueryParamsWithoutWhereClause(t *testing.T) {
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

	// Test query without filter - should work as before
	query := entity.Query("primary").Query("item1")

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Should not have FilterExpression
	if _, ok := params["FilterExpression"]; ok {
		t.Error("Expected no FilterExpression when Where is not called")
	}

	// Should still have required params
	if params["TableName"] == nil {
		t.Error("Expected TableName to be set")
	}

	if params["KeyConditionExpression"] == nil {
		t.Error("Expected KeyConditionExpression to be set")
	}
}
