package electrodb

import (
	"context"
)

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
	skFacets      []interface{} // SK facet values for begins_with prefix (like JS ElectroDB)
	skCondition   *sortKeyCondition
	filters       []string
	options       *QueryOptions
	filterBuilder *FilterBuilder
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
	// Split facets between PK and SK based on index definition
	// This allows ElectroDB-style queries like:
	//   .Query("byApp").Query(appId, "published")
	// where "published" is the first SK facet (status)
	pkFacetCount := len(qb.index.PK.Facets)

	var pkFacets, skFacets []interface{}
	for i, facet := range facets {
		if i < pkFacetCount {
			pkFacets = append(pkFacets, facet)
		} else {
			skFacets = append(skFacets, facet)
		}
	}

	return &QueryChain{
		entity:        qb.entity,
		accessPattern: qb.accessPattern,
		index:         qb.index,
		pkFacets:      pkFacets,
		skFacets:      skFacets,
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
func (qc *QueryChain) Where(callback WhereCallback) *QueryChain {
	fb := NewFilterBuilder(qc.entity.schema.Attributes)
	fb.Where(callback)

	// Merge with existing filter builder if present
	if qc.filterBuilder != nil {
		// Combine expressions with AND
		existingExpr, existingNames, existingValues := qc.filterBuilder.Build()
		newExpr, newNames, newValues := fb.Build()

		combined := NewFilterBuilder(qc.entity.schema.Attributes)
		combined.builder.expression = existingExpr + " AND " + newExpr

		// Merge names and values
		mergedNames, mergedValues := MergeExpressionAttributes(
			existingNames,
			existingValues,
			newNames,
			newValues,
		)
		combined.builder.names = mergedNames
		combined.builder.values = mergedValues

		qc.filterBuilder = combined
	} else {
		qc.filterBuilder = fb
	}

	return qc
}

// Filter adds a filter using a named filter from schema
func (qc *QueryChain) Filter(filterName string, params map[string]interface{}) *QueryChain {
	// Look up the named filter in the schema
	if qc.entity.schema.Filters == nil {
		return qc
	}

	filterFunc, exists := qc.entity.schema.Filters[filterName]
	if !exists {
		return qc
	}

	// Create a filter builder and execute the named filter
	fb := NewFilterBuilder(qc.entity.schema.Attributes)
	fb.Where(func(attrs map[string]*AttributeRef, ops *OperationBuilder) string {
		// Convert AttributeRef map to AttributeOperations for the filter function
		attrOps := make(AttributeOperations)
		for name, ref := range attrs {
			attrOps[name] = &AttributeOperator{
				name:    name,
				builder: ref.builder,
			}
		}
		return filterFunc(attrOps, params)
	})

	// Merge with existing filter builder if present
	if qc.filterBuilder != nil {
		// Combine expressions with AND
		existingExpr, existingNames, existingValues := qc.filterBuilder.Build()
		newExpr, newNames, newValues := fb.Build()

		combined := NewFilterBuilder(qc.entity.schema.Attributes)
		combined.builder.expression = existingExpr + " AND " + newExpr

		// Merge names and values
		mergedNames, mergedValues := MergeExpressionAttributes(
			existingNames,
			existingValues,
			newNames,
			newValues,
		)
		combined.builder.names = mergedNames
		combined.builder.values = mergedValues

		qc.filterBuilder = combined
	} else {
		qc.filterBuilder = fb
	}

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
	return executor.ExecuteQuery(context.Background(), qc.accessPattern, qc.pkFacets, qc.skFacets, qc.skCondition, qc.options, qc.filterBuilder)
}

// Params returns the DynamoDB parameters without executing
func (qc *QueryChain) Params() (map[string]interface{}, error) {
	builder := NewParamsBuilder(qc.entity)
	return builder.BuildQueryParams(qc.accessPattern, qc.pkFacets, qc.skFacets, qc.skCondition, qc.options, qc.filterBuilder)
}
