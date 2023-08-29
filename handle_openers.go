package main

import (
	"github.com/hillu/go-ntdll"
	"golang.org/x/sys/windows"
)

const (
	CONTROL_CODE_OPEN_PROTECTED_PROCESS = 0x8335003c
)

func OpenProcessHandle(pid int, handleMode string) (*windows.Handle, error) {
	if handleMode == HANDLEMODE_DIRECT {
		hProc, err := DirectOpenProc(pid)
		if err != nil {
			return nil, err
		}
		return hProc, nil
	}
	hProc, err := ProcExpOpenProc(pid)
	if err != nil {
		return nil, err
	}
	return hProc, nil
}

func DirectOpenProc(pid int) (*windows.Handle, error) {
	hProc, err := windows.OpenProcess(ntdll.PROCESS_ALL_ACCESS, false, uint32(pid))
	if err != nil {
		return nil, CreateError(err)
	}
	return &hProc, nil
}

func ProcExpOpenProc(pid int) (*windows.Handle, error) {
	hDriver, err := GetProcExpDriver()
	if err != nil {
		return nil, CreateError(err)
	}
	hProc, err := DriverOpenProcess(*hDriver, pid)
	if err != nil {
		return nil, CreateError(err)
	}
	windows.Close(*hDriver)
	return hProc, nil
}
