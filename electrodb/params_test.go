package electrodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestBuildGetItemParams(t *testing.T) {
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
				SK: &FacetDefinition{Field: "sk", Facets: []string{"name"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	builder := NewParamsBuilder(entity)

	keys := Keys{
		"id":   "test-123",
		"name": "John Doe",
	}

	params, err := builder.BuildGetItemParams(keys, nil)
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	if params["TableName"] != "TestTable" {
		t.Errorf("Expected TableName 'TestTable', got '%v'", params["TableName"])
	}

	keyMap, ok := params["Key"].(map[string]types.AttributeValue)
	if !ok {
		t.Fatal("Key is not a map")
	}

	pkVal, ok := keyMap["pk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("PK is not a string value")
	}

	if pkVal.Value != "$TestService#TestEntity#id_test-123" {
		t.Errorf("Expected PK '$TestService#TestEntity#id_test-123', got '%s'", pkVal.Value)
	}

	skVal, ok := keyMap["sk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("SK is not a string value")
	}

	if skVal.Value != "$TestService#TestEntity#name_John Doe" {
		t.Errorf("Expected SK '$TestService#TestEntity#name_John Doe', got '%s'", skVal.Value)
	}
}

func TestBuildPutItemParams(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"name":  {Type: AttributeTypeString, Required: true},
			"email": {Type: AttributeTypeString, Required: false},
			"count": {
				Type:     AttributeTypeNumber,
				Required: false,
				Default: func() interface{} {
					return 0
				},
			},
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

	builder := NewParamsBuilder(entity)

	item := Item{
		"id":    "test-123",
		"name":  "John Doe",
		"email": "john@example.com",
	}

	params, err := builder.BuildPutItemParams(item, nil)
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	if params["TableName"] != "TestTable" {
		t.Errorf("Expected TableName 'TestTable', got '%v'", params["TableName"])
	}

	itemMap, ok := params["Item"].(map[string]types.AttributeValue)
	if !ok {
		t.Fatal("Item is not a map")
	}

	// Check that pk was added
	if _, exists := itemMap["pk"]; !exists {
		t.Error("Expected pk to be added to item")
	}

	// Check that default was applied
	if _, exists := itemMap["count"]; !exists {
		t.Error("Expected count default to be applied")
	}
}

func TestBuildPutItemParamsMissingRequired(t *testing.T) {
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

	builder := NewParamsBuilder(entity)

	item := Item{
		"id": "test-123",
		// name is missing but required
	}

	_, err = builder.BuildPutItemParams(item, nil)
	if err == nil {
		t.Fatal("Expected error for missing required attribute")
	}

	electroErr, ok := err.(*ElectroError)
	if !ok {
		t.Fatal("Expected ElectroError type")
	}

	if electroErr.Code != "MissingAttribute" {
		t.Errorf("Expected error code 'MissingAttribute', got '%s'", electroErr.Code)
	}
}

func TestBuildUpdateItemParams(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"name":  {Type: AttributeTypeString, Required: false},
			"count": {Type: AttributeTypeNumber, Required: false},
			"tags":  {Type: AttributeTypeList, Required: false},
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

	builder := NewParamsBuilder(entity)

	keys := Keys{"id": "test-123"}
	setOps := map[string]interface{}{
		"name": "Updated Name",
	}
	addOps := map[string]interface{}{
		"count": 5,
	}
	remOps := []string{"tags"}

	params, err := builder.BuildUpdateItemParams(keys, setOps, addOps, remOps, nil)
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	if params["TableName"] != "TestTable" {
		t.Errorf("Expected TableName 'TestTable', got '%v'", params["TableName"])
	}

	updateExpr, ok := params["UpdateExpression"].(string)
	if !ok {
		t.Fatal("UpdateExpression is not a string")
	}

	// Check that SET, ADD, and REMOVE are in the expression
	if updateExpr == "" {
		t.Error("UpdateExpression is empty")
	}

	// Should have expression attribute names and values
	if _, exists := params["ExpressionAttributeNames"]; !exists {
		t.Error("ExpressionAttributeNames not found")
	}

	if _, exists := params["ExpressionAttributeValues"]; !exists {
		t.Error("ExpressionAttributeValues not found")
	}

	// Default return value should be ALL_NEW
	if params["ReturnValues"] != "ALL_NEW" {
		t.Errorf("Expected ReturnValues 'ALL_NEW', got '%v'", params["ReturnValues"])
	}
}

func TestBuildDeleteItemParams(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
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

	builder := NewParamsBuilder(entity)

	keys := Keys{"id": "test-123"}

	params, err := builder.BuildDeleteItemParams(keys, nil)
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	if params["TableName"] != "TestTable" {
		t.Errorf("Expected TableName 'TestTable', got '%v'", params["TableName"])
	}

	if _, exists := params["Key"]; !exists {
		t.Error("Key not found in params")
	}
}

func TestBuildQueryParams(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":       {Type: AttributeTypeString, Required: true},
			"mall":     {Type: AttributeTypeString, Required: true},
			"building": {Type: AttributeTypeString, Required: true},
			"unit":     {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
			"units": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"mall"}},
				SK:    &FacetDefinition{Field: "gsi1sk", Facets: []string{"building", "unit"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	builder := NewParamsBuilder(entity)

	pkFacets := []interface{}{"EastPointe"}

	params, err := builder.BuildQueryParams("units", pkFacets, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	if params["TableName"] != "TestTable" {
		t.Errorf("Expected TableName 'TestTable', got '%v'", params["TableName"])
	}

	if params["IndexName"] != "gsi1pk-gsi1sk-index" {
		t.Errorf("Expected IndexName 'gsi1pk-gsi1sk-index', got '%v'", params["IndexName"])
	}

	keyCondition, ok := params["KeyConditionExpression"].(string)
	if !ok {
		t.Fatal("KeyConditionExpression is not a string")
	}

	if keyCondition != "gsi1pk = :pk" {
		t.Errorf("Expected KeyConditionExpression 'gsi1pk = :pk', got '%s'", keyCondition)
	}
}

func TestBuildQueryParamsWithSortKey(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":       {Type: AttributeTypeString, Required: true},
			"mall":     {Type: AttributeTypeString, Required: true},
			"building": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"units": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK:    FacetDefinition{Field: "gsi1pk", Facets: []string{"mall"}},
				SK:    &FacetDefinition{Field: "gsi1sk", Facets: []string{"building"}},
			},
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	builder := NewParamsBuilder(entity)

	pkFacets := []interface{}{"EastPointe"}
	skCondition := &sortKeyCondition{
		operation: "begins_with",
		values:    []interface{}{"Building"},
	}

	params, err := builder.BuildQueryParams("units", pkFacets, skCondition, nil, nil)
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	keyCondition, ok := params["KeyConditionExpression"].(string)
	if !ok {
		t.Fatal("KeyConditionExpression is not a string")
	}

	expected := "gsi1pk = :pk AND begins_with(gsi1sk, :sk)"
	if keyCondition != expected {
		t.Errorf("Expected KeyConditionExpression '%s', got '%s'", expected, keyCondition)
	}
}
