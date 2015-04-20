package analysis

import (
	gr "github.com/maxnordlund/breamio/eyetracker"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	coordSource := make(chan *gr.ETData)
	cb := NewCoordBuffer(coordSource, 2*time.Second, 5) //Collect data for 10 seconds and assume we get no more than 1 datapoint per second.

	res := cb.GetCoords()
	if _, ok := <-res; ok {
		t.Fatal("The coordbuffer returns coords even though it is empty")
	}

	coords := make([]gr.XYer, 11)

	for k, _ := range coords {
		coords[k] = gr.Point2D{float64(k), float64(k)} // coords == {(0,0), (1,1), ... (10,10)}
	}

	for i := 0; i < 10; i++ {
		coordSource <- &gr.ETData{
			Filtered:  coords[i],
			Timestamp: time.Now(),
		}
	}
	time.Sleep(10 * time.Millisecond) //Make sure the buffer has time to append all coordinates

	res = cb.GetCoords()
	it := 0
	for c := range res {
		if c.Filtered.X() != float64(it) || c.Filtered.Y() != float64(it) {
			t.Fatal("The buffer is not returning the right coords")
		}
		it++
	}
	if it != 10 {
		t.Fatal("The buffer is returning the wrong number of coordinates")
	}

	for _, v := range coords {
		coordSource <- &gr.ETData{
			Filtered:  v,
			Timestamp: time.Now(),
		}
	}
	time.Sleep(10 * time.Millisecond) //Make sure the buffer has time to append all coordinates

	res = cb.GetCoords()
	it = 0
	rc := make([]*gr.ETData, 0, 12)
	for c := range res {
		rc = append(rc, c)
		it++
	}
	if it < 10 {
		t.Fatal("The buffer is returning the wrong number of coordinates")
	}
	for i := len(rc); i > len(rc)-10; i-- {
		if rc[i-1].Filtered.X() != float64(10-len(rc)+i) {
			t.Fatal("Wrong data is returned from the buffer")
		}
	}

	time.Sleep(3 * time.Second) //Make sure the buffer have time to forget all coordinates

	cb.GetCoords()
	if _, ok := <-res; ok {
		t.Fatal("The buffer is returning coordinates even though it should be empty")
	}

	cb.SetInterval(4 * time.Second)
	for _, v := range coords {
		coordSource <- &gr.ETData{
			Filtered:  v,
			Timestamp: time.Now(),
		}
		coordSource <- &gr.ETData{
			Filtered:  v,
			Timestamp: time.Now(),
		}
		coordSource <- &gr.ETData{
			Filtered:  v,
			Timestamp: time.Now(),
		}
	}
	time.Sleep(20 * time.Millisecond) //Make sure the buffer has time to append all coordinates

	res = cb.GetCoords()
	it = 0
	for c := range res {
		if c != nil {
			it++
		}
	}
	if it < 20 || it > 27 {
		t.Fatal("The buffer did not change interval properly")
	}

	cb.SetDesiredFreq(1)

	for _, v := range coords {
		coordSource <- &gr.ETData{
			Filtered:  v,
			Timestamp: time.Now(),
		}
		coordSource <- &gr.ETData{
			Filtered:  v,
			Timestamp: time.Now(),
		}
		coordSource <- &gr.ETData{
			Filtered:  v,
			Timestamp: time.Now(),
		}
	}
	time.Sleep(20 * time.Millisecond) //Make sure the buffer has time to append all coordinates

	res = cb.GetCoords()
	it = 0
	for c := range res {
		if c != nil {
			it++
		}
	}
	if it < 4 || it > 10 {
		t.Fatal("The buffer did not change frequenzy properly")
	}

	cb.Start()
	res = cb.GetCoords()
	it = 0
	for c := range res {
		if c != nil {
			it++
		}
	}
	if it == 0 {
		t.Fatal("The buffer flushed on start")
	}

	cb.Stop()
	res = cb.GetCoords()
	for c := range res {
		if c != nil {
			t.Fatal("Stop does not empty the buffer")
		}
	}

	coordSource <- &gr.ETData{
		Filtered:  coords[0],
		Timestamp: time.Now(),
	}
	time.Sleep(10 * time.Millisecond)

	res = cb.GetCoords()
	it = 0
	for c := range res {
		if c != nil {
			it++
		}
	}
	if it > 0 {
		t.Fatal("Stop does not stop appending of coords")
	}

	cb.Start()
	coordSource <- &gr.ETData{
		Filtered:  coords[0],
		Timestamp: time.Now(),
	}
	time.Sleep(10 * time.Millisecond)
	res = cb.GetCoords()
	it = 0
	for c := range res {
		if c != nil {
			it++
		}
	}
	if it == 0 {
		t.Fatal("Start does not start appending of coords")
	}

	cb.Flush()
	res = cb.GetCoords()
	it = 0
	for c := range res {
		if c != nil {
			it++
		}
	}
	if it != 0 {
		t.Fatal("Flush does not empty buffer")
	}
}
