package main

import (
	_ "github.com/maxnordlund/breamio/aioli/access"
	"github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	_ "github.com/maxnordlund/breamio/gorgonzola/mock"
	_ "github.com/maxnordlund/breamio/gorgonzola/tobii"
)

func main() {
	done := make(chan os.Signal, 1)

	l := beenleigh.New(briee.New)
	go func() {
		<-done
		l.Close()
	}()

	l.ListenAndServe()
}
