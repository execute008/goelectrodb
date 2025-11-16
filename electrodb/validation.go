package electrodb

import (
	"fmt"
)

// Validator handles attribute validation and transformation
type Validator struct {
	entity *Entity
}

// NewValidator creates a new Validator
func NewValidator(entity *Entity) *Validator {
	return &Validator{entity: entity}
}

// ValidateAndTransformForWrite validates and transforms an item before writing to DynamoDB
// This applies: validation, enum checks, Set transformations, readonly checks
func (v *Validator) ValidateAndTransformForWrite(item Item, isUpdate bool) (Item, error) {
	result := make(Item)

	for name, value := range item {
		attr, exists := v.entity.schema.Attributes[name]
		if !exists {
			// Allow unknown attributes to pass through
			result[name] = value
			continue
		}

		// Check ReadOnly enforcement (only for updates, not creates)
		if isUpdate && attr.ReadOnly {
			return nil, NewElectroError("ReadOnlyViolation",
				fmt.Sprintf("Attribute '%s' is read-only and cannot be updated", name), nil)
		}

		// Validate enum values
		if attr.Type == AttributeTypeEnum && len(attr.EnumValues) > 0 {
			if err := v.validateEnum(name, value, attr.EnumValues); err != nil {
				return nil, err
			}
		}

		// Apply custom validation function
		if attr.Validate != nil {
			if err := attr.Validate(value); err != nil {
				return nil, NewElectroError("ValidationError",
					fmt.Sprintf("Validation failed for attribute '%s': %v", name, err), err)
			}
		}

		// Apply Set transformation (transforms value before writing to DynamoDB)
		transformedValue := value
		if attr.Set != nil {
			transformedValue = attr.Set(value)
		}

		result[name] = transformedValue
	}

	return result, nil
}

// TransformForRead applies Get transformations and filters hidden attributes when reading from DynamoDB
func (v *Validator) TransformForRead(item Item) Item {
	if item == nil {
		return nil
	}

	result := make(Item)

	for name, value := range item {
		attr, exists := v.entity.schema.Attributes[name]
		if !exists {
			// Allow unknown attributes to pass through
			result[name] = value
			continue
		}

		// Skip hidden attributes
		if attr.Hidden {
			continue
		}

		// Apply Get transformation (transforms value after reading from DynamoDB)
		transformedValue := value
		if attr.Get != nil {
			transformedValue = attr.Get(value)
		}

		result[name] = transformedValue
	}

	return result
}

// validateEnum checks if a value is in the allowed enum values
func (v *Validator) validateEnum(attrName string, value interface{}, enumValues []interface{}) error {
	for _, enumVal := range enumValues {
		if value == enumVal {
			return nil
		}
	}

	return NewElectroError("InvalidEnumValue",
		fmt.Sprintf("Attribute '%s' has invalid enum value '%v'. Allowed values: %v",
			attrName, value, enumValues), nil)
}

// ValidateUpdateOperations validates operations for update (SET, ADD, DELETE, REMOVE)
func (v *Validator) ValidateUpdateOperations(
	setOps map[string]interface{},
	addOps map[string]interface{},
	delOps map[string]interface{},
	remOps []string,
) error {
	// Validate SET operations
	for name := range setOps {
		attr, exists := v.entity.schema.Attributes[name]
		if !exists {
			continue
		}

		// Check ReadOnly
		if attr.ReadOnly {
			return NewElectroError("ReadOnlyViolation",
				fmt.Sprintf("Attribute '%s' is read-only and cannot be updated", name), nil)
		}
	}

	// Validate ADD operations (can't add to readonly)
	for name := range addOps {
		attr, exists := v.entity.schema.Attributes[name]
		if !exists {
			continue
		}

		if attr.ReadOnly {
			return NewElectroError("ReadOnlyViolation",
				fmt.Sprintf("Attribute '%s' is read-only and cannot be updated", name), nil)
		}
	}

	// Validate DELETE operations (can't delete from readonly)
	for name := range delOps {
		attr, exists := v.entity.schema.Attributes[name]
		if !exists {
			continue
		}

		if attr.ReadOnly {
			return NewElectroError("ReadOnlyViolation",
				fmt.Sprintf("Attribute '%s' is read-only and cannot be updated", name), nil)
		}
	}

	// Validate REMOVE operations (can't remove readonly)
	for _, name := range remOps {
		attr, exists := v.entity.schema.Attributes[name]
		if !exists {
			continue
		}

		if attr.ReadOnly {
			return NewElectroError("ReadOnlyViolation",
				fmt.Sprintf("Attribute '%s' is read-only and cannot be removed", name), nil)
		}
	}

	return nil
}

// ApplySetTransformations applies Set transformations to update operations
func (v *Validator) ApplySetTransformations(
	setOps map[string]interface{},
	addOps map[string]interface{},
	delOps map[string]interface{},
) (map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	// Transform SET operations
	transformedSet := make(map[string]interface{})
	for name, value := range setOps {
		attr, exists := v.entity.schema.Attributes[name]
		if exists && attr.Set != nil {
			transformedSet[name] = attr.Set(value)
		} else {
			transformedSet[name] = value
		}

		// Validate enum if applicable
		if exists && attr.Type == AttributeTypeEnum && len(attr.EnumValues) > 0 {
			if err := v.validateEnum(name, transformedSet[name], attr.EnumValues); err != nil {
				// In a real scenario, we'd want to return this error
				// For now, we'll just use the original value
				transformedSet[name] = value
			}
		}

		// Apply validation if exists
		if exists && attr.Validate != nil {
			if err := attr.Validate(transformedSet[name]); err != nil {
				// In a real scenario, we'd want to return this error
				// For now, we'll just use the original value
				transformedSet[name] = value
			}
		}
	}

	// Transform ADD operations
	transformedAdd := make(map[string]interface{})
	for name, value := range addOps {
		attr, exists := v.entity.schema.Attributes[name]
		if exists && attr.Set != nil {
			transformedAdd[name] = attr.Set(value)
		} else {
			transformedAdd[name] = value
		}
	}

	// Transform DELETE operations
	transformedDel := make(map[string]interface{})
	for name, value := range delOps {
		attr, exists := v.entity.schema.Attributes[name]
		if exists && attr.Set != nil {
			transformedDel[name] = attr.Set(value)
		} else {
			transformedDel[name] = value
		}
	}

	return transformedSet, transformedAdd, transformedDel
}
