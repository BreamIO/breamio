package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	_ "github.com/maxnordlund/breamio/aioli/access"
	"github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	_ "github.com/maxnordlund/breamio/gorgonzola/mock"
	_ "github.com/maxnordlund/breamio/gorgonzola/tobii"
	_ "github.com/maxnordlund/breamio/swiss/proxy"
)

func main() {
	log.Println("Bream IO EyeStream ETFastForward Server \"Swiss Cheese\"")

	flag.Parse()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	l := beenleigh.New(briee.New)
	go func() {
		<-done
		l.Close()
	}()

	l.ListenAndServe()
	log.Println("Thanks to be of service!")
}
