package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_parseWindowWeekday(t *testing.T) {
	testcases := []struct {
		s    string
		want time.Weekday
	}{
		{
			s:    "Mon",
			want: time.Monday,
		},
		{
			s:    "MON",
			want: time.Monday,
		},
		{
			s:    "mon",
			want: time.Monday,
		},
		{
			s:    "tue",
			want: time.Tuesday,
		},
		{
			s:    "Wed",
			want: time.Wednesday,
		},
		{
			s:    "thu",
			want: time.Thursday,
		},
		{
			s:    "Fri",
			want: time.Friday,
		},
		{
			s:    "Sat",
			want: time.Saturday,
		},
		{
			s:    "Sun",
			want: time.Sunday,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got, err := parseWindowWeekday(tc.s)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, *got) {
				t.Errorf("unexpected weekday: %s", cmp.Diff(tc.want, *got))
			}
		})
	}
}

func Test_parseWindowAll(t *testing.T) {
	testcases := []struct {
		s string

		wantWeekday time.Weekday
		wantHour    int
		wantMinute  int
	}{
		{
			s:           "Mon:02:00",
			wantWeekday: time.Monday,
			wantHour:    2,
			wantMinute:  0,
		},
		{
			s:           "Mon:02:33",
			wantWeekday: time.Monday,
			wantHour:    2,
			wantMinute:  33,
		},
		{
			s:           "sun:19:33",
			wantWeekday: time.Sunday,
			wantHour:    19,
			wantMinute:  33,
		},
		{
			s:           "fri:23:59",
			wantWeekday: time.Friday,
			wantHour:    23,
			wantMinute:  59,
		},
		{
			s:           "thu:00:00",
			wantWeekday: time.Thursday,
			wantHour:    0,
			wantMinute:  0,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			gotWeekday, gotHour, gotMinute, err := parseWindowAll(tc.s)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.wantWeekday, *gotWeekday) {
				t.Errorf("unexpected weekday: %s", cmp.Diff(tc.wantWeekday, *gotWeekday))
			}
			if !cmp.Equal(tc.wantHour, *gotHour) {
				t.Errorf("unexpected hour: %s", cmp.Diff(tc.wantHour, *gotHour))
			}
			if !cmp.Equal(tc.wantMinute, *gotMinute) {
				t.Errorf("unexpected minute: %s", cmp.Diff(tc.wantMinute, *gotMinute))
			}
		})
	}
}

func Test_NewMaintenanceWindow(t *testing.T) {
	testcases := []struct {
		s string

		want *MaintenanceWindow
	}{
		{
			s: "Mon:02:00-Sun:17:00",
			want: &MaintenanceWindow{
				fromWeekday: time.Monday,
				fromHour:    2,
				fromMin:     0,

				toWeekday: time.Sunday,
				toHour:    17,
				toMin:     0,
			},
		},
		{
			s: "Sun:02:30-Sun:03:50",
			want: &MaintenanceWindow{
				fromWeekday: time.Sunday,
				fromHour:    2,
				fromMin:     30,

				toWeekday: time.Sunday,
				toHour:    3,
				toMin:     50,
			},
		},
		{
			s: "Sun:02:30-Fri:03:50",
			want: &MaintenanceWindow{
				fromWeekday: time.Sunday,
				fromHour:    2,
				fromMin:     30,

				toWeekday: time.Friday,
				toHour:    3,
				toMin:     50,
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			win, err := NewMaintenanceWindow(tc.s)
			if err != nil {
				t.Fatal(err)
			}

			// NOTE: we need to manually compare the attributes
			// because cmp can only do this for public ones
			if !cmp.Equal(tc.want.fromWeekday, win.fromWeekday) {
				t.Errorf("unexpected from weekday: %s", cmp.Diff(tc.want.fromWeekday, win.fromWeekday))
			}
			if !cmp.Equal(tc.want.fromHour, win.fromHour) {
				t.Errorf("unexpected from hour: %s", cmp.Diff(tc.want.fromHour, win.fromHour))
			}
			if !cmp.Equal(tc.want.fromMin, win.fromMin) {
				t.Errorf("unexpected from min: %s", cmp.Diff(tc.want.fromMin, win.fromMin))
			}
			if !cmp.Equal(tc.want.toWeekday, win.toWeekday) {
				t.Errorf("unexpected to weekday: %s", cmp.Diff(tc.want.toWeekday, win.toWeekday))
			}
			if !cmp.Equal(tc.want.toHour, win.toHour) {
				t.Errorf("unexpected to hour: %s", cmp.Diff(tc.want.toHour, win.toHour))
			}
			if !cmp.Equal(tc.want.toMin, win.toMin) {
				t.Errorf("unexpected to min: %s", cmp.Diff(tc.want.toMin, win.toMin))
			}
		})
	}
}

func Test_String(t *testing.T) {
	testcases := []struct {
		s string
	}{
		{
			s: "Mon:02:00-Sun:17:00",
		},
		{
			s: "Sun:02:30-Sun:03:50",
		},
		{
			s: "Sun:02:30-Fri:03:50",
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			win, err := NewMaintenanceWindow(tc.s)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.s, win.String()) {
				t.Errorf("unexpected string: %s", cmp.Diff(tc.s, win.String()))
			}
		})
	}
}
