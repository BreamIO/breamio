/*
Business Logic module

Beenleigh handles all business logic of BreamIO Eriver
It exposes a interface for interacting with this logic 
but not the actual implementation.
*/
package beenleigh

import (
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"log"
	"os"
	"io"
	"net"
)

const (
	tcpJSONaddr = ":3031"
)

// The interface of a BreamIO logic.
// It allows you get access to the primary EventEmitter it listens to.
// In order for it to listen to anything, the ListenAndServe method must first be called
type Logic interface {
	RootEmitter() briee.EventEmitter
	ListenAndServe()
	io.Closer
}

// Creates a instance of the current Logic implementation.
func New(eef func() briee.EventEmitter, io aioli.IOManager) Logic {
	return newBL(eef, io)
}

type handlerFunc func(Spec) error


// First actual implementation
// Allows creation of trackers and statistics modules using the "new" event.
type breamLogic struct {
	root briee.EventEmitter
	ioman aioli.IOManager
	logger *log.Logger
	closer chan struct{}
	onNewTrackerEvent handlerFunc
	eventEmitterConstructor func() briee.EventEmitter
}

func newBL(eef func() briee.EventEmitter, io aioli.IOManager) *breamLogic {
	logic := new(breamLogic)
	logic.logger = log.New(os.Stdout, "[Beenleigh]", log.Ldate)
	logic.closer = make(chan struct{})
	logic.eventEmitterConstructor = eef
	
	//Create the first event emitter
	logic.root = eef()
	
	if io != nil {
		//Hook it up to the io manager
		logic.ioman = io
		logic.ioman.AddEE(logic.root, 256)
	}
	
	logic.onNewTrackerEvent = func() handlerFunc {
		return func(spec Spec) error {
			return onNewTrackerEvent(logic, spec)
		}
	}()
	
	return logic
}

func (bl *breamLogic) RootEmitter() briee.EventEmitter {
	return bl.root
}

func (bl *breamLogic) ListenAndServe() {
	//Subscribe to events
	newEvents := bl.root.Subscribe("new", Spec{}).(<-chan Spec)
	shutdownEvents := bl.root.Subscribe("shutdown", struct{}{}).(<-chan struct{})
	
	go bl.ioman.Run()
	
	go func() {
		addr, err := net.ResolveTCPAddr("tcp", ":3031")
		if err != nil{
			bl.logger.Printf("Failed to resolve TCP address %s: %s\n", tcpJSONaddr, err)
			return
		}
		
		ln, err := net.ListenTCP("tcp", addr)
		if err != nil {
			bl.logger.Printf("Failed to listen on TCP address %s: %s\n", tcpJSONaddr, err)
			return
		}
		defer ln.Close()
		
		for {
			select {
				case <-bl.closer:
					return
				default:
			}
			in, err := ln.Accept()
			if err != nil {
				bl.logger.Printf("Failed to accept connection on TCP address %s: %s\n", tcpJSONaddr,  err)
				return
			}
			dec := aioli.NewDecoder(in)
			go bl.ioman.Listen(dec)
		}
	}()
	
	for {
		select {
			case event := <- newEvents:
				switch event.Type {
				case "tracker":
					if err := bl.onNewTrackerEvent(event); err != nil {
						bl.root.Dispatch("error:new:tracker", err)
					}
				case "statistics":
					/*if err := bl.onNewStatisticsEvent(event); err != nil {
						bl.root.Dispatch("error:new:tracker", err)
					}*/
				}
			case <- shutdownEvents:
				bl.logger.Println("Recieved shutdown event.")
				return
			case <- bl.closer:
				//bl.logger.Println("Time to close the shop!")
				return
		}
	}
}

func onNewTrackerEvent(bl *breamLogic, event Spec) error {
	bl.logger.Println("Recieved new:tracker event.")
	ee := bl.eventEmitterConstructor()
	bl.ioman.AddEE(ee, event.Emitter)
	tracker, err := gorgonzola.CreateFromURI(event.Data)
	if err != nil {
		bl.logger.Printf("Could not create new tracker with uri %s: %s", event.Data, err)
		return err
	}
	tracker.Connect()
	go tracker.Link(ee)
	bl.logger.Printf("Created a new tracker with uri %s on EE %d.\n", event.Data, event.Emitter)
	return nil
}

func (bl *breamLogic) Close() error {
	close(bl.closer)
	return nil
}

// A specification for creation of new objects.
// Type should be a type available for creation by the logic implementation.
// Data is a context sensitive string, which syntax depends on the type.
// Emitter is a integer, identifying the emitter number to link the new object to. 
type Spec struct {
	Type string
	Data string
	Emitter int
}