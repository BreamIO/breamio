package testing

import (
	. "github.com/maxnordlund/breamio/gorgonzola"
)

type MockEmitter struct {
	Pubsubs map[string]interface{}
	Unsubscribed map[string]bool
}

func (m *MockEmitter) create(eventID string, typ interface{}) {
	if _, ok := m.Pubsubs[eventID]; ok {
		return
	} else {
		switch typ.(type) {
			case *ETData:
				m.Pubsubs[eventID] = make(chan *ETData, 1)
			case Point2D:
				m.Pubsubs[eventID] = make(chan Point2D, 1)
			case Error:
				m.Pubsubs[eventID] = make(chan Error, 1)
			case struct{}:
				m.Pubsubs[eventID] = make(chan struct{}, 1)
			case float64:
				m.Pubsubs[eventID] = make(chan float64, 1)
		}
	}
}

func (m *MockEmitter) Publish(eventID string, typ interface{}) interface{} {
	m.create(eventID, typ)
	switch typ.(type) {
		case *ETData: return (chan<- *ETData)(m.Pubsubs[eventID].(chan *ETData))
		case Point2D: return (chan<- Point2D)(m.Pubsubs[eventID].(chan Point2D))
		case Error: return (chan<- Error)(m.Pubsubs[eventID].(chan Error))
		case struct{}:           return (chan<- struct{})(m.Pubsubs[eventID].(chan struct{}))
		case float64:            return (chan<- float64)(m.Pubsubs[eventID].(chan float64))
	}
	return m.Pubsubs[eventID]
}

func (m *MockEmitter) Subscribe(eventID string, typ interface{}) interface{} {
	m.create(eventID, typ)
	switch typ.(type) {
		case *ETData: return (<-chan *ETData)(m.Pubsubs[eventID].(chan *ETData))
		case Point2D: return (<-chan Point2D)(m.Pubsubs[eventID].(chan Point2D))
		case Error: return (<-chan Error)(m.Pubsubs[eventID].(chan Error))
		case struct{}:           return (<-chan struct{})(m.Pubsubs[eventID].(chan struct{}))
		case float64:
	}
	return (m.Pubsubs[eventID])
}

func (m *MockEmitter) Dispatch(eventID string, v interface{}) {
	if m.Pubsubs[eventID] != nil {
		m.Pubsubs[eventID].(chan struct{}) <- v.(struct{})
	}
}

func (m *MockEmitter) Unsubscribe(eventID string, typ interface{}) error {
	m.Unsubscribed[eventID] = true
	return nil
}