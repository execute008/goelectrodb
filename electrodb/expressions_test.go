package electrodb

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestExpressionBuilderEq(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"name":  {Type: AttributeTypeString},
		"age":   {Type: AttributeTypeNumber},
		"email": {Type: AttributeTypeString},
	}

	builder := NewExpressionBuilder(attributes)

	err := builder.BuildWhereExpression(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return attrs["name"].Eq("John Doe")
	})

	if err != nil {
		t.Fatalf("Failed to build expression: %v", err)
	}

	expr, names, values := builder.Build()

	if expr == "" {
		t.Error("Expected non-empty expression")
	}

	if !strings.Contains(expr, "=") {
		t.Errorf("Expected expression to contain '=', got: %s", expr)
	}

	if len(names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(names))
	}

	if len(values) != 1 {
		t.Errorf("Expected 1 value, got %d", len(values))
	}
}

func TestExpressionBuilderMultipleConditions(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"name":   {Type: AttributeTypeString},
		"age":    {Type: AttributeTypeNumber},
		"active": {Type: AttributeTypeBoolean},
	}

	builder := NewExpressionBuilder(attributes)

	err := builder.BuildWhereExpression(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		nameCondition := attrs["name"].Eq("John")
		ageCondition := attrs["age"].Gt(25)
		return nameCondition + " AND " + ageCondition
	})

	if err != nil {
		t.Fatalf("Failed to build expression: %v", err)
	}

	expr, names, values := builder.Build()

	if expr == "" {
		t.Error("Expected non-empty expression")
	}

	if !strings.Contains(expr, "AND") {
		t.Errorf("Expected expression to contain 'AND', got: %s", expr)
	}

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
}

func TestExpressionBuilderComparisons(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"score": {Type: AttributeTypeNumber},
		"name":  {Type: AttributeTypeString},
	}

	tests := []struct {
		name     string
		callback WhereCallback
		contains string
	}{
		{
			name: "greater than",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return attrs["score"].Gt(50)
			},
			contains: ">",
		},
		{
			name: "greater than or equal",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return attrs["score"].Gte(50)
			},
			contains: ">=",
		},
		{
			name: "less than",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return attrs["score"].Lt(100)
			},
			contains: "<",
		},
		{
			name: "less than or equal",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return attrs["score"].Lte(100)
			},
			contains: "<=",
		},
		{
			name: "not equal",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return attrs["name"].Ne("Admin")
			},
			contains: "<>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewExpressionBuilder(attributes)
			err := builder.BuildWhereExpression(tt.callback)
			if err != nil {
				t.Fatalf("Failed to build expression: %v", err)
			}

			expr, _, _ := builder.Build()
			if !strings.Contains(expr, tt.contains) {
				t.Errorf("Expected expression to contain '%s', got: %s", tt.contains, expr)
			}
		})
	}
}

func TestExpressionBuilderBetween(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"age": {Type: AttributeTypeNumber},
	}

	builder := NewExpressionBuilder(attributes)

	err := builder.BuildWhereExpression(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return attrs["age"].Between(18, 65)
	})

	if err != nil {
		t.Fatalf("Failed to build expression: %v", err)
	}

	expr, names, values := builder.Build()

	if !strings.Contains(expr, "BETWEEN") {
		t.Errorf("Expected expression to contain 'BETWEEN', got: %s", expr)
	}

	if !strings.Contains(expr, "AND") {
		t.Errorf("Expected expression to contain 'AND', got: %s", expr)
	}

	if len(names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(names))
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 values (start and end), got %d", len(values))
	}
}

func TestExpressionBuilderFunctions(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"email":       {Type: AttributeTypeString},
		"description": {Type: AttributeTypeString},
		"tags":        {Type: AttributeTypeList},
	}

	tests := []struct {
		name     string
		callback WhereCallback
		contains string
	}{
		{
			name: "begins_with",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return attrs["email"].Begins("john@")
			},
			contains: "begins_with",
		},
		{
			name: "contains",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return attrs["description"].Contains("important")
			},
			contains: "contains",
		},
		{
			name: "attribute_exists",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return ops.Exists(attrs["tags"])
			},
			contains: "attribute_exists",
		},
		{
			name: "attribute_not_exists",
			callback: func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
				return ops.NotExists(attrs["tags"])
			},
			contains: "attribute_not_exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewExpressionBuilder(attributes)
			err := builder.BuildWhereExpression(tt.callback)
			if err != nil {
				t.Fatalf("Failed to build expression: %v", err)
			}

			expr, _, _ := builder.Build()
			if !strings.Contains(expr, tt.contains) {
				t.Errorf("Expected expression to contain '%s', got: %s", tt.contains, expr)
			}
		})
	}
}

func TestExpressionBuilderComplexExpression(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"name":   {Type: AttributeTypeString},
		"age":    {Type: AttributeTypeNumber},
		"email":  {Type: AttributeTypeString},
		"active": {Type: AttributeTypeBoolean},
	}

	builder := NewExpressionBuilder(attributes)

	err := builder.BuildWhereExpression(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		// (name = "John" AND age > 25) AND (email begins_with "john@" AND active = true)
		condition1 := attrs["name"].Eq("John") + " AND " + attrs["age"].Gt(25)
		condition2 := attrs["email"].Begins("john@") + " AND " + attrs["active"].Eq(true)
		return "(" + condition1 + ") AND (" + condition2 + ")"
	})

	if err != nil {
		t.Fatalf("Failed to build expression: %v", err)
	}

	expr, names, values := builder.Build()

	if expr == "" {
		t.Error("Expected non-empty expression")
	}

	// Should have 4 attributes referenced
	if len(names) != 4 {
		t.Errorf("Expected 4 names, got %d", len(names))
	}

	// Should have 4 values
	if len(values) != 4 {
		t.Errorf("Expected 4 values, got %d", len(values))
	}

	// Check structure
	if !strings.Contains(expr, "begins_with") {
		t.Error("Expected expression to contain begins_with function")
	}
}

func TestFilterBuilder(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"status": {Type: AttributeTypeString},
		"count":  {Type: AttributeTypeNumber},
	}

	fb := NewFilterBuilder(attributes)

	err := fb.Where(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return attrs["status"].Eq("active") + " AND " + attrs["count"].Gt(0)
	})

	if err != nil {
		t.Fatalf("Failed to build filter: %v", err)
	}

	expr, names, values := fb.Build()

	if expr == "" {
		t.Error("Expected non-empty expression")
	}

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
}

func TestConditionBuilder(t *testing.T) {
	attributes := map[string]*AttributeDefinition{
		"id":      {Type: AttributeTypeString},
		"version": {Type: AttributeTypeNumber},
	}

	cb := NewConditionBuilder(attributes)

	err := cb.Where(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		return ops.NotExists(attrs["id"])
	})

	if err != nil {
		t.Fatalf("Failed to build condition: %v", err)
	}

	expr, _, _ := cb.Build()

	if !strings.Contains(expr, "attribute_not_exists") {
		t.Errorf("Expected expression to contain attribute_not_exists, got: %s", expr)
	}
}

func TestMergeExpressionAttributes(t *testing.T) {
	existingNames := map[string]string{
		"#attr0": "id",
	}

	existingValues := map[string]types.AttributeValue{
		":val0": &types.AttributeValueMemberS{Value: "123"},
	}

	newNames := map[string]string{
		"#attr1": "name",
	}

	newValues := map[string]types.AttributeValue{
		":val1": &types.AttributeValueMemberS{Value: "John"},
	}

	mergedNames, mergedValues := MergeExpressionAttributes(existingNames, existingValues, newNames, newValues)

	if len(mergedNames) != 2 {
		t.Errorf("Expected 2 merged names, got %d", len(mergedNames))
	}

	if len(mergedValues) != 2 {
		t.Errorf("Expected 2 merged values, got %d", len(mergedValues))
	}

	if mergedNames["#attr0"] != "id" {
		t.Error("Expected existing name to be preserved")
	}

	if mergedNames["#attr1"] != "name" {
		t.Error("Expected new name to be added")
	}
}
