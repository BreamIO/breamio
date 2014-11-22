package webber

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strconv"

	"code.google.com/p/go.net/websocket"
	"github.com/gorilla/mux"

	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/module"
)

const (
	Port    = "8080"
	Address = "localhost"
)

var Root = "web"

var webber = New(module.Constructor{
	Logger: bl.NewLoggerS("Webber"),
})

func init() {
	if installpath := os.Getenv("EYESTREAM"); installpath != "" {
		Root = path.Join(installpath, "web")
	}

	bl.Register(module.SingletonFactory{Name: "Webber", Module: webber})

}

func New(c module.Constructor) *Webber {
	w := &Webber{
		SimpleModule: module.NewSimpleModule("Webber", c),
		mux:          mux.NewRouter(),
	}

	w.Logger().Println("Initializing Webserver")
	//drawerTmpl := template.Must(template.ParseFiles(drawer))

	w.addServings()

	go func() {
		err := w.ListenAndServe()
		if err != nil {
			w.Logger().Println("Listen and Serve error:", err)
		}
		w.Logger().Println("Stopping Webserver")
	}()

	return w
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
	module.SimpleModule
	mux      *mux.Router
	listener net.Listener
}

func Instance() *Webber {
	return webber
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

	web.Logger().Printf("Listening on %s", listenAddress)
	go func() {
		web.Logger().Println(http.ListenAndServe("localhost:6060", nil))
	}()

	http.Serve(web.listener, web.mux)
	return nil
}

func (web *Webber) Handle(pattern string, publisher WebPublisher) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		formId := req.FormValue("id")
		if formId == "" {
			web.Logger().Println("Requires id parameter.")
			PublishError(w, Error{406, "Requires id parameter."})
			return
		}

		id, err := strconv.Atoi(formId)
		if err != nil {
			web.Logger().Println("id parameter should contain integer.")
			PublishError(w, Error{400, "id parameter should contain integer."})
			return
		}

		publisher.WebPublish(id, w, req)
	})
}

func (web *Webber) HandleStatic(pattern, file string) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		web.Logger().Printf("Static request for %s.", pattern)
		web.Logger().Println(file)
		http.ServeFile(w, req, path.Join(Root, file))
	})
}

func (web *Webber) HandleWebSocket(pattern string, handler websocket.Handler) {
	web.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Handle websocket requests separately, but still serve static files
		if r.Header.Get("Upgrade") == "websocket" && r.Header.Get("Connection") == "Upgrade" {
			web.Logger().Printf("Websocket request recieved on %s.", pattern)
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
	web.HandleStatic("/control", "control.html")
	web.HandleStatic("/consumer", "consumer.html")
	web.HandleStatic("/api/eyestream.js", "eyestream.js")
	web.HandleStatic("/dep/bluebird.js", "bluebird.js")
	web.HandleStatic("/crossdomain.xml", "crossdomain.xml")
	web.HandleStatic("/gui", "simple_web_gui.html")

	web.HandleStatic("/colorpicker.min.js", "colorpicker.min.js")
	web.HandleStatic("/colorpicker.min.css", "colorpicker.min.css")

	web.HandleStatic("/images/select.gif", "/images/select.gif")
	web.HandleStatic("/images/overlay.png", "/images/overlay.png")
	web.HandleStatic("/images/select_hue.png", "/images/select_hue.png")
	web.HandleStatic("/images/indic.gif", "/images/indic.gif")
	web.HandleStatic("/images/gradient_input.png", "/images/gradient_input.png")
	web.HandleStatic("/images/grabber.png", "/images/grabber.png")
	web.HandleStatic("/images/submit.png", "/images/submit.png")

	web.HandleStatic("/breamio.css", "breamio.css")
	web.HandleStatic("/LANENAR_.ttf", "LANENAR_.ttf")

	web.Handle("/trail", PublisherFunc(func(id int, w http.ResponseWriter, req *http.Request) *Error {
		drawerTmpl, err := template.ParseFiles(path.Join(Root, "trail.html"))
		if err != nil {
			web.Logger().Println("Template parse error:", err)
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
			web.Logger().Println("Template parse error:", err)
			PublishError(w, Error{500, "Template parse error"})
		}
		drwr := drawer{
			Id: id,
		}
		tmpl.Execute(w, drwr) //TODO catch any errors.
		return nil
	}))
	// web.HandleStatic("/stats", "stats.html")
	web.mux.HandleFunc("/calibrate", func(w http.ResponseWriter, req *http.Request) {
		calibrateTmpl, err := template.ParseFiles(path.Join(Root, calibrate))
		if err != nil {
			web.Logger().Println("Template parse error:", err)
			PublishError(w, Error{500, "Template parse error"})
		}
		normalizeSource, err := ioutil.ReadFile(normalize)
		if err != nil {
			web.Logger().Println("File read error:", err)
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
		web.Logger().Println("Serving request for calibrate")
		err = calibrateTmpl.Execute(w, cali)
	})
}
