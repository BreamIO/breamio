package licence

import (
	"testing"
	"time"
)

func TestGetGoogleTime(t *testing.T) {
	Gtime, err := getGoogleTime()
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

func TestCheckEvalPeriod(t *testing.T) {
	err := checkEvalPeriod("2 Jan 2006", "2 Jan 2006")
	if err == nil {
		t.Fatal("checkEvalPeriod does not return error on old dates")
	}
	err = checkEvalPeriod("2 Jan 2006", "2 Jan 2100") //This will be a bug in a distant future
	if err != nil {
		t.Fatal("checkEvalPeriod does return error on future dates")
	}
}
