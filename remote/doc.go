/*
Package remote declares and implements the IOManager interface. The package is used for decoding messages from active listerners and redirecting the wrapped content to selected registered event emitters.

Uses the event emitter package briee:
		import "github.com/maxnordlund/breamio/briee"
		import "github.com/maxnordlund/breamio/moduler"

Example use:
		ee := briee.New()
		go ee.Run()

		bl := moduler.New(briee.New) //Something to keep track of emitters.
		bl.ListenAndServe()

		ioman := New(bl)
		go ioman.Run()

		// Create decoder of io.Reader
		var network bytes.Buffer
		dec := NewDecoder(&network)

		// Listen on decoder
		logger := log.New(os.Stdout, "", os.LstdFlags)
		go ioman.Listen(dec, logger)
*/
package remote
