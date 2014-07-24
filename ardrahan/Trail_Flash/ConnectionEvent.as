package 
{
	import flash.events.Event;
	
	/**
	 * All the the evets 
	 * @author Bream iO	
	 * 
	 */
	public class ConnectionEvent extends Event
	{
		public static const CONNECTED:String = "connected";
		public static const CLOSED:String = "closed";
		public static const CONFIGDATA:String = "Confing:";
		public static const DATA:String = "tracker:etdata";
		public static const CONFIGSTATS:String = "Config:regionStatsDrawer";
		public static const STATAS:String = "Data:regionStatsDrawer"; 
		
		public var data:String = "";
		
		public function ConnectionEvent(type:String, bubbles:Boolean = false, cancelable:Boolean = false) {
			super(type, bubbles, cancelable);
		}
	}
}