package electrodb

import "context"

// QueryBuilder is an interface for building queries
type QueryBuilder interface {
	// Query starts a query with partition key facets
	Query(facets ...interface{}) *QueryChain
}

// QueryChain represents a chainable query
type QueryChain struct {
	entity        *Entity
	accessPattern string
	index         *IndexDefinition
	pkFacets      []interface{}
	skCondition   *sortKeyCondition
	filters       []string
	options       *QueryOptions
}

type sortKeyCondition struct {
	operation string
	values    []interface{}
}

// queryBuilderImpl implements QueryBuilder
type queryBuilderImpl struct {
	entity        *Entity
	accessPattern string
	index         *IndexDefinition
}

func newQueryBuilder(entity *Entity, accessPattern string, index *IndexDefinition) QueryBuilder {
	return &queryBuilderImpl{
		entity:        entity,
		accessPattern: accessPattern,
		index:         index,
	}
}

func (qb *queryBuilderImpl) Query(facets ...interface{}) *QueryChain {
	return &QueryChain{
		entity:        qb.entity,
		accessPattern: qb.accessPattern,
		index:         qb.index,
		pkFacets:      facets,
	}
}

// Eq adds an equals condition on the sort key
func (qc *QueryChain) Eq(value interface{}) *QueryChain {
	qc.skCondition = &sortKeyCondition{
		operation: "=",
		values:    []interface{}{value},
	}
	return qc
}

// Gt adds a greater-than condition on the sort key
func (qc *QueryChain) Gt(value interface{}) *QueryChain {
	qc.skCondition = &sortKeyCondition{
		operation: ">",
		values:    []interface{}{value},
	}
	return qc
}

// Gte adds a greater-than-or-equal condition on the sort key
func (qc *QueryChain) Gte(value interface{}) *QueryChain {
	qc.skCondition = &sortKeyCondition{
		operation: ">=",
		values:    []interface{}{value},
	}
	return qc
}

// Lt adds a less-than condition on the sort key
func (qc *QueryChain) Lt(value interface{}) *QueryChain {
	qc.skCondition = &sortKeyCondition{
		operation: "<",
		values:    []interface{}{value},
	}
	return qc
}

// Lte adds a less-than-or-equal condition on the sort key
func (qc *QueryChain) Lte(value interface{}) *QueryChain {
	qc.skCondition = &sortKeyCondition{
		operation: "<=",
		values:    []interface{}{value},
	}
	return qc
}

// Between adds a between condition on the sort key
func (qc *QueryChain) Between(start, end interface{}) *QueryChain {
	qc.skCondition = &sortKeyCondition{
		operation: "BETWEEN",
		values:    []interface{}{start, end},
	}
	return qc
}

// Begins adds a begins-with condition on the sort key
func (qc *QueryChain) Begins(value interface{}) *QueryChain {
	qc.skCondition = &sortKeyCondition{
		operation: "begins_with",
		values:    []interface{}{value},
	}
	return qc
}

// Where adds a custom filter expression
func (qc *QueryChain) Where(filterFn func(AttributeOperations) string) *QueryChain {
	// TODO: Implement filter function execution
	return qc
}

// Filter adds a filter using a named filter from schema
func (qc *QueryChain) Filter(filterName string, params map[string]interface{}) *QueryChain {
	// TODO: Implement named filter execution
	return qc
}

// Options sets query options
func (qc *QueryChain) Options(opts *QueryOptions) *QueryChain {
	qc.options = opts
	return qc
}

// Go executes the query
func (qc *QueryChain) Go() (*QueryResponse, error) {
	executor := NewExecutionHelper(qc.entity)
	return executor.ExecuteQuery(context.Background(), qc.accessPattern, qc.pkFacets, qc.skCondition, qc.options)
}

// Params returns the DynamoDB parameters without executing
func (qc *QueryChain) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(qc.entity)
	return builder.BuildQueryParams(qc.accessPattern, qc.pkFacets, qc.skCondition, qc.options)
}
