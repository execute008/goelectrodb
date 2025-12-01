package electrodb

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/execute008/goelectrodb/electrodb/internal"
)

// ParamsBuilder builds DynamoDB operation parameters
type ParamsBuilder struct {
	entity *Entity
}

// NewParamsBuilder creates a new ParamsBuilder
func NewParamsBuilder(entity *Entity) *ParamsBuilder {
	return &ParamsBuilder{entity: entity}
}

// BuildGetItemParams builds parameters for GetItem operation
func (pb *ParamsBuilder) BuildGetItemParams(keys Keys, options *GetOptions) (map[string]interface{}, error) {
	// Find the primary index (the one without an Index field set)
	var primaryIndex *IndexDefinition
	for _, index := range pb.entity.schema.Indexes {
		if index.Index == nil {
			primaryIndex = index
			break
		}
	}

	if primaryIndex == nil {
		return nil, NewElectroError("InvalidSchema", "No primary index found", nil)
	}

	// Build the partition key
	pkKey, err := pb.buildKey(primaryIndex.PK, keys)
	if err != nil {
		return nil, err
	}

	if !pkKey.Fulfilled {
		return nil, NewElectroError("InvalidKeys", "Partition key facets not fully provided", nil)
	}

	// Build the key map
	keyMap := map[string]types.AttributeValue{
		primaryIndex.PK.Field: &types.AttributeValueMemberS{Value: pkKey.Key},
	}

	// Add sort key if it exists
	if primaryIndex.SK != nil {
		skKey, err := pb.buildKeyWithType(*primaryIndex.SK, keys, true)
		if err != nil {
			return nil, err
		}

		if !skKey.Fulfilled {
			return nil, NewElectroError("InvalidKeys", "Sort key facets not fully provided", nil)
		}

		keyMap[primaryIndex.SK.Field] = &types.AttributeValueMemberS{Value: skKey.Key}
	}

	params := map[string]interface{}{
		"TableName": pb.getTableName(),
		"Key":       keyMap,
	}

	// Add projection expression if attributes are specified
	if options != nil && len(options.Attributes) > 0 {
		projectionExpression := ""
		for i, attr := range options.Attributes {
			if i > 0 {
				projectionExpression += ", "
			}
			projectionExpression += attr
		}
		params["ProjectionExpression"] = projectionExpression
	}

	return params, nil
}

// BuildPutItemParams builds parameters for PutItem operation
func (pb *ParamsBuilder) BuildPutItemParams(item Item, options *PutOptions) (map[string]interface{}, error) {
	// Validate required attributes
	if err := pb.validateRequiredAttributes(item); err != nil {
		return nil, err
	}

	// Apply defaults
	enrichedItem := pb.applyDefaults(item)

	// Apply automatic timestamps
	enrichedItem = ApplyTimestamps(enrichedItem, pb.entity.schema, false)

	// Apply attribute padding
	enrichedItem = ApplyPadding(enrichedItem, pb.entity.schema)

	// Validate and transform for write (validation, enum, Set transforms, readonly checks)
	validator := NewValidator(pb.entity)
	transformedItem, err := validator.ValidateAndTransformForWrite(enrichedItem, false)
	if err != nil {
		return nil, err
	}

	// Add keys to the item
	transformedItem, err = pb.addKeysToItem(transformedItem)
	if err != nil {
		return nil, err
	}

	// Convert to DynamoDB format
	av, err := attributevalue.MarshalMap(transformedItem)
	if err != nil {
		return nil, NewElectroError("MarshalError", "Failed to marshal item", err)
	}

	params := map[string]interface{}{
		"TableName": pb.getTableName(),
		"Item":      av,
	}

	// Add return values if specified
	if options != nil && options.Response != nil {
		params["ReturnValues"] = *options.Response
	}

	return params, nil
}

// BuildUpdateItemParams builds parameters for UpdateItem operation
func (pb *ParamsBuilder) BuildUpdateItemParams(
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
) (map[string]interface{}, error) {
	// Build key first
	getParams, err := pb.BuildGetItemParams(keys, nil)
	if err != nil {
		return nil, err
	}

	// Apply automatic timestamps to update operations
	setOps = ApplyUpdateTimestamps(setOps, pb.entity.schema)

	// Validate update operations (readonly checks)
	validator := NewValidator(pb.entity)
	if err := validator.ValidateUpdateOperations(setOps, addOps, delOps, remOps); err != nil {
		return nil, err
	}

	// Apply transformations and validations
	setOps, addOps, delOps = validator.ApplySetTransformations(setOps, addOps, delOps)

	// Build update expression
	updateExpr := ""
	exprAttrNames := make(map[string]string)
	exprAttrValues := make(map[string]types.AttributeValue)
	valueCounter := 0

	// SET operations
	if len(setOps) > 0 {
		updateExpr += "SET "
		first := true
		for attr, value := range setOps {
			if !first {
				updateExpr += ", "
			}
			first = false

			attrName := fmt.Sprintf("#attr%d", valueCounter)
			valueName := fmt.Sprintf(":val%d", valueCounter)
			valueCounter++

			updateExpr += fmt.Sprintf("%s = %s", attrName, valueName)
			exprAttrNames[attrName] = attr

			av, err := attributevalue.Marshal(value)
			if err != nil {
				return nil, NewElectroError("MarshalError", "Failed to marshal value", err)
			}
			exprAttrValues[valueName] = av
		}
	}

	// ADD operations
	if len(addOps) > 0 {
		if updateExpr != "" {
			updateExpr += " "
		}
		updateExpr += "ADD "
		first := true
		for attr, value := range addOps {
			if !first {
				updateExpr += ", "
			}
			first = false

			attrName := fmt.Sprintf("#attr%d", valueCounter)
			valueName := fmt.Sprintf(":val%d", valueCounter)
			valueCounter++

			updateExpr += fmt.Sprintf("%s %s", attrName, valueName)
			exprAttrNames[attrName] = attr

			av, err := attributevalue.Marshal(value)
			if err != nil {
				return nil, NewElectroError("MarshalError", "Failed to marshal value", err)
			}
			exprAttrValues[valueName] = av
		}
	}

	// DELETE operations (for removing values from sets)
	if len(delOps) > 0 {
		if updateExpr != "" {
			updateExpr += " "
		}
		updateExpr += "DELETE "
		first := true
		for attr, value := range delOps {
			if !first {
				updateExpr += ", "
			}
			first = false

			attrName := fmt.Sprintf("#attr%d", valueCounter)
			valueName := fmt.Sprintf(":val%d", valueCounter)
			valueCounter++

			updateExpr += fmt.Sprintf("%s %s", attrName, valueName)
			exprAttrNames[attrName] = attr

			av, err := attributevalue.Marshal(value)
			if err != nil {
				return nil, NewElectroError("MarshalError", "Failed to marshal value", err)
			}
			exprAttrValues[valueName] = av
		}
	}

	// REMOVE operations
	if len(remOps) > 0 {
		if updateExpr != "" {
			updateExpr += " "
		}
		updateExpr += "REMOVE "
		first := true
		for _, attr := range remOps {
			if !first {
				updateExpr += ", "
			}
			first = false

			attrName := fmt.Sprintf("#attr%d", valueCounter)
			valueCounter++

			updateExpr += attrName
			exprAttrNames[attrName] = attr
		}
	}

	// Handle APPEND operations (using list_append in SET clause)
	if len(appendOps) > 0 {
		for attr, value := range appendOps {
			if updateExpr == "" {
				updateExpr = "SET "
			} else if !contains(updateExpr, "SET") {
				updateExpr += " SET "
			} else {
				updateExpr += ", "
			}

			attrName := fmt.Sprintf("#attr%d", valueCounter)
			valueName := fmt.Sprintf(":val%d", valueCounter)
			valueCounter++

			// list_append(attribute, :value) appends to the end
			updateExpr += fmt.Sprintf("%s = list_append(%s, %s)", attrName, attrName, valueName)
			exprAttrNames[attrName] = attr

			av, err := attributevalue.Marshal(value)
			if err != nil {
				return nil, NewElectroError("MarshalError", "Failed to marshal value", err)
			}
			exprAttrValues[valueName] = av
		}
	}

	// Handle PREPEND operations (using list_append in SET clause with reversed order)
	if len(prependOps) > 0 {
		for attr, value := range prependOps {
			if updateExpr == "" {
				updateExpr = "SET "
			} else if !contains(updateExpr, "SET") {
				updateExpr += " SET "
			} else {
				updateExpr += ", "
			}

			attrName := fmt.Sprintf("#attr%d", valueCounter)
			valueName := fmt.Sprintf(":val%d", valueCounter)
			valueCounter++

			// list_append(:value, attribute) prepends to the beginning
			updateExpr += fmt.Sprintf("%s = list_append(%s, %s)", attrName, valueName, attrName)
			exprAttrNames[attrName] = attr

			av, err := attributevalue.Marshal(value)
			if err != nil {
				return nil, NewElectroError("MarshalError", "Failed to marshal value", err)
			}
			exprAttrValues[valueName] = av
		}
	}

	// Handle SUBTRACT operations (using subtraction in SET clause)
	if len(subtractOps) > 0 {
		for attr, value := range subtractOps {
			if updateExpr == "" {
				updateExpr = "SET "
			} else if !contains(updateExpr, "SET") {
				updateExpr += " SET "
			} else {
				updateExpr += ", "
			}

			attrName := fmt.Sprintf("#attr%d", valueCounter)
			valueName := fmt.Sprintf(":val%d", valueCounter)
			valueCounter++

			// attribute = attribute - :value
			updateExpr += fmt.Sprintf("%s = %s - %s", attrName, attrName, valueName)
			exprAttrNames[attrName] = attr

			av, err := attributevalue.Marshal(value)
			if err != nil {
				return nil, NewElectroError("MarshalError", "Failed to marshal value", err)
			}
			exprAttrValues[valueName] = av
		}
	}

	// Handle DATA operations (for removing specific list indices)
	// This uses REMOVE with indexed paths like attribute[0], attribute[1]
	if len(dataOps) > 0 {
		for attr, indices := range dataOps {
			if indexList, ok := indices.([]int); ok {
				for _, index := range indexList {
					if updateExpr != "" && !contains(updateExpr, "REMOVE") {
						updateExpr += " REMOVE "
					} else if contains(updateExpr, "REMOVE") {
						updateExpr += ", "
					} else {
						updateExpr = "REMOVE "
					}

					attrName := fmt.Sprintf("#attr%d", valueCounter)
					valueCounter++

					updateExpr += fmt.Sprintf("%s[%d]", attrName, index)
					exprAttrNames[attrName] = attr
				}
			}
		}
	}

	params := map[string]interface{}{
		"TableName":                 pb.getTableName(),
		"Key":                       getParams["Key"],
		"UpdateExpression":          updateExpr,
		"ExpressionAttributeNames":  exprAttrNames,
		"ExpressionAttributeValues": exprAttrValues,
	}

	// Add return values if specified
	if options != nil && options.Response != nil {
		params["ReturnValues"] = *options.Response
	} else {
		params["ReturnValues"] = "ALL_NEW"
	}

	return params, nil
}

// BuildDeleteItemParams builds parameters for DeleteItem operation
func (pb *ParamsBuilder) BuildDeleteItemParams(keys Keys, options *DeleteOptions) (map[string]interface{}, error) {
	getParams, err := pb.BuildGetItemParams(keys, nil)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"TableName": pb.getTableName(),
		"Key":       getParams["Key"],
	}

	// Add return values if specified
	if options != nil && options.Response != nil {
		params["ReturnValues"] = *options.Response
	}

	return params, nil
}

// BuildQueryParams builds parameters for Query operation
func (pb *ParamsBuilder) BuildQueryParams(
	indexName string,
	pkFacets []interface{},
	skCondition *sortKeyCondition,
	options *QueryOptions,
	filterBuilder *FilterBuilder,
) (map[string]interface{}, error) {
	index, exists := pb.entity.schema.Indexes[indexName]
	if !exists {
		return nil, NewElectroError("InvalidIndex", fmt.Sprintf("Index '%s' not found", indexName), nil)
	}

	// Build facets map from array
	facetsMap := make(map[string]interface{})
	for i, facet := range index.PK.Facets {
		if i < len(pkFacets) {
			facetsMap[facet] = pkFacets[i]
		}
	}

	// Build partition key
	pkKey, err := pb.buildKey(index.PK, facetsMap)
	if err != nil {
		return nil, err
	}

	if !pkKey.Fulfilled {
		return nil, NewElectroError("InvalidKeys", "Partition key facets not fully provided", nil)
	}

	// Build key condition expression
	keyCondition := fmt.Sprintf("%s = :pk", index.PK.Field)
	exprAttrValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: pkKey.Key},
	}

	// Add sort key condition if provided, OR add entity prefix filter if SK exists
	if index.SK != nil {
		skField := index.SK.Field
		if skCondition != nil {
			// Explicit SK condition provided
			switch skCondition.operation {
			case "=":
				keyCondition += fmt.Sprintf(" AND %s = :sk", skField)
				exprAttrValues[":sk"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[0]),
				}
			case ">":
				keyCondition += fmt.Sprintf(" AND %s > :sk", skField)
				exprAttrValues[":sk"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[0]),
				}
			case ">=":
				keyCondition += fmt.Sprintf(" AND %s >= :sk", skField)
				exprAttrValues[":sk"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[0]),
				}
			case "<":
				keyCondition += fmt.Sprintf(" AND %s < :sk", skField)
				exprAttrValues[":sk"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[0]),
				}
			case "<=":
				keyCondition += fmt.Sprintf(" AND %s <= :sk", skField)
				exprAttrValues[":sk"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[0]),
				}
			case "BETWEEN":
				keyCondition += fmt.Sprintf(" AND %s BETWEEN :sk1 AND :sk2", skField)
				exprAttrValues[":sk1"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[0]),
				}
				exprAttrValues[":sk2"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[1]),
				}
			case "begins_with":
				keyCondition += fmt.Sprintf(" AND begins_with(%s, :sk)", skField)
				exprAttrValues[":sk"] = &types.AttributeValueMemberS{
					Value: fmt.Sprintf("%v", skCondition.values[0]),
				}
			}
		} else {
			// No explicit SK condition - add entity prefix to filter by entity type
			// This is critical for single-table design where multiple entities share the same PK
			// TypeScript ElectroDB format: $<entity>_<version>#<firstFacetLabel>_
			// Example: $contentlike_1#likeid_
			skPrefix := internal.BuildSortKeyPrefix(pb.entity.schema.Entity, pb.entity.schema.Version)
			// Add the first SK facet label to match TypeScript ElectroDB format
			if len(index.SK.Facets) > 0 {
				firstFacet := strings.ToLower(index.SK.Facets[0])
				skPrefix = fmt.Sprintf("%s#%s_", skPrefix, firstFacet)
			}
			keyCondition += fmt.Sprintf(" AND begins_with(%s, :sk)", skField)
			exprAttrValues[":sk"] = &types.AttributeValueMemberS{Value: skPrefix}
		}
	}

	params := map[string]interface{}{
		"TableName":                 pb.getTableName(),
		"KeyConditionExpression":    keyCondition,
		"ExpressionAttributeValues": exprAttrValues,
	}

	// Add index name if it's a GSI
	if index.Index != nil {
		params["IndexName"] = *index.Index
	}

	// Add options
	if options != nil {
		if options.Limit != nil {
			params["Limit"] = *options.Limit
		}
		if options.Order != nil && *options.Order == "desc" {
			params["ScanIndexForward"] = false
		}
	}

	// Add filter expression if provided
	if filterBuilder != nil {
		filterExpr, filterNames, filterValues := filterBuilder.Build()
		if filterExpr != "" {
			params["FilterExpression"] = filterExpr

			// Merge expression attribute names and values
			existingNames := make(map[string]string)
			if params["ExpressionAttributeNames"] != nil {
				existingNames = params["ExpressionAttributeNames"].(map[string]string)
			}

			existingValues := params["ExpressionAttributeValues"].(map[string]types.AttributeValue)

			mergedNames, mergedValues := MergeExpressionAttributes(
				existingNames,
				existingValues,
				filterNames,
				filterValues,
			)

			if len(mergedNames) > 0 {
				params["ExpressionAttributeNames"] = mergedNames
			}
			params["ExpressionAttributeValues"] = mergedValues
		}
	}

	return params, nil
}

// Helper methods

func (pb *ParamsBuilder) buildKey(facetDef FacetDefinition, supplied map[string]interface{}) (internal.KeyResult, error) {
	return pb.buildKeyWithType(facetDef, supplied, false)
}

func (pb *ParamsBuilder) buildKeyWithType(facetDef FacetDefinition, supplied map[string]interface{}, isSortKey bool) (internal.KeyResult, error) {
	var prefix string
	if isSortKey {
		// SK prefix: $<entity>_<version>
		prefix = internal.BuildSortKeyPrefix(pb.entity.schema.Entity, pb.entity.schema.Version)
	} else {
		// PK prefix: $<service>
		prefix = internal.BuildPartitionKeyPrefix(pb.entity.schema.Service)
	}

	labels := internal.BuildLabels(facetDef.Facets)

	options := internal.KeyOptions{
		Prefix:           prefix,
		IsCustom:         false,
		ExcludeLabelTail: false,
	}

	if facetDef.Casing != nil {
		options.Casing = facetDef.Casing
	}

	return internal.MakeKey(options, facetDef.Facets, supplied, labels), nil
}

func (pb *ParamsBuilder) getTableName() string {
	if pb.entity.config.Table != nil {
		return *pb.entity.config.Table
	}
	return pb.entity.schema.Table
}

func (pb *ParamsBuilder) validateRequiredAttributes(item Item) error {
	for name, attr := range pb.entity.schema.Attributes {
		if attr.Required {
			if _, exists := item[name]; !exists {
				return NewElectroError("MissingAttribute",
					fmt.Sprintf("Required attribute '%s' is missing", name), nil)
			}
		}
	}
	return nil
}

func (pb *ParamsBuilder) applyDefaults(item Item) Item {
	result := make(Item)
	for k, v := range item {
		result[k] = v
	}

	for name, attr := range pb.entity.schema.Attributes {
		if _, exists := result[name]; !exists && attr.Default != nil {
			result[name] = attr.Default()
		}
	}

	return result
}

func (pb *ParamsBuilder) addKeysToItem(item Item) (Item, error) {
	result := make(Item)
	for k, v := range item {
		result[k] = v
	}

	// Add keys for all indexes
	for _, index := range pb.entity.schema.Indexes {
		// Build partition key
		pkKey, err := pb.buildKey(index.PK, item)
		if err != nil {
			return nil, err
		}
		result[index.PK.Field] = pkKey.Key

		// Build sort key if it exists
		if index.SK != nil {
			skKey, err := pb.buildKeyWithType(*index.SK, item, true)
			if err != nil {
				return nil, err
			}
			if skKey.Fulfilled {
				result[index.SK.Field] = skKey.Key
			}
		}
	}

	return result, nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
