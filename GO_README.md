# Go ElectroDB

A Go port of [ElectroDB](https://github.com/tywalch/electrodb) - a DynamoDB library to ease the use of modeling complex hierarchical relationships and implementing a Single Table Design while keeping your query code readable.

## Status

This is an **initial port** of ElectroDB to Go. The core functionality has been implemented including:

✅ Entity schema definition and validation
✅ Composite key building for single-table design
✅ CRUD operations (Get, Put, Update, Delete)
✅ Query operations with fluent API
✅ DynamoDB parameter generation
✅ Comprehensive test coverage

### Planned Features

The following features from the JavaScript version are planned for future releases:

- 🔄 Filter expressions and where clauses
- 🔄 Service (multi-entity) support
- 🔄 Collections for cross-entity queries
- 🔄 Batch operations (BatchGet, BatchWrite)
- 🔄 Transaction support
- 🔄 Pagination with cursors
- 🔄 Custom attribute validators
- 🔄 Attribute watch functionality

## Installation

```bash
go get github.com/execute008/go-electrodb
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/execute008/go-electrodb/electrodb"
)

func main() {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	// Define your entity schema
	schema := &electrodb.Schema{
		Service: "MallDirectory",
		Entity:  "MallStore",
		Table:   "StoreDirectory",
		Version: "1",
		Attributes: map[string]*electrodb.AttributeDefinition{
			"id": {
				Type:     electrodb.AttributeTypeString,
				Required: false,
				Field:    "storeLocationId",
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
				Type:       electrodb.AttributeTypeEnum,
				Required:   true,
				EnumValues: []interface{}{"food", "clothing", "electronics"},
			},
		},
		Indexes: map[string]*electrodb.IndexDefinition{
			"store": {
				PK: electrodb.FacetDefinition{
					Field:  "pk",
					Facets: []string{"id"},
				},
			},
			"units": {
				Index: stringPtr("gsi1pk-gsi1sk-index"),
				PK: electrodb.FacetDefinition{
					Field:  "gsi1pk",
					Facets: []string{"mall"},
				},
				SK: &electrodb.FacetDefinition{
					Field:  "gsi1sk",
					Facets: []string{"building", "unit", "store"},
				},
			},
		},
	}

	// Create entity
	entity, err := electrodb.NewEntity(schema, &electrodb.Config{
		Client: client,
		Table:  stringPtr("StoreDirectory"),
	})
	if err != nil {
		panic(err)
	}

	// Put an item
	item := electrodb.Item{
		"id":       "123",
		"mall":     "EastPointe",
		"store":    "LatteLarrys",
		"building": "BuildingA",
		"unit":     "B54",
		"category": "food",
	}

	putParams, err := entity.Put(item).Params()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Put params: %+v\n", putParams)

	// Get an item
	keys := electrodb.Keys{
		"id": "123",
	}

	getParams, err := entity.Get(keys).Params()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Get params: %+v\n", getParams)

	// Query by mall
	queryParams, err := entity.Query("units").Query("EastPointe").Params()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Query params: %+v\n", queryParams)

	// Update an item
	updateParams, err := entity.Update(keys).
		Set(map[string]interface{}{
			"category": "electronics",
		}).
		Params()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Update params: %+v\n", updateParams)
}

func stringPtr(s string) *string {
	return &s
}
```

## Key Concepts

### Single Table Design

Go ElectroDB helps you implement DynamoDB single-table design by:

1. **Composite Keys**: Automatically building partition and sort keys from your entity facets
2. **Entity Isolation**: Prefixing keys with service and entity names to prevent collisions
3. **Access Patterns**: Defining indexes that map to your query patterns

### Schema Definition

Define your entity schema with:

- **Attributes**: Field definitions with types, validation, and defaults
- **Indexes**: Primary and secondary indexes with composite facets
- **Facets**: Attributes that compose your keys

### Operations

#### Put/Create
```go
// Put (create or replace)
entity.Put(item).Go()

// Create (fails if exists)
entity.Create(item).Go()
```

#### Get
```go
entity.Get(keys).Go()
```

#### Update
```go
entity.Update(keys).
	Set(map[string]interface{}{"name": "New Name"}).
	Add(map[string]interface{}{"count": 1}).
	Remove([]string{"oldField"}).
	Go()
```

#### Delete
```go
entity.Delete(keys).Go()
```

#### Query
```go
// Query by partition key
entity.Query("indexName").Query(facet1, facet2).Go()

// With sort key conditions
entity.Query("indexName").Query(facet1).
	Begins("prefix").
	Go()

// Available sort key conditions:
// .Eq(value)
// .Gt(value)
// .Gte(value)
// .Lt(value)
// .Lte(value)
// .Between(start, end)
// .Begins(prefix)
```

### Parameter Generation

All operations support `.Params()` to generate DynamoDB parameters without executing:

```go
params, err := entity.Put(item).Params()
// params contains the DynamoDB PutItem parameters
```

This is useful for:
- Debugging
- Batch operations
- Transaction building
- Custom execution logic

## Architecture

The library is organized into:

- **electrodb**: Main package with Entity, operations, and types
- **electrodb/internal**: Internal utilities for key building

### Key Building

Keys are automatically built from facets using the format:
```
$service#entity#facet1_value1#facet2_value2
```

Example:
- Service: `MallDirectory`
- Entity: `MallStore`
- Facets: `[mall: "EastPointe", building: "A"]`
- Result: `$MallDirectory#MallStore#mall_EastPointe#building_A`

## Testing

The library includes comprehensive tests:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./electrodb
go test ./electrodb/internal
```

Current test coverage:
- ✅ Schema validation
- ✅ Entity creation
- ✅ Key building with various configurations
- ✅ Parameter generation for all operations
- ✅ Required attribute validation
- ✅ Default value application

## Differences from JavaScript ElectroDB

### Type Safety

Go's static typing provides compile-time safety that JavaScript can't:
- Schema definitions are type-checked
- Operation parameters are validated
- Attribute types are enforced

### API Design

The Go port maintains the fluent API design while adapting to Go idioms:
- Method chaining with value receivers
- Explicit error handling
- Context support (planned)
- Options pattern for configuration

### Not Yet Implemented

Features from the JS version not yet in the Go port:
- Service (multi-entity management)
- Collections
- Filter expressions
- Batch operations
- Transactions
- Pagination
- Attribute watchers
- Custom validators beyond basic type checking

## Examples

See the `examples/` directory for complete examples:

- Basic CRUD operations
- Query patterns
- Schema design
- Testing strategies

## Contributing

This is an initial port with room for improvement. Contributions welcome!

### Development

```bash
# Clone the repository
git clone https://github.com/execute008/go-electrodb
cd go-electrodb

# Run tests
go test ./...

# Run specific tests
go test -run TestEntityCreation ./electrodb
```

### Roadmap

1. **Phase 1** (✅ Complete): Core CRUD + Query
2. **Phase 2** (🔄 In Progress): Filters, Service, Collections
3. **Phase 3** (📋 Planned): Batch, Transactions, Pagination
4. **Phase 4** (📋 Planned): Advanced features (watchers, custom validators)

## License

MIT License - see LICENSE file

## Credits

- Original [ElectroDB](https://github.com/tywalch/electrodb) by Tyler Walch
- Go port by execute008

## Resources

- [Original ElectroDB Documentation](https://electrodb.dev/)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
- [Single Table Design](https://www.alexdebrie.com/posts/dynamodb-single-table/)
