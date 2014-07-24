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
		
		
	}
	
}