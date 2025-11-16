# ElectroDB - Architecture Quick Reference

## Key Files and Their Purposes

| File | Size | Purpose |
|------|------|---------|
| `entity.js` | ~1545 lines | Main Entity class - handles CRUD operations, query building, response formatting |
| `service.js` | ~900 lines | Service class - manages multiple entities, collection queries |
| `schema.js` | ~1340 lines | Schema parsing, attribute definitions, validation |
| `clauses.js` | ~1685 lines | Query chain building - the fluent API engine |
| `operations.js` | ~400 lines | Expression builders (UpdateExpression, FilterExpression, ExpressionState) |
| `types.js` | ~200 lines | Constants, enums, type definitions |
| `filters.js` | ~100 lines | Filter factory and clause injection |
| `where.js` | ~180 lines | Where/Filter expression class |
| `filterOperations.js` | ~125 lines | Filter operation templates (eq, gt, contains, etc.) |
| `updateOperations.js` | ~200 lines | Update operation templates (set, add, append, delete) |
| `client.js` | ~300 lines | DynamoDB client wrapper (v2 and v3 support) |
| `conversions.js` | ~70 lines | Cursor/key conversion utilities |
| `validations.js` | ~400 lines | Validation utilities |
| `util.js` | ~290 lines | Helper functions |
| `errors.js` | ~250 lines | Error definitions |
| `events.js` | ~70 lines | Event management |
| `transaction.js` | ~150 lines | Transaction handling |

## Core Concepts

### 1. Entity Definition
```javascript
const Task = new Entity({
  model: { entity: "task", version: "1", service: "app" },
  attributes: {
    taskId: { type: "string", required: true },
    projectId: { type: "string", required: true },
    status: { type: ["open", "closed"], default: "open" }
  },
  indexes: {
    primary: {
      pk: { field: "pk", composite: ["taskId"] },
      sk: { field: "sk", composite: ["projectId"] }
    },
    byProject: {
      index: "gsi1pk-gsi1sk-index",
      pk: { field: "gsi1pk", composite: ["projectId"] },
      sk: { field: "gsi1sk", composite: ["status"] }
    }
  }
});
```

### 2. Query Building (Fluent Chain)
```javascript
// Step 1: Access pattern
Task.query.byProject({ projectId: "proj1" })

// Step 2: Sort key condition (optional)
  .eq({ status: "open" })

// Step 3: Filter (optional)
  .where(({ taskId }, { begins }) => begins(taskId, "TASK"))

// Step 4: Options (optional)
  .options({ limit: 10 })

// Step 5: Execute
  .go()
```

### 3. State Management
ChainState object tracks:
- Current index being queried
- Composite key facets (pk/sk values)
- Sort key conditions
- Filter expressions
- Update operations
- Query options

### 4. Expression Building
- ExpressionAttributeNames map: `{ "#attr": "attributeName" }`
- ExpressionAttributeValues map: `{ ":val0": "value" }`
- Final expressions: `"#attr = :val0 AND #other > :val1"`

## Main Classes

```
Entity
├── model (parsed schema)
├── schema (attribute definitions)
├── query (access pattern methods - dynamically generated)
└── Methods: get, put, patch, update, delete, scan, etc.

Service
├── entities (map of Entity instances)
├── collections (collection configurations)
└── transaction (transaction manager)

ChainState
├── query (current query state)
├── params (built DynamoDB parameters)
└── Methods: init, setPK, setSK, setMethod, etc.

UpdateExpression extends ExpressionState
├── operations (SET, ADD, REMOVE, DELETE, SUBTRACT sets)
├── composites (composite attributes being updated)
└── Methods: set, add, remove, build

FilterExpression extends ExpressionState
├── expression (accumulated filter/condition string)
└── Methods: add, unsafeSet, build

Attribute
├── name, field, type, required
├── default, validate, format/unformat
└── Handles validation, formatting, field mapping
```

## Method Chainability

Clauses in `clauses.js` define method chains with allowed children:

```javascript
clauses.index.children = [
  "check", "get", "delete", "update", "query", "upsert",
  "put", "scan", "collection", "create", "remove", "patch", "batchPut", ...
]

clauses.query.children = ["params", "go", "commit", "where", "filter", "gt", "gte", "lt", "lte", "eq", "begins", "between"]

clauses.update.children = [
  "data", "set", "append", "add", "updateRemove", "updateDelete",
  "go", "params", "subtract", "commit", "composite", "ifNotExists", "where"
]
```

## DynamoDB Operations Mapping

| Method | DynamoDB Op | Example |
|--------|------------|---------|
| `.get()` | GetItem | `Task.get({ taskId })`  |
| `.put()` | PutItem | `Task.put({ taskId, projectId })` |
| `.update()` / `.patch()` | UpdateItem | `Task.patch({ taskId }).set({ status })` |
| `.delete()` / `.remove()` | DeleteItem | `Task.delete({ taskId })` |
| `.query.*()` | Query | `Task.query.byProject({ projectId }).eq({ status })` |
| `.scan` | Scan | `Task.scan.go()` |
| `.get([keys])` | BatchGetItem | `Task.get([{taskId: 1}, {taskId: 2}])` |
| `.put([items])` | BatchWriteItem | `Task.put([{...}, {...}])` |

## Response Formatting

Responses formatted as:
```javascript
{
  data: [...],           // Formatted entity items
  cursor: "...",         // Pagination cursor (if applicable)
  lastEvaluatedKey: {}, // DynamoDB LEK (if applicable)
  count: 10,
  scannedCount: 10
}
```

## Error Handling

Errors structured as:
```javascript
ElectroError
├── code (1000-5000 range)
├── section (for help docs)
├── name (error type)
└── message (detailed description)
```

## Key Patterns for Go Port

### Pattern 1: Chainable Builder
```go
// JavaScript example - return new state with methods
current[methodName] = (...args) => {
  state.prev = state.self
  state.self = clauseName
  let result = clause.action(entity, state, ...args)
  if (clause.children.length) {
    return state.init(entity, allClauses, clause)
  }
}

// Go equivalent would use:
// - Receiver methods on State
// - Return *State for chaining
// - Interface for available methods
```

### Pattern 2: Dynamic Method Registration
```go
// JavaScript: Creates query.indexName methods dynamically
for (let accessPattern in model.indexes) {
  entity.query[accessPattern] = (...values) => {
    return entity._makeChain(index, clauses).query(...values)
  }
}

// Go equivalent:
// - Use type assertions or reflection
// - Or generate methods at compile-time with code generation
// - Or use map[string]func(...) for dynamic dispatch
```

### Pattern 3: Attribute-Based Expressions
```go
// JavaScript: Callback receives proxy object
.where(({ status, taskId }, { eq, begins }) => {
  return eq(status, "open") + " AND " + begins(taskId, "TASK")
})

// Go equivalent:
// - Use struct with field getters
// - Return expression string
// - Or use builder object for type safety
```

## Testing Strategy

- **Offline tests**: Test query building and parameter generation without DynamoDB
- **Connected tests**: Test actual DynamoDB operations
- **Type tests**: Validate TypeScript inference
- **Regression tests**: Github issue test cases

## Performance Considerations

1. **Batch operations**: Handle concurrency with configurable limits
2. **Pagination**: Use cursors to avoid re-scanning
3. **Sparse indexes**: Support partial composite keys for query efficiency
4. **Index selection**: Automatic selection based on provided facets
5. **Lazy evaluation**: Parameters built only when `.go()` or `.params()` called

