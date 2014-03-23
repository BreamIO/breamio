package main

import (
	"code.google.com/p/go.net/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
)

type WebData struct {
	EyeTrackers []int
}

func main() {
	data := WebData{
		EyeTrackers: {1, 2, 3},
	}
	logger := log.New(os.Stdout, "[Server]", log.LstdFlags)
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Failed to get current working directory: %s\n", err)
	}
	index := template.Must(template.New("index").ParseFiles(path.Join(pwd, "index.html.go")))
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		index.Execute(rw, data)
	})
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
