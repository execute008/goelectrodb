package electrodb

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ExpressionBuilder builds DynamoDB filter and condition expressions
type ExpressionBuilder struct {
	names      map[string]string
	values     map[string]types.AttributeValue
	expression string
	nameCount  int
	valueCount int
	attributes map[string]*AttributeDefinition
}

// NewExpressionBuilder creates a new expression builder
func NewExpressionBuilder(attributes map[string]*AttributeDefinition) *ExpressionBuilder {
	return &ExpressionBuilder{
		names:      make(map[string]string),
		values:     make(map[string]types.AttributeValue),
		expression: "",
		nameCount:  0,
		valueCount: 0,
		attributes: attributes,
	}
}

// AttributeRef represents a reference to an attribute in an expression
type AttributeRef struct {
	builder *ExpressionBuilder
	name    string
}

// OperationBuilder provides filter operation methods
type OperationBuilder struct {
	builder *ExpressionBuilder
}

// WhereCallback is a function that builds a where clause
type WhereCallback func(attrs map[string]*AttributeRef, ops *OperationBuilder) string

// addName adds an attribute name to the expression
func (eb *ExpressionBuilder) addName(name string) string {
	placeholder := fmt.Sprintf("#attr%d", eb.nameCount)
	eb.nameCount++
	eb.names[placeholder] = name
	return placeholder
}

// addValue adds a value to the expression
func (eb *ExpressionBuilder) addValue(value interface{}) (string, error) {
	placeholder := fmt.Sprintf(":val%d", eb.valueCount)
	eb.valueCount++

	av, err := marshalValue(value)
	if err != nil {
		return "", err
	}

	eb.values[placeholder] = av
	return placeholder, nil
}

// marshalValue marshals a Go value to a DynamoDB attribute value
func marshalValue(value interface{}) (types.AttributeValue, error) {
	switch v := value.(type) {
	case string:
		return &types.AttributeValueMemberS{Value: v}, nil
	case int:
		return &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", v)}, nil
	case int32:
		return &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", v)}, nil
	case int64:
		return &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", v)}, nil
	case float32:
		return &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", v)}, nil
	case float64:
		return &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", v)}, nil
	case bool:
		return &types.AttributeValueMemberBOOL{Value: v}, nil
	case nil:
		return &types.AttributeValueMemberNULL{Value: true}, nil
	default:
		return &types.AttributeValueMemberS{Value: fmt.Sprintf("%v", v)}, nil
	}
}

// AddExpression adds an expression to the builder
func (eb *ExpressionBuilder) AddExpression(expr string) {
	if eb.expression == "" {
		eb.expression = expr
	} else {
		// Wrap existing expression in parentheses if not already
		if !strings.HasPrefix(eb.expression, "(") {
			eb.expression = fmt.Sprintf("(%s)", eb.expression)
		}
		eb.expression = fmt.Sprintf("%s AND (%s)", eb.expression, expr)
	}
}

// Build returns the built expression and attributes
func (eb *ExpressionBuilder) Build() (string, map[string]string, map[string]types.AttributeValue) {
	return eb.expression, eb.names, eb.values
}

// GetExpression returns just the expression string
func (eb *ExpressionBuilder) GetExpression() string {
	return eb.expression
}

// Eq creates an equals condition
func (ar *AttributeRef) Eq(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s = %s", nameRef, valueRef)
}

// Ne creates a not-equals condition
func (ar *AttributeRef) Ne(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s <> %s", nameRef, valueRef)
}

// Gt creates a greater-than condition
func (ar *AttributeRef) Gt(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s > %s", nameRef, valueRef)
}

// Gte creates a greater-than-or-equal condition
func (ar *AttributeRef) Gte(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s >= %s", nameRef, valueRef)
}

// Lt creates a less-than condition
func (ar *AttributeRef) Lt(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s < %s", nameRef, valueRef)
}

// Lte creates a less-than-or-equal condition
func (ar *AttributeRef) Lte(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s <= %s", nameRef, valueRef)
}

// Between creates a between condition
func (ar *AttributeRef) Between(start, end interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	startRef, err := ar.builder.addValue(start)
	if err != nil {
		return ""
	}
	endRef, err := ar.builder.addValue(end)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("(%s BETWEEN %s AND %s)", nameRef, startRef, endRef)
}

// Contains creates a contains condition
func (ar *AttributeRef) Contains(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("contains(%s, %s)", nameRef, valueRef)
}

// Begins creates a begins_with condition
func (ar *AttributeRef) Begins(value interface{}) string {
	nameRef := ar.builder.addName(ar.name)
	valueRef, err := ar.builder.addValue(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("begins_with(%s, %s)", nameRef, valueRef)
}

// Operation methods

// Exists creates an attribute_exists condition
func (ob *OperationBuilder) Exists(attr *AttributeRef) string {
	nameRef := ob.builder.addName(attr.name)
	return fmt.Sprintf("attribute_exists(%s)", nameRef)
}

// NotExists creates an attribute_not_exists condition
func (ob *OperationBuilder) NotExists(attr *AttributeRef) string {
	nameRef := ob.builder.addName(attr.name)
	return fmt.Sprintf("attribute_not_exists(%s)", nameRef)
}

// Size returns the size of an attribute
func (ob *OperationBuilder) Size(attr *AttributeRef) string {
	nameRef := ob.builder.addName(attr.name)
	return fmt.Sprintf("size(%s)", nameRef)
}

// AttributeType checks the type of an attribute
func (ob *OperationBuilder) AttributeType(attr *AttributeRef, typeName string) string {
	nameRef := ob.builder.addName(attr.name)
	typeRef, _ := ob.builder.addValue(typeName)
	return fmt.Sprintf("attribute_type(%s, %s)", nameRef, typeRef)
}

// buildAttributeRefs builds attribute references for the where callback
func (eb *ExpressionBuilder) buildAttributeRefs() map[string]*AttributeRef {
	refs := make(map[string]*AttributeRef)
	for name := range eb.attributes {
		refs[name] = &AttributeRef{
			builder: eb,
			name:    name,
		}
	}
	return refs
}

// BuildWhereExpression builds an expression from a where callback
func (eb *ExpressionBuilder) BuildWhereExpression(callback WhereCallback) error {
	attrs := eb.buildAttributeRefs()
	ops := &OperationBuilder{builder: eb}

	expression := callback(attrs, ops)
	if expression == "" {
		return nil
	}

	eb.AddExpression(expression)
	return nil
}

// FilterBuilder adds filter expression support to operations
type FilterBuilder struct {
	builder    *ExpressionBuilder
	filterExpr string
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder(attributes map[string]*AttributeDefinition) *FilterBuilder {
	return &FilterBuilder{
		builder: NewExpressionBuilder(attributes),
	}
}

// Where adds a where clause
func (fb *FilterBuilder) Where(callback WhereCallback) error {
	return fb.builder.BuildWhereExpression(callback)
}

// Build returns the filter expression and attributes
func (fb *FilterBuilder) Build() (string, map[string]string, map[string]types.AttributeValue) {
	return fb.builder.Build()
}

// ConditionBuilder adds condition expression support
type ConditionBuilder struct {
	builder *ExpressionBuilder
}

// NewConditionBuilder creates a new condition builder
func NewConditionBuilder(attributes map[string]*AttributeDefinition) *ConditionBuilder {
	return &ConditionBuilder{
		builder: NewExpressionBuilder(attributes),
	}
}

// Where adds a condition expression
func (cb *ConditionBuilder) Where(callback WhereCallback) error {
	return cb.builder.BuildWhereExpression(callback)
}

// Build returns the condition expression and attributes
func (cb *ConditionBuilder) Build() (string, map[string]string, map[string]types.AttributeValue) {
	return cb.builder.Build()
}

// Merge merges expression names and values into existing maps
func MergeExpressionAttributes(
	existingNames map[string]string,
	existingValues map[string]types.AttributeValue,
	newNames map[string]string,
	newValues map[string]types.AttributeValue,
) (map[string]string, map[string]types.AttributeValue) {
	// Merge names
	for k, v := range newNames {
		existingNames[k] = v
	}

	// Merge values
	for k, v := range newValues {
		existingValues[k] = v
	}

	return existingNames, existingValues
}
