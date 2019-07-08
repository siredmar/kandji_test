package types

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
				FromWeekday: time.Monday,
				FromHour:    2,
				FromMin:     0,

				ToWeekday: time.Sunday,
				ToHour:    17,
				ToMin:     0,
			},
		},
		{
			s: "Sun:02:30-Sun:03:50",
			want: &MaintenanceWindow{
				FromWeekday: time.Sunday,
				FromHour:    2,
				FromMin:     30,

				ToWeekday: time.Sunday,
				ToHour:    3,
				ToMin:     50,
			},
		},
		{
			s: "Sun:02:30-Fri:03:50",
			want: &MaintenanceWindow{
				FromWeekday: time.Sunday,
				FromHour:    2,
				FromMin:     30,

				ToWeekday: time.Friday,
				ToHour:    3,
				ToMin:     50,
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
			if !cmp.Equal(tc.want.FromWeekday, win.FromWeekday) {
				t.Errorf("unexpected from weekday: %s", cmp.Diff(tc.want.FromWeekday, win.FromWeekday))
			}
			if !cmp.Equal(tc.want.FromHour, win.FromHour) {
				t.Errorf("unexpected from hour: %s", cmp.Diff(tc.want.FromHour, win.FromHour))
			}
			if !cmp.Equal(tc.want.FromMin, win.FromMin) {
				t.Errorf("unexpected from min: %s", cmp.Diff(tc.want.FromMin, win.FromMin))
			}
			if !cmp.Equal(tc.want.ToWeekday, win.ToWeekday) {
				t.Errorf("unexpected to weekday: %s", cmp.Diff(tc.want.ToWeekday, win.ToWeekday))
			}
			if !cmp.Equal(tc.want.ToHour, win.ToHour) {
				t.Errorf("unexpected to hour: %s", cmp.Diff(tc.want.ToHour, win.ToHour))
			}
			if !cmp.Equal(tc.want.ToMin, win.ToMin) {
				t.Errorf("unexpected to min: %s", cmp.Diff(tc.want.ToMin, win.ToMin))
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

func Test_IsActive(t *testing.T) {
	testcases := []struct {
		now    time.Time
		window *MaintenanceWindow
		want   bool
	}{
		{ // now represents 02.01.2019 - 05:00 (Tuesday)
			now: time.Date(2019, 1, 1, 5, 0, 0, 0, time.UTC),
			window: &MaintenanceWindow{
				FromWeekday: time.Monday,
				FromHour:    4,
				FromMin:     0,

				ToWeekday: time.Monday,
				ToHour:    8,
				ToMin:     0,
			},
			want: false,
		},
		{ // now represents 02.01.2019 - 05:00 (Tuesday)
			now: time.Date(2019, 1, 1, 5, 0, 0, 0, time.UTC),
			window: &MaintenanceWindow{
				FromWeekday: time.Tuesday,
				FromHour:    4,
				FromMin:     0,

				ToWeekday: time.Tuesday,
				ToHour:    6,
				ToMin:     0,
			},
			want: true,
		},
		{ // now represents 02.01.2019 - 05:00 (Tuesday)
			now: time.Date(2019, 1, 1, 5, 0, 0, 0, time.UTC),
			window: &MaintenanceWindow{
				FromWeekday: time.Monday,
				FromHour:    20,
				FromMin:     0,

				ToWeekday: time.Wednesday,
				ToHour:    2,
				ToMin:     0,
			},
			want: true,
		},
		{ // now represents 02.01.2019 - 05:30 (Tuesday)
			now: time.Date(2019, 1, 1, 5, 30, 0, 0, time.UTC),
			window: &MaintenanceWindow{
				FromWeekday: time.Tuesday,
				FromHour:    5,
				FromMin:     31,

				ToWeekday: time.Tuesday,
				ToHour:    8,
				ToMin:     0,
			},
			want: false,
		},
		{ // now represents 02.01.2019 - 05:30 (Tuesday)
			now: time.Date(2019, 1, 1, 5, 30, 0, 0, time.UTC),
			window: &MaintenanceWindow{
				FromWeekday: time.Tuesday,
				FromHour:    5,
				FromMin:     29,

				ToWeekday: time.Tuesday,
				ToHour:    8,
				ToMin:     0,
			},
			want: true,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tc.window.IsActive(tc.now)

			if got != tc.want {
				t.Fatalf("unexpected response")
			}
		})
	}
}
