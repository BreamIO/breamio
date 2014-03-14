/*
Package briee defines and implements the EventEmitter interface.

Example use:

		ee := New()
		go ee.Run()

		dataSend := MyStruct{...}
		var dataRecv MyStruct

		// Note explicit type assertion
		sendChan := ee.Publish("event string identifier", MyStruct{}).(chan<- MyStruct)
		recvChan := ee.Subscribe("event string identifier", MyStruct{}).(<-chan MyStruct)

		go func(){
			sendChan <- dataSend
		}()

		dataRecv = (<-recvChan)

		// dataSend == dataRecv

		ee.Close() // Will terminate ee.Run() goroutine

*/
package briee
