package model

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_labelsComputation(t *testing.T) {
	testcases := []struct {
		current map[string]string
		update  map[string]string
		want    map[string]string
	}{
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
			},
			update: map[string]string{
				"gridx.de/channel": "stable",
			},
			want: map[string]string{
				"gridx.de/channel": "stable",
			},
		},
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
			},
			update: map[string]string{
				"gridx.de/area": "west",
			},
			want: map[string]string{
				"gridx.de/channel": "stable",
				"gridx.de/area":    "west",
			},
		},
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
			},
			update: map[string]string{
				"gridx.de/channel": "alpha",
			},
			want: map[string]string{
				"gridx.de/channel": "alpha",
			},
		},
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
				"gridx.de/area":    "west",
			},
			update: map[string]string{
				"gridx.de/channel": "alpha",
				"gridx.de/area":    "east",
			},
			want: map[string]string{
				"gridx.de/channel": "alpha",
				"gridx.de/area":    "east",
			},
		},
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
				"gridx.de/area":    "west",
			},
			update: map[string]string{
				"gridx.de/channel-": "",
			},
			want: map[string]string{
				"gridx.de/area": "west",
			},
		},
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
				"gridx.de/area":    "west",
			},
			update: map[string]string{
				"gridx.de/channel-": "",
				"gridx.de/area-":    "",
			},
			want: map[string]string{},
		},
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
				"gridx.de/area":    "west",
			},
			update: map[string]string{
				"gridx.de/channel_NotExists-": "stable",
			},
			want: map[string]string{
				"gridx.de/channel": "stable",
				"gridx.de/area":    "west",
			},
		},
		{
			current: nil,
			update: map[string]string{
				"gridx.de/channel": "stable",
			},
			want: map[string]string{
				"gridx.de/channel": "stable",
			},
		},
		{
			current: map[string]string{
				"gridx.de/channel": "stable",
			},
			update: nil,
			want: map[string]string{
				"gridx.de/channel": "stable",
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := ComputeLabels(tc.current, tc.update)

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected labels map: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}
