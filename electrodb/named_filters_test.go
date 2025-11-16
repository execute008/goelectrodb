package electrodb

import (
	"strings"
	"testing"
)

func TestNamedFilter(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"category":  {Type: AttributeTypeString, Required: true},
			"price":     {Type: AttributeTypeNumber, Required: false},
			"inStock":   {Type: AttributeTypeBoolean, Required: false},
			"rating":    {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"productId"}},
			},
			"byCategory": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"category"}},
				SK:    &FacetDefinition{Field: "gsi1sk", Facets: []string{"productId"}},
			},
		},
		Filters: map[string]FilterFunc{
			"affordable": func(attr AttributeOperations, params map[string]interface{}) string {
				maxPrice := params["maxPrice"]
				return attr["price"].Lte(maxPrice)
			},
			"inStock": func(attr AttributeOperations, params map[string]interface{}) string {
				return attr["inStock"].Eq(true)
			},
			"highlyRated": func(attr AttributeOperations, params map[string]interface{}) string {
				minRating := params["minRating"]
				return attr["rating"].Gte(minRating)
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test query with named filter
	query := entity.Query("byCategory").Query("electronics").
		Filter("affordable", map[string]interface{}{
			"maxPrice": 500,
		})

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Verify filter expression exists
	filterExpr, ok := params["FilterExpression"].(string)
	if !ok || filterExpr == "" {
		t.Fatal("Expected FilterExpression to be set")
	}

	// Verify filter expression contains expected operator
	if !strings.Contains(filterExpr, "<=") {
		t.Errorf("Expected filter expression to contain '<=', got: %s", filterExpr)
	}
}

func TestMultipleNamedFilters(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"category":  {Type: AttributeTypeString, Required: true},
			"price":     {Type: AttributeTypeNumber, Required: false},
			"inStock":   {Type: AttributeTypeBoolean, Required: false},
			"rating":    {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"byCategory": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"category"}},
			},
		},
		Filters: map[string]FilterFunc{
			"affordable": func(attr AttributeOperations, params map[string]interface{}) string {
				maxPrice := params["maxPrice"]
				return attr["price"].Lte(maxPrice)
			},
			"inStock": func(attr AttributeOperations, params map[string]interface{}) string {
				return attr["inStock"].Eq(true)
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test query with multiple named filters
	query := entity.Query("byCategory").Query("electronics").
		Filter("affordable", map[string]interface{}{
			"maxPrice": 500,
		}).
		Filter("inStock", nil)

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Verify filter expression exists
	filterExpr, ok := params["FilterExpression"].(string)
	if !ok || filterExpr == "" {
		t.Fatal("Expected FilterExpression to be set")
	}

	// Verify both conditions are in the expression with AND
	if !strings.Contains(filterExpr, "AND") {
		t.Errorf("Expected filter expression to contain 'AND', got: %s", filterExpr)
	}
}

func TestNamedFilterWithWhereClause(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId":   {Type: AttributeTypeString, Required: true},
			"category":    {Type: AttributeTypeString, Required: true},
			"price":       {Type: AttributeTypeNumber, Required: false},
			"inStock":     {Type: AttributeTypeBoolean, Required: false},
			"description": {Type: AttributeTypeString, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"byCategory": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"category"}},
			},
		},
		Filters: map[string]FilterFunc{
			"affordable": func(attr AttributeOperations, params map[string]interface{}) string {
				maxPrice := params["maxPrice"]
				return attr["price"].Lte(maxPrice)
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test combining named filter with Where clause
	query := entity.Query("byCategory").Query("electronics").
		Filter("affordable", map[string]interface{}{
			"maxPrice": 500,
		}).
		Where(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
			return attrs["description"].Contains("premium")
		})

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Verify filter expression exists
	filterExpr, ok := params["FilterExpression"].(string)
	if !ok || filterExpr == "" {
		t.Fatal("Expected FilterExpression to be set")
	}

	// Verify both conditions are combined
	if !strings.Contains(filterExpr, "AND") {
		t.Errorf("Expected filter expression to contain 'AND', got: %s", filterExpr)
	}

	// Should have both the named filter and the where clause
	if !strings.Contains(filterExpr, "contains") {
		t.Errorf("Expected filter expression to contain 'contains', got: %s", filterExpr)
	}
}

func TestNonExistentNamedFilter(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"category":  {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"byCategory": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"category"}},
			},
		},
		Filters: map[string]FilterFunc{
			"existing": func(attr AttributeOperations, params map[string]interface{}) string {
				return attr["productId"].Eq(params["id"])
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test query with non-existent named filter (should be ignored)
	query := entity.Query("byCategory").Query("electronics").
		Filter("nonExistent", map[string]interface{}{})

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Should not have FilterExpression since the filter doesn't exist
	if _, ok := params["FilterExpression"]; ok {
		t.Error("Expected no FilterExpression for non-existent filter")
	}
}

func TestNamedFilterWithComplexExpression(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"category":  {Type: AttributeTypeString, Required: true},
			"price":     {Type: AttributeTypeNumber, Required: false},
			"rating":    {Type: AttributeTypeNumber, Required: false},
			"inStock":   {Type: AttributeTypeBoolean, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"byCategory": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"category"}},
			},
		},
		Filters: map[string]FilterFunc{
			"premium": func(attr AttributeOperations, params map[string]interface{}) string {
				minPrice := params["minPrice"]
				minRating := params["minRating"]
				return "(" + attr["price"].Gte(minPrice) + " AND " + attr["rating"].Gte(minRating) + ") AND " + attr["inStock"].Eq(true)
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test query with complex named filter
	query := entity.Query("byCategory").Query("electronics").
		Filter("premium", map[string]interface{}{
			"minPrice":  100,
			"minRating": 4.5,
		})

	params, err := query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// Verify filter expression exists
	filterExpr, ok := params["FilterExpression"].(string)
	if !ok || filterExpr == "" {
		t.Fatal("Expected FilterExpression to be set")
	}

	// Verify complex expression structure
	if !strings.Contains(filterExpr, "(") || !strings.Contains(filterExpr, ")") {
		t.Errorf("Expected filter expression to contain parentheses, got: %s", filterExpr)
	}

	if !strings.Contains(filterExpr, "AND") {
		t.Errorf("Expected filter expression to contain 'AND', got: %s", filterExpr)
	}
}
