package time

import (
	"net/http"
	"time"
	"log"
)

var (
	defaultClient = &Client{} //a default client used by http for GET, HEAD and POST 
)

func GetGoogleTime() time.Time, error  {
	resp, err := defaultClient.Head(www.google.com)
	if err != nil {
		return nil, err
	}
	date:=resp.Header["Date"]
	log.Println(date)
}