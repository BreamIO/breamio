/*
Package aioli declares and implements the IOManager interface. The package is used for decoding messages from active listerners and redirecting the wrapped content to selected registered event emitters.

Uses the event emitter package briee:
		import "github.com/maxnordlund/breamio/briee"

Example use:
		ee := briee.NewEventEmitter()
		go ee.Run()

		ioman := NewIOManager()
		go ioman.Run()

		// Add event emitter
		err := ioman.AddEE(&ee, 1)
		if err != nil {
			fmt.Printf("Unable to add event emitter")
		}

		// Listen on io.Reader network
		var network bytes.Buffer
		go ioman.Listen(&network)
*/
package aioli
