package 
{
	import flash.display.MovieClip;
	import flash.display.Shape;
	import flash.display.Sprite;
	import flash.display.DisplayObject;
	import flash.display.StageScaleMode;
	import flash.display.StageAlign;
	import flash.events.Event;
	import flash.text.engine.JustificationStyle;
	import flash.text.Font;
	import flash.text.TextField;
	import flash.text.TextFormat;
	import flash.utils.ByteArray;
	import mx.utils.Base64Decoder;
	import com.adobe.serialization.json.*;
	import flash.external.ExternalInterface;
	import flash.net.LocalConnection;
	import flash.utils.Timer;
	import flash.xml.*;
	import flash.events.TimerEvent;
	import flash.system.Security;
	/**
	 * ...
	 * @author David Karlbom
	 */
	public class Main extends Sprite 
	{
		private var ost:String;
		private var _colorOfIcon:uint = 0x123456;
		private var _thickness:uint = 3;
		private var _alpha:uint = 1;
		private var _radius:uint = 15;
		private var _trailColours:Vector.<uint> = new Vector.<uint>();
		private var _amountOfColors:uint = 1;
		private var first:uint = 0;
		private var current: uint = 0; 
		private var conn : Connection;
		private var host : String = "130.229.172.21";
		private var pluginSWFLC:LocalConnection;	
		
		public function Main():void 
		{
			Security.allowDomain(host);
			Security.allowInsecureDomain(host);
			Security.loadPolicyFile("http://" + host +":8080/crossdomain.xml");
			if (stage) init();
			else addEventListener(Event.ADDED_TO_STAGE, init);
			
		}
		private function init(e:Event = null):void {
			removeEventListener(Event.ADDED_TO_STAGE, init);
			stage.scaleMode = StageScaleMode.NO_SCALE;				
			stage.align = StageAlign.TOP_LEFT;
			
			conn= new Connection(host, 4041);
			conn.addEventListener(ConnectionEvent.CONFIGDATA, updateConfig);
			conn.addEventListener(ConnectionEvent.DATA, updateData);
			/*
			ExternalInterface.addCallback("SetConnectionChannel",createLocalConnection);
			ExternalInterface.addCallback("SetConfiguration", setConfiguration);
			var timer1:Timer = new Timer(100,1);
			timer1.addEventListener(TimerEvent.TIMER, sendConfiguration);
			timer1.start();
			*/
		}
		private function createLocalConnection(pluginSWFLCName:String):void {
			
			pluginSWFLC = new LocalConnection();
			try
			{
				pluginSWFLC.connect(pluginSWFLCName);
				pluginSWFLC.client = this;
			}
			catch (error:ArgumentError) {
			this.addChild(generateImg(20, 30));
			}
		}
		
		private function updateConfig(configData: ConnectionEvent):void {
			var jsonParsed:Object = JSON.decode(configData.data);
			_thickness = jsonParsed.data.thickness;
			_alpha = jsonParsed.data.alpha;
			_radius = jsonParsed.data.radius;
			_colorOfIcon = jsonParsed.data.color[0];
			_amountOfColors = jsonParsed.data.color.length;
			_trailColours = new Vector.<uint>();
			for (var i:uint = 1; i < _amountOfColors; i++) {			
				_trailColours.push(jsonParsed.data.color[i]);
			}
		}
		
		private function updateData(traildata:ConnectionEvent):void {
			trace("NEWDATA");
			var jsonParsed:Object = JSON.decode(traildata.data);
			var _x:Number = jsonParsed.Filtered.Xf;
			var _y:Number = jsonParsed.Filtered.Yf;
			var tmp: Sprite =generateImg(_x*this.stage.stageWidth, _y*this.stage.stageHeight);
			stage.addChild(tmp);
			if (_amountOfColors > 1) {
				tmp = trail(tmp);
			}
			if (first == 1 ) {
				//trace((current - 1 + 2) % 2);
				stage.removeChildAt((current - 1 + 2) % 2);
			}
			stage.addChild(tmp);
			stage.setChildIndex(tmp, current);
			current = (current+1)%2;
			first = 1;
			trace(_x * this.stage.stageHeight + "," + _y * this.stage.stageWidth);
		}
		
		/**
		 * Genrates a new circle according to the current config.
		 * @param	x
		 * @param	y
		 * @return
		 */
		private function generateImg(x:uint = 0, y:uint = 0) : Sprite {
			var s:Sprite = new Sprite();
			s.graphics.lineStyle(_thickness, _colorOfIcon);
			s.graphics.drawCircle(x, y, _radius);
			s.alpha = _alpha;
			return s;
		}
		
		/**
		 * If trail is pressent this function should be run when new trail data 
		 * arrives 
		 * @param	sprite
		 * @return
		 */
		private function trail(sprite:Sprite): Sprite {
			return sprite;
		}	
		
		
		private function addDebuggText(text:String = "", y:uint = 0) : TextField{		
			var label:TextField = new TextField(); 
			label.width = 4000; 
			label.text = text;
			label.textColor = 0x00FFFF;
			label.x = 0; 
			label.y = y;
			return label; 

		}	
		
		
		
		public function setConfiguration(_color:String):void {
			_colorOfIcon = uint(_color);
		}
		
		
	}
	
}