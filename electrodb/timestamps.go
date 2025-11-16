package electrodb

import (
	"time"
)

// ApplyTimestamps applies automatic timestamps to an item
// This is called during Put/Create operations
func ApplyTimestamps(item Item, schema *Schema, isUpdate bool) Item {
	if schema.Timestamps == nil {
		return item
	}

	result := make(Item)
	for k, v := range item {
		result[k] = v
	}

	now := time.Now().Unix()

	// Set createdAt only on create (not on update)
	if !isUpdate && schema.Timestamps.CreatedAt != "" {
		// Only set if not already present (user may have provided it)
		if _, exists := result[schema.Timestamps.CreatedAt]; !exists {
			result[schema.Timestamps.CreatedAt] = now
		}
	}

	// Set updatedAt on both create and update
	if schema.Timestamps.UpdatedAt != "" {
		result[schema.Timestamps.UpdatedAt] = now
	}

	return result
}

// ApplyUpdateTimestamps applies automatic timestamps to update operations
// This adds updatedAt to SET operations
func ApplyUpdateTimestamps(setOps map[string]interface{}, schema *Schema) map[string]interface{} {
	if schema.Timestamps == nil || schema.Timestamps.UpdatedAt == "" {
		return setOps
	}

	result := make(map[string]interface{})
	for k, v := range setOps {
		result[k] = v
	}

	// Always set updatedAt on updates
	result[schema.Timestamps.UpdatedAt] = time.Now().Unix()

	return result
}
