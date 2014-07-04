package webber

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strconv"
)

const (
	Port    = "8080"
	Address = "localhost"
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

		go func() {
			err := webber.ListenAndServe()
			if err != nil {
				webber.logger.Println("Listen and Serve error:", err)
			}
		}()
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

func PublishError(w http.ResponseWriter, e Error) *Error {
	http.Error(w, e.Cause, e.StatusCode)
	return &e
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

func (web *Webber) ListenAndServe() (err error) {
	var (
		port    = os.Getenv("PORT")
		address = os.Getenv("ADDRESS")
	)

	if port == "" {
		port = Port
	}
	if address == "" {
		address = Address
	}

	listenAddress := fmt.Sprintf("%s:%s", address, port)
	web.listener, err = net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	web.logger.Printf("Listening on %s", listenAddress)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	http.Serve(web.listener, web.mux)
	return nil
}

func (web *Webber) Handle(pattern string, publisher WebPublisher) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		formId := req.FormValue("id")
		if formId == "" {
			web.logger.Println("Requires id parameter.")
			PublishError(w, Error{406, "Requires id parameter."})
			return
		}

		id, err := strconv.Atoi(formId)
		if err != nil {
			log.Println("id parameter should contain integer.")
			PublishError(w, Error{400, "id parameter should contain integer."})
			return
		}

		publisher.WebPublish(id, w, req)
	})
}

func (web *Webber) HandleStatic(pattern, file string) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		web.logger.Printf("Static request for %s.", pattern)
		web.logger.Println(file)
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
	web.HandleStatic("/consumer", path.Join(Root, "consumer.html"))
	web.HandleStatic("/api/eyestream.js", path.Join(Root, "eyestream.js"))
	web.HandleStatic("/dep/bluebird.js", path.Join(Root, "bluebird.js"))
	web.Handle("/trail", PublisherFunc(func(id int, w http.ResponseWriter, req *http.Request) *Error {
		drawerTmpl, err := template.ParseFiles(path.Join(Root, "trail.html"))
		if err != nil {
			web.logger.Println("Template parse error:", err)
			return PublishError(w, Error{500, "Template parse error"})
		}
		drwr := drawer{
			Id: id,
		}
		drawerTmpl.Execute(w, drwr) //TODO catch any errors.
		return nil
	}))
	web.Handle("/stats", PublisherFunc(func(id int, w http.ResponseWriter, req *http.Request) *Error {
		tmpl, err := template.ParseFiles(path.Join(Root, "stats.html"))
		if err != nil {
			web.logger.Println("Template parse error:", err)
			PublishError(w, Error{500, "Template parse error"})
		}
		drwr := drawer{
			Id: id,
		}
		tmpl.Execute(w, drwr) //TODO catch any errors.
		return nil
	}))
	// web.HandleStatic("/stats", path.Join(Root, "stats.html"))
	web.mux.HandleFunc("/calibrate", func(w http.ResponseWriter, req *http.Request) {
		calibrateTmpl, err := template.ParseFiles(path.Join(Root, calibrate))
		if err != nil {
			web.logger.Println("Template parse error:", err)
			PublishError(w, Error{500, "Template parse error"})
		}
		normalizeSource, err := ioutil.ReadFile(path.Join(Root, normalize))
		if err != nil {
			web.logger.Println("File read error:", err)
			PublishError(w, Error{500, "File read error"})
		}
		cali := Calibrate{
			Id: 1,
			EyeTrackers: map[int]string{
				1: "", // done
				2: "", // in-progress
				3: "",
				4: "",
				5: "",
			},
			Normalize: template.CSS(string(normalizeSource)),
		}
		web.logger.Println("Serving request for calibrate")
		err = calibrateTmpl.Execute(w, cali)
	})
}
