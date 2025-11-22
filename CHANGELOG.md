# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-22

### Added

- **Core Entity Operations**
  - `Get` - Retrieve items by primary key
  - `Put` - Create or replace items
  - `Create` - Create items with existence check
  - `Update` - Update items with SET, ADD, REMOVE, DELETE operations
  - `Delete` - Delete items by primary key
  - `Scan` - Full table scans with pagination

- **Query Operations**
  - Fluent query builder with index selection
  - Sort key conditions: `eq`, `gt`, `gte`, `lt`, `lte`, `between`, `begins`
  - Filter expressions with `Where` clause
  - Named filters defined in schema
  - Ascending/descending order support

- **Batch Operations**
  - `BatchGet` - Retrieve up to 100 items in a single request
  - `BatchWrite` - Write up to 25 items in a single request
  - Automatic handling of unprocessed items

- **Transaction Support**
  - `TransactWrite` - Atomic write transactions (Put, Update, Delete)
  - `TransactGet` - Atomic read transactions
  - Condition expressions for transactional operations
  - Transaction cancellation reason parsing

- **Pagination**
  - Automatic pagination with `Pages()` iterator
  - Manual cursor-based pagination
  - Configurable page limits

- **Services and Collections**
  - Group entities into services
  - Collection queries across multiple entities
  - Service-level batch operations

- **Schema Features**
  - Automatic key composition (PK/SK)
  - Multiple indexes (GSI support)
  - Attribute validation and transformation
  - `Get`/`Set` transformers for attribute values
  - Enum validation
  - ReadOnly and Hidden attributes
  - Attribute padding for sortable keys

- **Automatic Timestamps**
  - `createdAt` - Set on item creation
  - `updatedAt` - Set on item creation and updates

- **TTL Support**
  - Configurable TTL attribute for automatic item expiration

- **Advanced Update Operations**
  - `Append` - Append to list attributes
  - `Prepend` - Prepend to list attributes
  - `Add` - Add to numeric or set attributes
  - `Subtract` - Subtract from numeric attributes
  - `AddToSet` - Add elements to sets
  - `DeleteFromSet` - Remove elements from sets
  - `Remove` - Remove attributes
  - `Data` - Remove specific list indices

- **Condition Expressions**
  - Conditional mutations with `Condition()` builder
  - Support for attribute existence, comparisons, and functions

- **Error Handling**
  - Structured `ElectroError` type with code, message, and cause
  - Exported error code constants for programmatic handling

- **Developer Experience**
  - `Params()` method on all operations for debugging
  - Full godoc documentation
  - Comprehensive examples

### Notes

- This is the initial stable release of GoElectroDB
- Full feature parity with ElectroDB JavaScript library
- Requires Go 1.21 or later
- Compatible with AWS SDK for Go v2
