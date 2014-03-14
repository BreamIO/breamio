/*
Package aioli declares and implements the IOManager interface. The package is used for decoding messages from active listerners and redirecting the wrapped content to selected registered event emitters.

Uses the event emitter package briee:
		import "github.com/maxnordlund/breamio/briee"

Example use:
		ee := briee.New()
		go ee.Run()

		ioman := New()
		go ioman.Run()

		// Add event emitter
		err := ioman.AddEE(&ee, 1)
		if err != nil {
			fmt.Printf("Unable to add event emitter")
		}

		// Create decoder of io.Reader
		var network bytes.Buffer
		dec := NewDecoder(&network)

		// Listen on decoder
		logger := log.New(os.Stdout, "", os.LstdFlags)
		go ioman.Listen(dec, logger)
*/
package aioli
