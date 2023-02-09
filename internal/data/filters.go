package data

import (
	"math"
	"strings"

	"github.com/narinderv/blockbuster/internal/validator"
)

// Structure to hold the pagination details
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

type Filters struct {
	Page     int
	PageSize int
	Sort     string
	SortList []string // List of allowed values for the sort field
}

func (filter Filters) getSortColumn() string {

	for _, col := range filter.SortList {
		if filter.Sort == col {
			return strings.TrimPrefix(filter.Sort, "-")
		}
	}

	// Will panic if sort value is not present in the sort list
	panic("Unexpected sort value: " + filter.Sort)
}

func (filter Filters) getSortDirection() string {

	if strings.HasPrefix(filter.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (filter Filters) getLimit() int {
	return filter.PageSize
}

func (filter Filters) getOffset() int {
	return (filter.Page - 1) * filter.PageSize
}

func ValidateFilters(val *validator.Validator, filters *Filters) {

	// Page
	val.Check(filters.Page > 0, "page", "must be greater than zero")
	val.Check(filters.Page <= validator.MAX_PAGE_LEN, "page", "must not be more than 10 million")

	// Page Size
	val.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	val.Check(filters.PageSize <= 100, "page_size", "must not be more than 100")

	// Sort
	val.Check(validator.Permittedvalues(filters.Sort, filters.SortList...), "sort", "invalid sort value")
}

func CalculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
