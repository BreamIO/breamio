package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/maxnordlund/breamio/briee"
	_ "github.com/maxnordlund/breamio/eyetracker/mock"
	_ "github.com/maxnordlund/breamio/eyetracker/tobii"
	"github.com/maxnordlund/breamio/moduler"
	_ "github.com/maxnordlund/breamio/proxy/proxy"
	_ "github.com/maxnordlund/breamio/remote/access"
)

func main() {
	log.Println("Bream IO EyeStream ETFastForward Server \"Swiss Cheese\"")

	flag.Parse()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	l := moduler.New(briee.New)
	go func() {
		<-done
		l.Close()
	}()

	l.ListenAndServe()
	log.Println("Thanks to be of service!")
}
