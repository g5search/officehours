package officehours

import (
	"strings"
	"testing"
	"time"
)

func TestSchedule(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Error("expected America/Los_Angeles timezone to load properly")
	}

	suite := []struct {
		Name         string
		Location     string
		Schedule     map[string][]string
		Err          string
		Expectations map[string]bool
	}{
		{
			Name: "with a working schedule",
			Schedule: map[string][]string{
				"Monday": []string{"9:00AM", "5:00PM"},
				"Friday": []string{"9:00AM", "1:00PM"},
			},
			Expectations: map[string]bool{
				"Fri, 11 Aug 2017 11:00:00 PST": true,  // in schedule on day
				"Fri, 11 Aug 2017 20:00:00 PST": false, // out of schedule on day
				"Thu, 10 Aug 2017 12:00:00 PST": false, // schedule for day undefined
				"Mon, 07 Aug 2017 17:00:00 UTC": true,  // other zone in schedule
				"Mon, 07 Aug 2017 10:00:00 UTC": false, // other zone out of schedule
			},
		},
		{
			Name: "with a working schedule using lowercase day names",
			Schedule: map[string][]string{
				"monday": []string{"9:00AM", "5:00PM"},
				"friday": []string{"9:00AM", "1:00PM"},
			},
			Expectations: map[string]bool{
				"Fri, 11 Aug 2017 11:00:00 PST": true,  // in schedule on day
				"Fri, 11 Aug 2017 20:00:00 PST": false, // out of schedule on day
			},
		},
		{
			Name: "with a bad weekday name",
			Schedule: map[string][]string{
				"Shmursday": []string{"9:00AM", "5:00PM"},
				"Friday":    []string{"9:00AM", "1:00PM"},
			},
			Err: "unknown weekday name: Shmursday",
		},
		{
			Name:     "with a bad location name",
			Location: "West Testakota",
			Err:      "problem parsing zone 'West Testakota': cannot find",
		},
		{
			Name: "with a bad weekday name",
			Schedule: map[string][]string{
				"Monday": []string{"9:00AM"},
			},
			Err: "day schedule must have a start and end time",
		},
		{
			Name: "with a bad time format",
			Schedule: map[string][]string{
				"Monday": []string{"NINE AM", "TEN AT NIGHT"},
			},
			Err: "can't parse schedule: parsing time \"NINE AM\"",
		},
	}

	for _, test := range suite {
		t.Run(test.Name, func(t *testing.T) {
			// This is going to have daily savings applied or not depending on when
			// you run this. Fun times!
			location := "America/Los_Angeles"
			if test.Location != "" {
				location = test.Location
			}
			scheduleMap := map[string][]string{
				"Friday": []string{"9:00AM", "1:00PM"},
			}
			if test.Schedule != nil {
				scheduleMap = test.Schedule
			}
			schedule, err := NewSchedule(scheduleMap, location)
			if test.Err == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Error("error unexpectedly nil")
				} else {
					if !strings.Contains(err.Error(), test.Err) {
						t.Errorf("expected error message to contain '%s', got '%s'", test.Err, err.Error())
					}
				}
			}

			for s, expected := range test.Expectations {
				if schedule == nil {
					t.Error("schedule is unexpectedly nil")
					return
				}

				parsed, err := time.ParseInLocation(time.RFC1123, s, la)
				if err != nil {
					t.Errorf("parsing time '%s': %v", s, err)
				}
				if schedule.InSchedule(parsed) != expected {
					t.Errorf("expected time '%s' IsSchedule to be %v, was not", s, expected)
				}
			}
		})
	}
}

func TestSchedules(t *testing.T) {
	laMorning, err := NewSchedule(
		map[string][]string{"Monday": []string{"9:00AM", "12:00PM"}},
		"America/Los_Angeles",
	)
	if err != nil {
		t.Error("expected morning schedule to create")
	}

	laAfternoon, err := NewSchedule(
		map[string][]string{"Monday": []string{"12:00PM", "5:00PM"}},
		"America/Los_Angeles",
	)
	if err != nil {
		t.Error("expected afternoon schedule to create")
	}

	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Error("expected America/Los_Angeles timezone to load properly")
	}

	morning, err := time.ParseInLocation(time.RFC1123, "Mon, 07 Aug 2017 10:00:00 PST", la)
	if err != nil {
		t.Error("expected time to parse")
	}
	afternoon, err := time.ParseInLocation(time.RFC1123, "Mon, 07 Aug 2017 15:00:00 PST", la)
	if err != nil {
		t.Error("expected time to parse")
	}
	night, err := time.ParseInLocation(time.RFC1123, "Mon, 07 Aug 2017 20:00:00 PST", la)
	if err != nil {
		t.Error("expected time to parse")
	}
	scheduled := Schedules([]*Schedule{laMorning, laAfternoon})

	if !scheduled.InAny(morning) {
		t.Error("expected morning to be in schedule")
	}
	if !scheduled.InAny(afternoon) {
		t.Error("expected afternoon to be in schedule")
	}
	if scheduled.InAny(night) {
		t.Error("expected night to not be in schedule")
	}
}
