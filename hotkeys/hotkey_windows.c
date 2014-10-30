#include "hotkey_windows.h"

int ES_RegisterHotKey(unsigned int modifiers, unsigned int key) {
	static int id = 1;
	if (RegisterHotKey(NULL, id, modifiers, key)) {
		return id;
	}
	return 0;
}

int ES_GetMessage(MSG * msg) {
	//Param1: The message to be filled.
	//Param2: Do not link to a particular window.
	//Param3: Message should be at least a Hotkey event
	//Param4: Message should be at most a Hotkey event
	//Param5: Remove message afterwards.
	// return PeekMessage(msg, NULL, 0, 0, 1);

	//Same as above but blocking
	return GetMessage(msg, NULL, 0, 0);
}
