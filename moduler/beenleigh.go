/*
Business Logic module

Beenleigh handles all business logic of Bream IO Eye Stream
It exposes a interface for interacting with this logic
but not the actual implementation.
*/
package moduler

import (
	"fmt"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/config"
	"path/filepath"

	"errors"
	"io"
	"log"
	"os"
	"sync"
)

var factories = make(map[string]Factory)

// Allows a module to register a constructor to be called during startup.
// The system also allows for destructors through the Close() error method.
// This is typically used to register global events and similar.
func Register(c Factory) {
	factories[c.String()] = c
}

// The interface of a BreamIO logic.
// It allows you get access to the primary EventEmitter it listens to.
// In order for it to listen to anything, the ListenAndServe method must first be called
type Logic interface {
	RootEmitter() briee.EventEmitter
	CreateEmitter(id int) briee.EventEmitter
	EmitterLookup(int) (briee.EventEmitter, error) //Cant use remote.EmitterLookuper due to circluar dependecy
	ListenAndServe()
	Logger() Logger
	io.Closer
}

// Creates a instance of the current Logic implementation.
func New(eef func() briee.EventEmitter) Logic {
	return newBL(eef)
}

// First actual implementation
// Allows creation of trackers and statistics modules using the "new" event.
type breamLogic struct {
	root                    briee.EventEmitter
	logger                  Logger
	closer                  chan struct{}
	wg                      *sync.WaitGroup
	eventEmitterConstructor func() briee.EventEmitter
	emitters                map[int]briee.EventEmitter
	lock                    sync.RWMutex
}

func newBL(eef func() briee.EventEmitter) *breamLogic {
	logic := new(breamLogic)
	logic.logger = NewLogger(logic)
	logic.closer = make(chan struct{})
	logic.emitters = make(map[int]briee.EventEmitter)
	logic.eventEmitterConstructor = eef
	logic.wg = new(sync.WaitGroup)

	//Create the first event emitter
	logic.root = eef()
	logic.emitters[256] = logic.root

	return logic
}

func (bl *breamLogic) RootEmitter() briee.EventEmitter {
	return bl.root
}

func (bl *breamLogic) ListenAndServe() {
	defer bl.root.Close()

	bl.Logger().Println("Loading configuration")
	err := bl.LoadConfig()
	if err != nil {
		bl.logger.Fatalln(err)
	}

	//Subscribe to events
	for _, f := range factories {
		bl.Logger().Printf("Starting module %s.", f)
		if closer, ok := f.(io.Closer); ok {
			defer closer.Close()
		}
		if runner, ok := f.(Runner); ok {
			// Legacy module or simply require special behaviour
			go runner.Run(bl)
		} else {
			//Default behaviour
			go RunFactory(bl, f)
		}
	}

	shutdownEvents := bl.root.Subscribe("shutdown", struct{}{}).(<-chan struct{})

	//Set up servers.
	//ts := remote.NewTCPServer(ioman, log.New(os.Stdout, "[TCPServer] ", log.LstdFlags))
	//ws := remote.NewWSServer(ioman, log.New(os.Stdout, "[WSServer] ", log.LstdFlags))
	//go ts.Listen()
	//go ws.Listen()

	for {
		select {
		case <-shutdownEvents:
			bl.logger.Println("Recieved shutdown event.")
			bl.Close()
			return
		case <-bl.closer:
			//bl.logger.Println("Time to close the shop!")
			return
		}
	}
}

func (bl *breamLogic) Close() error {
	defer bl.lock.Unlock()
	bl.lock.Lock()

	for _, emitter := range bl.emitters {
		emitter.Close()
	}

	close(bl.closer)
	bl.Logger().Println("Alive")
	bl.wg.Wait()
	bl.Logger().Println("Alive")
	return nil
}

// Creates a new emitter on the specified id if no such emitter exists.
// Regardless of pre-existence status, the emitter of that id is returned.
func (bl *breamLogic) CreateEmitter(id int) briee.EventEmitter {
	defer bl.lock.Unlock()
	bl.lock.Lock()
	emitter, ok := bl.emitters[id]
	if !ok {
		emitter = bl.eventEmitterConstructor()
		bl.wg.Add(1)
		bl.emitters[id] = emitter
		go func(emitter briee.EventEmitter) {
			emitter.Wait()
			bl.wg.Done()
		}(emitter)

	}
	return emitter
}

func (bl *breamLogic) EmitterLookup(id int) (briee.EventEmitter, error) {
	defer bl.lock.RUnlock()
	bl.lock.RLock()
	if v, ok := bl.emitters[id]; ok {
		return v, nil
	}
	return nil, errors.New("No emitter with that id.")
}

func (breamLogic) LoadConfig() error {
	configFile := config.DefaultConfigFile
	if os.Getenv("EYESTREAM") != "" {
		configFile = filepath.Join(os.Getenv("EYESTREAM"), configFile)
	}

	f, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return config.Load(f)
}

func (bl breamLogic) Logger() Logger {
	return bl.logger
}

func (breamLogic) String() string {
	return "Beenleigh"
}

func NewLogger(n fmt.Stringer) *log.Logger {
	return NewLoggerS(n.String())
}

func NewLoggerS(name string) *log.Logger {
	return log.New(os.Stderr, "[ "+name+" ] ", log.LstdFlags|log.Lshortfile)
}
