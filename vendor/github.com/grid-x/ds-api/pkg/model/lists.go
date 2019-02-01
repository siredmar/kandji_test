package model

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	// PerPageMax is a internal maximum amount of results
	PerPageMax = 100
	// PerPageDefault overrides the api default value (20) of pages
	PerPageDefault = 20
	// PerPageMin is a internal minimum amount of results
	PerPageMin = 1

	// OrderAsc is a valid value for the api request
	OrderAsc = "asc"
	// OrderDesc is a valid value for the api request
	OrderDesc = "desc"

	// LinkHeaderKey represets the key for requests. Its used to construct valid
	// URLS for a request.
	LinkHeaderKey = "Link"
	// LinkHeaderValue is a format-string to create a valid request.
	// it does include pagination information
	// Example: <https://api.github.com/user/repos?page=50&per_page=100>; rel="last"
	LinkHeaderValue = `<%s?page=%d&per_page=%d&sort=%s&order=%s%s>; rel="%s"`
)

var (
	// AllOrders is global slice which contains two filter for
	// ascending and descending by default
	AllOrders = []string{OrderAsc, OrderDesc}
)

// ListOptions specifies the optional parameters to various List
// methods that support pagination.
type ListOptions struct {
	Page    int
	PerPage int
}

// SortedListOptions specifies the optional parameters to various
// List methods thatsupport pagination and sorting by a key and order.
// Extends *ListOptions.
type SortedListOptions struct {
	Sort  string
	Order string
	*ListOptions
}

// NewListOptions parses v url.Values into *ListOptions that can be used to do
// pagination on any function support ListOptions.
func NewListOptions(v url.Values) *ListOptions {
	options := &ListOptions{
		Page:    1,
		PerPage: PerPageDefault,
	}

	if page, err := strconv.Atoi(v.Get("page")); err == nil {
		options.Page = page
		if options.Page < 1 {
			options.Page = 1
		}
	}

	if perPage, err := strconv.Atoi(v.Get("per_page")); err == nil {
		options.PerPage = perPage
		if options.PerPage > PerPageMax {
			options.PerPage = PerPageMax
		}
		if options.PerPage < PerPageMin {
			options.PerPage = PerPageMin
		}
	}

	return options
}

// NewSortedListOptions parses v url.Values into *SortedListOptions that can be used to do
// pagination as well as sorting by a specific key specified in the
// "sort" GET parameter. sortDefault ist the default sortColumn
func NewSortedListOptions(v url.Values, sortDefault string, sortMap map[string]string) *SortedListOptions {

	options := &SortedListOptions{
		ListOptions: NewListOptions(v),
		Sort:        sortDefault,
		Order:       OrderAsc,
	}

	if order := v.Get("order"); order == OrderAsc || order == OrderDesc {
		options.Order = order
	}

	if sort, ok := sortMap[v.Get("sort")]; ok {
		options.Sort = sort
	}

	return options
}

func validOrDefault(input string, list []string, def string) string {
	for _, l := range list {
		if strings.EqualFold(input, l) {
			return l
		}
	}
	return def
}

// Sanitize fills missing data in the SortedListOptions to ensure that no errors occur.
// The rest api however uses the NewSortedListOptions function above so no sanitizing has to be done
// and this check becomes redundant.
func (s *SortedListOptions) Sanitize(sortDefault string, sortableFields []string) *SortedListOptions {

	if s == nil {
		return &SortedListOptions{
			ListOptions: &ListOptions{
				Page:    1,
				PerPage: PerPageDefault,
			},
			Sort:  sortDefault,
			Order: OrderAsc,
		}
	}

	if s.ListOptions == nil {
		s.ListOptions = &ListOptions{
			Page:    1,
			PerPage: PerPageDefault,
		}
	}

	// Make sure that we only have small letters
	s.Order = strings.ToLower(s.Order)

	if s.Order != OrderAsc && s.Order != OrderDesc {
		s.Order = OrderAsc
	}

	if s.Page < 1 {
		s.Page = 1
	}

	if s.PerPage > PerPageMax {
		s.PerPage = PerPageMax
	}

	if s.PerPage < PerPageMin {
		s.PerPage = PerPageMin
	}

	s.Sort = validOrDefault(s.Sort, sortableFields, sortDefault)
	return s
}

// ConstructLinkHeader returns a tuple of strings where the first element represents the name
// of the link header and the second field represents the value of the link header - this is
// list as outlined at https://developer.github.com/v3/#pagination
func ConstructLinkHeader(endpoint string, hasNext bool, sortMap map[string]string, opts *SortedListOptions, additional []string) (string, string) {

	var sort string
	for key, value := range sortMap {
		if value == opts.Sort {
			sort = key
		}
	}

	// If we have additional data, we prepend one more blank
	// string to force the firs & in the url
	if additional != nil {
		additional = append([]string{""}, additional...)
	}

	constructLink := func(relname string, f func(x int) int) string {
		return fmt.Sprintf(
			LinkHeaderValue,
			endpoint,
			f(opts.Page),
			opts.PerPage,
			sort,
			opts.Order,
			strings.Join(additional, "&"), // TODO (juliusmh): fix ugliness
			relname,
		)
	}

	headers := []string{
		constructLink("first", func(x int) int { return 1 }),
	}

	if hasNext {
		headers = append(headers, constructLink("next", func(x int) int { return x + 1 }))
	}

	if opts.Page > 1 {
		headers = append(headers, constructLink("prev", func(x int) int { return x - 1 }))
	}

	return LinkHeaderKey, strings.Join(headers, ", ")
}
