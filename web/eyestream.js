EyeStream = function() {
	var socket
	var events = {}

	var onmessage = function (message) {
		'use strict';
		var event = JSON.parse(message.data);
		// console.log(event);
		if (event.Error !== undefined && event.Error !== null) {
			console.log(event.Error);
			return;
		}
		
		var data = JSON.parse(atob(event.Data));
		if (events[event.Event] !== undefined) {
			events[event.Event](event)
		}
	};

	var onopen = function() {
		'use strict';
		console.log("EyeStream socket opened.");
	}

	return {
		ROOT: 256,
		connect: function(addr) {
			'use strict';
			socket = new WebSocket(addr);
			socket.onopen = onopen

			socket.onmessage = onmessage

			socket.onclose = function() {
				'use strict';
				console.log("EyeStream socket closed.")
			};
		},

		subscribe: function(event) {
			if (socket === undefined) {
				console.log("Socket is not initialized. Did you forget to connect first?")
			}

			socket.send(JSON.stringify({
				Event: event,
				Subscribe: true,
				ID: emitter
			}));
		},

		command: function(event, emitter, data) {
			if (socket === undefined) {
				console.log("Socket is not initialized. Did you forget to connect first?")
			}

			console.log({
				raw: JSON.stringify(data),
				btoa: atob(btoa(JSON.stringify(data))),
			})

			socket.send(JSON.stringify({
				Event: event,
				ID: emitter,
				Data: btoa(JSON.stringify(data))
			}));
		}
	}
}()