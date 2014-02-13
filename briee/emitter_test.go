package briee

import (
	"sync"
	"testing"
)

type A struct {
	Int int
	Str string
}

type B struct {
	Float float64
	Int int
}

func TestEmitter(t *testing.T) {
	var ee *LocalEventEmitter
	ee = NewEventEmitter()
	go ee.Run()

	PublA1 := ee.Publish("A", A{}).(chan<- A)
	PublA2 := ee.Publish("A", A{}).(chan<- A)

	SubsB1 := ee.Subscribe("B", B{}).(<-chan B)
	SubsA1 := ee.Subscribe("A", A{}).(<-chan A)

	PublB1 := ee.Publish("B", B{}).(chan<- B)
	SubsA2 := ee.Subscribe("A", A{}).(<-chan A)

	Adata := A{42, "A data"}
	Bdata := B{13.37, 7}

	var recvA1, recvA2 A
	var recvB1 B

	var wg sync.WaitGroup
	wg.Add(2)

	go func(){
		recvA1 = <-SubsA1
		recvA2 = <-SubsA2
		recvA1 = <-SubsA1
		recvA2 = <-SubsA2
		recvB1 = <-SubsB1
		wg.Done()
	}()

	go func(){
		PublA1 <- Adata
		PublA2 <- Adata
		PublB1 <- Bdata
		wg.Done()
	}()

	wg.Wait()

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

