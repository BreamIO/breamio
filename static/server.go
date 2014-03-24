package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/maxnordlund/breamio/aioli"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
)

type testData struct {
	EyeTrackers map[int]string
	logger      *log.Logger
}

func newTest() *testData {
	return &testData{
		EyeTrackers: map[int]string{
			1: "done",
			2: "",
			3: "in-progress",
		},
		logger: log.New(os.Stdout, "[Server]", log.LstdFlags),
	}
}

func (data *testData) must(err error) {
	if err != nil {
		data.logger.Fatalf("JSON error: %s\n", err)
	}
}

func (data *testData) mustMarshal(val interface{}) (encoded []byte) {
	encoded, err := json.Marshal(val)
	data.must(err)
	return
}

var (
	calibrate = adder{
		prefix: "calibrate",
		times:  5,
		end: func(val *aioli.ExtPkg, ws *websocket.Conn, data *testData) {
			sec, err := time.ParseDuration("7s")
			if err != nil {
				data.logger.Fatalf("Duration parsing error: %s", err)
			}
			time.AfterFunc(sec, func() {
				websocket.JSON.Send(ws, aioli.ExtPkg{
					Event: "validate:start",
					ID:    val.ID,
					Data:  []byte{},
				})
			})
		},
	}
	validate = adder{
		prefix: "validate",
		times:  3,
		end: func(val *aioli.ExtPkg, ws *websocket.Conn, data *testData) {
			data.logger.Printf("Got final package: %v", val)
		},
	}
)

func (data *testData) handler(ws *websocket.Conn) {
	go func(ws *websocket.Conn) {
		for {
			val := new(aioli.ExtPkg)
			data.must(websocket.JSON.Receive(ws, val))
			data.logger.Printf("Event: %v", val)
			switch val.Event {
			case "calibrate:start":
				websocket.JSON.Send(ws, aioli.ExtPkg{
					Event: "calibrate:next",
					ID:    val.ID,
					Data:  val.Data,
				})
			case "calibrate:add":
				calibrate.handle(val, ws, data)
			case "validate:add":
				validate.handle(val, ws, data)
			}
		}
	}(ws)
}

type adder struct {
	prefix string
	times  int
	end    func(val *aioli.ExtPkg, ws *websocket.Conn, data *testData)
}

func (add *adder) handle(val *aioli.ExtPkg, ws *websocket.Conn, data *testData) {
	turns := 0
	data.must(json.Unmarshal(val.Data, &turns))
	if turns < add.times {
		websocket.JSON.Send(ws, aioli.ExtPkg{
			Event: add.prefix + ":next",
			ID:    val.ID,
			Data:  data.mustMarshal(turns + 1),
		})
	} else {
		websocket.JSON.Send(ws, aioli.ExtPkg{
			Event: add.prefix + ":end",
			ID:    val.ID,
			Data:  val.Data,
		})
		if add.end != nil {
			add.end(val, ws, data)
		}
	}
}

func main() {
	data := newTest()
	wsHandler := websocket.Handler(data.handler)
	pwd, err := os.Getwd()
	if err != nil {
		data.logger.Fatalf("Failed to get current working directory: %s\n", err)
	}
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		// Handle websocket requests separately, but still serve index.html.go
		if req.Header.Get("Upgrade") == "websocket" && req.Header.Get("Connection") == "Upgrade" {
			data.logger.Printf("WebSocket Upgrade request for %s\n", req.URL)
			wsHandler.ServeHTTP(rw, req)
		} else {
			data.logger.Printf("HTTP %s request for %s\n", req.Method, req.URL)
			index := template.Must(template.ParseFiles(path.Join(pwd, "index.html.go")))
			if err := index.Execute(rw, data); err != nil {
				data.logger.Fatalf("Template rendering error: %s\n", err)
			}
		}
	})
	http.ListenAndServe(":8080", nil)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
