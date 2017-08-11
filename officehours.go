// Package officehours allows you to define a weekly schedule which is timezone
// aware, and determine if times (in any timezone) fall within the schedule.
package officehours

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var days = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

// A Schedule holds a weekly schedule of time ranges and days of the week,
// which can be queried with a time to see if that time falls in or out of the
// schedule.
type Schedule struct {
	daily    map[string][]string
	location *time.Location
}

// NewSchedule instantiates a new schedule. The passed-in map must have valid
// full day-of-the-week names as keys (e.g. Monday), and the values must be a
// slice with a length of exactly two. They correspond to the start and end
// time for that day, and the format must be time.Kitchen (e.g. 3:00PM). The
// passed-in zone name is required, and must be known to the operating system
// (e.g. "America/Los_Angeles", "MST").
func NewSchedule(daily map[string][]string, zoneName string) (*Schedule, error) {
	location, err := time.LoadLocation(zoneName)
	if err != nil {
		return nil, fmt.Errorf("problem parsing zone '%s': %v", zoneName, err)
	}
Days:
	for provided := range daily {
		for _, allowed := range days {
			if provided == allowed {
				continue Days
			}
		}
		return nil, fmt.Errorf("unknown weekday name: %s", provided)
	}

	for _, times := range daily {
		if len(times) != 2 {
			return nil, errors.New("day schedule must have a start and end time")
		}
		if _, err := time.Parse(time.Kitchen, times[0]); err != nil {
			return nil, fmt.Errorf("can't parse schedule: %v", err)
		}
		if _, err := time.Parse(time.Kitchen, times[1]); err != nil {
			return nil, fmt.Errorf("can't parse schedule: %v", err)
		}
	}

	return &Schedule{daily: daily, location: location}, nil
}

// InSchedule takes a time and determines if it falls under the weekly
// schedule. You should be intentional about setting the timezone of the
// passed-in time, because the comparison is timezone aware.
func (s Schedule) InSchedule(t time.Time) bool {
	localized := t.In(s.location)
	times, found := s.daily[localized.Weekday().String()]
	if !found {
		return false
	}

	// these were all validated good in the constructor
	start, _ := time.Parse(time.Kitchen, times[0])
	end, _ := time.Parse(time.Kitchen, times[1])
	startOnDay := relativeDayTime(localized, start.Hour(), start.Minute())
	endOnDay := relativeDayTime(localized, end.Hour(), end.Minute())

	return localized.After(startOnDay) && localized.Before(endOnDay)
}

// Generates a new time for the same day as localized, in the same zone, but
// using the passed-in hour and minute.
func relativeDayTime(localized time.Time, hour, minute int) time.Time {
	_, offsetSeconds := localized.Zone()
	offsetSeconds = int(math.Abs(float64(offsetSeconds)))

	// as far as I know, this can't fail
	parsed, _ := time.Parse(
		time.RFC3339,
		fmt.Sprintf(
			"%04d-%02d-%02dT%02d:%02d:00-%02d:00",
			localized.Year(),
			localized.Month(),
			localized.Day(),
			hour,
			minute,
			offsetSeconds/60/60,
		),
	)

	return parsed
}
