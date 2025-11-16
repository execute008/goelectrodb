package internal

import (
	"testing"
)

func TestMakeKey(t *testing.T) {
	tests := []struct {
		name     string
		options  KeyOptions
		facets   []string
		supplied map[string]interface{}
		labels   []FacetLabel
		expected KeyResult
	}{
		{
			name: "simple key with all facets",
			options: KeyOptions{
				Prefix: "$service#entity",
			},
			facets: []string{"mall", "building", "unit"},
			supplied: map[string]interface{}{
				"mall":     "EastPointe",
				"building": "BuildingA",
				"unit":     "B54",
			},
			labels: []FacetLabel{
				{Name: "mall", Label: "mall"},
				{Name: "building", Label: "building"},
				{Name: "unit", Label: "unit"},
			},
			expected: KeyResult{
				Key:       "$service#entity#mall_EastPointe#building_BuildingA#unit_B54",
				Fulfilled: true,
			},
		},
		{
			name: "partial key - missing last facet",
			options: KeyOptions{
				Prefix:          "$service#entity",
				ExcludeLabelTail: true,
			},
			facets: []string{"mall", "building", "unit"},
			supplied: map[string]interface{}{
				"mall":     "EastPointe",
				"building": "BuildingA",
			},
			labels: []FacetLabel{
				{Name: "mall", Label: "mall"},
				{Name: "building", Label: "building"},
				{Name: "unit", Label: "unit"},
			},
			expected: KeyResult{
				Key:       "$service#entity#mall_EastPointe#building_BuildingA",
				Fulfilled: false,
			},
		},
		{
			name: "single facet key",
			options: KeyOptions{
				Prefix: "$service#entity",
			},
			facets: []string{"id"},
			supplied: map[string]interface{}{
				"id": "12345",
			},
			labels: []FacetLabel{
				{Name: "id", Label: "id"},
			},
			expected: KeyResult{
				Key:       "$service#entity#id_12345",
				Fulfilled: true,
			},
		},
		{
			name: "key with postfix",
			options: KeyOptions{
				Prefix:  "$service#entity",
				Postfix: stringPtr("#suffix"),
			},
			facets: []string{"mall"},
			supplied: map[string]interface{}{
				"mall": "EastPointe",
			},
			labels: []FacetLabel{
				{Name: "mall", Label: "mall"},
			},
			expected: KeyResult{
				Key:       "$service#entity#mall_EastPointe#suffix",
				Fulfilled: true,
			},
		},
		{
			name: "key with uppercase casing",
			options: KeyOptions{
				Prefix: "$service#entity",
				Casing: stringPtr("upper"),
			},
			facets: []string{"mall"},
			supplied: map[string]interface{}{
				"mall": "EastPointe",
			},
			labels: []FacetLabel{
				{Name: "mall", Label: "mall"},
			},
			expected: KeyResult{
				Key:       "$SERVICE#ENTITY#MALL_EASTPOINTE",
				Fulfilled: true,
			},
		},
		{
			name: "key with lowercase casing",
			options: KeyOptions{
				Prefix: "$service#entity",
				Casing: stringPtr("lower"),
			},
			facets: []string{"mall"},
			supplied: map[string]interface{}{
				"mall": "EastPointe",
			},
			labels: []FacetLabel{
				{Name: "mall", Label: "mall"},
			},
			expected: KeyResult{
				Key:       "$service#entity#mall_eastpointe",
				Fulfilled: true,
			},
		},
		{
			name: "empty key - no supplied values",
			options: KeyOptions{
				Prefix:          "$service#entity",
				ExcludeLabelTail: true,
			},
			facets:   []string{"mall", "building"},
			supplied: map[string]interface{}{},
			labels: []FacetLabel{
				{Name: "mall", Label: "mall"},
				{Name: "building", Label: "building"},
			},
			expected: KeyResult{
				Key:       "$service#entity",
				Fulfilled: false,
			},
		},
		{
			name: "numeric values",
			options: KeyOptions{
				Prefix: "$service#entity",
			},
			facets: []string{"year", "month", "day"},
			supplied: map[string]interface{}{
				"year":  2023,
				"month": 12,
				"day":   25,
			},
			labels: []FacetLabel{
				{Name: "year", Label: "year"},
				{Name: "month", Label: "month"},
				{Name: "day", Label: "day"},
			},
			expected: KeyResult{
				Key:       "$service#entity#year_2023#month_12#day_25",
				Fulfilled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeKey(tt.options, tt.facets, tt.supplied, tt.labels)

			if result.Key != tt.expected.Key {
				t.Errorf("Expected key '%s', got '%s'", tt.expected.Key, result.Key)
			}

			if result.Fulfilled != tt.expected.Fulfilled {
				t.Errorf("Expected fulfilled=%v, got %v", tt.expected.Fulfilled, result.Fulfilled)
			}
		})
	}
}

func TestBuildPrefix(t *testing.T) {
	tests := []struct {
		service  string
		entity   string
		expected string
	}{
		{
			service:  "MallStoreDirectory",
			entity:   "MallStores",
			expected: "$MallStoreDirectory#MallStores",
		},
		{
			service:  "UserService",
			entity:   "User",
			expected: "$UserService#User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.service+"/"+tt.entity, func(t *testing.T) {
			result := BuildPrefix(tt.service, tt.entity)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildLabels(t *testing.T) {
	facets := []string{"mall", "building", "unit"}
	labels := BuildLabels(facets)

	if len(labels) != 3 {
		t.Fatalf("Expected 3 labels, got %d", len(labels))
	}

	for i, facet := range facets {
		if labels[i].Name != facet {
			t.Errorf("Expected name '%s', got '%s'", facet, labels[i].Name)
		}
		if labels[i].Label != facet {
			t.Errorf("Expected label '%s', got '%s'", facet, labels[i].Label)
		}
	}
}

func TestFormatKeyCasing(t *testing.T) {
	tests := []struct {
		key      string
		casing   string
		expected string
	}{
		{
			key:      "HelloWorld",
			casing:   "upper",
			expected: "HELLOWORLD",
		},
		{
			key:      "HelloWorld",
			casing:   "lower",
			expected: "helloworld",
		},
		{
			key:      "HelloWorld",
			casing:   "none",
			expected: "HelloWorld",
		},
		{
			key:      "HelloWorld",
			casing:   "default",
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.casing, func(t *testing.T) {
			result := formatKeyCasing(tt.key, tt.casing)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
