package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/maxnordlund/breamio/aioli"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
)

const (
	Company = "Bream IO"
	Product = "Eriver"
	Version = "v2.0"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	fmt.Println("Welcome to", Company, Product, Version)
	logic := bl.New(briee.New, aioli.New())

	go func() {
		<-done
		logic.Close()
	}()

	logic.ListenAndServe()
	fmt.Println("Thank you for using our product.")
}

func init() {
	runtime.GOMAXPROCS(2)
}
