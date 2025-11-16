package electrodb

import (
	"context"
	"fmt"
)

// Service manages multiple entities in a single table
type Service struct {
	name       string
	entities   map[string]*Entity
	client     DynamoDBClient
	table      *string
	config     *Config
	collection map[string]*Collection
}

// ServiceConfig holds configuration for a service
type ServiceConfig struct {
	Client DynamoDBClient
	Table  *string
}

// Collection represents a cross-entity query collection
type Collection struct {
	name     string
	index    string
	entities []string
	service  *Service
}

// NewService creates a new Service instance
func NewService(name string, config *ServiceConfig) *Service {
	if config == nil {
		config = &ServiceConfig{}
	}

	return &Service{
		name:       name,
		entities:   make(map[string]*Entity),
		client:     config.Client,
		table:      config.Table,
		config:     &Config{Client: config.Client, Table: config.Table},
		collection: make(map[string]*Collection),
	}
}

// Join adds an entity to the service
func (s *Service) Join(entity *Entity) error {
	if entity == nil {
		return NewElectroError("InvalidEntity", "Entity cannot be nil", nil)
	}

	entityName := entity.schema.Entity
	if entityName == "" {
		return NewElectroError("InvalidEntity", "Entity must have a name", nil)
	}

	// Check if entity already exists
	if _, exists := s.entities[entityName]; exists {
		return NewElectroError("DuplicateEntity",
			fmt.Sprintf("Entity '%s' already exists in service", entityName), nil)
	}

	// Update entity config with service's client and table
	if entity.config == nil {
		entity.config = &Config{}
	}

	if entity.config.Client == nil && s.client != nil {
		entity.config.Client = s.client
		entity.client = s.client
	}

	if entity.config.Table == nil && s.table != nil {
		entity.config.Table = s.table
	}

	// Add to entities map
	s.entities[entityName] = entity

	// Build collections from entity indexes
	s.buildCollections(entity)

	return nil
}

// buildCollections creates collections from entity indexes
func (s *Service) buildCollections(entity *Entity) {
	for indexName, index := range entity.schema.Indexes {
		collectionName := indexName
		if index.Collection != nil {
			collectionName = *index.Collection
		}

		// Create or update collection
		collection, exists := s.collection[collectionName]
		if !exists {
			indexStr := ""
			if index.Index != nil {
				indexStr = *index.Index
			}

			collection = &Collection{
				name:     collectionName,
				index:    indexStr,
				entities: make([]string, 0),
				service:  s,
			}
			s.collection[collectionName] = collection
		}

		// Add entity to collection if not already present
		found := false
		for _, e := range collection.entities {
			if e == entity.schema.Entity {
				found = true
				break
			}
		}
		if !found {
			collection.entities = append(collection.entities, entity.schema.Entity)
		}
	}
}

// Entities returns all entities in the service
func (s *Service) Entities() map[string]*Entity {
	return s.entities
}

// Entity returns a specific entity by name
func (s *Service) Entity(name string) (*Entity, error) {
	entity, exists := s.entities[name]
	if !exists {
		return nil, NewElectroError("EntityNotFound",
			fmt.Sprintf("Entity '%s' not found in service", name), nil)
	}
	return entity, nil
}

// Collections returns all collections in the service
func (s *Service) Collections() map[string]*Collection {
	return s.collection
}

// Collection returns a specific collection by name
func (s *Service) Collection(name string) (*Collection, error) {
	collection, exists := s.collection[name]
	if !exists {
		return nil, NewElectroError("CollectionNotFound",
			fmt.Sprintf("Collection '%s' not found in service", name), nil)
	}
	return collection, nil
}

// CollectionQuery represents a query across multiple entities in a collection
type CollectionQuery struct {
	collection  *Collection
	pkFacets    []interface{}
	skCondition *sortKeyCondition
	options     *QueryOptions
	ctx         context.Context
}

// Query starts a collection query
func (c *Collection) Query(facets ...interface{}) *CollectionQuery {
	return &CollectionQuery{
		collection: c,
		pkFacets:   facets,
		ctx:        context.Background(),
	}
}

// Eq adds an equals condition on the sort key
func (cq *CollectionQuery) Eq(value interface{}) *CollectionQuery {
	cq.skCondition = &sortKeyCondition{
		operation: "=",
		values:    []interface{}{value},
	}
	return cq
}

// Gt adds a greater-than condition on the sort key
func (cq *CollectionQuery) Gt(value interface{}) *CollectionQuery {
	cq.skCondition = &sortKeyCondition{
		operation: ">",
		values:    []interface{}{value},
	}
	return cq
}

// Gte adds a greater-than-or-equal condition on the sort key
func (cq *CollectionQuery) Gte(value interface{}) *CollectionQuery {
	cq.skCondition = &sortKeyCondition{
		operation: ">=",
		values:    []interface{}{value},
	}
	return cq
}

// Lt adds a less-than condition on the sort key
func (cq *CollectionQuery) Lt(value interface{}) *CollectionQuery {
	cq.skCondition = &sortKeyCondition{
		operation: "<",
		values:    []interface{}{value},
	}
	return cq
}

// Lte adds a less-than-or-equal condition on the sort key
func (cq *CollectionQuery) Lte(value interface{}) *CollectionQuery {
	cq.skCondition = &sortKeyCondition{
		operation: "<=",
		values:    []interface{}{value},
	}
	return cq
}

// Between adds a between condition on the sort key
func (cq *CollectionQuery) Between(start, end interface{}) *CollectionQuery {
	cq.skCondition = &sortKeyCondition{
		operation: "BETWEEN",
		values:    []interface{}{start, end},
	}
	return cq
}

// Begins adds a begins-with condition on the sort key
func (cq *CollectionQuery) Begins(value interface{}) *CollectionQuery {
	cq.skCondition = &sortKeyCondition{
		operation: "begins_with",
		values:    []interface{}{value},
	}
	return cq
}

// Options sets query options
func (cq *CollectionQuery) Options(opts *QueryOptions) *CollectionQuery {
	cq.options = opts
	return cq
}

// CollectionQueryResponse represents a collection query response
type CollectionQueryResponse struct {
	Data   map[string][]map[string]interface{}
	Cursor *string
}

// Go executes the collection query
func (cq *CollectionQuery) Go() (*CollectionQueryResponse, error) {
	if cq.collection.service.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the service", nil)
	}

	result := &CollectionQueryResponse{
		Data: make(map[string][]map[string]interface{}),
	}

	// Query each entity in the collection
	for _, entityName := range cq.collection.entities {
		entity, err := cq.collection.service.Entity(entityName)
		if err != nil {
			return nil, err
		}

		// Find the index name for this collection in this entity
		var indexName string
		for idx, indexDef := range entity.schema.Indexes {
			collName := idx
			if indexDef.Collection != nil {
				collName = *indexDef.Collection
			}
			if collName == cq.collection.name {
				indexName = idx
				break
			}
		}

		if indexName == "" {
			continue
		}

		// Execute query for this entity
		queryBuilder := entity.Query(indexName)
		if queryBuilder == nil {
			continue
		}

		queryResp, err := queryBuilder.Query(cq.pkFacets...).Go()
		if err != nil {
			return nil, err
		}

		result.Data[entityName] = queryResp.Data
	}

	return result, nil
}

// Params returns the DynamoDB parameters for the collection query
func (cq *CollectionQuery) Params() (map[string]interface{}, error) {
	params := make(map[string]interface{})
	params["entities"] = make(map[string]interface{})

	// Generate params for each entity
	for _, entityName := range cq.collection.entities {
		entity, err := cq.collection.service.Entity(entityName)
		if err != nil {
			return nil, err
		}

		// Find the index name for this collection in this entity
		var indexName string
		for idx, indexDef := range entity.schema.Indexes {
			collName := idx
			if indexDef.Collection != nil {
				collName = *indexDef.Collection
			}
			if collName == cq.collection.name {
				indexName = idx
				break
			}
		}

		if indexName == "" {
			continue
		}

		// Generate params for this entity
		queryBuilder := entity.Query(indexName)
		if queryBuilder == nil {
			continue
		}

		entityParams, err := queryBuilder.Query(cq.pkFacets...).Params()
		if err != nil {
			return nil, err
		}

		params["entities"].(map[string]interface{})[entityName] = entityParams
	}

	return params, nil
}
