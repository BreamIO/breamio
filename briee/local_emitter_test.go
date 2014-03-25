package briee

import (
	"reflect"
	"sync"
	//"log"
	"testing"
)

type A struct {
	Int int
	Str string
}

type B struct {
	Float float64
	Int   int
}

func TestNewEmitter(t *testing.T) {
	ee := New()

	PublA1 := ee.Publish("A", A{}).(chan<- A)

	SubsA1 := ee.Subscribe("A", A{}).(<-chan A)

	Adata := A{42, "A data"}
	var recvA1 A

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		recvA1 = <-SubsA1
		wg.Done()
	}()

	go func() {
		PublA1 <- Adata
		wg.Done()
		close(PublA1)
	}()

	wg.Wait()

	if Adata != recvA1 {
		t.Errorf("Got data %v, want %v", recvA1, Adata)
	}
}

func TestEmitter(t *testing.T) {
	ee := New()

	PublA1 := ee.Publish("A", A{}).(chan<- A)
	PublA2 := ee.Publish("A", A{}).(chan<- A)

	SubsB1 := ee.Subscribe("B", &B{}).(<-chan *B)
	SubsA1 := ee.Subscribe("A", A{}).(<-chan A)

	PublB1 := ee.Publish("B", &B{}).(chan<- *B)
	SubsA2 := ee.Subscribe("A", A{}).(<-chan A)

	Adata := A{42, "A data"}
	Bdata := &B{13.37, 7}

	var recvA1, recvA2 A
	var recvB1 *B

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		recvA1 = <-SubsA1
		recvA2 = <-SubsA2
		recvA1 = <-SubsA1
		recvA2 = <-SubsA2
		recvB1 = <-SubsB1
		wg.Done()
	}()

	go func() {
		PublA1 <- Adata
		PublA2 <- Adata
		PublB1 <- Bdata
		wg.Done()
	}()

	wg.Wait()

	close(PublA1)
	close(PublA2)
	close(PublB1)

	if Adata != recvA1 {
		t.Errorf("Got data %v, want %v", recvA1, Adata)
	}

	if Adata != recvA2 {
		t.Errorf("Got data %v, want %v", recvA2, Adata)
	}

	if Bdata != recvB1 {
		t.Errorf("Got data %v, want %v", recvB1, Bdata)
	}
}

func testNilPublisher(t *testing.T) {
	ee := New()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Nil type in Publish did not trigger panic")
		}
	}()
	_ = ee.Publish("A", nil).(chan<- A)
}

func testNotification(t *testing.T) {
	ee := New()

	publ := ee.Publish("Notification", struct{}{}).(chan<- struct{})
	subs := ee.Subscribe("Notification", struct{}{}).(<-chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		publ <- struct{}{}
		wg.Done()
	}()

	go func() {
		<-subs
		wg.Done()
	}()

	wg.Done()
}

func TestTypeOf(t *testing.T) {
	ee := New()

	_ = ee.Publish("A", A{}).(chan<- A)
	Adata := A{42, "A data"}
	// Test type of event
	atype, err := ee.TypeOf("A")

	if err != nil {
		t.Errorf("Unknown event identifer")
	}

	if atype != reflect.TypeOf(Adata) {
		t.Errorf("Unmatched types")
	}

	_, err = ee.TypeOf("B")
	if err == nil {
		t.Errorf("TypeOf an unregistered event shall cause an error")
	}

}

func TestTypes(t *testing.T) {
	ee := New()

	_ = ee.Publish("Map", map[string]A{}).(chan<- map[string]A)
	_ = ee.Subscribe("Map", map[string]A{}).(<-chan map[string]A)
	_ = ee.Publish("Slice", []A{}).(chan<- []A)
	_ = ee.Subscribe("Slice", []A{}).(<-chan []A)

}

/*
func TestUnsubscribe(t *testing.T) {
	ee := New()
	sub := ee.Subscribe("event", A{}).(<-chan A)
	err := ee.Unsubscribe("event", sub)
	if err != nil {
		t.Errorf("error unsubscribing, %v", err)
	}
}
*/
/*
func TestDispatch(t *testing.T) {
	ee := New()
	sub := ee.Subscribe("event", struct{}{}).(<-chan struct{})
	ee.Dispatch("event", struct{}{})
	ee.Dispatch("another event", struct{}{})
	(<-sub)

}
*/
func TestPanicPublisher(t *testing.T) {
	ee := New()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Publishing on an existing event with wrong types should cause a panic")
		}
	}()

	_ = ee.Subscribe("event", A{}).(<-chan A)
	_ = ee.Publish("event", struct{}{}).(chan<- struct{})
}

func TestPanicSubscriber(t *testing.T) {
	ee := New()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Subscribing on an existing event with wrong types should cause a panic")
		}
	}()

	_ = ee.Publish("event", struct{}{}).(chan<- struct{})
	_ = ee.Subscribe("event", A{}).(<-chan A)
}

/*
func TestUnsubscribeWrongEE(t *testing.T) {
	ee1 := New()
	ee2 := New()
	sub := ee1.Subscribe("event", struct{}{}).(<-chan struct{})
	_ = ee2.Subscribe("event", struct{}{}).(<-chan struct{})
	err := ee2.Unsubscribe("event", sub)
	if err == nil {
		t.Errorf(err.Error())
	}
}

func TestUnsubscribeNoEvent(t *testing.T) {
	ee := New()
	sub := ee.Subscribe("event", struct{}{}).(<-chan struct{})
	err := ee.Unsubscribe("another event", sub)
	if err == nil {
		t.Errorf(err.Error())
	}
}
*/
