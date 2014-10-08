package time

import (
	"net/http"
	"time"
)

var (
	defaultClient = &http.Client{} //a default client used by http for GET, HEAD and POST
)

func GetGoogleTime() (time.Time, error) {
	resp, err := defaultClient.Head("http://www.google.com")
	if err != nil {
		return time.Now(), err
	}
	date := resp.Header["Date"]
	return time.Now(), nil
}
