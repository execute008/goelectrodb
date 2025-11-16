package internal

import (
	"fmt"
	"strings"
)

// KeyOptions defines options for key building
type KeyOptions struct {
	Prefix          string
	IsCustom        bool
	Casing          *string
	Postfix         *string
	ExcludeLabelTail bool
	ExcludePostfix  bool
}

// FacetLabel represents a facet with its label
type FacetLabel struct {
	Name  string
	Label string
}

// KeyResult contains the built key and whether all facets were fulfilled
type KeyResult struct {
	Key       string
	Fulfilled bool
}

// MakeKey builds a composite key from facets and supplied values
// This is the Go port of ElectroDB's _makeKey function
func MakeKey(
	options KeyOptions,
	facets []string,
	supplied map[string]interface{},
	labels []FacetLabel,
) KeyResult {
	key := options.Prefix
	foundCount := 0

	for i := 0; i < len(labels); i++ {
		label := labels[i]
		value, exists := supplied[label.Name]

		// If value is undefined and we should exclude tail, break
		if !exists && options.ExcludeLabelTail {
			break
		}

		// Build the key part with label
		if options.IsCustom {
			key = fmt.Sprintf("%s%s", key, label.Label)
		} else {
			key = fmt.Sprintf("%s#%s_", key, label.Label)
		}

		// If value is undefined, we can't build any more of the key
		if !exists {
			break
		}

		foundCount++

		// Append the value
		key = fmt.Sprintf("%s%v", key, value)
	}

	// Check if all facets were fulfilled
	fulfilled := foundCount == len(labels)

	// Apply postfix if fulfilled and not excluded
	if fulfilled && options.Postfix != nil && !options.ExcludePostfix {
		key = fmt.Sprintf("%s%s", key, *options.Postfix)
	}

	// Apply casing if specified
	if options.Casing != nil {
		key = formatKeyCasing(key, *options.Casing)
	}

	return KeyResult{
		Key:       key,
		Fulfilled: fulfilled,
	}
}

// formatKeyCasing applies casing transformations to a key
func formatKeyCasing(key string, casing string) string {
	switch strings.ToLower(casing) {
	case "upper":
		return strings.ToUpper(key)
	case "lower":
		return strings.ToLower(key)
	case "none", "default":
		return key
	default:
		return key
	}
}

// BuildPrefix builds the entity/service prefix for keys
func BuildPrefix(service, entity string) string {
	return fmt.Sprintf("$%s#%s", service, entity)
}

// BuildLabels creates FacetLabel array from facet names
func BuildLabels(facets []string) []FacetLabel {
	labels := make([]FacetLabel, len(facets))
	for i, facet := range facets {
		labels[i] = FacetLabel{
			Name:  facet,
			Label: facet,
		}
	}
	return labels
}
