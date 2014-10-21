#include <iostream>
#include <X11/Xlib.h>
#include <X11/Xutil.h>

#include "hotkey.h"

int main() {
	Display* dpy = XOpenDisplay(0);
	Window root = DefaultRootWindow(dpy);
	XEvent ev;

	unsigned int modifiers = ControlMask;
	int keycode            = XKeysymToKeycode(dpy, XK_Y);
	Window grab_window     = root;
	Bool owner_events      = False;
	int pointer_mode       = GrabModeAsync;
	int keyboard_mode      = GrabModeAsync;

	XGrabKey(dpy, keycode, modifiers, grab_window, owner_events, pointer_mode, keyboard_mode);

	XSelectInput(dpy, root, KeyPressMask);

	bool quit = false;
	while (true) {
		XNextEvent(dpy, &ev);
		switch(ev.type) {
			case KeyPress:
				std::cout << "Hotkey pressed!" << std::endl;
				XUngrabKey(dpy, keycode, modifiers, grab_window);
				quit = true;
			default:
				break;
		}
		if (quit) {
			break;
		}
	}
	XCloseDisplay(dpy);
	return 0;
}
