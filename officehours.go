// Package officehours allows you to define a weekly schedule which is timezone
// aware, and determine if times (in any timezone) fall within the schedule.
package officehours

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var days = []string{
	"sunday",
	"monday",
	"tuesday",
	"wednesday",
	"thursday",
	"friday",
	"saturday",
}

// Schedules is a collection of Schedule objects.
type Schedules []*Schedule

// InAny will return true if any Schedule in the collection has the time in its
// Schedule.
func (s Schedules) InAny(t time.Time) bool {
	for _, schedule := range s {
		if schedule.InSchedule(t) {
			return true
		}
	}

	return false
}

// InAnyWithOffsets will return true if any Schedule in the collection has the
// time in its Schedule, using InScheduleWithOffsets.
func (s Schedules) InAnyWithOffsets(t time.Time, before, after time.Duration) bool {
	for _, schedule := range s {
		if schedule.InScheduleWithOffsets(t, before, after) {
			return true
		}
	}

	return false
}

// A Schedule holds a weekly schedule of time ranges and days of the week,
// which can be queried with a time to see if that time falls in or out of the
// schedule.
type Schedule struct {
	daily    map[string][]string
	location *time.Location
}

// NewSchedule instantiates a new schedule. The passed-in map must have valid
// full day-of-the-week names as keys, though case is ignored (e.g. Monday and
// monday are valid), and the values must be a slice with a length of exactly
// two. They correspond to the start and end time for that day, and the format
// must be time.Kitchen (e.g. 3:00PM). The passed-in zone name is required, and
// must be known to the operating system (e.g. "America/Los_Angeles", "MST").
func NewSchedule(daily map[string][]string, zoneName string) (*Schedule, error) {
	location, err := time.LoadLocation(zoneName)
	if err != nil {
		return nil, fmt.Errorf("problem parsing zone '%s': %v", zoneName, err)
	}
Days:
	for provided := range daily {
		for _, allowed := range days {
			if strings.ToLower(provided) == allowed {
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

	// we just lowercase all the day names so that it doesn't matter what case
	// they were provided with.
	normalizedCase := make(map[string][]string)
	for day, time := range daily {
		normalizedCase[strings.ToLower(day)] = time
	}

	return &Schedule{daily: normalizedCase, location: location}, nil
}

// InSchedule takes a time and determines if it falls under the weekly
// schedule. You should be intentional about setting the timezone of the
// passed-in time, because the comparison is timezone aware.
func (s Schedule) InSchedule(t time.Time) bool {
	return s.InScheduleWithOffsets(t, 0, 0)
}

// InScheduleWithOffsets checks to see if the passed-in time is within the
// schedule, tweaking the schedule to move the start and end times by the
// passed-in duration. To move a time forward, pass a negative time.Duration.
// Can be used to allow multiple objects to use the same schedule, but have
// some of them always shut down a little earlier and start up a little later.
func (s Schedule) InScheduleWithOffsets(t time.Time, before time.Duration, after time.Duration) bool {
	localized := t.In(s.location)
	times, found := s.daily[strings.ToLower(localized.Weekday().String())]
	if !found {
		return false
	}

	// these were all validated good in the constructor
	start, _ := time.Parse(time.Kitchen, times[0])
	start = start.Add(before)
	end, _ := time.Parse(time.Kitchen, times[1])
	end = end.Add(after)
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
