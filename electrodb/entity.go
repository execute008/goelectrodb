package electrodb

import (
	"context"
	"fmt"
)

// Entity represents a DynamoDB entity with schema and operations
type Entity struct {
	schema *Schema
	config *Config
	client DynamoDBClient
	query  map[string]QueryBuilder
}

// NewEntity creates a new Entity instance
func NewEntity(schema *Schema, config *Config) (*Entity, error) {
	if schema == nil {
		return nil, NewElectroError("InvalidSchema", "Schema cannot be nil", nil)
	}

	if err := validateSchema(schema); err != nil {
		return nil, err
	}

	if config == nil {
		config = &Config{}
	}

	entity := &Entity{
		schema: schema,
		config: config,
		client: config.Client,
		query:  make(map[string]QueryBuilder),
	}

	// Initialize query builders for each index
	for accessPattern, index := range schema.Indexes {
		entity.query[accessPattern] = newQueryBuilder(entity, accessPattern, index)
	}

	return entity, nil
}

// validateSchema validates the entity schema
func validateSchema(schema *Schema) error {
	if schema.Service == "" {
		return NewElectroError("InvalidSchema", "Service name is required", nil)
	}

	if schema.Entity == "" {
		return NewElectroError("InvalidSchema", "Entity name is required", nil)
	}

	if schema.Table == "" {
		return NewElectroError("InvalidSchema", "Table name is required", nil)
	}

	if schema.Attributes == nil || len(schema.Attributes) == 0 {
		return NewElectroError("InvalidSchema", "At least one attribute is required", nil)
	}

	if schema.Indexes == nil || len(schema.Indexes) == 0 {
		return NewElectroError("InvalidSchema", "At least one index is required", nil)
	}

	// Validate that all facets reference existing attributes
	for indexName, index := range schema.Indexes {
		for _, facet := range index.PK.Facets {
			if _, exists := schema.Attributes[facet]; !exists {
				return NewElectroError("InvalidSchema",
					fmt.Sprintf("PK facet '%s' in index '%s' references non-existent attribute", facet, indexName), nil)
			}
		}

		if index.SK != nil {
			for _, facet := range index.SK.Facets {
				if _, exists := schema.Attributes[facet]; !exists {
					return NewElectroError("InvalidSchema",
						fmt.Sprintf("SK facet '%s' in index '%s' references non-existent attribute", facet, indexName), nil)
				}
			}
		}
	}

	return nil
}

// Get retrieves an item by its key
func (e *Entity) Get(keys Keys) *GetOperation {
	return &GetOperation{
		entity: e,
		keys:   keys,
		ctx:    context.Background(),
	}
}

// Put creates or replaces an item
func (e *Entity) Put(item Item) *PutOperation {
	return &PutOperation{
		entity: e,
		item:   item,
		ctx:    context.Background(),
	}
}

// Create creates a new item (fails if exists)
func (e *Entity) Create(item Item) *PutOperation {
	op := &PutOperation{
		entity: e,
		item:   item,
		ctx:    context.Background(),
	}

	// Add condition to prevent overwrite - only create if primary key doesn't exist
	// Find the primary key field from the primary index
	primaryIndex, exists := e.schema.Indexes["primary"]
	if exists && len(primaryIndex.PK.Facets) > 0 {
		// Use the first facet as the primary key attribute for the condition
		pkAttr := primaryIndex.PK.Facets[0]
		op.Condition(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
			return ops.NotExists(attrs[pkAttr])
		})
	}

	return op
}

// Upsert creates a new item or updates an existing one
// This is an alias for Put, which naturally upserts in DynamoDB
func (e *Entity) Upsert(item Item) *PutOperation {
	return e.Put(item)
}

// Update updates an existing item
func (e *Entity) Update(keys Keys) *UpdateOperation {
	return &UpdateOperation{
		entity:  e,
		keys:    keys,
		setOps:  make(map[string]interface{}),
		addOps:  make(map[string]interface{}),
		delOps:  make(map[string]interface{}),
		remOps:  []string{},
		ctx:     context.Background(),
	}
}

// Patch partially updates an item
func (e *Entity) Patch(keys Keys) *UpdateOperation {
	return e.Update(keys)
}

// Upsert creates or updates an item using UpdateItem
// Unlike Upsert(), which uses Put, this uses Update which allows partial updates
// This will create the item if it doesn't exist, or update specific attributes if it does
func (e *Entity) UpsertUpdate(keys Keys) *UpdateOperation {
	// UpdateItem in DynamoDB naturally upserts - it creates if doesn't exist
	return e.Update(keys)
}

// Delete deletes an item
func (e *Entity) Delete(keys Keys) *DeleteOperation {
	return &DeleteOperation{
		entity: e,
		keys:   keys,
		ctx:    context.Background(),
	}
}

// Remove is an alias for Delete
func (e *Entity) Remove(keys Keys) *DeleteOperation {
	return e.Delete(keys)
}

// Scan performs a table scan
func (e *Entity) Scan() *ScanOperation {
	return &ScanOperation{
		entity: e,
		ctx:    context.Background(),
	}
}

// Query returns a query builder for the specified access pattern
func (e *Entity) Query(accessPattern string) QueryBuilder {
	if qb, exists := e.query[accessPattern]; exists {
		return qb
	}
	return nil
}

// Schema returns the entity schema
func (e *Entity) Schema() *Schema {
	return e.schema
}

// GetOperation represents a get operation
type GetOperation struct {
	entity  *Entity
	keys    Keys
	options *GetOptions
	ctx     context.Context
}

// Go executes the get operation
func (g *GetOperation) Go() (*GetResponse, error) {
	executor := NewExecutionHelper(g.entity)
	return executor.ExecuteGetItem(g.ctx, g.keys, g.options)
}

// Params returns the DynamoDB parameters without executing
func (g *GetOperation) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(g.entity)
	return builder.BuildGetItemParams(g.keys, g.options)
}

// PutOperation represents a put operation
type PutOperation struct {
	entity           *Entity
	item             Item
	options          *PutOptions
	ctx              context.Context
	conditionBuilder *ConditionBuilder
}

// Condition adds a condition expression to the put operation
func (p *PutOperation) Condition(callback WhereCallback) *PutOperation {
	cb := NewConditionBuilder(p.entity.schema.Attributes)
	cb.Where(callback)
	p.conditionBuilder = cb
	return p
}

// Go executes the put operation
func (p *PutOperation) Go() (*PutResponse, error) {
	executor := NewExecutionHelper(p.entity)
	return executor.ExecutePutItem(p.ctx, p.item, p.options)
}

// Params returns the DynamoDB parameters without executing
func (p *PutOperation) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(p.entity)
	return builder.BuildPutItemParams(p.item, p.options)
}

// UpdateOperation represents an update operation
type UpdateOperation struct {
	entity           *Entity
	keys             Keys
	setOps           map[string]interface{}
	addOps           map[string]interface{}
	delOps           map[string]interface{}
	remOps           []string
	options          *UpdateOptions
	ctx              context.Context
	conditionBuilder *ConditionBuilder
}

// Set sets an attribute value
func (u *UpdateOperation) Set(updates map[string]interface{}) *UpdateOperation {
	for key, value := range updates {
		u.setOps[key] = value
	}
	return u
}

// Add adds to an attribute (for numbers and sets)
func (u *UpdateOperation) Add(updates map[string]interface{}) *UpdateOperation {
	for key, value := range updates {
		u.addOps[key] = value
	}
	return u
}

// Remove removes attributes
func (u *UpdateOperation) Remove(attributes []string) *UpdateOperation {
	u.remOps = append(u.remOps, attributes...)
	return u
}

// Condition adds a condition expression to the update operation
func (u *UpdateOperation) Condition(callback WhereCallback) *UpdateOperation {
	cb := NewConditionBuilder(u.entity.schema.Attributes)
	cb.Where(callback)
	u.conditionBuilder = cb
	return u
}

// Go executes the update operation
func (u *UpdateOperation) Go() (*UpdateResponse, error) {
	executor := NewExecutionHelper(u.entity)
	return executor.ExecuteUpdateItem(u.ctx, u.keys, u.setOps, u.addOps, u.remOps, u.options)
}

// Params returns the DynamoDB parameters without executing
func (u *UpdateOperation) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(u.entity)
	return builder.BuildUpdateItemParams(u.keys, u.setOps, u.addOps, u.remOps, u.options)
}

// DeleteOperation represents a delete operation
type DeleteOperation struct {
	entity           *Entity
	keys             Keys
	options          *DeleteOptions
	ctx              context.Context
	conditionBuilder *ConditionBuilder
}

// Condition adds a condition expression to the delete operation
func (d *DeleteOperation) Condition(callback WhereCallback) *DeleteOperation {
	cb := NewConditionBuilder(d.entity.schema.Attributes)
	cb.Where(callback)
	d.conditionBuilder = cb
	return d
}

// Go executes the delete operation
func (d *DeleteOperation) Go() (*DeleteResponse, error) {
	executor := NewExecutionHelper(d.entity)
	return executor.ExecuteDeleteItem(d.ctx, d.keys, d.options)
}

// Params returns the DynamoDB parameters without executing
func (d *DeleteOperation) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(d.entity)
	return builder.BuildDeleteItemParams(d.keys, d.options)
}

// ScanOperation represents a scan operation
type ScanOperation struct {
	entity  *Entity
	options *QueryOptions
	ctx     context.Context
}

// Go executes the scan operation
func (s *ScanOperation) Go() (*ScanResponse, error) {
	executor := NewExecutionHelper(s.entity)
	return executor.ExecuteScan(s.ctx, s.options)
}

// Params returns the DynamoDB parameters without executing
func (s *ScanOperation) Params() (map[string]interface{}, error) {
	// TODO: Build and return Scan parameters
	return make(map[string]interface{}), nil
}
