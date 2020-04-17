package common

import (
	"testing"
	"time"
)

func TestSchedule(t *testing.T) {
	schedule := NewSchedule()
	schedule.Add(time.Now().Add(time.Minute), "now")
	schedule.Add(time.Now().Add(24*time.Hour), "tomoror")
	<-time.After(time.Second)
	t.Log(schedule.GetSchedule())
}
