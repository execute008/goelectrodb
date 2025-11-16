package electrodb

import "context"

// Page represents a single page of query results
type Page struct {
	Data   []map[string]interface{}
	Cursor *string
}

// Pages returns all pages of results by automatically following cursors
// This is a convenience method that handles pagination automatically
func (qc *QueryChain) Pages(opts ...PagesOptions) ([]map[string]interface{}, error) {
	var allItems []map[string]interface{}
	var cursor *string
	maxPages := 0
	limit := int32(0)

	// Parse options if provided
	if len(opts) > 0 {
		if opts[0].MaxPages > 0 {
			maxPages = opts[0].MaxPages
		}
		if opts[0].Limit > 0 {
			limit = opts[0].Limit
		}
	}

	pageCount := 0

	for {
		// Build options for this page
		queryOpts := &QueryOptions{
			Cursor: cursor,
		}

		if limit > 0 {
			queryOpts.Limit = &limit
		}

		// Override options if they were set on the query chain
		if qc.options != nil {
			if qc.options.Limit != nil && limit == 0 {
				queryOpts.Limit = qc.options.Limit
			}
			if qc.options.Order != nil {
				queryOpts.Order = qc.options.Order
			}
			if qc.options.Raw {
				queryOpts.Raw = qc.options.Raw
			}
		}

		// Execute query with cursor
		tempChain := &QueryChain{
			entity:        qc.entity,
			accessPattern: qc.accessPattern,
			index:         qc.index,
			pkFacets:      qc.pkFacets,
			skCondition:   qc.skCondition,
			filterBuilder: qc.filterBuilder,
			options:       queryOpts,
		}

		result, err := tempChain.Go()
		if err != nil {
			return nil, err
		}

		// Append items from this page
		allItems = append(allItems, result.Data...)

		// Update cursor for next page
		cursor = result.Cursor

		pageCount++

		// Stop if no more pages or max pages reached
		if cursor == nil || *cursor == "" {
			break
		}

		if maxPages > 0 && pageCount >= maxPages {
			break
		}
	}

	return allItems, nil
}

// PagesIterator provides an iterator interface for paginating through results
type PagesIterator struct {
	query     *QueryChain
	cursor    *string
	options   *QueryOptions
	maxPages  int
	pageCount int
	done      bool
	err       error
}

// Page returns an iterator for manually controlling pagination
func (qc *QueryChain) Page(opts ...PagesOptions) *PagesIterator {
	maxPages := 0
	var queryOpts *QueryOptions

	if len(opts) > 0 {
		if opts[0].MaxPages > 0 {
			maxPages = opts[0].MaxPages
		}
		if opts[0].Limit > 0 {
			queryOpts = &QueryOptions{
				Limit: &opts[0].Limit,
			}
		}
	}

	if queryOpts == nil {
		queryOpts = &QueryOptions{}
	}

	// Inherit options from query chain
	if qc.options != nil {
		if qc.options.Limit != nil && queryOpts.Limit == nil {
			queryOpts.Limit = qc.options.Limit
		}
		if qc.options.Order != nil {
			queryOpts.Order = qc.options.Order
		}
		if qc.options.Raw {
			queryOpts.Raw = qc.options.Raw
		}
	}

	return &PagesIterator{
		query:     qc,
		options:   queryOpts,
		maxPages:  maxPages,
		pageCount: 0,
		done:      false,
	}
}

// Next retrieves the next page of results
// Returns (page, hasMore, error)
func (pi *PagesIterator) Next() (*Page, bool, error) {
	if pi.done {
		return nil, false, pi.err
	}

	// Check if max pages reached
	if pi.maxPages > 0 && pi.pageCount >= pi.maxPages {
		pi.done = true
		return nil, false, nil
	}

	// Build query options with cursor
	opts := &QueryOptions{
		Cursor: pi.cursor,
	}

	if pi.options.Limit != nil {
		opts.Limit = pi.options.Limit
	}
	if pi.options.Order != nil {
		opts.Order = pi.options.Order
	}
	if pi.options.Raw {
		opts.Raw = pi.options.Raw
	}

	// Execute query
	tempChain := &QueryChain{
		entity:        pi.query.entity,
		accessPattern: pi.query.accessPattern,
		index:         pi.query.index,
		pkFacets:      pi.query.pkFacets,
		skCondition:   pi.query.skCondition,
		filterBuilder: pi.query.filterBuilder,
		options:       opts,
	}

	result, err := tempChain.Go()
	if err != nil {
		pi.done = true
		pi.err = err
		return nil, false, err
	}

	// Update cursor for next iteration
	pi.cursor = result.Cursor
	pi.pageCount++

	// Create page response
	page := &Page{
		Data:   result.Data,
		Cursor: result.Cursor,
	}

	// Check if there are more pages
	hasMore := result.Cursor != nil && *result.Cursor != ""
	if !hasMore {
		pi.done = true
	}

	return page, hasMore, nil
}

// PagesOptions configures pagination behavior
type PagesOptions struct {
	MaxPages int   // Maximum number of pages to retrieve
	Limit    int32 // Items per page
}

// ScanPages returns all pages of scan results
func (s *ScanOperation) Pages(opts ...PagesOptions) ([]map[string]interface{}, error) {
	var allItems []map[string]interface{}
	var cursor *string
	maxPages := 0
	limit := int32(0)

	// Parse options if provided
	if len(opts) > 0 {
		if opts[0].MaxPages > 0 {
			maxPages = opts[0].MaxPages
		}
		if opts[0].Limit > 0 {
			limit = opts[0].Limit
		}
	}

	pageCount := 0

	for {
		// Build options for this page
		queryOpts := &QueryOptions{
			Cursor: cursor,
		}

		if limit > 0 {
			queryOpts.Limit = &limit
		}

		// Override options if they were set
		if s.options != nil {
			if s.options.Limit != nil && limit == 0 {
				queryOpts.Limit = s.options.Limit
			}
			if s.options.Raw {
				queryOpts.Raw = s.options.Raw
			}
		}

		// Execute scan with cursor
		executor := NewExecutionHelper(s.entity)
		result, err := executor.ExecuteScan(s.ctx, queryOpts)
		if err != nil {
			return nil, err
		}

		// Append items from this page
		allItems = append(allItems, result.Data...)

		// Update cursor for next page
		cursor = result.Cursor

		pageCount++

		// Stop if no more pages or max pages reached
		if cursor == nil || *cursor == "" {
			break
		}

		if maxPages > 0 && pageCount >= maxPages {
			break
		}
	}

	return allItems, nil
}

// ScanPagesIterator provides an iterator interface for scan pagination
type ScanPagesIterator struct {
	scan      *ScanOperation
	cursor    *string
	options   *QueryOptions
	maxPages  int
	pageCount int
	done      bool
	err       error
}

// Page returns an iterator for manually controlling scan pagination
func (s *ScanOperation) Page(opts ...PagesOptions) *ScanPagesIterator {
	maxPages := 0
	var queryOpts *QueryOptions

	if len(opts) > 0 {
		if opts[0].MaxPages > 0 {
			maxPages = opts[0].MaxPages
		}
		if opts[0].Limit > 0 {
			queryOpts = &QueryOptions{
				Limit: &opts[0].Limit,
			}
		}
	}

	if queryOpts == nil {
		queryOpts = &QueryOptions{}
	}

	// Inherit options from scan operation
	if s.options != nil {
		if s.options.Limit != nil && queryOpts.Limit == nil {
			queryOpts.Limit = s.options.Limit
		}
		if s.options.Raw {
			queryOpts.Raw = s.options.Raw
		}
	}

	return &ScanPagesIterator{
		scan:      s,
		options:   queryOpts,
		maxPages:  maxPages,
		pageCount: 0,
		done:      false,
	}
}

// Next retrieves the next page of scan results
func (spi *ScanPagesIterator) Next() (*Page, bool, error) {
	if spi.done {
		return nil, false, spi.err
	}

	// Check if max pages reached
	if spi.maxPages > 0 && spi.pageCount >= spi.maxPages {
		spi.done = true
		return nil, false, nil
	}

	// Build query options with cursor
	opts := &QueryOptions{
		Cursor: spi.cursor,
	}

	if spi.options.Limit != nil {
		opts.Limit = spi.options.Limit
	}
	if spi.options.Raw {
		opts.Raw = spi.options.Raw
	}

	// Execute scan
	executor := NewExecutionHelper(spi.scan.entity)
	result, err := executor.ExecuteScan(context.Background(), opts)
	if err != nil {
		spi.done = true
		spi.err = err
		return nil, false, err
	}

	// Update cursor for next iteration
	spi.cursor = result.Cursor
	spi.pageCount++

	// Create page response
	page := &Page{
		Data:   result.Data,
		Cursor: result.Cursor,
	}

	// Check if there are more pages
	hasMore := result.Cursor != nil && *result.Cursor != ""
	if !hasMore {
		spi.done = true
	}

	return page, hasMore, nil
}
