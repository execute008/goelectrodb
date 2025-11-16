package main

import (
	"fmt"
	"log"

	"github.com/execute008/goelectrodb/electrodb"
)

// Simple quick start example showing basic CRUD operations

func main() {
	fmt.Println("ElectroDB Go - Quick Start Example")
	fmt.Println("===================================\n")

	// 1. Define a simple user schema
	userSchema := &electrodb.Schema{
		Service: "MyApp",
		Entity:  "User",
		Table:   "users-table",

		Attributes: map[string]*electrodb.AttributeDefinition{
			// Required attributes
			"userId": {
				Type:     electrodb.AttributeTypeString,
				Required: true,
			},
			"email": {
				Type:     electrodb.AttributeTypeString,
				Required: true,
			},

			// Optional attributes
			"firstName": {Type: electrodb.AttributeTypeString},
			"lastName":  {Type: electrodb.AttributeTypeString},
			"age":       {Type: electrodb.AttributeTypeNumber},
			"status": {
				Type:       electrodb.AttributeTypeEnum,
				EnumValues: []interface{}{"active", "inactive"},
				Default:    func() interface{} { return "active" },
			},
		},

		// Define primary key structure
		Indexes: map[string]*electrodb.IndexDefinition{
			"primary": {
				PK: electrodb.FacetDefinition{
					Field:  "pk",
					Facets: []string{"userId"},
				},
			},
			"byEmail": {
				Index: stringPtr("gsi1"),
				PK: electrodb.FacetDefinition{
					Field:  "gsi1pk",
					Facets: []string{"email"},
				},
			},
		},

		// Enable automatic timestamps
		Timestamps: &electrodb.TimestampsConfig{
			CreatedAt: "createdAt",
			UpdatedAt: "updatedAt",
		},
	}

	// 2. Create the entity
	entity, err := electrodb.NewEntity(userSchema, nil)
	if err != nil {
		log.Fatalf("Failed to create entity: %v", err)
	}

	fmt.Println("✅ Entity created successfully\n")

	// 3. Create (Put) a new user
	fmt.Println("📝 Creating user...")
	putParams, err := entity.Put(electrodb.Item{
		"userId":    "user-123",
		"email":     "john@example.com",
		"firstName": "John",
		"lastName":  "Doe",
		"age":       30,
	}).Params()

	if err != nil {
		log.Fatalf("Failed to create put params: %v", err)
	}

	fmt.Printf("   Put operation params generated\n")
	fmt.Printf("   Table: %s\n", putParams["TableName"])
	fmt.Printf("   Item includes: userId, email, firstName, lastName, age\n")
	fmt.Printf("   Auto-added: createdAt, updatedAt, pk\n\n")

	// 4. Get a user
	fmt.Println("🔍 Getting user...")
	getParams, err := entity.Get(electrodb.Keys{
		"userId": "user-123",
	}).Params()

	if err != nil {
		log.Fatalf("Failed to create get params: %v", err)
	}

	fmt.Printf("   Get operation params generated\n")
	fmt.Printf("   Table: %s\n", getParams["TableName"])
	fmt.Printf("   Key: {userId: user-123}\n\n")

	// 5. Update a user
	fmt.Println("✏️  Updating user...")
	updateParams, err := entity.Update(electrodb.Keys{
		"userId": "user-123",
	}).
		Set(map[string]interface{}{
			"firstName": "Johnny",
			"lastName":  "Doe",
		}).
		Add(map[string]interface{}{
			"age": 1, // Increment age
		}).
		Params()

	if err != nil {
		log.Fatalf("Failed to create update params: %v", err)
	}

	fmt.Printf("   Update operation params generated\n")
	fmt.Printf("   Table: %s\n", updateParams["TableName"])
	fmt.Printf("   Expression: %s\n", updateParams["UpdateExpression"])
	fmt.Printf("   Note: updatedAt automatically updated\n\n")

	// 6. Query users
	fmt.Println("🔎 Querying users...")
	queryParams, err := entity.Query("primary").
		Query("user-123").
		Params()

	if err != nil {
		log.Fatalf("Failed to create query params: %v", err)
	}

	fmt.Printf("   Query operation params generated\n")
	fmt.Printf("   Table: %s\n", queryParams["TableName"])
	fmt.Printf("   Key Condition: %s\n\n", queryParams["KeyConditionExpression"])

	// 7. Delete a user
	fmt.Println("🗑️  Deleting user...")
	deleteParams, err := entity.Delete(electrodb.Keys{
		"userId": "user-123",
	}).Params()

	if err != nil {
		log.Fatalf("Failed to create delete params: %v", err)
	}

	fmt.Printf("   Delete operation params generated\n")
	fmt.Printf("   Table: %s\n", deleteParams["TableName"])
	fmt.Printf("   Key: {userId: user-123}\n\n")

	// 8. Batch operations
	fmt.Println("📦 Batch operations...")

	// Batch Get
	batchGetOp := entity.BatchGet([]electrodb.Keys{
		{"userId": "user-1"},
		{"userId": "user-2"},
		{"userId": "user-3"},
	})
	fmt.Printf("   Batch get operation created for 3 users\n")

	// Batch Write
	batchWriteOp := entity.BatchWrite().
		Put([]electrodb.Item{
			{"userId": "user-4", "email": "user4@example.com"},
			{"userId": "user-5", "email": "user5@example.com"},
		}).
		Delete([]electrodb.Keys{
			{"userId": "user-6"},
		})
	fmt.Printf("   Batch write operation created (2 puts, 1 delete)\n")

	_, _ = batchGetOp, batchWriteOp // Use variables to avoid warnings

	fmt.Println("\n✨ Quick start complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("- See examples/comprehensive for all features")
	fmt.Println("- Read README_GO.md for detailed documentation")
	fmt.Println("- Add AWS DynamoDB client to actually execute operations")
	fmt.Println("- Check out validation, transformations, and advanced features")
}

func stringPtr(s string) *string {
	return &s
}
