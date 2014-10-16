#ifdef WINDOWS
#include <windows.h>
int ES_RegisterHotKey(unsigned int modifiers, unsigned int key);
int ES_GetMessage(MSG * msg);
#endif

#ifdef LINUX
int main();
#endif
