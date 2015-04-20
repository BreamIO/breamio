package moduler

import (
	"reflect"
	"time"

	"github.com/maxnordlund/breamio/remote"
	//"github.com/maxnordlund/breamio/briee"
)

/*
	HERE BEGINS THE LANDS OF THE MOCK IMPLEMENTATIONS.
	Seriously. Go on only if invited or if you know what you are doing.
*/

type mockEmitter struct {
	pubsubs map[string]interface{}
	subs    map[string]chan bool
}

func newMockEmitter() *mockEmitter {
	return &mockEmitter{
		make(map[string]interface{}),
		map[string](chan bool){
			"new:tracker": make(chan bool, 1),
			"new:ancient": make(chan bool, 1),
			"shutdown":    make(chan bool, 1),
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
	default:
		return nil
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
	default:
		return nil
	}
}

func (m *mockEmitter) Unsubscribe(chid string, v interface{}) error {
	delete(m.pubsubs, chid)
	return nil
}

func (m *mockEmitter) Close() error {
	return nil
}

func (m *mockEmitter) Wait() {
}

func (m *mockEmitter) TypeOf(chid string) (reflect.Type, error) {
	return reflect.TypeOf(Spec{}), nil
}

func (m *mockEmitter) subscribedTo(chid string) bool {
	select {
	case <-m.subs[chid]:
		return true
	case <-time.After(100 * time.Millisecond):
		return false
	}
	return false
}

type mockIOManager struct {
	remote.IOManager
	started bool
}

func newMockIOManager(el remote.EmitterLookuper) *mockIOManager {
	return &mockIOManager{
		remote.New(el),
		false,
	}
}

func (m *mockIOManager) Run() {
	m.started = true
	m.IOManager.Run()
}
