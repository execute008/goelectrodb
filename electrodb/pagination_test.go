package electrodb

import (
	"testing"
)

func TestPagesMethod(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Product",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"productId": {Type: AttributeTypeString, Required: true},
			"category":  {Type: AttributeTypeString, Required: true},
			"name":      {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
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

	// Test that Pages method exists and can be called
	query := entity.Query("byCategory").Query("electronics")

	// Since we don't have a real DynamoDB client, we can't actually execute
	// but we can verify the method exists and builds correctly
	_, err = query.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	// The method should exist (compilation test)
	_ = query.Pages
}

func TestPageIteratorMethod(t *testing.T) {
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
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test that Page method exists and returns an iterator
	query := entity.Query("byCategory").Query("electronics")
	iterator := query.Page()

	if iterator == nil {
		t.Fatal("Expected non-nil iterator")
	}

	// Verify iterator has the query reference
	if iterator.query == nil {
		t.Error("Expected iterator to have query reference")
	}

	// Verify iterator options
	if iterator.options == nil {
		t.Error("Expected iterator to have options")
	}

	// Verify not done initially
	if iterator.done {
		t.Error("Expected iterator to not be done initially")
	}
}

func TestPageIteratorWithOptions(t *testing.T) {
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
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test Page with options
	query := entity.Query("byCategory").Query("electronics")
	iterator := query.Page(PagesOptions{
		MaxPages: 5,
		Limit:    10,
	})

	if iterator.maxPages != 5 {
		t.Errorf("Expected maxPages to be 5, got %d", iterator.maxPages)
	}

	if iterator.options.Limit == nil || *iterator.options.Limit != 10 {
		t.Errorf("Expected limit to be 10, got %v", iterator.options.Limit)
	}
}

func TestScanPagesMethod(t *testing.T) {
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

	// Test that Pages method exists on Scan
	scan := entity.Scan()

	// The method should exist (compilation test)
	_ = scan.Pages
}

func TestScanPageIterator(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Item",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"itemId": {Type: AttributeTypeString, Required: true},
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

	// Test that Page method exists on Scan and returns an iterator
	scan := entity.Scan()
	iterator := scan.Page()

	if iterator == nil {
		t.Fatal("Expected non-nil iterator")
	}

	// Verify iterator has the scan reference
	if iterator.scan == nil {
		t.Error("Expected iterator to have scan reference")
	}

	// Verify not done initially
	if iterator.done {
		t.Error("Expected iterator to not be done initially")
	}
}

func TestPagesOptionsDefaults(t *testing.T) {
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
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test Page without options (should use defaults)
	query := entity.Query("byCategory").Query("electronics")
	iterator := query.Page()

	// Should have default values
	if iterator.maxPages != 0 {
		t.Errorf("Expected maxPages to be 0 (unlimited), got %d", iterator.maxPages)
	}

	if iterator.options == nil {
		t.Error("Expected options to be initialized")
	}
}

func TestQueryWithLimitAndPages(t *testing.T) {
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
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test that query options are preserved when using Page
	limit := int32(25)
	query := entity.Query("byCategory").Query("electronics").
		Options(&QueryOptions{
			Limit: &limit,
		})

	iterator := query.Page()

	// The limit from query options should be inherited
	if iterator.options.Limit == nil || *iterator.options.Limit != 25 {
		t.Errorf("Expected inherited limit to be 25, got %v", iterator.options.Limit)
	}
}

func TestPageOptionsOverrideQueryOptions(t *testing.T) {
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
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test that PagesOptions can override query options
	queryLimit := int32(25)
	query := entity.Query("byCategory").Query("electronics").
		Options(&QueryOptions{
			Limit: &queryLimit,
		})

	iterator := query.Page(PagesOptions{
		Limit: 50,
	})

	// The PagesOptions limit should override the query limit
	if iterator.options.Limit == nil || *iterator.options.Limit != 50 {
		t.Errorf("Expected overridden limit to be 50, got %v", iterator.options.Limit)
	}
}
