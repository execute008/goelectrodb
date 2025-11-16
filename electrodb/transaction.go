package electrodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TransactionItem represents a single transaction operation
type TransactionItem interface {
	// BuildTransactItem builds the DynamoDB transaction item
	BuildTransactItem() (types.TransactWriteItem, error)
	// BuildTransactGetItem builds the DynamoDB transaction get item
	BuildTransactGetItem() (types.TransactGetItem, error)
}

// TransactWriteBuilder builds a transaction write request
type TransactWriteBuilder struct {
	service *Service
	items   []TransactionItem
}

// TransactWrite creates a new transaction write builder
func (s *Service) TransactWrite(fn func(entities map[string]*Entity) []TransactionItem) *TransactWriteBuilder {
	items := fn(s.entities)
	return &TransactWriteBuilder{
		service: s,
		items:   items,
	}
}

// TransactWriteResponse represents a transaction write response
type TransactWriteResponse struct {
	Canceled bool
	Data     []TransactResult
}

// TransactResult represents a single transaction result
type TransactResult struct {
	Rejected bool
	Code     string
	Message  string
	Item     map[string]interface{}
}

// Go executes the transaction write
func (twb *TransactWriteBuilder) Go() (*TransactWriteResponse, error) {
	return twb.GoWithContext(context.Background())
}

// GoWithContext executes the transaction write with a context
func (twb *TransactWriteBuilder) GoWithContext(ctx context.Context) (*TransactWriteResponse, error) {
	if twb.service.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the service", nil)
	}

	if len(twb.items) == 0 {
		return &TransactWriteResponse{
			Canceled: false,
			Data:     make([]TransactResult, 0),
		}, nil
	}

	// Build transaction items
	transactItems := make([]types.TransactWriteItem, 0, len(twb.items))
	for _, item := range twb.items {
		transactItem, err := item.BuildTransactItem()
		if err != nil {
			return nil, err
		}
		transactItems = append(transactItems, transactItem)
	}

	// Execute transaction
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	}

	_, err := twb.service.client.TransactWriteItems(ctx, input)
	if err != nil {
		// Check if it's a transaction canceled exception
		// TODO: Parse cancellation reasons and return them
		return nil, NewElectroError("TransactionError", "Transaction failed", err)
	}

	// Successful transaction
	return &TransactWriteResponse{
		Canceled: false,
		Data:     make([]TransactResult, len(twb.items)),
	}, nil
}

// Params returns the DynamoDB parameters without executing
func (twb *TransactWriteBuilder) Params() (map[string]interface{}, error) {
	transactItems := make([]types.TransactWriteItem, 0, len(twb.items))
	for _, item := range twb.items {
		transactItem, err := item.BuildTransactItem()
		if err != nil {
			return nil, err
		}
		transactItems = append(transactItems, transactItem)
	}

	tableName := twb.service.table
	if tableName == nil && len(twb.service.entities) > 0 {
		for _, entity := range twb.service.entities {
			tableName = &entity.schema.Table
			break
		}
	}

	return map[string]interface{}{
		"TransactItems": transactItems,
		"TableName":     tableName,
	}, nil
}

// TransactGetBuilder builds a transaction get request
type TransactGetBuilder struct {
	service *Service
	items   []TransactionItem
}

// TransactGet creates a new transaction get builder
func (s *Service) TransactGet(fn func(entities map[string]*Entity) []TransactionItem) *TransactGetBuilder {
	items := fn(s.entities)
	return &TransactGetBuilder{
		service: s,
		items:   items,
	}
}

// TransactGetResponse represents a transaction get response
type TransactGetResponse struct {
	Canceled bool
	Data     []TransactResult
}

// Go executes the transaction get
func (tgb *TransactGetBuilder) Go() (*TransactGetResponse, error) {
	return tgb.GoWithContext(context.Background())
}

// GoWithContext executes the transaction get with a context
func (tgb *TransactGetBuilder) GoWithContext(ctx context.Context) (*TransactGetResponse, error) {
	if tgb.service.client == nil {
		return nil, NewElectroError("NoClientProvided",
			"No DynamoDB client was provided to the service", nil)
	}

	if len(tgb.items) == 0 {
		return &TransactGetResponse{
			Canceled: false,
			Data:     make([]TransactResult, 0),
		}, nil
	}

	// Build transaction get items
	transactItems := make([]types.TransactGetItem, 0, len(tgb.items))
	for _, item := range tgb.items {
		transactItem, err := item.BuildTransactGetItem()
		if err != nil {
			return nil, err
		}
		transactItems = append(transactItems, transactItem)
	}

	// Execute transaction
	input := &dynamodb.TransactGetItemsInput{
		TransactItems: transactItems,
	}

	result, err := tgb.service.client.TransactGetItems(ctx, input)
	if err != nil {
		return nil, NewElectroError("TransactionError", "Transaction failed", err)
	}

	// Parse responses
	results := make([]TransactResult, len(result.Responses))
	for i, response := range result.Responses {
		var item map[string]interface{}
		if response.Item != nil {
			err = attributevalue.UnmarshalMap(response.Item, &item)
			if err != nil {
				return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
			}
		}

		results[i] = TransactResult{
			Rejected: false,
			Item:     item,
		}
	}

	return &TransactGetResponse{
		Canceled: false,
		Data:     results,
	}, nil
}

// Params returns the DynamoDB parameters without executing
func (tgb *TransactGetBuilder) Params() (map[string]interface{}, error) {
	transactItems := make([]types.TransactGetItem, 0, len(tgb.items))
	for _, item := range tgb.items {
		transactItem, err := item.BuildTransactGetItem()
		if err != nil {
			return nil, err
		}
		transactItems = append(transactItems, transactItem)
	}

	tableName := tgb.service.table
	if tableName == nil && len(tgb.service.entities) > 0 {
		for _, entity := range tgb.service.entities {
			tableName = &entity.schema.Table
			break
		}
	}

	return map[string]interface{}{
		"TransactItems": transactItems,
		"TableName":     tableName,
	}, nil
}

// TransactPutItem wraps a put operation for transactions
type TransactPutItem struct {
	entity           *Entity
	item             Item
	conditionBuilder *ConditionBuilder
}

// Commit prepares a put operation for a transaction
func (p *PutOperation) Commit() TransactionItem {
	return &TransactPutItem{
		entity:           p.entity,
		item:             p.item,
		conditionBuilder: p.conditionBuilder,
	}
}

// BuildTransactItem builds the transaction write item
func (tpi *TransactPutItem) BuildTransactItem() (types.TransactWriteItem, error) {
	builder := NewParamsBuilder(tpi.entity)
	params, err := builder.BuildPutItemParams(tpi.item, nil)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	tableName := tpi.entity.config.Table
	if tableName == nil {
		tableName = &tpi.entity.schema.Table
	}

	put := &types.Put{
		TableName: tableName,
		Item:      params["Item"].(map[string]types.AttributeValue),
	}

	// Add condition expression if provided
	if tpi.conditionBuilder != nil {
		condExpr, condNames, condValues := tpi.conditionBuilder.Build()
		if condExpr != "" {
			put.ConditionExpression = &condExpr

			if len(condNames) > 0 {
				put.ExpressionAttributeNames = condNames
			}

			if len(condValues) > 0 {
				put.ExpressionAttributeValues = condValues
			}
		}
	}

	return types.TransactWriteItem{
		Put: put,
	}, nil
}

// BuildTransactGetItem is not supported for put operations
func (tpi *TransactPutItem) BuildTransactGetItem() (types.TransactGetItem, error) {
	return types.TransactGetItem{}, NewElectroError("InvalidOperation",
		"Put operations cannot be used in TransactGet", nil)
}

// TransactUpdateItem wraps an update operation for transactions
type TransactUpdateItem struct {
	entity           *Entity
	keys             Keys
	setOps           map[string]interface{}
	addOps           map[string]interface{}
	delOps           map[string]interface{}
	remOps           []string
	conditionBuilder *ConditionBuilder
}

// Commit prepares an update operation for a transaction
func (u *UpdateOperation) Commit() TransactionItem {
	return &TransactUpdateItem{
		entity:           u.entity,
		keys:             u.keys,
		setOps:           u.setOps,
		addOps:           u.addOps,
		delOps:           u.delOps,
		remOps:           u.remOps,
		conditionBuilder: u.conditionBuilder,
	}
}

// BuildTransactItem builds the transaction write item
func (tui *TransactUpdateItem) BuildTransactItem() (types.TransactWriteItem, error) {
	builder := NewParamsBuilder(tui.entity)
	params, err := builder.BuildUpdateItemParams(tui.keys, tui.setOps, tui.addOps, tui.delOps, tui.remOps, nil)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	tableName := tui.entity.config.Table
	if tableName == nil {
		tableName = &tui.entity.schema.Table
	}

	update := &types.Update{
		TableName:                 tableName,
		Key:                       params["Key"].(map[string]types.AttributeValue),
		UpdateExpression:          stringPtr(params["UpdateExpression"].(string)),
		ExpressionAttributeNames:  params["ExpressionAttributeNames"].(map[string]string),
		ExpressionAttributeValues: params["ExpressionAttributeValues"].(map[string]types.AttributeValue),
	}

	// Add condition expression if provided
	if tui.conditionBuilder != nil {
		condExpr, condNames, condValues := tui.conditionBuilder.Build()
		if condExpr != "" {
			update.ConditionExpression = &condExpr

			// Merge expression attribute names and values
			mergedNames, mergedValues := MergeExpressionAttributes(
				update.ExpressionAttributeNames,
				update.ExpressionAttributeValues,
				condNames,
				condValues,
			)
			update.ExpressionAttributeNames = mergedNames
			update.ExpressionAttributeValues = mergedValues
		}
	}

	return types.TransactWriteItem{
		Update: update,
	}, nil
}

// BuildTransactGetItem is not supported for update operations
func (tui *TransactUpdateItem) BuildTransactGetItem() (types.TransactGetItem, error) {
	return types.TransactGetItem{}, NewElectroError("InvalidOperation",
		"Update operations cannot be used in TransactGet", nil)
}

// TransactDeleteItem wraps a delete operation for transactions
type TransactDeleteItem struct {
	entity           *Entity
	keys             Keys
	conditionBuilder *ConditionBuilder
}

// Commit prepares a delete operation for a transaction
func (d *DeleteOperation) Commit() TransactionItem {
	return &TransactDeleteItem{
		entity:           d.entity,
		keys:             d.keys,
		conditionBuilder: d.conditionBuilder,
	}
}

// BuildTransactItem builds the transaction write item
func (tdi *TransactDeleteItem) BuildTransactItem() (types.TransactWriteItem, error) {
	builder := NewParamsBuilder(tdi.entity)
	params, err := builder.BuildDeleteItemParams(tdi.keys, nil)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	tableName := tdi.entity.config.Table
	if tableName == nil {
		tableName = &tdi.entity.schema.Table
	}

	del := &types.Delete{
		TableName: tableName,
		Key:       params["Key"].(map[string]types.AttributeValue),
	}

	// Add condition expression if provided
	if tdi.conditionBuilder != nil {
		condExpr, condNames, condValues := tdi.conditionBuilder.Build()
		if condExpr != "" {
			del.ConditionExpression = &condExpr

			if len(condNames) > 0 {
				del.ExpressionAttributeNames = condNames
			}

			if len(condValues) > 0 {
				del.ExpressionAttributeValues = condValues
			}
		}
	}

	return types.TransactWriteItem{
		Delete: del,
	}, nil
}

// BuildTransactGetItem is not supported for delete operations
func (tdi *TransactDeleteItem) BuildTransactGetItem() (types.TransactGetItem, error) {
	return types.TransactGetItem{}, NewElectroError("InvalidOperation",
		"Delete operations cannot be used in TransactGet", nil)
}

// TransactGetItem wraps a get operation for transactions
type TransactGetItem struct {
	entity *Entity
	keys   Keys
}

// Commit prepares a get operation for a transaction
func (g *GetOperation) Commit() TransactionItem {
	return &TransactGetItem{
		entity: g.entity,
		keys:   g.keys,
	}
}

// BuildTransactItem is not supported for get operations
func (tgi *TransactGetItem) BuildTransactItem() (types.TransactWriteItem, error) {
	return types.TransactWriteItem{}, NewElectroError("InvalidOperation",
		"Get operations cannot be used in TransactWrite", nil)
}

// BuildTransactGetItem builds the transaction get item
func (tgi *TransactGetItem) BuildTransactGetItem() (types.TransactGetItem, error) {
	builder := NewParamsBuilder(tgi.entity)
	params, err := builder.BuildGetItemParams(tgi.keys, nil)
	if err != nil {
		return types.TransactGetItem{}, err
	}

	tableName := tgi.entity.config.Table
	if tableName == nil {
		tableName = &tgi.entity.schema.Table
	}

	return types.TransactGetItem{
		Get: &types.Get{
			TableName: tableName,
			Key:       params["Key"].(map[string]types.AttributeValue),
		},
	}, nil
}
