package electrodb

import (
	"testing"
	"time"
)

func TestPutWithTTL(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Session",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"sessionId": {Type: AttributeTypeString, Required: true},
			"userId":    {Type: AttributeTypeString, Required: true},
			"ttl":       {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"sessionId"}},
			},
		},
		TTL: &TTLConfig{
			Attribute: "ttl",
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test Put with TTL
	putOp := entity.Put(Item{
		"sessionId": "session123",
		"userId":    "user123",
	}).WithTTL(24 * time.Hour)

	// Verify TTL was added to item
	if putOp.item["ttl"] == nil {
		t.Error("Expected ttl attribute to be set")
	}

	ttl, ok := putOp.item["ttl"].(int64)
	if !ok {
		t.Fatal("Expected ttl to be int64")
	}

	// TTL should be approximately 24 hours from now
	expectedTTL := time.Now().Add(24 * time.Hour).Unix()
	if ttl < expectedTTL-10 || ttl > expectedTTL+10 {
		t.Errorf("Expected TTL to be approximately %d, got %d", expectedTTL, ttl)
	}
}

func TestPutWithTTLTimestamp(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Token",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"tokenId": {Type: AttributeTypeString, Required: true},
			"expires": {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"tokenId"}},
			},
		},
		TTL: &TTLConfig{
			Attribute: "expires",
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Use specific timestamp
	specificTime := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	expectedTTL := specificTime.Unix()

	putOp := entity.Put(Item{
		"tokenId": "token123",
	}).WithTTLTimestamp(expectedTTL)

	if putOp.item["expires"] == nil {
		t.Error("Expected expires attribute to be set")
	}

	ttl, ok := putOp.item["expires"].(int64)
	if !ok {
		t.Fatal("Expected expires to be int64")
	}

	if ttl != expectedTTL {
		t.Errorf("Expected TTL to be %d, got %d", expectedTTL, ttl)
	}
}

func TestUpdateWithTTL(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Cache",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"cacheKey": {Type: AttributeTypeString, Required: true},
			"value":    {Type: AttributeTypeString, Required: false},
			"ttl":      {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"cacheKey"}},
			},
		},
		TTL: &TTLConfig{
			Attribute: "ttl",
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test Update with TTL
	updateOp := entity.Update(Keys{"cacheKey": "cache123"}).
		Set(map[string]interface{}{
			"value": "cached data",
		}).
		WithTTL(1 * time.Hour)

	// Verify TTL was added to setOps
	if updateOp.setOps["ttl"] == nil {
		t.Error("Expected ttl to be in setOps")
	}

	ttl, ok := updateOp.setOps["ttl"].(int64)
	if !ok {
		t.Fatal("Expected ttl to be int64")
	}

	expectedTTL := time.Now().Add(1 * time.Hour).Unix()
	if ttl < expectedTTL-10 || ttl > expectedTTL+10 {
		t.Errorf("Expected TTL to be approximately %d, got %d", expectedTTL, ttl)
	}
}

func TestRemoveTTL(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Record",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"recordId": {Type: AttributeTypeString, Required: true},
			"ttl":      {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"recordId"}},
			},
		},
		TTL: &TTLConfig{
			Attribute: "ttl",
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test removing TTL
	updateOp := entity.Update(Keys{"recordId": "rec123"}).
		RemoveTTL()

	// Verify TTL attribute is in remOps
	found := false
	for _, attr := range updateOp.remOps {
		if attr == "ttl" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected ttl to be in remOps")
	}
}

func TestTTLWithoutConfiguration(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Item",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"itemId": {Type: AttributeTypeString, Required: true},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"itemId"}},
			},
		},
		// No TTL configured
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test Put with TTL when TTL not configured - should be silently ignored
	putOp := entity.Put(Item{
		"itemId": "item123",
	}).WithTTL(1 * time.Hour)

	// Should not have added any TTL attribute
	if len(putOp.item) > 1 {
		t.Error("Expected no TTL attribute to be added when TTL not configured")
	}

	// Test Update with TTL when TTL not configured
	updateOp := entity.Update(Keys{"itemId": "item123"}).
		WithTTL(1 * time.Hour)

	if len(updateOp.setOps) > 0 {
		t.Error("Expected no TTL in setOps when TTL not configured")
	}
}

func TestTTLHelperFunctions(t *testing.T) {
	// Test TTLFromNow
	duration := 24 * time.Hour
	ttl := TTLFromNow(duration)
	expected := time.Now().Add(duration).Unix()

	if ttl < expected-10 || ttl > expected+10 {
		t.Errorf("Expected TTL to be approximately %d, got %d", expected, ttl)
	}

	// Test TTLFromTime
	specificTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	ttl = TTLFromTime(specificTime)
	expected = specificTime.Unix()

	if ttl != expected {
		t.Errorf("Expected TTL to be %d, got %d", expected, ttl)
	}

	// Test IsTTLExpired with expired TTL
	expiredTTL := time.Now().Add(-1 * time.Hour).Unix()
	if !IsTTLExpired(expiredTTL) {
		t.Error("Expected TTL to be expired")
	}

	// Test IsTTLExpired with future TTL
	futureTTL := time.Now().Add(1 * time.Hour).Unix()
	if IsTTLExpired(futureTTL) {
		t.Error("Expected TTL to not be expired")
	}

	// Test TimeUntilTTL
	futureTTL = time.Now().Add(2 * time.Hour).Unix()
	duration = TimeUntilTTL(futureTTL)

	if duration < 1*time.Hour+50*time.Minute || duration > 2*time.Hour+10*time.Minute {
		t.Errorf("Expected duration to be approximately 2 hours, got %v", duration)
	}
}

func TestCombinedTTLAndSetOperations(t *testing.T) {
	schema := &Schema{
		Service: "TestService",
		Entity:  "Document",
		Table:   "TestTable",
		Attributes: map[string]*AttributeDefinition{
			"docId": {Type: AttributeTypeString, Required: true},
			"tags":  {Type: AttributeTypeSet, Required: false},
			"data":  {Type: AttributeTypeString, Required: false},
			"ttl":   {Type: AttributeTypeNumber, Required: false},
		},
		Indexes: map[string]*IndexDefinition{
			"primary": {
				PK: FacetDefinition{Field: "pk", Facets: []string{"docId"}},
			},
		},
		TTL: &TTLConfig{
			Attribute: "ttl",
		},
	}

	entity, err := NewEntity(schema, nil)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Test combining TTL with other operations
	updateOp := entity.Update(Keys{"docId": "doc123"}).
		Set(map[string]interface{}{
			"data": "important data",
		}).
		AddToSet("tags", []string{"verified"}).
		WithTTL(7 * 24 * time.Hour) // 7 days

	// Verify all operations
	if updateOp.setOps["data"] == nil {
		t.Error("Expected data in setOps")
	}

	if updateOp.setOps["ttl"] == nil {
		t.Error("Expected ttl in setOps")
	}

	if updateOp.addOps["tags"] == nil {
		t.Error("Expected tags in addOps")
	}

	// Build params to verify it all works together
	params, err := updateOp.Params()
	if err != nil {
		t.Fatalf("Failed to build params: %v", err)
	}

	if params["UpdateExpression"] == nil {
		t.Error("Expected UpdateExpression to be set")
	}
}
