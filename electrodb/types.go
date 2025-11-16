package electrodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// AttributeType represents the type of an attribute
type AttributeType string

const (
	AttributeTypeString  AttributeType = "string"
	AttributeTypeNumber  AttributeType = "number"
	AttributeTypeBoolean AttributeType = "boolean"
	AttributeTypeEnum    AttributeType = "enum"
	AttributeTypeAny     AttributeType = "any"
	AttributeTypeList    AttributeType = "list"
	AttributeTypeMap     AttributeType = "map"
	AttributeTypeSet     AttributeType = "set"
)

// ValidationFunc is a function that validates an attribute value
type ValidationFunc func(value interface{}) error

// DefaultFunc is a function that returns a default value for an attribute
type DefaultFunc func() interface{}

// GetFunc is a function that transforms a value when reading from DynamoDB
type GetFunc func(value interface{}) interface{}

// SetFunc is a function that transforms a value when writing to DynamoDB
type SetFunc func(value interface{}) interface{}

// AttributeDefinition defines a single attribute in the schema
type AttributeDefinition struct {
	Type       AttributeType
	Required   bool
	Default    DefaultFunc
	Validate   ValidationFunc
	Field      string // DynamoDB field name (if different from attribute name)
	Get        GetFunc
	Set        SetFunc
	ReadOnly   bool
	Watch      []string // Attributes to watch for changes
	Label      string
	Cast       string
	Padding    *PaddingConfig
	Hidden     bool
	EnumValues []interface{} // For enum type
}

// PaddingConfig defines padding configuration for attributes
type PaddingConfig struct {
	Length int
	Char   string
}

// FacetDefinition defines partition or sort key facets
type FacetDefinition struct {
	Field   string
	Facets  []string
	Casing  *string // optional: "upper", "lower", "none", "default"
	Template *string
}

// IndexDefinition defines a primary or secondary index
type IndexDefinition struct {
	Index      *string          // GSI name (nil for primary index)
	PK         FacetDefinition  `json:"pk"`
	SK         *FacetDefinition `json:"sk,omitempty"`
	Collection *string          // Collection name for this index
	Type       *string          // "isolated" or "clustered"
}

// Schema defines the entity schema
type Schema struct {
	Service    string
	Entity     string
	Table      string
	Version    string
	Attributes map[string]*AttributeDefinition
	Indexes    map[string]*IndexDefinition
	Filters    map[string]FilterFunc
}

// FilterFunc is a custom filter function
type FilterFunc func(attr AttributeOperations, params map[string]interface{}) string

// AttributeOperations provides operations for filter building
type AttributeOperations map[string]*AttributeOperator

// AttributeOperator provides comparison operations for an attribute
type AttributeOperator struct {
	name    string
	builder *ExpressionBuilder
}

// Eq generates an equals filter expression
func (a *AttributeOperator) Eq(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return nameRef + " = " + valueRef
}

// Ne generates a not-equals filter expression
func (a *AttributeOperator) Ne(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return nameRef + " <> " + valueRef
}

// Gt generates a greater-than filter expression
func (a *AttributeOperator) Gt(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return nameRef + " > " + valueRef
}

// Gte generates a greater-than-or-equal filter expression
func (a *AttributeOperator) Gte(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return nameRef + " >= " + valueRef
}

// Lt generates a less-than filter expression
func (a *AttributeOperator) Lt(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return nameRef + " < " + valueRef
}

// Lte generates a less-than-or-equal filter expression
func (a *AttributeOperator) Lte(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return nameRef + " <= " + valueRef
}

// Between generates a BETWEEN filter expression
func (a *AttributeOperator) Between(low, high interface{}) string {
	nameRef := a.builder.addName(a.name)
	lowRef, _ := a.builder.addValue(low)
	highRef, _ := a.builder.addValue(high)
	return nameRef + " BETWEEN " + lowRef + " AND " + highRef
}

// Begins generates a begins_with filter expression
func (a *AttributeOperator) Begins(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return "begins_with(" + nameRef + ", " + valueRef + ")"
}

// Contains generates a contains filter expression
func (a *AttributeOperator) Contains(value interface{}) string {
	nameRef := a.builder.addName(a.name)
	valueRef, _ := a.builder.addValue(value)
	return "contains(" + nameRef + ", " + valueRef + ")"
}

// Config holds entity configuration
type Config struct {
	Client      DynamoDBClient
	Table       *string
	Listeners   []EventListener
	Logger      Logger
	Identifiers *IdentifierConfig
}

// IdentifierConfig defines entity identifiers
type IdentifierConfig struct {
	Entity  string
	Version string
}

// DynamoDBClient is an interface for DynamoDB operations
type DynamoDBClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	TransactGetItems(ctx context.Context, params *dynamodb.TransactGetItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactGetItemsOutput, error)
}

// EventListener is an interface for event listeners
type EventListener interface {
	OnQuery(params map[string]interface{})
	OnResults(results interface{})
}

// Logger is an interface for logging
type Logger interface {
	Info(message string, data map[string]interface{})
	Warn(message string, data map[string]interface{})
	Error(message string, data map[string]interface{})
}

// QueryOptions defines options for query execution
type QueryOptions struct {
	Limit       *int32
	Pages       *int
	Cursor      *string
	Raw         bool
	Attributes  []string
	Order       *string // "asc" or "desc"
	Concurrent  *int
	IgnoreCursor bool
}

// PutOptions defines options for put operations
type PutOptions struct {
	Response   *string // "none", "all_old", "all_new"
	Attributes []string
	Raw        bool
}

// UpdateOptions defines options for update operations
type UpdateOptions struct {
	Response   *string
	Attributes []string
	Raw        bool
}

// DeleteOptions defines options for delete operations
type DeleteOptions struct {
	Response   *string
	Attributes []string
	Raw        bool
}

// GetOptions defines options for get operations
type GetOptions struct {
	Attributes []string
	Raw        bool
}

// QueryResponse represents a query response
type QueryResponse struct {
	Data   []map[string]interface{}
	Cursor *string
}

// PutResponse represents a put response
type PutResponse struct {
	Data map[string]interface{}
}

// UpdateResponse represents an update response
type UpdateResponse struct {
	Data map[string]interface{}
}

// DeleteResponse represents a delete response
type DeleteResponse struct {
	Data map[string]interface{}
}

// GetResponse represents a get response
type GetResponse struct {
	Data map[string]interface{}
}

// ScanResponse represents a scan response
type ScanResponse struct {
	Data   []map[string]interface{}
	Cursor *string
}

// BatchGetResponse represents a batch get response
type BatchGetResponse struct {
	Data        []map[string]interface{}
	Unprocessed []Keys
}

// BatchWriteResponse represents a batch write response
type BatchWriteResponse struct {
	Unprocessed struct {
		Puts    []Item
		Deletes []Keys
	}
}

// TransactionResponse represents a transaction response
type TransactionResponse struct {
	Data []map[string]interface{}
}

// Item represents a DynamoDB item
type Item map[string]interface{}

// Keys represents the keys for a DynamoDB item
type Keys map[string]interface{}

// UpdateData represents update data
type UpdateData map[string]interface{}

// ElectroError represents an error from ElectroDB
type ElectroError struct {
	Code    string
	Message string
	Cause   error
	Time    time.Time
}

func (e *ElectroError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// NewElectroError creates a new ElectroError
func NewElectroError(code, message string, cause error) *ElectroError {
	return &ElectroError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Time:    time.Now(),
	}
}
