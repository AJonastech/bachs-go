package bachs

import "iter"

// ListMeta is the pagination block returned alongside every list response. It
// mirrors the API's pagination object.
type ListMeta struct {
	// NextCursor, when non-nil, fetches the following page. Nil on the last
	// page. Used by cursor-paginated endpoints.
	NextCursor *string `json:"next_cursor"`
	// PrevCursor, when non-nil, fetches the previous page. Nil on the first
	// page.
	PrevCursor *string `json:"prev_cursor"`
	// HasMore is true when more results exist beyond this page.
	HasMore bool `json:"has_more"`
	// Limit is the page size actually applied, after clamping to the maximum.
	Limit int `json:"limit"`
	// Offset is the record offset this page starts from. Used by
	// offset-paginated endpoints.
	Offset int `json:"offset"`
	// Returned is the number of items in this page.
	Returned int `json:"returned"`
	// Total is the total number of records matching the query.
	Total int `json:"total"`
}

// Page is one page of a list response: the items plus the pagination block.
type Page[T any] struct {
	Items      []T      `json:"items"`
	Pagination ListMeta `json:"pagination"`
}

// paginate turns a paged endpoint into an iterator over individual items. It
// calls fetch for the current page, yields each item, then calls advance with
// the page's ListMeta; advance updates the captured page state (offset or
// cursor) and reports whether another page should be fetched.
//
// The first error stops iteration after being yielded once. An empty page also
// stops iteration, guarding against a malformed has_more that never advances.
func paginate[T any](
	fetch func() (*Page[T], error),
	advance func(ListMeta) bool,
) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for {
			page, err := fetch()
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
			if len(page.Items) == 0 || !advance(page.Pagination) {
				return
			}
		}
	}
}

// Collect drains an item iterator into a slice, stopping at the first error.
// It is a convenience for callers who want every result at once rather than
// streaming:
//
//	all, err := bachs.Collect(client.Customers.All(ctx, nil))
//
// Beware that it loads every page into memory and makes one request per page.
func Collect[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var out []T
	for item, err := range seq {
		if err != nil {
			return out, err
		}
		out = append(out, item)
	}
	return out, nil
}
