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
	// TODO: Add condition to prevent overwrite
	return op
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
	if g.entity.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the entity", nil)
	}

	// TODO: Implement actual DynamoDB get operation
	return &GetResponse{
		Data: make(map[string]interface{}),
	}, nil
}

// Params returns the DynamoDB parameters without executing
func (g *GetOperation) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(g.entity)
	return builder.BuildGetItemParams(g.keys, g.options)
}

// PutOperation represents a put operation
type PutOperation struct {
	entity  *Entity
	item    Item
	options *PutOptions
	ctx     context.Context
}

// Go executes the put operation
func (p *PutOperation) Go() (*PutResponse, error) {
	if p.entity.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the entity", nil)
	}

	// TODO: Implement actual DynamoDB put operation
	return &PutResponse{
		Data: make(map[string]interface{}),
	}, nil
}

// Params returns the DynamoDB parameters without executing
func (p *PutOperation) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(p.entity)
	return builder.BuildPutItemParams(p.item, p.options)
}

// UpdateOperation represents an update operation
type UpdateOperation struct {
	entity  *Entity
	keys    Keys
	setOps  map[string]interface{}
	addOps  map[string]interface{}
	delOps  map[string]interface{}
	remOps  []string
	options *UpdateOptions
	ctx     context.Context
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

// Go executes the update operation
func (u *UpdateOperation) Go() (*UpdateResponse, error) {
	if u.entity.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the entity", nil)
	}

	// TODO: Implement actual DynamoDB update operation
	return &UpdateResponse{
		Data: make(map[string]interface{}),
	}, nil
}

// Params returns the DynamoDB parameters without executing
func (u *UpdateOperation) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(u.entity)
	return builder.BuildUpdateItemParams(u.keys, u.setOps, u.addOps, u.remOps, u.options)
}

// DeleteOperation represents a delete operation
type DeleteOperation struct {
	entity  *Entity
	keys    Keys
	options *DeleteOptions
	ctx     context.Context
}

// Go executes the delete operation
func (d *DeleteOperation) Go() (*DeleteResponse, error) {
	if d.entity.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the entity", nil)
	}

	// TODO: Implement actual DynamoDB delete operation
	return &DeleteResponse{
		Data: make(map[string]interface{}),
	}, nil
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
	if s.entity.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the entity", nil)
	}

	// TODO: Implement actual DynamoDB scan operation
	return &ScanResponse{
		Data: make([]map[string]interface{}, 0),
	}, nil
}

// Params returns the DynamoDB parameters without executing
func (s *ScanOperation) Params() (map[string]interface{}, error) {
	// TODO: Build and return Scan parameters
	return make(map[string]interface{}), nil
}
