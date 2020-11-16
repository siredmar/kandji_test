package filter

type CompositeFilter struct {
	filters []Filter
}

func NewCompositeFilter(filters ...Filter) *CompositeFilter {
	return &CompositeFilter{
		filters: filters,
	}
}

func (f *CompositeFilter) Eval(in interface{}) (bool, error) {
	for _, f := range f.filters {
		include, err := f.Eval(in)
		if err != nil {
			return false, err
		}
		if !include {
			return false, nil
		}
	}
	return true, nil
}
