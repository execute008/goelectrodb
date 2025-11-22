// Package electrodb provides a Go implementation of ElectroDB, a DynamoDB library
// that simplifies working with single-table design patterns.
//
// GoElectroDB brings the power of ElectroDB's JavaScript library to Go developers,
// offering a fluent API for defining entities, building queries, and managing
// complex access patterns in DynamoDB.
//
// # Key Features
//
//   - Single-table design support with automatic key composition
//   - Fluent query builder with type-safe operations
//   - Batch operations (BatchGet, BatchWrite)
//   - Transaction support (TransactWrite, TransactGet)
//   - Automatic pagination with cursor-based navigation
//   - Filter expressions with named filters
//   - Condition expressions for conditional mutations
//   - Automatic timestamps (createdAt, updatedAt)
//   - Attribute validation and transformation
//   - TTL (Time-To-Live) support
//
// # Quick Start
//
// Define an entity schema:
//
//	schema := &electrodb.Schema{
//		Service: "myapp",
//		Entity:  "user",
//		Table:   "my-table",
//		Attributes: map[string]*electrodb.AttributeDefinition{
//			"userId":    {Type: electrodb.AttributeTypeString, Required: true},
//			"email":     {Type: electrodb.AttributeTypeString, Required: true},
//			"name":      {Type: electrodb.AttributeTypeString},
//			"createdAt": {Type: electrodb.AttributeTypeString},
//		},
//		Indexes: map[string]*electrodb.IndexDefinition{
//			"primary": {
//				PK: electrodb.FacetDefinition{Field: "pk", Facets: []string{"userId"}},
//				SK: &electrodb.FacetDefinition{Field: "sk", Facets: []string{"email"}},
//			},
//		},
//	}
//
// Create an entity and perform operations:
//
//	entity := electrodb.NewEntity(schema, &electrodb.Config{
//		Client: dynamoClient,
//	})
//
//	// Create an item
//	resp, err := entity.Put(electrodb.Item{
//		"userId": "user-123",
//		"email":  "user@example.com",
//		"name":   "John Doe",
//	}).Go()
//
//	// Query items
//	results, err := entity.Query().Primary(electrodb.Keys{
//		"userId": "user-123",
//	}).Go()
//
// # Error Handling
//
// All operations return an *ElectroError on failure. Use the Code field to
// programmatically handle specific error types:
//
//	resp, err := entity.Get(keys).Go()
//	if err != nil {
//		var electroErr *electrodb.ElectroError
//		if errors.As(err, &electroErr) {
//			switch electroErr.Code {
//			case electrodb.ErrNoClientProvided:
//				// Handle missing client
//			case electrodb.ErrDynamoDB:
//				// Handle DynamoDB errors
//			}
//		}
//	}
//
// # Services and Collections
//
// Group related entities into services for collection queries:
//
//	service := electrodb.NewService(&electrodb.ServiceConfig{
//		Table:  "my-table",
//		Client: dynamoClient,
//	})
//	service.AddEntity("user", userEntity)
//	service.AddEntity("order", orderEntity)
//
//	// Query a collection across entities
//	results, err := service.Collection("orders_by_user").
//		Query(electrodb.Keys{"userId": "user-123"}).Go()
//
// For more information, see the README and examples in the repository.
package electrodb
