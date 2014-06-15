(function(exports, undefined) {
	"use strict";
	var EyeStream = function EyeStream(addr) {
		this.connected = false
		this.events = {}
		this.ttl    = 0
		this.socket = new WebSocket(addr)
		this.socket.onerror = this.error.bind(this)
		this.socket.onopen = this.opened.bind(this)
		this.socket.onmessage = this.received.bind(this)
		this.socket.onclose = this.closed.bind(this)
	}

	EyeStream.ROOT = 256

	EyeStream.prototype.error = function error(event) {
		console.error("WebSocket error:", event)
		this.socket.close(4000, event.message)
	}

	EyeStream.prototype.opened = function opened(event) {
		console.info("WebSocket opened:", event)
		this.connected = true
	}

	EyeStream.prototype.closed = function closed(event) {
		console.warn("WebSocket closed:", event)
		this.connected = false
	}

	EyeStream.prototype.received = function received(event) {
		var data, handler, message = JSON.parse(event.data)
		this.ttl = Date.now()
		if (message.Error !== undefined && message.Error !== null) {
			console.error("Received server error:", event.Error)
			return
		}

		try {
			data = JSON.parse(atob(message.Data))
		} catch (err) {
			console.error("Received parse error:", err)
			return
		}

		handler = this.events[message.Event]
		if (typeof handler === "function") {
			handler(data)
		}
	}

	EyeStream.prototype.subscribe = function subscribe(event, emitter, handler) {
		console.info("Subscribe to:", event, "@", emitter)
		this.events[event] = handler
		this.socket.send(JSON.stringify({
			Event: event,
			Subscribe: true,
			ID: emitter
		}));
	}

	EyeStream.prototype.command = function command(event, emitter, data) {
		var pkg = {
			Event: event,
			ID: emitter,
			Data: data
		}
		console.info("Sending command:", pkg)
		pkg.Data = btoa(JSON.stringify(data))
		this.socket.send(JSON.stringify(pkg));
	}

	// ws://localhost:8080/api/json
	exports.EyeStream = new EyeStream(([
		(exports.location.protocol === "http:") ? "ws": "wss",
		"://",
		exports.location.host,
		"/api/json"
	]).join(""))
})(window)
