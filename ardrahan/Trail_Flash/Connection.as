package 
{
	import ConnectionEvent;
	import flash.events.EventDispatcher;
	import flash.events.Event;
	import flash.events.TimerEvent;
	import flash.net.Socket;
	import flash.utils.ByteArray;
	import flash.events.ProgressEvent;
	import flash.errors.IOError;
	import flash.events.Event;
	import flash.events.EventDispatcher;
	import flash.events.IOErrorEvent;
	import flash.events.ProgressEvent;
	import flash.events.SecurityErrorEvent;
	import mx.utils.*;
	import flash.utils.Timer;
	import com.adobe.serialization.json.*;
	
	public class Connection extends EventDispatcher {
		private static const debugging:Boolean = true;
		private var typeOfConfig:String;
		private var typeOfEvent:String;
		private var _host: String;
		private var _port: int;
		private var base64number: String;
		private var socket:Socket;
		private var _timer:Timer;
		private var y:Number = 0.1;
		private var x:Number = 0.1;
		
		public function Connection (host:String = null, port:int = 0, type:String = null )
		{ 
			_port = port;
			_host = host;
			super();
			socket = new Socket();
			configureSocketListener();
			socket.connect(host, port);	
		}
		
		/**
		 * Setting up event Listener for the socket. 
		 * 
		 */
		private function configureSocketListener():void {
			socket.addEventListener(Event.CLOSE, closeHandler)						 // function to run when a Close event is called
			socket.addEventListener(Event.CONNECT, connectHandler) 					 // function to run when a Connection is esablished 
			socket.addEventListener(IOErrorEvent.IO_ERROR, ioHandler) 				 // function to run when a IO error occurs
			socket.addEventListener(ProgressEvent.SOCKET_DATA,readData) 				 // function to run when New incomming Data is on the socket
			socket.addEventListener(SecurityErrorEvent.SECURITY_ERROR,securityError) 	 // function to run when a Secutirity error occurs
		}
		
		private function readData(event:Event): void {
			var tmp : ByteArray = new ByteArray();
			var length:uint = socket.bytesAvailable;
			trace(length);
			socket.readBytes(tmp, 0, length);
			var parsedJson:Object = JSON.decode(tmp.readUTFBytes(tmp.length));
			var eventType:String = parsedJson.Event;
			var _event:ConnectionEvent = new ConnectionEvent(eventType);
			_event.data = decode64Byte(parsedJson.Data);
			this.dispatchEvent(_event);	
		}
		
		private function ioHandler(event:Event):void {
			if (debugging) {
				trace("An io error occured" + Event);
			}
		}
		
		private function connectHandler(event:Event):void {
			subscribe();  // maybe should be called on the object instead and so can be done from the outside after connect instead. 
			return
		}
		
		private function securityError(event:Event):void { 
			if (debugging) {
				trace("A security error occured" + Event);
			}
		}
		
		private function closeHandler(event:Event):void {
			if (debugging) {
				trace("Server closed the connection " + Event);
			}
			socket.close();
			return
		}
		
		private function decode64Byte(input:String): String {
			var tmp : ByteArray;
			var decoder:Base64Decoder = new Base64Decoder();
			decoder.decode(input)
			tmp = decoder.toByteArray();
			return tmp.readUTFBytes(tmp.length);
		}
		
		private function subscribe():void {
			var sub:ExtPkg = new ExtPkg("tracker:etdata", true, 1);
			var tosend:String = JSON.encode(sub as Object);
			socket.writeUTFBytes(tosend);
		}
	}
}