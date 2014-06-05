package webber

import (
	"strconv"
	"fmt"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"net/http"
)

func getInstance() *Webber {
	return webber
}

var webber = New()

func init() {
	bl.Register(bl.NewRunHandler(func(logic Logic, closer <-chan struct{}) {

	}))
}

type Error struct {
	StatusCode int
	Cause      string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Cause)
}

type WebPublisher interface {
	WebPublish(int, http.ResponseWriter, *http.Request) *Error
}

type PublisherFunc func(int, http.ResponseWriter, *http.Request) *Error

type Webber struct {
	mux http.ServeMux
}

func New() *Webber {
	return &Webber{
		mux: http.NewServeMux()
	}
}

func (web *Webber) Handle(pattern string, publisher WebPublisher) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		formId := req.FormValue("id")
		if formId = "" {
			PublishError(w, Error {406, "Requires id parameter."})
		}

		id, err := strconv.Atoi(formId)
		if err != nil {
			PublishError(w, Error {400, "id parameter should contains integer."})
		}

		publisher.WebPublish(id, w, req)
	})
}

func (web *Webber) HandleStatic(pattern, file string) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, file)
	})
}

func PublishError(w http.ResponseWriter, e Error) {
	http.Error(w, e.Cause, e.StatusCode)
}
