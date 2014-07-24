package 
{
	import flash.utils.ByteArray;
	import ExtPkg;
	/**
	 * ...
	 * @author David Karlbom
	 */
	public class ExtPkg {
		private var _Event:String;
		private var _Subscribe:Boolean;
		private var _ID:int;
		private var _Data:ByteArray;
		
		public function ExtPkg(Event:String, Sub:Boolean, ID:int) {
			_Event = Event;
			_Subscribe = Sub;
			_ID = ID;
		}
		
		public function get Event():String{
			return _Event;
		}
		
		public function get Subscribe():Boolean {
			return _Subscribe;
		}
		
		public function get ID():int {
			return _ID;
		}
	}

	
}