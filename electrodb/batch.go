package electrodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	// MaxBatchGetItems is the maximum number of items in a batch get request
	MaxBatchGetItems = 100
	// MaxBatchWriteItems is the maximum number of items in a batch write request
	MaxBatchWriteItems = 25
)

// BatchGetRequest represents a batch get request
type BatchGetRequest struct {
	entity *Entity
	keys   []Keys
	ctx    context.Context
}

// BatchGet creates a new batch get request
func (e *Entity) BatchGet(keys []Keys) *BatchGetRequest {
	return &BatchGetRequest{
		entity: e,
		keys:   keys,
		ctx:    context.Background(),
	}
}

// Go executes the batch get operation
func (bgr *BatchGetRequest) Go() (*BatchGetResponse, error) {
	if bgr.entity.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the entity", nil)
	}

	if len(bgr.keys) == 0 {
		return &BatchGetResponse{
			Data:        make([]map[string]interface{}, 0),
			Unprocessed: make([]Keys, 0),
		}, nil
	}

	tableName := bgr.entity.config.Table
	if tableName == nil {
		tableName = &bgr.entity.schema.Table
	}

	result := &BatchGetResponse{
		Data:        make([]map[string]interface{}, 0),
		Unprocessed: make([]Keys, 0),
	}

	// Process in batches of MaxBatchGetItems
	for i := 0; i < len(bgr.keys); i += MaxBatchGetItems {
		end := i + MaxBatchGetItems
		if end > len(bgr.keys) {
			end = len(bgr.keys)
		}

		batchKeys := bgr.keys[i:end]
		batchResult, err := bgr.executeBatch(batchKeys, *tableName)
		if err != nil {
			return nil, err
		}

		result.Data = append(result.Data, batchResult.Data...)
		result.Unprocessed = append(result.Unprocessed, batchResult.Unprocessed...)
	}

	return result, nil
}

func (bgr *BatchGetRequest) executeBatch(keys []Keys, tableName string) (*BatchGetResponse, error) {
	// Build keys for this batch
	keyItems := make([]map[string]types.AttributeValue, 0, len(keys))
	builder := NewParamsBuilder(bgr.entity)

	for _, keySet := range keys {
		params, err := builder.BuildGetItemParams(keySet, nil)
		if err != nil {
			return nil, err
		}

		keyItems = append(keyItems, params["Key"].(map[string]types.AttributeValue))
	}

	// Execute batch get
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			tableName: {
				Keys: keyItems,
			},
		},
	}

	response, err := bgr.entity.client.BatchGetItem(bgr.ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute BatchGetItem", err)
	}

	// Parse response
	result := &BatchGetResponse{
		Data:        make([]map[string]interface{}, 0),
		Unprocessed: make([]Keys, 0),
	}

	if items, ok := response.Responses[tableName]; ok {
		for _, item := range items {
			var parsedItem map[string]interface{}
			err = attributevalue.UnmarshalMap(item, &parsedItem)
			if err != nil {
				return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
			}

			// Remove internal keys
			executor := NewExecutionHelper(bgr.entity)
			parsedItem = executor.removeInternalKeys(parsedItem)

			result.Data = append(result.Data, parsedItem)
		}
	}

	// Handle unprocessed keys
	if unprocessed, ok := response.UnprocessedKeys[tableName]; ok && len(unprocessed.Keys) > 0 {
		for _, unprocessedKey := range unprocessed.Keys {
			var parsedKey Keys
			err = attributevalue.UnmarshalMap(unprocessedKey, &parsedKey)
			if err != nil {
				// If unmarshaling fails, append empty keys to preserve count
				result.Unprocessed = append(result.Unprocessed, Keys{})
				continue
			}
			result.Unprocessed = append(result.Unprocessed, parsedKey)
		}
	}

	return result, nil
}

// BatchWriteRequest represents a batch write request
type BatchWriteRequest struct {
	entity  *Entity
	puts    []Item
	deletes []Keys
	ctx     context.Context
}

// BatchWrite creates a new batch write request
func (e *Entity) BatchWrite() *BatchWriteRequest {
	return &BatchWriteRequest{
		entity:  e,
		puts:    make([]Item, 0),
		deletes: make([]Keys, 0),
		ctx:     context.Background(),
	}
}

// Put adds put operations to the batch
func (bwr *BatchWriteRequest) Put(items []Item) *BatchWriteRequest {
	bwr.puts = append(bwr.puts, items...)
	return bwr
}

// Delete adds delete operations to the batch
func (bwr *BatchWriteRequest) Delete(keys []Keys) *BatchWriteRequest {
	bwr.deletes = append(bwr.deletes, keys...)
	return bwr
}

// Go executes the batch write operation
func (bwr *BatchWriteRequest) Go() (*BatchWriteResponse, error) {
	totalOps := len(bwr.puts) + len(bwr.deletes)
	if totalOps == 0 {
		return &BatchWriteResponse{}, nil
	}

	if totalOps > MaxBatchWriteItems {
		return nil, NewElectroError("BatchTooLarge",
			fmt.Sprintf("Batch write cannot exceed %d items, got %d", MaxBatchWriteItems, totalOps), nil)
	}

	if bwr.entity.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the entity", nil)
	}

	tableName := bwr.entity.config.Table
	if tableName == nil {
		tableName = &bwr.entity.schema.Table
	}

	// Build write requests
	writeRequests := make([]types.WriteRequest, 0, totalOps)
	builder := NewParamsBuilder(bwr.entity)

	// Add put requests
	for _, item := range bwr.puts {
		params, err := builder.BuildPutItemParams(item, nil)
		if err != nil {
			return nil, err
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: params["Item"].(map[string]types.AttributeValue),
			},
		})
	}

	// Add delete requests
	for _, keys := range bwr.deletes {
		params, err := builder.BuildDeleteItemParams(keys, nil)
		if err != nil {
			return nil, err
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: params["Key"].(map[string]types.AttributeValue),
			},
		})
	}

	// Execute batch write
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			*tableName: writeRequests,
		},
	}

	response, err := bwr.entity.client.BatchWriteItem(bwr.ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute BatchWriteItem", err)
	}

	result := &BatchWriteResponse{}

	// Handle unprocessed items
	if unprocessed, ok := response.UnprocessedItems[*tableName]; ok && len(unprocessed) > 0 {
		result.Unprocessed.Puts = make([]Item, 0)
		result.Unprocessed.Deletes = make([]Keys, 0)

		for _, writeReq := range unprocessed {
			if writeReq.PutRequest != nil {
				var parsedItem Item
				err := attributevalue.UnmarshalMap(writeReq.PutRequest.Item, &parsedItem)
				if err != nil {
					// If unmarshaling fails, append empty item to preserve count
					result.Unprocessed.Puts = append(result.Unprocessed.Puts, Item{})
					continue
				}
				result.Unprocessed.Puts = append(result.Unprocessed.Puts, parsedItem)
			}
			if writeReq.DeleteRequest != nil {
				var parsedKey Keys
				err := attributevalue.UnmarshalMap(writeReq.DeleteRequest.Key, &parsedKey)
				if err != nil {
					// If unmarshaling fails, append empty keys to preserve count
					result.Unprocessed.Deletes = append(result.Unprocessed.Deletes, Keys{})
					continue
				}
				result.Unprocessed.Deletes = append(result.Unprocessed.Deletes, parsedKey)
			}
		}
	}

	return result, nil
}

// BatchGetService creates a batch get request across multiple entities in a service
type BatchGetService struct {
	service  *Service
	requests map[string][]Keys
	ctx      context.Context
}

// BatchGet creates a new batch get request for the service
func (s *Service) BatchGet() *BatchGetService {
	return &BatchGetService{
		service:  s,
		requests: make(map[string][]Keys),
		ctx:      context.Background(),
	}
}

// Get adds get requests for a specific entity
func (bgs *BatchGetService) Get(entityName string, keys []Keys) *BatchGetService {
	if _, exists := bgs.requests[entityName]; !exists {
		bgs.requests[entityName] = make([]Keys, 0)
	}
	bgs.requests[entityName] = append(bgs.requests[entityName], keys...)
	return bgs
}

// BatchGetServiceResponse represents a batch get service response
type BatchGetServiceResponse struct {
	Data        map[string][]map[string]interface{}
	Unprocessed map[string][]Keys
}

// Go executes the batch get operation across entities
func (bgs *BatchGetService) Go() (*BatchGetServiceResponse, error) {
	result := &BatchGetServiceResponse{
		Data:        make(map[string][]map[string]interface{}),
		Unprocessed: make(map[string][]Keys),
	}

	// Execute batch get for each entity
	for entityName, keys := range bgs.requests {
		entity, err := bgs.service.Entity(entityName)
		if err != nil {
			return nil, err
		}

		entityResult, err := entity.BatchGet(keys).Go()
		if err != nil {
			return nil, err
		}

		result.Data[entityName] = entityResult.Data
		if len(entityResult.Unprocessed) > 0 {
			result.Unprocessed[entityName] = entityResult.Unprocessed
		}
	}

	return result, nil
}

// BatchWriteService creates a batch write request across multiple entities in a service
type BatchWriteService struct {
	service *Service
	puts    map[string][]Item
	deletes map[string][]Keys
	ctx     context.Context
}

// BatchWrite creates a new batch write request for the service
func (s *Service) BatchWrite() *BatchWriteService {
	return &BatchWriteService{
		service: s,
		puts:    make(map[string][]Item),
		deletes: make(map[string][]Keys),
		ctx:     context.Background(),
	}
}

// Put adds put operations for a specific entity
func (bws *BatchWriteService) Put(entityName string, items []Item) *BatchWriteService {
	if _, exists := bws.puts[entityName]; !exists {
		bws.puts[entityName] = make([]Item, 0)
	}
	bws.puts[entityName] = append(bws.puts[entityName], items...)
	return bws
}

// Delete adds delete operations for a specific entity
func (bws *BatchWriteService) Delete(entityName string, keys []Keys) *BatchWriteService {
	if _, exists := bws.deletes[entityName]; !exists {
		bws.deletes[entityName] = make([]Keys, 0)
	}
	bws.deletes[entityName] = append(bws.deletes[entityName], keys...)
	return bws
}

// BatchWriteServiceResponse represents a batch write service response
type BatchWriteServiceResponse struct {
	Unprocessed map[string]struct {
		Puts    []Item
		Deletes []Keys
	}
}

// Go executes the batch write operation across entities
func (bws *BatchWriteService) Go() (*BatchWriteServiceResponse, error) {
	result := &BatchWriteServiceResponse{
		Unprocessed: make(map[string]struct {
			Puts    []Item
			Deletes []Keys
		}),
	}

	// Execute batch write for each entity (puts)
	for entityName, items := range bws.puts {
		entity, err := bws.service.Entity(entityName)
		if err != nil {
			return nil, err
		}

		entityResult, err := entity.BatchWrite().Put(items).Go()
		if err != nil {
			return nil, err
		}

		if len(entityResult.Unprocessed.Puts) > 0 || len(entityResult.Unprocessed.Deletes) > 0 {
			result.Unprocessed[entityName] = entityResult.Unprocessed
		}
	}

	// Execute batch write for each entity (deletes)
	for entityName, keys := range bws.deletes {
		entity, err := bws.service.Entity(entityName)
		if err != nil {
			return nil, err
		}

		entityResult, err := entity.BatchWrite().Delete(keys).Go()
		if err != nil {
			return nil, err
		}

		if len(entityResult.Unprocessed.Puts) > 0 || len(entityResult.Unprocessed.Deletes) > 0 {
			if existing, ok := result.Unprocessed[entityName]; ok {
				existing.Puts = append(existing.Puts, entityResult.Unprocessed.Puts...)
				existing.Deletes = append(existing.Deletes, entityResult.Unprocessed.Deletes...)
				result.Unprocessed[entityName] = existing
			} else {
				result.Unprocessed[entityName] = entityResult.Unprocessed
			}
		}
	}

	return result, nil
}
