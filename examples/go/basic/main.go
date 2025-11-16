package main

import (
	"fmt"

	"github.com/execute008/go-electrodb/electrodb"
)

func main() {
	// Define a schema for a mall store directory
	schema := &electrodb.Schema{
		Service: "MallDirectory",
		Entity:  "Store",
		Table:   "StoreDirectory",
		Version: "1",
		Attributes: map[string]*electrodb.AttributeDefinition{
			"id": {
				Type:     electrodb.AttributeTypeString,
				Required: false,
				Default: func() interface{} {
					return "auto-generated-id"
				},
			},
			"mall": {
				Type:     electrodb.AttributeTypeString,
				Required: true,
			},
			"store": {
				Type:     electrodb.AttributeTypeString,
				Required: true,
			},
			"building": {
				Type:     electrodb.AttributeTypeString,
				Required: true,
			},
			"unit": {
				Type:     electrodb.AttributeTypeString,
				Required: true,
			},
			"category": {
				Type:     electrodb.AttributeTypeString,
				Required: true,
			},
			"rent": {
				Type:     electrodb.AttributeTypeNumber,
				Required: false,
				Default: func() interface{} {
					return 0
				},
			},
		},
		Indexes: map[string]*electrodb.IndexDefinition{
			// Primary index - access by store ID
			"store": {
				PK: electrodb.FacetDefinition{
					Field:  "pk",
					Facets: []string{"id"},
				},
			},
			// GSI - query stores by mall and location
			"units": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK: electrodb.FacetDefinition{
					Field:  "gsi1pk",
					Facets: []string{"mall"},
				},
				SK: &electrodb.FacetDefinition{
					Field:  "gsi1sk",
					Facets: []string{"building", "unit"},
				},
			},
			// GSI - query stores by mall and category
			"categories": {
				Index: stringPtr("gsi2pk-gsi2sk-index"),
				PK: electrodb.FacetDefinition{
					Field:  "gsi2pk",
					Facets: []string{"mall"},
				},
				SK: &electrodb.FacetDefinition{
					Field:  "gsi2sk",
					Facets: []string{"category", "store"},
				},
			},
		},
	}

	// Create entity (without client for this example)
	entity, err := electrodb.NewEntity(schema, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Go ElectroDB Example ===\n")

	// Example 1: Put Item Parameters
	fmt.Println("1. PUT ITEM")
	putItem := electrodb.Item{
		"id":       "store-001",
		"mall":     "EastPointe",
		"store":    "LatteLarrys",
		"building": "BuildingA",
		"unit":     "B54",
		"category": "coffee",
		"rent":     2500,
	}

	putParams, err := entity.Put(putItem).Params()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Put Parameters:\n")
	fmt.Printf("  TableName: %v\n", putParams["TableName"])
	fmt.Printf("  Item includes keys: pk, gsi1pk, gsi1sk, gsi2pk, gsi2sk\n\n")

	// Example 2: Get Item Parameters
	fmt.Println("2. GET ITEM")
	getParams, err := entity.Get(electrodb.Keys{
		"id": "store-001",
	}).Params()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Get Parameters:\n")
	fmt.Printf("  TableName: %v\n", getParams["TableName"])
	fmt.Printf("  Key: %v\n\n", getParams["Key"])

	// Example 3: Query by Mall (using units index)
	fmt.Println("3. QUERY STORES BY MALL")
	queryParams, err := entity.Query("units").Query("EastPointe").Params()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Query Parameters:\n")
	fmt.Printf("  TableName: %v\n", queryParams["TableName"])
	fmt.Printf("  IndexName: %v\n", queryParams["IndexName"])
	fmt.Printf("  KeyConditionExpression: %v\n\n", queryParams["KeyConditionExpression"])

	// Example 4: Query with Sort Key Condition
	fmt.Println("4. QUERY STORES IN SPECIFIC BUILDING")
	queryParams2, err := entity.Query("units").
		Query("EastPointe").
		Begins("BuildingA").
		Params()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Query with SK Parameters:\n")
	fmt.Printf("  TableName: %v\n", queryParams2["TableName"])
	fmt.Printf("  IndexName: %v\n", queryParams2["IndexName"])
	fmt.Printf("  KeyConditionExpression: %v\n\n", queryParams2["KeyConditionExpression"])

	// Example 5: Update Item
	fmt.Println("5. UPDATE ITEM")
	updateParams, err := entity.Update(electrodb.Keys{
		"id": "store-001",
	}).Set(map[string]interface{}{
		"category": "cafe",
	}).Add(map[string]interface{}{
		"rent": 100,
	}).Params()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Update Parameters:\n")
	fmt.Printf("  TableName: %v\n", updateParams["TableName"])
	fmt.Printf("  UpdateExpression: %v\n", updateParams["UpdateExpression"])
	fmt.Printf("  ReturnValues: %v\n\n", updateParams["ReturnValues"])

	// Example 6: Delete Item
	fmt.Println("6. DELETE ITEM")
	deleteParams, err := entity.Delete(electrodb.Keys{
		"id": "store-001",
	}).Params()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Delete Parameters:\n")
	fmt.Printf("  TableName: %v\n", deleteParams["TableName"])
	fmt.Printf("  Key: %v\n\n", deleteParams["Key"])

	// Example 7: Demonstrate Key Building
	fmt.Println("7. KEY BUILDING DEMONSTRATION")
	fmt.Println("Schema:")
	fmt.Println("  Service: MallDirectory")
	fmt.Println("  Entity: Store")
	fmt.Println("  PK Facets: [id]")
	fmt.Println("  SK Facets (units): [mall, building, unit]")
	fmt.Println("\nGenerated Keys:")
	fmt.Println("  PK: $MallDirectory#Store#id_store-001")
	fmt.Println("  GSI1PK: $MallDirectory#Store#mall_EastPointe")
	fmt.Println("  GSI1SK: $MallDirectory#Store#building_BuildingA#unit_B54")
	fmt.Println("\nThis allows:")
	fmt.Println("  - Get by store ID (primary index)")
	fmt.Println("  - Query all stores in a mall (GSI1)")
	fmt.Println("  - Query stores in a specific building (GSI1 + begins_with)")
	fmt.Println("  - Query stores by category (GSI2)")

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nTo use with real DynamoDB:")
	fmt.Println("1. Create AWS DynamoDB client")
	fmt.Println("2. Pass client in Config when creating entity")
	fmt.Println("3. Use .Go() instead of .Params() to execute operations")
}

func stringPtr(s string) *string {
	return &s
}
