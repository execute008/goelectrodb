# ElectroDB - Comprehensive Code Analysis

## 1. What is ElectroDB? What Problem Does It Solve?

**ElectroDB** is a DynamoDB library that simplifies working with multiple entities and complex hierarchical relationships in a single DynamoDB table using the "single-table design" pattern.

### Problem It Solves:
- **Query Complexity**: DynamoDB raw queries require manually constructing KeyConditionExpressions, FilterExpressions, ExpressionAttributeNames, and ExpressionAttributeValues
- **Entity Isolation**: Managing multiple entity types in a single table without conflicts
- **Index Design**: Simplifying the creation and querying of complex hierarchical access patterns through Global Secondary Indexes (GSIs)
- **Update Expressions**: Building complex update operations with SET, ADD, APPEND, DELETE operations in a type-safe way
- **Pagination**: Generating URL-safe cursors for pagination
- **Cross-Entity Queries**: Querying data across multiple entities in a single request via "Collections"

### Key Value Propositions:
- Converts intuitive, chainable method calls into proper DynamoDB SDK parameters
- Provides schema validation and attribute type enforcement
- Automatic index selection based on provided query parameters
- Strong TypeScript support with type inference

---

## 2. Main Features and Capabilities

### Core Features:
1. **Entity Definition & Isolation**
   - Define entities with schema, attributes, and indexes
   - Automatic entity identification in responses
   - Single-table isolation using configurable identifiers

2. **Attribute Schema Enforcement**
   - Type validation (string, number, boolean, enum, map, set, list, any, custom)
   - Default values (static or computed)
   - Custom validation functions
   - Field name mapping (attribute name → DynamoDB field name)
   - Read-only, hidden, and required attributes

3. **Index Design**
   - Primary index on main table (pk, sk)
   - Global Secondary Indexes (GSIs) with custom pk/sk
   - Composite facets for building hierarchical keys
   - Isolated and clustered collection indexes
   - Support for sparse indexes

4. **Query Building**
   - Chainable API for constructing queries
   - Access pattern-based querying (query.indexName(facets))
   - Sort key conditions: eq, lt, gt, lte, gte, begins, between
   - Query operations via .query(), .scan(), .get(), .match(), .find()

5. **Filtering & Conditions**
   - FilterExpression builders
   - ConditionExpression builders for mutations
   - Readable filter syntax with type-safe operations
   - Support for complex boolean logic

6. **Update Operations**
   - SET, ADD, APPEND, DELETE, SUBTRACT operations
   - if_not_exists conditional updates
   - Composite attribute updates
   - UPSERT operations

7. **Batch Operations**
   - Batch PUT (batchPut)
   - Batch GET (batchGet)
   - Batch WRITE (batchWrite)
   - Configurable concurrency

8. **Transactions**
   - TransactWrite for multi-item writes
   - TransactGet for multi-item reads
   - Condition checking within transactions

9. **Pagination**
   - Automatic cursor generation (url-safe)
   - Multiple pagination modes: raw, named, item, cursor
   - Async iteration support
   - Fine-grain pagination control (limit, pages)

10. **Collections**
    - Query across multiple entities in single request
    - Isolated collections (one entity per collection lookup)
    - Clustered collections (multiple entities per collection lookup)
    - Hydration support for loading full item details

11. **Response Formatting**
    - Include/exclude keys from responses
    - Raw vs. formatted responses
    - Entity type inference
    - Custom data options

---

## 3. Core API Surface and Public Methods

### Entity Methods:

**CRUD Operations:**
- `entity.get(keys)` - Get single item by primary key
- `entity.put(attributes)` / `entity.put([items])` - Insert/replace items
- `entity.create(attributes)` - Create new item (condition: item doesn't exist)
- `entity.upsert(attributes)` - Insert or update
- `entity.update(keys)` - Update item
- `entity.patch(keys)` - Patch item (update with data)
- `entity.delete(keys)` / `entity.delete([keys])` - Delete items
- `entity.remove(keys)` - Remove item (alias for delete)

**Query Operations:**
- `entity.query.accessPatternName(facets)` - Query via specific access pattern
- `entity.scan` - Scan entire table
- `entity.match(filter)` - Find matching items
- `entity.find(filter)` - Find first matching item

**Batch Operations:**
- `entity.put([items])` - Batch put
- `entity.get([keys])` - Batch get
- `entity.delete([keys])` - Batch delete

**Query Chainable Methods:**
- `.query(facets)` / `.gt(sk)` / `.lt(sk)` / `.gte(sk)` / `.lte(sk)` / `.eq(sk)` / `.begins(sk)` / `.between(sk1, sk2)` - Sort key conditions
- `.where((attrs, ops) => ops.eq(attr, val))` - Custom conditions
- `.filter((attrs, ops) => ops.lte(attr, val))` - Filter expressions
- `.set({data})` / `.add({data})` / `.append({data})` / `.subtract({data})` - Update operations
- `.options({})` - Query options (limit, pages, concurrency, etc.)
- `.go()` - Execute query
- `.params()` - Get DynamoDB parameters without executing

**Pagination:**
- `.page(cursor)` - Get next page
- `.params({pager: 'all'})` - Get all pages parameters

**Transactions:**
- `entity.transactWrite()` - Transaction write
- `entity.transactGet()` - Transaction get

### Service Methods:

**Collection Queries:**
- `service.collections.collectionName(facets)` - Query across entities
- `service.entities.entityName` - Access individual entity

**Transaction Management:**
- `service.transaction.write(fn)` - Transaction write callback
- `service.transaction.get(fn)` - Transaction get callback

---

## 4. Directory Structure

```
/home/user/go-electrodb/
├── src/                           # Main source code (~13K lines)
│   ├── entity.js                 # Entity class (1545 lines) - main CRUD/query logic
│   ├── service.js                # Service class (900+ lines) - multi-entity management
│   ├── schema.js                 # Schema parsing and validation (1340+ lines)
│   ├── clauses.js                # Query chain building (1685 lines)
│   ├── operations.js             # Expression builders (400+ lines)
│   ├── update.js                 # Update expression builder
│   ├── where.js                  # Filter expression builder
│   ├── filters.js                # Filter factory
│   ├── filterOperations.js       # Filter operation templates
│   ├── updateOperations.js       # Update operation templates
│   ├── types.js                  # Constants and types
│   ├── client.js                 # DynamoDB client wrapper
│   ├── conversions.js            # Key/cursor conversions
│   ├── validations.js            # Validation utilities (400+ lines)
│   ├── util.js                   # Utility functions
│   ├── errors.js                 # Error definitions
│   ├── events.js                 # Event management
│   └── transaction.js            # Transaction handling
│
├── test/                         # Test suite
│   ├── offline.entity.spec.js    # Entity tests without DynamoDB
│   ├── offline.service.spec.js   # Service tests
│   ├── offline.pagination.spec.js # Pagination tests
│   ├── offline.*.spec.js         # More offline tests
│   ├── connected.*.spec.js       # Tests with live DynamoDB
│   ├── ts_connected.*.spec.ts    # TypeScript tests
│   ├── definitions/              # Test entity definitions
│   ├── generation/               # Test data generation
│   ├── conversions.json          # Test conversion data
│   └── *.test-d.ts              # Type testing files
│
├── examples/                     # Example implementations
│   ├── taskManager/              # Task management example
│   │   ├── models/
│   │   │   ├── task.ts
│   │   │   ├── employee.ts
│   │   │   ├── office.ts
│   │   │   └── index.ts
│   │   ├── load.ts
│   │   └── query.ts
│   ├── library/                  # Library management example
│   ├── versionControl/           # Version control example
│   ├── locks/                    # Locking patterns example
│   ├── provisionTable/           # Table provisioning example
│   └── common/                   # Shared utilities
│
├── index.js                      # Main entry point
├── index.d.ts                    # TypeScript definitions
├── package.json                  # Dependencies and scripts
└── README.md                     # Documentation
```

**Key Locations:**
- **Source Code**: `/home/user/go-electrodb/src/`
- **Tests**: `/home/user/go-electrodb/test/`
- **Examples**: `/home/user/go-electrodb/examples/`

---

## 5. Key Abstractions and Components

### A. Entity Class
**Purpose**: Represents a single entity type in DynamoDB

**Key Components:**
- `model` - Parsed schema definition
- `schema` - Attribute definitions
- `identifiers` - Entity type identifiers for single-table isolation
- `query` - Access pattern methods (dynamically generated from indexes)
- `config` - Configuration options (client, table name, listeners)
- Methods: get, put, patch, update, delete, scan, query via access patterns

**Key Methods:**
- `_makeChain()` - Creates a chainable query builder
- `_parseModel()` - Parses entity schema
- `formatResponse()` - Formats DynamoDB response to entity format
- `executeQuery()` - Executes query with pagination support
- `executeBulkGet()` - Batch get with concurrency control

### B. Service Class
**Purpose**: Manages multiple entities and provides cross-entity queries via Collections

**Key Components:**
- `entities` - Map of Entity instances
- `collections` - Collection configurations
- `compositeAttributes` - Shared composite attributes across entities
- `transaction` - Transaction management

**Key Methods:**
- `_betaConstructor()` / `_v1Constructor()` - Version-specific initialization
- Access pattern methods for querying across entities

### C. Schema and Attributes
**Schema Classes:**
- `Schema` - Main schema builder
- `Attribute` - Individual attribute definition
- `AttributeTraverser` - Traverses nested attribute paths

**Attribute Properties:**
- `name` - Attribute name
- `field` - DynamoDB field name
- `type` - Attribute type (string, number, boolean, enum, map, set, list, any, custom, static)
- `required` - Is required
- `default` - Default value (static or function)
- `validate` - Custom validation function
- `format`/`unformat` - Key serialization/deserialization
- `indexes` - Indexes this attribute participates in

### D. Query Chain Building (ChainState)
**Purpose**: Builds DynamoDB queries through a chainable API

**Pattern**: Builder pattern with fluent interface
- Each method returns a new set of available methods
- State accumulates query parameters as methods are called
- Terminal operations (.go() or .params()) execute/return the query

**Key State Properties:**
```
query: {
  index: string,              // Index name
  type: string,              // Query type (eq, begins, between, etc.)
  method: string,            // DynamoDB method (query, scan, get, put, update, delete)
  facets: {},                // Composite key attributes
  keys: {                    // Key values
    pk: {},
    sk: []
  },
  update: UpdateExpression,  // Update operations
  filter: {                  // Filter expressions
    ConditionExpression: FilterExpression,
    FilterExpression: FilterExpression
  },
  options: {}                // Query options
}
```

**Methods Flow:**
1. Access pattern method (query.accessPattern(pk))
2. Sort key condition (.gt(), .begins(), etc.) - optional
3. Conditions (.where(), .filter())
4. Update operations (.set(), .add()) - for mutations
5. Query options (.options())
6. Terminal operation (.go() or .params())

### E. Expression Builders

**ExpressionState** (Base class for building expressions)
- `setName(paths, name, value)` - Create named placeholder for attributes
- `setValue(name, value)` - Create value placeholder
- Methods track names and values for ExpressionAttributeNames/Values

**UpdateExpression**
- Manages SET, ADD, REMOVE, DELETE, SUBTRACT operations
- Tracks composite attributes that are updated
- Builds final UpdateExpression string

**FilterExpression**
- Manages filter/condition expressions
- Combines expressions with AND/OR logic
- Supports both named filters and inline conditions

### F. Operations & Filter Operations

**UpdateOperations:**
- `set` - SET operation
- `add` - ADD operation (numbers, sets)
- `append` - LIST_APPEND operation
- `delete` - DELETE operation (sets)
- `subtract` - SET operation for numeric subtraction
- `ifNotExists` - if_not_exists() function
- etc.

**FilterOperations:**
- `eq` - Equality
- `ne` - Not equal
- `gt`, `gte`, `lt`, `lte` - Comparisons
- `begins` - begins_with()
- `between` - BETWEEN operator
- `contains` - contains()
- `exists`, `notExists` - attribute_exists/not_exists
- `size` - size()
- `type` - attribute_type()
- etc.

### G. Client Wrapper
**Purpose**: Abstracts DynamoDB SDK versions (v2 and v3)

**Supports:**
- AWS SDK v2 DocumentClient
- AWS SDK v3 DynamoDB with Document Client wrapper
- Transact operations error handling
- Automatic unmarshalling

---

## 6. Test Structure and Organization

### Test Files by Type:

**Offline Tests** (no DynamoDB connection required):
- `offline.entity.spec.js` - Entity query building and validation
- `offline.service.spec.js` - Service and collection functionality
- `offline.pagination.spec.js` - Pagination logic
- `offline.batch.spec.js` - Batch operations
- `offline.where.spec.js` - Where clause building
- `offline.filters.spec.js` - Filter expressions
- `offline.attribute.spec.js` - Attribute handling
- `offline.options.spec.js` - Query options
- `offline.validations.spec.js` - Validation logic
- `offline.util.spec.js` - Utility functions

**Connected Tests** (with live DynamoDB):
- `connected.crud.spec.js` - Live CRUD operations (177K, largest test)
- `connected.update.spec.js` - Live update operations (69K)
- `connected.service.spec.js` - Live service operations (43K)
- `connected.page.spec.js` - Live pagination (57K)
- `connected.batch.spec.js` - Live batch operations
- `connected.where.spec.js` - Live where clause testing
- `connected.filters.spec.js` - Live filter testing
- `connected.issues.spec.js` - GitHub issue regression tests

**TypeScript Tests** (ts_connected.*.spec.ts):
- Type validation tests
- Ensure TypeScript inference works correctly
- Test cases for complex scenarios

**Type Definition Tests** (.test-d.ts):
- `entity.test-d.ts` - Entity type inference
- `indexes.test-d.ts` - Index type inference
- `mocks.test.ts` - Mock entity definitions

### Test Structure:
- **Mocha** test framework
- **Chai** for assertions
- **ts-node** for TypeScript support
- **nyc** for code coverage
- Test initialization script creates actual DynamoDB tables

---

## 7. Dependencies and External Libraries

### Production Dependencies:
```json
{
  "@aws-sdk/lib-dynamodb": "^3.654.0",      // DynamoDB Document Client
  "@aws-sdk/util-dynamodb": "^3.654.0",     // DynamoDB utilities (unmarshalling)
  "jsonschema": "1.2.7"                     // JSON schema validation
}
```

### Key Dev Dependencies:
- **Testing**: mocha, chai, jest, nyc (code coverage)
- **TypeScript**: typescript, ts-node, tsd (for TypeScript definitions)
- **AWS**: @aws-sdk/client-dynamodb, aws-cdk-lib, aws-sdk (v2)
- **Code Quality**: prettier (code formatting)
- **Examples**: moment (date handling), uuid (ID generation), faker (test data)
- **Documentation**: browserify (browser bundle)

### AWS SDK Integration:
- **v2 Support**: Uses AWS SDK v2 DocumentClient methods
- **v3 Support**: Wraps AWS SDK v3 DynamoDB client with custom logic
- Both versions abstracted through DocumentClientV2Wrapper and DocumentClientV3Wrapper

---

## 8. Key Design Patterns

### A. Builder Pattern
Query building uses a fluent, chainable builder pattern:
```
entity.query.indexName(facets)
  .gt(sortKeyValue)
  .where((attrs, ops) => ops.eq(attr, val))
  .options({ limit: 10 })
  .go()
```

### B. State Machine Pattern
ChainState maintains query state and only allows valid method sequences based on the current state.

### C. Expression Building Pattern
ExpressionState and its subclasses (UpdateExpression, FilterExpression) build expressions incrementally, tracking:
- Attribute name placeholders (#attr)
- Value placeholders (:value)
- Expression strings

### D. Factory Pattern
- FilterFactory creates filter builders
- WhereFactory creates where condition builders
- These inject clauses into the query chain

### E. Proxy Pattern
AttributeOperationProxy proxies attribute access for type-safe update/filter operations.

### F. Template Method Pattern
Operations (UpdateOperations, FilterOperations) define templates for building expressions.

---

## 9. Key Data Flows

### Query Execution Flow:
1. **Entity Definition** → Schema parsing → Index extraction → Attribute definitions
2. **Access Pattern Call** → Creates ChainState with initial facets
3. **Query Methods** → Accumulate state (sort key conditions, filters, options)
4. **Go/Params** → Build DynamoDB parameters from state
5. **Execute** → Call DynamoDB client
6. **Format Response** → Parse DynamoDB response into entity format

### Update Expression Building:
1. Call `.set({field: value})`, `.add({field: value})`, etc.
2. Operations added to UpdateExpression.operations[type] set
3. Generate placeholders in ExpressionAttributeNames/Values
4. Build final UpdateExpression string combining all operations
5. Composite attributes automatically updated with sort key components

### Pagination Flow:
1. Query returns response with LastEvaluatedKey
2. Cursor generated from LastEvaluatedKey (URL-safe encoding)
3. Next page call uses cursor → convert to ExclusiveStartKey
4. Process continues until no LastEvaluatedKey or page limit reached

---

## 10. Notable Architectural Features

### Single-Table Design Support
- Entity identifiers automatically prepended to pk/sk
- Facet prefixes separate different entities and access patterns
- Collection names used to group related entities
- Format: `$serviceName#entityName#facet1_value1#facet2_value2`

### Index Management
- **Primary Index**: Automatic, based on main entity facets
- **GSI**: Defined per access pattern
- **Collections**: Special index grouping for multi-entity queries
- **Sparse Indexes**: Support for partial composite keys

### Attribute Watching/Watcher Pattern
Attributes can depend on other attributes for updates (computed fields).

### Custom Attributes
Support for custom attribute types with serialization/deserialization.

### Event System
Entity/Service emit events for:
- Query building
- Results processing
- Logging

---

## 11. Version Considerations

### Model Versions:
- **Beta** - Original model format
- **v1** - Current production format
- **v2** - Future format placeholder

### Entity/Service Versions:
- Currently v1

### SDK Versions:
- v2 - AWS SDK v2 DocumentClient (legacy)
- v3 - AWS SDK v3 (current)

---

## Summary for Go Port

### What Needs to Be Ported:
1. **Core Classes**: Entity, Service, Schema, Attribute
2. **Query Builder**: ChainState and clause system
3. **Expression Builders**: UpdateExpression, FilterExpression, ExpressionState
4. **Operations**: UpdateOperations, FilterOperations templates
5. **Client Abstraction**: Wrapper around Go DynamoDB client (aws-sdk-go-v2)
6. **Schema Validation**: Attribute validation, type checking
7. **Response Formatting**: DynamoDB response to entity struct conversion

### Go-Specific Considerations:
- Use interfaces instead of JavaScript's dynamic property definition
- Use struct field tags for schema definition
- Use generics for type safety
- Implement builder pattern with method chaining returning *State
- Use channels or callbacks for pagination
- DynamoDB client: github.com/aws/aws-sdk-go-v2/service/dynamodb
- Expression building with aws-sdk-go-v2/feature/dynamodb/expression

### Complexity Level:
- **Medium-High**: The query building logic is sophisticated with state management
- **Type Safety**: Go's type system is more restrictive than JavaScript but provides better safety
- **Estimated Code Size**: Likely 3000-5000 lines of Go code for core functionality

