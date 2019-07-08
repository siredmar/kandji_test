package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MaintenanceWindow represents a maintenance window, i.e. a time window in
// which we can perform maintenance for a device
type MaintenanceWindow struct {
	FromWeekday time.Weekday `json:"fromWeekday,omitempty"`
	FromHour    int          `json:"fromHour,omitempty"`
	FromMin     int          `json:"fromMin,omitempty"`

	ToWeekday time.Weekday `json:"toWeekday,omitempty"`
	ToHour    int          `json:"toHour,omitempty"`
	ToMin     int          `json:"toMin,omitempty"`
}

func makeWindowWeekday(wd time.Weekday) *time.Weekday {
	res := new(time.Weekday)
	*res = wd
	return res
}

func parseWindowWeekday(s string) (*time.Weekday, error) {
	if len(s) != 3 {
		return nil, fmt.Errorf("invalid format for weekday: %s", s)
	}
	switch strings.ToLower(s) {
	case "mon":
		return makeWindowWeekday(time.Monday), nil
	case "tue":
		return makeWindowWeekday(time.Tuesday), nil
	case "wed":
		return makeWindowWeekday(time.Wednesday), nil
	case "thu":
		return makeWindowWeekday(time.Thursday), nil
	case "fri":
		return makeWindowWeekday(time.Friday), nil
	case "sat":
		return makeWindowWeekday(time.Saturday), nil
	case "sun":
		return makeWindowWeekday(time.Sunday), nil
	}

	return nil, fmt.Errorf("invalid format for weekday: %s", s)
}

func parseWindowHour(s string) (*int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("invalid value for hour in %s: %+v", s, err)
	}
	if i < 0 || i > 23 {
		return nil, fmt.Errorf("invalid value for hour in %s", s)
	}
	return &i, nil
}

func parseWindowMinute(s string) (*int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("invalid value for hour in %s: %+v", s, err)
	}
	if i < 0 || i > 59 {
		return nil, fmt.Errorf("invalid value for hour in %s", s)
	}
	return &i, nil
}

func parseWindowAll(str string) (*time.Weekday, *int, *int, error) {
	parts := strings.Split(str, ":")
	if len(parts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid format in %s", str)
	}

	weekday, err := parseWindowWeekday(parts[0])
	if err != nil {
		return nil, nil, nil, err
	}

	hour, err := parseWindowHour(parts[1])
	if err != nil {
		return nil, nil, nil, err
	}

	min, err := parseWindowMinute(parts[2])
	if err != nil {
		return nil, nil, nil, err
	}

	return weekday, hour, min, nil
}

// NewMaintenanceWindow creates a new MaintenanceWindow by parsing the given
// string in the format "Tue:01:23-Fri:23:58"
func NewMaintenanceWindow(s string) (*MaintenanceWindow, error) {
	minWindowSize := 1 * time.Hour

	if len(s) != 19 {
		return nil, fmt.Errorf("invalid format: %s", s)
	}
	split := strings.Split(s, "-")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid format: %s", s)
	}

	window := MaintenanceWindow{}

	fromWeekday, fromHour, fromMin, err := parseWindowAll(split[0])
	if err != nil {
		return nil, err
	}
	window.FromWeekday = *fromWeekday
	window.FromHour = *fromHour
	window.FromMin = *fromMin

	toWeekday, toHour, toMin, err := parseWindowAll(split[1])
	if err != nil {
		return nil, err
	}
	window.ToWeekday = *toWeekday
	window.ToHour = *toHour
	window.ToMin = *toMin

	if window.Duration() < minWindowSize {
		return nil, fmt.Errorf("given window is too small. Need at least duration: %s", minWindowSize.String())
	}
	return &window, nil
}

func (w *MaintenanceWindow) asDates() (time.Time, time.Time) {
	from := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	from = from.Add(time.Duration(w.FromHour) * time.Hour)
	from = from.Add(time.Duration(w.FromMin) * time.Minute)

	for from.Weekday() != w.FromWeekday {
		from = from.Add(24 * time.Hour)
	}

	to := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	to = to.Add(time.Duration(w.ToHour) * time.Hour)
	to = to.Add(time.Duration(w.ToMin) * time.Minute)

	for to.Weekday() != w.ToWeekday {
		to = to.Add(24 * time.Hour)
	}

	return from, to
}

// Duration returns the duration of the maintenance window
func (w *MaintenanceWindow) Duration() time.Duration {
	from, to := w.asDates()
	return to.Sub(from)
}

func dayToString(wd time.Weekday) string {
	switch wd {
	case time.Monday:
		return "Mon"
	case time.Tuesday:
		return "Tue"
	case time.Wednesday:
		return "Wed"
	case time.Thursday:
		return "Thu"
	case time.Friday:
		return "Fri"
	case time.Saturday:
		return "Sat"
	case time.Sunday:
		return "Sun"
	}
	return ""
}

// String returns the string representation of the maintenance window
func (w *MaintenanceWindow) String() string {
	fromDay := dayToString(w.FromWeekday)
	toDay := dayToString(w.ToWeekday)
	return fmt.Sprintf("%s:%02d:%02d-%s:%02d:%02d", fromDay, w.FromHour, w.FromMin, toDay, w.ToHour, w.ToMin)
}

// IsActive returns true if the target time is within the from/to range
func (w *MaintenanceWindow) IsActive(target time.Time) bool {
	from, to := w.asDates()

	now := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	now = now.Add(time.Duration(target.Hour()) * time.Hour)
	now = now.Add(time.Duration(target.Minute()) * time.Minute)

	for now.Weekday() != target.Weekday() {
		now = now.Add(24 * time.Hour)
	}

	if now.After(from) && now.Before(to) {
		return true
	}
	return false
}
