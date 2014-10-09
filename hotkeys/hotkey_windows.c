#include "hotkey.h"

int ES_RegisterHotKey(unsigned int modifiers, unsigned int key) {
	static int id = 1;
	if (RegisterHotKey(NULL, id, modifiers, key)) {
		return id;
	}
	return 0;
}

int ES_GetMessage(MSG * msg) {
	return GetMessage(msg, NULL, 0, 0);
}

