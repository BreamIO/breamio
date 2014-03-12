package beenleigh

import (
	"time"
	"reflect"
	"errors"
	
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
)

/*
	HERE BEGINS THE LANDS OF THE MOCK IMPLEMENTATIONS.
	Seriously. Go on only if invited or if you know what you are doing.
*/

type mockEmitter struct {
	pubsubs map[string]interface{}
	subs map[string]chan bool
}

func newMockEmitter() *mockEmitter {
	return &mockEmitter{
		make(map[string]interface{}),
		map[string](chan bool){
			"new": make(chan bool, 1),
			"shutdown": make(chan bool, 1),
		},
	}
}

func (m *mockEmitter) Publish(chid string, v interface{}) interface{} {
	if m.pubsubs[chid] != nil {
		return (chan<- interface{})(m.pubsubs[chid].(chan interface{}))
	}
	switch v.(type) {
		case Spec:
			ch := make(chan Spec)
			m.pubsubs[chid] = ch
			return (chan<- Spec)(ch)
		case *gorgonzola.ETData:
			ch := make(chan *gorgonzola.ETData)
			return (chan<- *gorgonzola.ETData)(ch)
		default: return nil
	}
}

func (m *mockEmitter) Dispatch(eventID string, v interface{}) {
	if m.pubsubs[eventID] != nil {
		m.pubsubs[eventID].(chan interface{}) <- v
	}
}

func (m *mockEmitter) Subscribe(chid string, v interface{}) interface{} {
	if m.subs[chid] != nil {
		m.subs[chid] <- true
	}
	switch v.(type) {
		case Spec:
			ch := make(chan Spec)
			m.pubsubs[chid] = ch
			return (<-chan Spec)(ch)
		case struct{}:
			ch := make(chan struct{})
			m.pubsubs[chid] = ch
			return (<-chan struct{})(ch)
		default: return nil
	}
}

func (m *mockEmitter) Close() error {
	return nil
}

func (m *mockEmitter) Run() {
}

func (m *mockEmitter) TypeOf(chid string) (reflect.Type, error) {
	return reflect.TypeOf(Spec{}), nil
}

func (m *mockEmitter) subscribedTo(chid string) bool {
	select {
		case <-m.subs[chid]:
			return true
		case <-time.After(50*time.Millisecond):
			return false
	}
	return false
}


type mockIOManager struct {
	aioli.IOManager
	ees map[int]briee.EventEmitter
	started bool
}

func newMockIOManager() *mockIOManager {
	return &mockIOManager{
		aioli.New(),
		make(map[int]briee.EventEmitter),
		false,
	}
}

func (m *mockIOManager) AddEE(ee briee.EventEmitter, id int) error {
	m.ees[id] = ee
	return m.IOManager.AddEE(ee, id)
}

func (m *mockIOManager) RemoveEE(id int) error {
	delete(m.ees, id)
	return m.IOManager.RemoveEE(id)
}

func (m *mockIOManager) Run() {
	m.started = true
	m.IOManager.Run()
} 

type BLMockTrackerDriver struct {
	gorgonzola.Tracker
	created bool
	id string
}

func (m *BLMockTrackerDriver) Create() (gorgonzola.Tracker, error) {
	return m.CreateFromId("test")
}

func (m *BLMockTrackerDriver) CreateFromId(id string) (gorgonzola.Tracker, error){
	if (id == "error") {
		return nil, errors.New("Nope.")
	}
	m.created = true
	m.id = id
	return m, nil
}

func (m *BLMockTrackerDriver) List() []string {
	return []string{"test"}
}