package main

import (
	"fmt"
	"time"
)

// Timeslot is the unit of an activity.
type Timeslot struct {
	StartTime time.Time `firestore:"start_time"`
	EndTime   time.Time `firestore:"end_time"`
	Activity  string    `firestore:"activity"`
}

func (t Timeslot) String() string {
	return fmt.Sprintf("%v - %v: %v", t.StartTime, t.EndTime, t.Activity)
}

// Duration calculates the duration a Timeslot spans.
func (t Timeslot) Duration() time.Duration {
	return t.EndTime.Sub(t.StartTime)
}
