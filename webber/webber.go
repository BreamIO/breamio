package webber

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
)

const (
	ListenAddress = ":8080"
)

var Root = "web"

func GetInstance() *Webber {
	return webber
}

var webber = New()

func init() {
	if installpath := os.Getenv("EYESTREAM"); installpath != "" {
		Root = path.Join(installpath, "web")
	}

	bl.Register(bl.NewRunHandler(func(logic bl.Logic, closer <-chan struct{}) {
		webber.logger.Println("Initializing Webserver")
		//drawerTmpl := template.Must(template.ParseFiles(path.Join(Root, drawer)))

		webber.addServings()

		go webber.ListenAndServe()
		<-closer
		webber.logger.Println("Stopping Webserver")
		webber.Close()
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

func (f PublisherFunc) WebPublish(id int, w http.ResponseWriter, req *http.Request) *Error {
	return f(id, w, req)
}

type Webber struct {
	mux *http.ServeMux

	logger   *log.Logger
	listener net.Listener
}

func New() *Webber {
	return &Webber{
		mux:    http.NewServeMux(),
		logger: log.New(os.Stdout, "[Webber] ", log.LstdFlags),
	}
}

func (web *Webber) ListenAndServe() error {
	var err error
	web.listener, err = net.Listen("tcp", ListenAddress)
	if err != nil {
		return err
	}
	web.logger.Printf("Listening on %s", ListenAddress)
	http.Serve(web.listener, web.mux)
	return nil
}

func (web *Webber) Handle(pattern string, publisher WebPublisher) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		formId := req.FormValue("id")
		if formId == "" {
			PublishError(w, Error{406, "Requires id parameter."})
			return
		}

		id, err := strconv.Atoi(formId)
		if err != nil {
			PublishError(w, Error{400, "id parameter should contain integer."})
			return
		}

		publisher.WebPublish(id, w, req)
	})
}

func (web *Webber) HandleStatic(pattern, file string) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		web.logger.Printf("Static request for %s.", pattern)
		http.ServeFile(w, req, file)
	})
}

func (web *Webber) HandleWebSocket(pattern string, handler websocket.Handler) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Handle websocket requests separately, but still serve static files
		if r.Header.Get("Upgrade") == "websocket" && r.Header.Get("Connection") == "Upgrade" {
			web.logger.Printf("Websocket request recieved on %s.", pattern)
			handler.ServeHTTP(w, r)
		} else {
			PublishError(w, Error{405, "Upgrade Required"})
			return
		}
	})
}

func (web *Webber) Close() error {
	if web.listener != nil {
		return web.listener.Close()
	}
	return nil
}

func (web *Webber) addServings() {
	web.HandleStatic("/control", path.Join(Root, "control.html"))
	web.HandleStatic("/api/eyestream.js", path.Join(Root, "eyestream.js"))
	web.Handle("/trail", PublisherFunc(func(id int, w http.ResponseWriter, req *http.Request) *Error {
		drawerTmpl, err := template.ParseFiles(path.Join(Root, "trail.html"))
		if err != nil {
			log.Println(err)
			PublishError(w, Error{500, "Template Parse Error"})
		}
		drwr := drawer{
			Id: id,
		}
		drawerTmpl.Execute(w, drwr) //TODO catch any errors.
		return nil
	}))
	web.HandleStatic("/stats", path.Join(Root, "stats.html"))
}

func PublishError(w http.ResponseWriter, e Error) {
	http.Error(w, e.Cause, e.StatusCode)
}
