package data

import (
	"strings"

	"github.com/jseow5177/greenlight/internal/validator"
)


type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

// Check that the client-provided Sort field matche one of the entries in our safelist.
// If it does, extract the column name from the Sort field by stripping the leading hyphen character (if it exists).
// sortColumn() is constructed to panic if the client-provided Sort value does not match one of the entries in the safelist.
// Technically, this should not happen because the Sort value would have been checked by ValidateFilters().
// But it is a sensible failsafe to stop SQL injection.
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe soft parameter: " + f.Sort)
}

// Return the sort direction ("ASC" or "DESC") depending on the prefix character of the Sort field.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check that the sort parameter matches a value in the safelist
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}