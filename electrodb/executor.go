package electrodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ExecutionHelper helps execute DynamoDB operations
type ExecutionHelper struct {
	entity *Entity
}

// NewExecutionHelper creates a new ExecutionHelper
func NewExecutionHelper(entity *Entity) *ExecutionHelper {
	return &ExecutionHelper{entity: entity}
}

// ExecuteGetItem executes a GetItem operation
func (eh *ExecutionHelper) ExecuteGetItem(ctx context.Context, keys Keys, options *GetOptions) (*GetResponse, error) {
	if eh.entity.client == nil {
		return nil, NewElectroError("NoClientProvided", "No DynamoDB client was provided to the entity", nil)
	}

	builder := NewParamsBuilder(eh.entity)
	params, err := builder.BuildGetItemParams(keys, options)
	if err != nil {
		return nil, err
	}

	// Convert to DynamoDB GetItemInput
	input := &dynamodb.GetItemInput{
		TableName: stringPtr(params["TableName"].(string)),
		Key:       params["Key"].(map[string]types.AttributeValue),
	}

	if projExpr, ok := params["ProjectionExpression"].(string); ok && projExpr != "" {
		input.ProjectionExpression = &projExpr
	}

	// Execute
	result, err := eh.entity.client.GetItem(ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute GetItem", err)
	}

	// Parse response
	var item map[string]interface{}
	if result.Item != nil {
		err = attributevalue.UnmarshalMap(result.Item, &item)
		if err != nil {
			return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
		}
	}

	// Remove internal keys if not raw mode
	if options == nil || !options.Raw {
		item = eh.removeInternalKeys(item)
		// Apply Get transformations and filter hidden attributes
		validator := NewValidator(eh.entity)
		item = validator.TransformForRead(item)
	}

	return &GetResponse{Data: item}, nil
}

// ExecutePutItem executes a PutItem operation
func (eh *ExecutionHelper) ExecutePutItem(ctx context.Context, item Item, options *PutOptions) (*PutResponse, error) {
	if eh.entity.client == nil {
		return nil, NewElectroError("NoClientProvided", "No DynamoDB client was provided to the entity", nil)
	}

	builder := NewParamsBuilder(eh.entity)
	params, err := builder.BuildPutItemParams(item, options)
	if err != nil {
		return nil, err
	}

	// Convert to DynamoDB PutItemInput
	input := &dynamodb.PutItemInput{
		TableName: stringPtr(params["TableName"].(string)),
		Item:      params["Item"].(map[string]types.AttributeValue),
	}

	if returnValues, ok := params["ReturnValues"].(string); ok {
		input.ReturnValues = types.ReturnValue(returnValues)
	}

	// Execute
	result, err := eh.entity.client.PutItem(ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute PutItem", err)
	}

	// Parse response
	var responseItem map[string]interface{}
	if result.Attributes != nil {
		err = attributevalue.UnmarshalMap(result.Attributes, &responseItem)
		if err != nil {
			return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
		}
	}

	// Remove internal keys if not raw mode
	if options == nil || !options.Raw {
		responseItem = eh.removeInternalKeys(responseItem)
		// Apply Get transformations and filter hidden attributes
		validator := NewValidator(eh.entity)
		responseItem = validator.TransformForRead(responseItem)
	}

	return &PutResponse{Data: responseItem}, nil
}

// ExecuteUpdateItem executes an UpdateItem operation
func (eh *ExecutionHelper) ExecuteUpdateItem(
	ctx context.Context,
	keys Keys,
	setOps map[string]interface{},
	addOps map[string]interface{},
	delOps map[string]interface{},
	remOps []string,
	appendOps map[string]interface{},
	prependOps map[string]interface{},
	subtractOps map[string]interface{},
	dataOps map[string]interface{},
	options *UpdateOptions,
) (*UpdateResponse, error) {
	if eh.entity.client == nil {
		return nil, NewElectroError("NoClientProvided", "No DynamoDB client was provided to the entity", nil)
	}

	builder := NewParamsBuilder(eh.entity)
	params, err := builder.BuildUpdateItemParams(keys, setOps, addOps, delOps, remOps, appendOps, prependOps, subtractOps, dataOps, options)
	if err != nil {
		return nil, err
	}

	// Convert to DynamoDB UpdateItemInput
	input := &dynamodb.UpdateItemInput{
		TableName:                 stringPtr(params["TableName"].(string)),
		Key:                       params["Key"].(map[string]types.AttributeValue),
		UpdateExpression:          stringPtr(params["UpdateExpression"].(string)),
		ExpressionAttributeNames:  params["ExpressionAttributeNames"].(map[string]string),
		ExpressionAttributeValues: params["ExpressionAttributeValues"].(map[string]types.AttributeValue),
		ReturnValues:              types.ReturnValue(params["ReturnValues"].(string)),
	}

	// Execute
	result, err := eh.entity.client.UpdateItem(ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute UpdateItem", err)
	}

	// Parse response
	var responseItem map[string]interface{}
	if result.Attributes != nil {
		err = attributevalue.UnmarshalMap(result.Attributes, &responseItem)
		if err != nil {
			return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
		}
	}

	// Remove internal keys if not raw mode
	if options == nil || !options.Raw {
		responseItem = eh.removeInternalKeys(responseItem)
		// Apply Get transformations and filter hidden attributes
		validator := NewValidator(eh.entity)
		responseItem = validator.TransformForRead(responseItem)
	}

	return &UpdateResponse{Data: responseItem}, nil
}

// ExecuteDeleteItem executes a DeleteItem operation
func (eh *ExecutionHelper) ExecuteDeleteItem(ctx context.Context, keys Keys, options *DeleteOptions) (*DeleteResponse, error) {
	if eh.entity.client == nil {
		return nil, NewElectroError("NoClientProvided", "No DynamoDB client was provided to the entity", nil)
	}

	builder := NewParamsBuilder(eh.entity)
	params, err := builder.BuildDeleteItemParams(keys, options)
	if err != nil {
		return nil, err
	}

	// Convert to DynamoDB DeleteItemInput
	input := &dynamodb.DeleteItemInput{
		TableName: stringPtr(params["TableName"].(string)),
		Key:       params["Key"].(map[string]types.AttributeValue),
	}

	if returnValues, ok := params["ReturnValues"].(string); ok {
		input.ReturnValues = types.ReturnValue(returnValues)
	}

	// Execute
	result, err := eh.entity.client.DeleteItem(ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute DeleteItem", err)
	}

	// Parse response
	var responseItem map[string]interface{}
	if result.Attributes != nil {
		err = attributevalue.UnmarshalMap(result.Attributes, &responseItem)
		if err != nil {
			return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
		}
	}

	// Remove internal keys if not raw mode
	if options == nil || !options.Raw {
		responseItem = eh.removeInternalKeys(responseItem)
		// Apply Get transformations and filter hidden attributes
		validator := NewValidator(eh.entity)
		responseItem = validator.TransformForRead(responseItem)
	}

	return &DeleteResponse{Data: responseItem}, nil
}

// ExecuteQuery executes a Query operation
func (eh *ExecutionHelper) ExecuteQuery(
	ctx context.Context,
	indexName string,
	pkFacets []interface{},
	skCondition *sortKeyCondition,
	options *QueryOptions,
	filterBuilder *FilterBuilder,
) (*QueryResponse, error) {
	if eh.entity.client == nil {
		return nil, NewElectroError("NoClientProvided", "No DynamoDB client was provided to the entity", nil)
	}

	builder := NewParamsBuilder(eh.entity)
	params, err := builder.BuildQueryParams(indexName, pkFacets, skCondition, options, filterBuilder)
	if err != nil {
		return nil, err
	}

	// Convert to DynamoDB QueryInput
	input := &dynamodb.QueryInput{
		TableName:                 stringPtr(params["TableName"].(string)),
		KeyConditionExpression:    stringPtr(params["KeyConditionExpression"].(string)),
		ExpressionAttributeValues: params["ExpressionAttributeValues"].(map[string]types.AttributeValue),
	}

	if indexName, ok := params["IndexName"].(string); ok {
		input.IndexName = &indexName
	}

	if filterExpr, ok := params["FilterExpression"].(string); ok {
		input.FilterExpression = &filterExpr
	}

	if exprAttrNames, ok := params["ExpressionAttributeNames"].(map[string]string); ok {
		input.ExpressionAttributeNames = exprAttrNames
	}

	if options != nil {
		if options.Limit != nil {
			input.Limit = options.Limit
		}
		if scanForward, ok := params["ScanIndexForward"].(bool); ok {
			input.ScanIndexForward = &scanForward
		}
		if options.Cursor != nil {
			exclusiveStartKey, err := decodeCursor(*options.Cursor)
			if err != nil {
				return nil, err
			}
			input.ExclusiveStartKey = exclusiveStartKey
		}
	}

	// Execute
	result, err := eh.entity.client.Query(ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute Query", err)
	}

	// Parse response
	items := make([]map[string]interface{}, 0, len(result.Items))
	validator := NewValidator(eh.entity)
	for _, item := range result.Items {
		var parsedItem map[string]interface{}
		err = attributevalue.UnmarshalMap(item, &parsedItem)
		if err != nil {
			return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
		}

		// Remove internal keys if not raw mode
		if options == nil || !options.Raw {
			parsedItem = eh.removeInternalKeys(parsedItem)
			// Apply Get transformations and filter hidden attributes
			parsedItem = validator.TransformForRead(parsedItem)
		}

		items = append(items, parsedItem)
	}

	// Generate cursor from LastEvaluatedKey
	var cursor *string
	if result.LastEvaluatedKey != nil {
		encoded, err := encodeCursor(result.LastEvaluatedKey)
		if err != nil {
			return nil, err
		}
		if encoded != "" {
			cursor = &encoded
		}
	}

	return &QueryResponse{
		Data:   items,
		Cursor: cursor,
	}, nil
}

// ExecuteScan executes a Scan operation
func (eh *ExecutionHelper) ExecuteScan(ctx context.Context, options *QueryOptions) (*ScanResponse, error) {
	if eh.entity.client == nil {
		return nil, NewElectroError("NoClientProvided", "No DynamoDB client was provided to the entity", nil)
	}

	// Build scan input
	tableName := eh.entity.config.Table
	if tableName == nil {
		tableName = &eh.entity.schema.Table
	}

	input := &dynamodb.ScanInput{
		TableName: tableName,
	}

	if options != nil {
		if options.Limit != nil {
			input.Limit = options.Limit
		}
		if options.Cursor != nil {
			exclusiveStartKey, err := decodeCursor(*options.Cursor)
			if err != nil {
				return nil, err
			}
			input.ExclusiveStartKey = exclusiveStartKey
		}
	}

	// Execute
	result, err := eh.entity.client.Scan(ctx, input)
	if err != nil {
		return nil, NewElectroError("DynamoDBError", "Failed to execute Scan", err)
	}

	// Parse response
	items := make([]map[string]interface{}, 0, len(result.Items))
	validator := NewValidator(eh.entity)
	for _, item := range result.Items {
		var parsedItem map[string]interface{}
		err = attributevalue.UnmarshalMap(item, &parsedItem)
		if err != nil {
			return nil, NewElectroError("UnmarshalError", "Failed to unmarshal response", err)
		}

		// Remove internal keys if not raw mode
		if options == nil || !options.Raw {
			parsedItem = eh.removeInternalKeys(parsedItem)
			// Apply Get transformations and filter hidden attributes
			parsedItem = validator.TransformForRead(parsedItem)
		}

		items = append(items, parsedItem)
	}

	// Generate cursor from LastEvaluatedKey
	var cursor *string
	if result.LastEvaluatedKey != nil {
		encoded, err := encodeCursor(result.LastEvaluatedKey)
		if err != nil {
			return nil, err
		}
		if encoded != "" {
			cursor = &encoded
		}
	}

	return &ScanResponse{
		Data:   items,
		Cursor: cursor,
	}, nil
}

// removeInternalKeys removes internal DynamoDB keys from the response
func (eh *ExecutionHelper) removeInternalKeys(item map[string]interface{}) map[string]interface{} {
	if item == nil {
		return nil
	}

	result := make(map[string]interface{})

	// Copy all user-defined attributes
	for name := range eh.entity.schema.Attributes {
		if val, exists := item[name]; exists {
			result[name] = val
		}
	}

	return result
}
