package filter

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"

	"github.com/grid-x/gxctl/pkg/api"
)

func TestFilters(t *testing.T) {
	testcases := []struct {
		desc        string
		filter      Filter
		in          interface{}
		wantInclude bool
		wantError   bool
	}{
		{
			desc:   "label no filter",
			filter: NewLabelFilter(""),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantInclude: true,
			wantError:   false,
		},
		{
			desc:        "wrong type",
			filter:      NewLabelFilter(""),
			in:          "not a device, lol",
			wantInclude: false,
			wantError:   true,
		},
		{
			desc:   "label key exists",
			filter: NewLabelFilter("foo"),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantInclude: true,
			wantError:   false,
		},
		{
			desc:   "label key does not exist",
			filter: NewLabelFilter("goo"),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantInclude: false,
			wantError:   false,
		},
		{
			desc:   "key=value exists",
			filter: NewLabelFilter("foo=bar"),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantInclude: true,
			wantError:   false,
		},
		{
			desc:   "label key=value does not exist",
			filter: NewLabelFilter("goo=baz"),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantInclude: false,
			wantError:   false,
		},
		{
			desc:   "label key exists, value is wrong",
			filter: NewLabelFilter("foo=baz"),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantInclude: false,
			wantError:   false,
		},
		{
			desc:   "label and matches",
			filter: NewLabelFilter("foo,goo=baz"),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "ANYTHING",
						"goo": "baz",
					},
				},
			},
			wantInclude: true,
			wantError:   false,
		},
		{
			desc:   "label and does not match",
			filter: NewLabelFilter("foo=bar,goo=baz"),
			in: api.Device{
				Metadata: types.Metadata{
					Labels: map[string]string{
						"foo": "bar",
						"goo": "NOPE",
					},
				},
			},
			wantInclude: false,
			wantError:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
			include, err := tc.filter.Eval(tc.in)

			didErr := false
			if tc.wantError && err == nil || !tc.wantError && err != nil {
				t.Errorf("wanted error=%v, got error=%v", tc.wantError, err)
				didErr = true
			}
			if tc.wantInclude != include {
				t.Errorf("wanted include=%v, got include=%v", tc.wantInclude, include)
				didErr = true
			}
			if didErr {
				t.Errorf("have: %v", include)
			}
		})
	}
}
