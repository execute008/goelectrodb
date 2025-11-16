# ElectroDB Go Port

A comprehensive Go port of [ElectroDB](https://github.com/tywalch/electrodb) - a DynamoDB library to ease the use of having multiple entities and complex hierarchical relationships in a single DynamoDB table.

## Features

### ✅ Core Features (Complete)

- **Entity & Service Management** - Define entities and services with full schema support
- **Query & Scan Operations** - Powerful query builder with filter support
- **Batch Operations** - BatchGet and BatchWrite with automatic chunking
- **Transactions** - Full transactional write support
- **Collections** - Query across multiple entities in a single request
- **Cursor-based Pagination** - Automatic and manual pagination support

### ✅ Validation & Transformation (Complete)

- **Custom Validation** - Define validation functions per attribute
- **Get/Set Transformations** - Transform values on read/write
- **Enum Validation** - Enforce allowed values for enum types
- **ReadOnly Attributes** - Prevent modification of immutable fields
- **Hidden Attributes** - Automatically filter sensitive data from responses
- **Default Values** - Auto-apply defaults for missing attributes

### ✅ Advanced Update Operations (Complete)

- **Append/Prepend** - Add to beginning/end of lists
- **Subtract** - Subtract from numeric attributes
- **Data** - Remove specific list elements by index
- **Set Operations** - AddToSet/DeleteFromSet for DynamoDB sets
- **Patch/Upsert** - Flexible update patterns

### ✅ Automation Features (Complete)

- **Automatic Timestamps** - Auto-manage createdAt/updatedAt
- **Attribute Padding** - Zero-pad numbers for proper string sorting
- **TTL Support** - Automatic item expiration

### ✅ Query Features (Complete)

- **Named Filters** - Reusable filter functions defined in schema
- **Filter Expressions** - Complex filter composition
- **Condition Expressions** - Conditional mutations
- **Pagination** - `.Pages()` for automatic, `.Page()` for manual control

## Installation

```bash
go get github.com/execute008/goelectrodb
```

## Quick Start

```go
package main

import (
    "github.com/execute008/goelectrodb/electrodb"
)

func main() {
    // Define schema
    schema := &electrodb.Schema{
        Service: "MyApp",
        Entity:  "User",
        Table:   "my-table",
        Attributes: map[string]*electrodb.AttributeDefinition{
            "userId": {Type: electrodb.AttributeTypeString, Required: true},
            "email":  {Type: electrodb.AttributeTypeString, Required: true},
            "status": {
                Type:       electrodb.AttributeTypeEnum,
                EnumValues: []interface{}{"active", "inactive"},
                Default:    func() interface{} { return "active" },
            },
        },
        Indexes: map[string]*electrodb.IndexDefinition{
            "primary": {
                PK: electrodb.FacetDefinition{Field: "pk", Facets: []string{"userId"}},
            },
        },
        Timestamps: &electrodb.TimestampsConfig{
            CreatedAt: "createdAt",
            UpdatedAt: "updatedAt",
        },
    }

    // Create entity
    entity, _ := electrodb.NewEntity(schema, nil)

    // Put item
    entity.Put(electrodb.Item{
        "userId": "user-123",
        "email":  "john@example.com",
    }).Go()

    // Get item
    result, _ := entity.Get(electrodb.Keys{"userId": "user-123"}).Go()

    // Update item
    entity.Update(electrodb.Keys{"userId": "user-123"}).
        Set(map[string]interface{}{"email": "newemail@example.com"}).
        Go()

    // Query
    results, _ := entity.Query("primary").Query("user-123").Go()
}
```

## Feature Examples

### Validation & Transformations

```go
schema := &electrodb.Schema{
    Attributes: map[string]*electrodb.AttributeDefinition{
        "email": {
            Type: electrodb.AttributeTypeString,
            Validate: func(value interface{}) error {
                email, ok := value.(string)
                if !ok || !strings.Contains(email, "@") {
                    return errors.New("invalid email")
                }
                return nil
            },
            Set: func(value interface{}) interface{} {
                // Lowercase email before storing
                if email, ok := value.(string); ok {
                    return strings.ToLower(email)
                }
                return value
            },
        },
        "price": {
            Type: electrodb.AttributeTypeNumber,
            // Store as cents, display as dollars
            Set: func(value interface{}) interface{} {
                if dollars, ok := value.(float64); ok {
                    return int(dollars * 100)
                }
                return value
            },
            Get: func(value interface{}) interface{} {
                if cents, ok := value.(int); ok {
                    return float64(cents) / 100.0
                }
                return value
            },
        },
        "password": {
            Type:   electrodb.AttributeTypeString,
            Hidden: true, // Filtered from all responses
        },
        "createdAt": {
            Type:     electrodb.AttributeTypeNumber,
            ReadOnly: true, // Can't be updated
        },
    },
}
```

### Advanced Update Operations

```go
// Append to list
entity.Update(keys).
    Append(map[string]interface{}{
        "tags": []string{"new-tag"},
    }).Go()

// Prepend to list
entity.Update(keys).
    Prepend(map[string]interface{}{
        "items": []string{"first-item"},
    }).Go()

// Subtract from number
entity.Update(keys).
    Subtract(map[string]interface{}{
        "balance": 50,
    }).Go()

// Remove list elements by index
entity.Update(keys).
    Data(map[string]interface{}{
        "items": []int{0, 2}, // Remove items at index 0 and 2
    }).Go()

// Set operations
entity.Update(keys).
    AddToSet("tags", []string{"tag1", "tag2"}).
    DeleteFromSet("blockedIds", []string{"id3"}).
    Go()

// Combined operations
entity.Update(keys).
    Set(map[string]interface{}{"name": "John"}).
    Add(map[string]interface{}{"count": 1}).
    Subtract(map[string]interface{}{"balance": 10}).
    Append(map[string]interface{}{"tags": []string{"verified"}}).
    Go()
```

### Automatic Timestamps

```go
schema := &electrodb.Schema{
    // ... attributes ...
    Timestamps: &electrodb.TimestampsConfig{
        CreatedAt: "createdAt",
        UpdatedAt: "updatedAt",
    },
}

// createdAt set on create, updatedAt set on create and update
entity.Put(item).Go() // Sets both createdAt and updatedAt
entity.Update(keys).Set(updates).Go() // Updates updatedAt only
```

### Attribute Padding

```go
schema := &electrodb.Schema{
    Attributes: map[string]*electrodb.AttributeDefinition{
        "orderNumber": {
            Type: electrodb.AttributeTypeNumber,
            Padding: &electrodb.PaddingConfig{
                Length: 10,
                Char:   "0",
            },
        },
    },
}

// Write: 42 → "0000000042" (padded)
// Read:  "0000000042" → 42 (unpadded)
```

### TTL (Time-To-Live)

```go
schema := &electrodb.Schema{
    TTL: &electrodb.TTLConfig{
        Attribute: "expiresAt",
    },
}

// Set TTL with duration
entity.Put(item).WithTTL(24 * time.Hour).Go()

// Set TTL with explicit timestamp
entity.Put(item).WithTTLTimestamp(futureTimestamp).Go()

// Remove TTL
entity.Update(keys).RemoveTTL().Go()
```

### Named Filters

```go
schema := &electrodb.Schema{
    Filters: map[string]electrodb.FilterFunc{
        "activeUsers": func(attr electrodb.AttributeOperations, params map[string]interface{}) string {
            return attr["status"].Eq("active")
        },
        "premiumUsers": func(attr electrodb.AttributeOperations, params map[string]interface{}) string {
            minBalance := params["minBalance"]
            return attr["balance"].Gte(minBalance)
        },
    },
}

// Use named filter
entity.Query("primary").
    Query("user-123").
    Filter("activeUsers", nil).
    Go()

// Use named filter with params
entity.Query("primary").
    Query("user-123").
    Filter("premiumUsers", map[string]interface{}{
        "minBalance": 1000,
    }).
    Go()
```

### Pagination

```go
// Automatic pagination (fetches all pages)
allItems, err := entity.Query("primary").
    Query("user-123").
    Pages(electrodb.PagesOptions{
        MaxPages: 10,
        Limit:    50,
    })

// Manual pagination
iterator := entity.Query("primary").
    Query("user-123").
    Page(electrodb.PagesOptions{Limit: 20})

for {
    page, hasMore, err := iterator.Next()
    if err != nil || !hasMore {
        break
    }
    // Process page.Data
}
```

### Batch Operations

```go
// Batch Get
results, _ := entity.BatchGet([]electrodb.Keys{
    {"userId": "user-1"},
    {"userId": "user-2"},
    {"userId": "user-3"},
}).Go()

// Batch Write
entity.BatchWrite().
    Put([]electrodb.Item{
        {"userId": "user-4", "email": "user4@example.com"},
        {"userId": "user-5", "email": "user5@example.com"},
    }).
    Delete([]electrodb.Keys{
        {"userId": "user-6"},
    }).
    Go()
```

### Transactions

```go
service.Transaction().
    Write(entity1.Put(item1).Commit()).
    Write(entity2.Update(keys).Set(updates).Commit()).
    Write(entity3.Delete(keys).Commit()).
    Commit()
```

### Upsert Operations

```go
// Create (fails if exists)
entity.Create(item).Go() // Has condition: attribute_not_exists(pk)

// Upsert (creates or replaces)
entity.Upsert(item).Go() // No condition

// Upsert with Update (partial update, creates if not exists)
entity.UpsertUpdate(keys).Set(updates).Go()
```

## Data Processing Pipeline

### Write Path
```
Item → Defaults → Timestamps → Padding → Validation → Set Transform → Keys → DynamoDB
```

### Read Path
```
DynamoDB → Unmarshal → Remove Internal Keys → Remove Padding → Get Transform → Filter Hidden → Response
```

## Comparison with ElectroDB (JavaScript)

| Feature | JavaScript | Go | Status |
|---------|-----------|-----|--------|
| Entity & Service | ✅ | ✅ | Complete |
| Query & Scan | ✅ | ✅ | Complete |
| Batch Operations | ✅ | ✅ | Complete |
| Transactions | ✅ | ✅ | Complete |
| Collections | ✅ | ✅ | Complete |
| Pagination | ✅ | ✅ | Complete |
| Validation | ✅ | ✅ | Complete |
| Get/Set Transforms | ✅ | ✅ | Complete |
| Enum Validation | ✅ | ✅ | Complete |
| ReadOnly/Hidden | ✅ | ✅ | Complete |
| Named Filters | ✅ | ✅ | Complete |
| TTL Support | ✅ | ✅ | Complete |
| Timestamps | ✅ | ✅ | Complete |
| Padding | ✅ | ✅ | Complete |
| Append/Prepend | ✅ | ✅ | Complete |
| Subtract | ✅ | ✅ | Complete |
| Set Operations | ✅ | ✅ | Complete |
| Upsert | ✅ | ✅ | Complete |

## Architecture

### Key Components

- **Entity** - Represents a single entity type with schema and operations
- **Service** - Manages multiple entities in a single table
- **Schema** - Defines attributes, indexes, filters, timestamps, TTL
- **ParamsBuilder** - Builds DynamoDB operation parameters
- **Executor** - Executes operations against DynamoDB
- **Validator** - Handles validation and transformations

### Operation Flow

1. **Create Operation** → Validate → Transform → Build Params → Execute
2. **Read Operation** → Execute → Transform → Filter → Return
3. **Update Operation** → Validate → Transform → Build Expression → Execute

## Testing

The library includes comprehensive test coverage:

- **132 tests** covering all features
- Unit tests for each operation type
- Integration tests for combined features
- Edge case testing

Run tests:
```bash
go test ./electrodb
```

Run specific tests:
```bash
go test -v ./electrodb -run TestValidation
go test -v ./electrodb -run TestAppend
```

## Examples

See `examples/comprehensive/main.go` for a complete example demonstrating all features.

## API Reference

### Entity Operations

- `entity.Get(keys)` - Get item by key
- `entity.Put(item)` - Put item
- `entity.Create(item)` - Put with condition (fails if exists)
- `entity.Upsert(item)` - Put without condition
- `entity.Update(keys)` - Update item
- `entity.Patch(keys)` - Alias for Update
- `entity.UpsertUpdate(keys)` - Update that creates if not exists
- `entity.Delete(keys)` - Delete item
- `entity.Query(index)` - Start query on index
- `entity.Scan()` - Scan table
- `entity.BatchGet(keys)` - Batch get operation
- `entity.BatchWrite()` - Batch write operation

### Update Methods

- `.Set(updates)` - Set attribute values
- `.Add(updates)` - Add to numbers or sets
- `.Subtract(updates)` - Subtract from numbers
- `.Append(updates)` - Append to lists
- `.Prepend(updates)` - Prepend to lists
- `.AddToSet(attr, values)` - Add values to set
- `.DeleteFromSet(attr, values)` - Remove values from set
- `.Remove(attrs)` - Remove attributes
- `.Data(updates)` - Remove list elements by index
- `.Condition(callback)` - Add condition expression
- `.WithTTL(duration)` - Set TTL
- `.RemoveTTL()` - Remove TTL

### Query Methods

- `.Query(facets...)` - Set partition key values
- `.Eq(value)` - Sort key equals
- `.Gt(value)` - Sort key greater than
- `.Gte(value)` - Sort key greater than or equal
- `.Lt(value)` - Sort key less than
- `.Lte(value)` - Sort key less than or equal
- `.Between(start, end)` - Sort key between
- `.Begins(value)` - Sort key begins with
- `.Where(callback)` - Add filter expression
- `.Filter(name, params)` - Use named filter
- `.Pages(opts)` - Automatic pagination
- `.Page(opts)` - Manual pagination
- `.Go()` - Execute operation
- `.Params()` - Get DynamoDB parameters

## Contributing

Contributions are welcome! The library aims for feature parity with the JavaScript ElectroDB.

## License

Same as ElectroDB (MIT)

## Credits

Based on [ElectroDB](https://github.com/tywalch/electrodb) by Tyler Walch.

Go port implementation with comprehensive feature coverage.
