package globalTime

import (
	"testing"
	"time"
)

func testGetGoogleTime(t *testing.T) {
	time, err := time.GetGoogleTime()
	if err != nil {
		t.Fatal(err, " Crashed")
	}
	local := time.Now()
	var dur time.Duration
	if local.After(time) {
		dur = local.Sub(time)
	} else {
		dur = time.Sub(time)
	}
	if dur.Minutes() > 2.0 { // checking if the clocks are more than 2 minutes apart
		t.Fatal("time is not synched")
	}
}
