package data

import (
	"math"
	"strings"

	"github.com/jseow5177/greenlight/internal/validator"
)

// Define a new Metadata struct for holding the pagination metadata
type Metadata struct {
	CurrentPage int `json:"current_page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	FirstPage int `json:"first_page,omitempty"`
	LastPage int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

// calculateMetadata() function calculates the appropriate pagination metadata values given the total number of records, 
// current page, and page size values.
// The last page value is calculated using the math.Ceil() function. If there were 12 records in total and a page size of 5,
// the last page would be math.Ceil(12/5) = 3.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		// Return an empty metadata if there are no records
		return Metadata{}
	}
	
	return Metadata{
		CurrentPage: page,
		PageSize: pageSize,
		FirstPage: 1,
		LastPage: int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
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

// Return the pagination page size
func (f Filters) limit() int {
	return f.PageSize
}

// Return the paginated page requested by user.
// There is a theoretical risk of an integer overflow as two int values are multiplied together.
// But, it is mitigated by the validation rules in ValidateFilters().
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
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