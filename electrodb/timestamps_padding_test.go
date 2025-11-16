package electrodb

import (
	"testing"
	"time"
)

// Test automatic timestamps on create
func TestTimestampsOnCreate(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":        {Type: AttributeTypeString, Required: true},
			"name":      {Type: AttributeTypeString},
			"createdAt": {Type: AttributeTypeNumber},
			"updatedAt": {Type: AttributeTypeNumber},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
		Timestamps: &TimestampsConfig{
			CreatedAt: "createdAt",
			UpdatedAt: "updatedAt",
		},
	}

	// Test applying timestamps to new item
	item := Item{
		"id":   "123",
		"name": "Test Item",
	}

	result := ApplyTimestamps(item, schema, false)

	// createdAt should be set
	if _, exists := result["createdAt"]; !exists {
		t.Error("createdAt should be automatically set")
	}

	// updatedAt should be set
	if _, exists := result["updatedAt"]; !exists {
		t.Error("updatedAt should be automatically set")
	}

	// Values should be recent Unix timestamps
	createdAt, ok := result["createdAt"].(int64)
	if !ok {
		t.Error("createdAt should be int64")
	}

	now := time.Now().Unix()
	if createdAt < now-5 || createdAt > now+5 {
		t.Errorf("createdAt should be recent, got %d, expected around %d", createdAt, now)
	}
}

// Test timestamps on update
func TestTimestampsOnUpdate(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":        {Type: AttributeTypeString, Required: true},
			"updatedAt": {Type: AttributeTypeNumber},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
		Timestamps: &TimestampsConfig{
			UpdatedAt: "updatedAt",
		},
	}

	setOps := map[string]interface{}{
		"name": "Updated Name",
	}

	result := ApplyUpdateTimestamps(setOps, schema)

	// updatedAt should be added to setOps
	if _, exists := result["updatedAt"]; !exists {
		t.Error("updatedAt should be automatically added to update")
	}

	// Original setOps should not be modified
	if _, exists := setOps["updatedAt"]; exists {
		t.Error("Original setOps should not be modified")
	}

	// updatedAt should be recent
	updatedAt, ok := result["updatedAt"].(int64)
	if !ok {
		t.Error("updatedAt should be int64")
	}

	now := time.Now().Unix()
	if updatedAt < now-5 || updatedAt > now+5 {
		t.Errorf("updatedAt should be recent, got %d, expected around %d", updatedAt, now)
	}
}

// Test that user-provided timestamps are not overwritten
func TestUserProvidedTimestampsNotOverwritten(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":        {Type: AttributeTypeString, Required: true},
			"createdAt": {Type: AttributeTypeNumber},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
		Timestamps: &TimestampsConfig{
			CreatedAt: "createdAt",
		},
	}

	userTimestamp := int64(1234567890)
	item := Item{
		"id":        "123",
		"createdAt": userTimestamp,
	}

	result := ApplyTimestamps(item, schema, false)

	// User-provided createdAt should not be overwritten
	if result["createdAt"] != userTimestamp {
		t.Errorf("User-provided createdAt should not be overwritten, got %v", result["createdAt"])
	}
}

// Test timestamps without config
func TestTimestampsWithoutConfig(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
		// No Timestamps config
	}

	item := Item{
		"id": "123",
	}

	result := ApplyTimestamps(item, schema, false)

	// Should return item unchanged
	if len(result) != 1 || result["id"] != "123" {
		t.Error("Item should be unchanged when no timestamp config")
	}
}

// Test attribute padding
func TestAttributePadding(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":        {Type: AttributeTypeString, Required: true},
			"version":   {Type: AttributeTypeNumber, Padding: &PaddingConfig{Length: 5, Char: "0"}},
			"orderNum":  {Type: AttributeTypeNumber, Padding: &PaddingConfig{Length: 10, Char: "0"}},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	item := Item{
		"id":       "123",
		"version":  42,
		"orderNum": 1234,
	}

	result := ApplyPadding(item, schema)

	// version should be padded to 5 characters
	if result["version"] != "00042" {
		t.Errorf("Expected version to be '00042', got %v", result["version"])
	}

	// orderNum should be padded to 10 characters
	if result["orderNum"] != "0000001234" {
		t.Errorf("Expected orderNum to be '0000001234', got %v", result["orderNum"])
	}
}

// Test padding removal
func TestPaddingRemoval(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":      {Type: AttributeTypeString, Required: true},
			"version": {Type: AttributeTypeNumber, Padding: &PaddingConfig{Length: 5, Char: "0"}},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	item := Item{
		"id":      "123",
		"version": "00042",
	}

	result := RemovePadding(item, schema)

	// version should be converted back to number
	version, ok := result["version"].(int64)
	if !ok {
		t.Fatalf("Expected version to be int64, got %T", result["version"])
	}

	if version != 42 {
		t.Errorf("Expected version to be 42, got %d", version)
	}
}

// Test padding with custom character
func TestPaddingCustomChar(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"code": {Type: AttributeTypeString, Padding: &PaddingConfig{Length: 8, Char: "#"}},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	item := Item{
		"id":   "123",
		"code": "ABC",
	}

	result := ApplyPadding(item, schema)

	// code should be padded with # characters
	if result["code"] != "#####ABC" {
		t.Errorf("Expected code to be '#####ABC', got %v", result["code"])
	}
}

// Test padding on string values
func TestPaddingOnStrings(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"code": {Type: AttributeTypeString, Padding: &PaddingConfig{Length: 6, Char: "0"}},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	item := Item{
		"id":   "123",
		"code": "42",
	}

	result := ApplyPadding(item, schema)

	// code should be padded
	if result["code"] != "000042" {
		t.Errorf("Expected code to be '000042', got %v", result["code"])
	}
}

// Test padding when value already meets length
func TestPaddingAlreadyMeetsLength(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":   {Type: AttributeTypeString, Required: true},
			"code": {Type: AttributeTypeNumber, Padding: &PaddingConfig{Length: 4, Char: "0"}},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	item := Item{
		"id":   "123",
		"code": 12345, // Already longer than padding length
	}

	result := ApplyPadding(item, schema)

	// code should not be truncated, just converted to string
	if result["code"] != "12345" {
		t.Errorf("Expected code to be '12345', got %v", result["code"])
	}
}

// Test combined timestamps and padding
func TestCombinedTimestampsAndPadding(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":        {Type: AttributeTypeString, Required: true},
			"version":   {Type: AttributeTypeNumber, Padding: &PaddingConfig{Length: 5, Char: "0"}},
			"createdAt": {Type: AttributeTypeNumber},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
		Timestamps: &TimestampsConfig{
			CreatedAt: "createdAt",
		},
	}

	item := Item{
		"id":      "123",
		"version": 3,
	}

	// Apply timestamps first
	withTimestamps := ApplyTimestamps(item, schema, false)

	// Then apply padding
	result := ApplyPadding(withTimestamps, schema)

	// version should be padded
	if result["version"] != "00003" {
		t.Errorf("Expected version to be '00003', got %v", result["version"])
	}

	// createdAt should still be present
	if _, exists := result["createdAt"]; !exists {
		t.Error("createdAt should be present")
	}
}

// Test padding removal with zero value
func TestPaddingRemovalZero(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"count": {Type: AttributeTypeNumber, Padding: &PaddingConfig{Length: 5, Char: "0"}},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	item := Item{
		"id":    "123",
		"count": "00000", // All zeros
	}

	result := RemovePadding(item, schema)

	// Should return 0, not empty string
	count, ok := result["count"].(int64)
	if !ok {
		t.Fatalf("Expected count to be int64, got %T", result["count"])
	}

	if count != 0 {
		t.Errorf("Expected count to be 0, got %d", count)
	}
}

// Test padding without config
func TestPaddingWithoutConfig(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "TestEntity",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"id":    {Type: AttributeTypeString, Required: true},
			"value": {Type: AttributeTypeNumber}, // No padding config
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"id"}},
			},
		},
	}

	item := Item{
		"id":    "123",
		"value": 42,
	}

	result := ApplyPadding(item, schema)

	// value should remain unchanged
	if result["value"] != 42 {
		t.Errorf("Expected value to remain 42, got %v", result["value"])
	}
}
