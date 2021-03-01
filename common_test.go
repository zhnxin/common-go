package common

import (
	"fmt"
	"testing"
	"time"
)

func TestSchedule(t *testing.T) {
	schedule := NewSchedule()
	for i := 0; i < 10; i++ {
		schedule.Add(time.Now().Add(3*time.Second), fmt.Sprint(i))
	}
	time.Sleep(time.Second)
	schedule.Add(time.Now().Add(time.Second), "extra")
	go func() {
		time.Sleep(3 * time.Second)
		schedule.Stop()
	}()
	for res := range schedule.Chan() {
		t.Log(res)
	}
}
