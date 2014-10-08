package licence

import (
	"log"
	"net/http"
	"os"
	"time"
)

var (
	defaultClient = &http.Client{} //a default client used by http for GET, HEAD and POST
)

func getGoogleTime() (time.Time, error) {
	resp, err := defaultClient.Head("http://www.google.com")
	if err != nil {
		return time.Now(), err
	}
	date := resp.Header["Date"]
	layout := "Mon, 2 Jan 2006 15:04:05 MST"
	t, err := time.Parse(layout, date[0])
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

//This function is blocking
func RepeatedlyCheckEvalTime(evalLayout, evalEndDate string) {
	err := checkEvalPeriod(evalLayout, evalEndDate)
	if err != nil {
		log.Println("Failed to check evaluation date, please verify your internet connection")
		log.Println("Now exiting program")
		os.Exit(0)
	}
	for _ = range time.Tick(24 * time.Hour) {
		err := checkEvalPeriod(evalLayout, evalEndDate)
		if err != nil {
			//Try 3 more times or exit the program
			it := 0
			for _ = range time.Tick(4 * time.Minute) {
				it++
				err = checkEvalPeriod(evalLayout, evalEndDate)
				if err == nil {
					break
				}
				if it >= 3 {
					log.Println("Failed to check evaluation date, please verify your internet connection")
					log.Println("Now exiting program")
					os.Exit(0)
				}
			}
		}
	}
}

//Using googles servertime checks if  evalEndDate have passed.
//Returns error iff it fails to verify that the date has not passed.
//Layout is the layout of the date that is parsable by time.Parse(...)
func checkEvalPeriod(evalLayout, evalEndDate string) error {
	// Evaluation period check
	googleTime, err := getGoogleTime()
	if err != nil {
		return err
	}
	endDate, err := time.Parse(evalLayout, evalEndDate)
	if err != nil {
		return err
	}
	if googleTime.After(endDate) {
		log.Println("Evaluation period is over. It ended", evalEndDate)
		os.Exit(0)
	}
	return nil
}
