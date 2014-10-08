package licence

import (
	"testing"
	"time"
)

func TestGetGoogleTime(t *testing.T) {
	Gtime, err := GetGoogleTime()
	if err != nil {
		t.Fatal(err, " Crashed")
	}
	local := time.Now()
	var dur time.Duration
	if local.After(Gtime) {
		dur = local.Sub(Gtime)
	} else {
		dur = Gtime.Sub(local)
	}
	if dur.Minutes() > 2.0 { // checking if the clocks are more than 2 minutes apart
		t.Fatal("time is not synched")
	}
}
