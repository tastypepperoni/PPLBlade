package main

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	ErrServiceStartPending = "SERVICE PENDING"
)

func SetUpService(serviceName string, driverFullPath string) error {
	connSCM, err := mgr.Connect()
	service := CheckService(*connSCM, serviceName)
	if service == nil {
		service, err = CreateService(*connSCM, serviceName, driverFullPath)
		if err != nil {
			return CreateError(err)
		}
	}
	if !VerifyServiceConfig(service, driverFullPath) {
		connSCM.Disconnect()
		service.Close()
		if err = RemoveService(serviceName, driverFullPath); err != nil {
			return CreateError(err)
		}
		return SetUpService(serviceName, driverFullPath)
	}
	if err = service.Start(); err != nil {
		return CreateError(err)
	}
	return nil
}

func CheckService(connSCM mgr.Mgr, serviceName string) *mgr.Service {
	if service, err := connSCM.OpenService(serviceName); err == nil {
		return service
	}
	return nil
}

func CreateService(connSCM mgr.Mgr, serviceName string, driverPath string) (*mgr.Service, error) {
	serviceConfig := mgr.Config{
		ServiceType:  windows.SERVICE_KERNEL_DRIVER,
		StartType:    windows.SERVICE_DEMAND_START,
		ErrorControl: windows.SERVICE_ERROR_IGNORE,
	}
	service, err := CreateServiceImported(&connSCM, serviceName, driverPath, serviceConfig)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func VerifyServiceConfig(service *mgr.Service, driverPath string) bool {
	serviceConfig, err := service.Config()
	if err != nil {
		return false
	}
	if serviceConfig.ServiceType != windows.SERVICE_KERNEL_DRIVER {
		return false
	}
	if serviceConfig.ErrorControl != windows.SERVICE_ERROR_IGNORE {
		return false
	}
	if serviceConfig.BinaryPathName != fmt.Sprintf("\\??\\%s", driverPath) {
		return false
	}
	return true
}

func VerifyServiceRunning(serviceName string) error {
	connSCM, err := mgr.Connect()
	if err != nil {
		return CreateError(err)
	}
	service, err := connSCM.OpenService(serviceName)
	if err != nil {
		return CreateError(err)
	}
	if serviceStatus, _ := service.Query(); serviceStatus.State == windows.SERVICE_START_PENDING {
		return errors.New(ErrServiceStartPending)
	} else if serviceStatus.State != windows.SERVICE_RUNNING {
		return CreateError(errors.New("service was not started correctly"))
	}
	return nil
}

func RemoveService(serviceName string, driverFullPath string) error {
	connSCM, err := mgr.Connect()
	if err != nil {
		return CreateError(err)
	}
	service, err := connSCM.OpenService(serviceName)
	if err != nil {
		return CreateError(err)
	}
	if !VerifyServiceConfig(service, driverFullPath) {
		return CreateError(errors.New("invalid service"))
	}
	if _, err = service.Control(svc.Stop); err != nil {
		return CreateError(err)
	}
	return CreateError(service.Delete())
}

func CreateServiceImported(m *mgr.Mgr, name, exepath string, c mgr.Config, args ...string) (*mgr.Service, error) {
	if c.StartType == 0 {
		c.StartType = mgr.StartManual
	}
	if c.ServiceType == 0 {
		c.ServiceType = windows.SERVICE_WIN32_OWN_PROCESS
	}
	h, err := windows.CreateService(m.Handle, toPtrImported(name), toPtrImported(c.DisplayName),
		windows.SERVICE_ALL_ACCESS, c.ServiceType,
		c.StartType, c.ErrorControl, toPtrImported(exepath), toPtrImported(c.LoadOrderGroup),
		nil, toStringBlockImported(c.Dependencies), toPtrImported(c.ServiceStartName), toPtrImported(c.Password))
	if err != nil {
		return nil, err
	}
	if c.SidType != windows.SERVICE_SID_TYPE_NONE {
		err = updateSidTypeImported(h, c.SidType)
		if err != nil {
			windows.DeleteService(h)
			windows.CloseServiceHandle(h)
			return nil, err
		}
	}
	if c.Description != "" {
		err = updateDescriptionImported(h, c.Description)
		if err != nil {
			windows.DeleteService(h)
			windows.CloseServiceHandle(h)
			return nil, err
		}
	}
	if c.DelayedAutoStart {
		err = updateStartUpImported(h, c.DelayedAutoStart)
		if err != nil {
			windows.DeleteService(h)
			windows.CloseServiceHandle(h)
			return nil, err
		}
	}
	return &mgr.Service{Name: name, Handle: h}, nil
}

func toPtrImported(s string) *uint16 {
	if len(s) == 0 {
		return nil
	}
	return syscall.StringToUTF16Ptr(s)
}

func toStringBlockImported(ss []string) *uint16 {
	if len(ss) == 0 {
		return nil
	}
	t := ""
	for _, s := range ss {
		if s != "" {
			t += s + "\x00"
		}
	}
	if t == "" {
		return nil
	}
	t += "\x00"
	return &utf16.Encode([]rune(t))[0]
}

func updateSidTypeImported(handle windows.Handle, sidType uint32) error {
	return windows.ChangeServiceConfig2(handle, windows.SERVICE_CONFIG_SERVICE_SID_INFO, (*byte)(unsafe.Pointer(&sidType)))
}

func updateDescriptionImported(handle windows.Handle, desc string) error {
	d := windows.SERVICE_DESCRIPTION{Description: toPtrImported(desc)}
	return windows.ChangeServiceConfig2(handle,
		windows.SERVICE_CONFIG_DESCRIPTION, (*byte)(unsafe.Pointer(&d)))
}

func updateStartUpImported(handle windows.Handle, isDelayed bool) error {
	var d windows.SERVICE_DELAYED_AUTO_START_INFO
	if isDelayed {
		d.IsDelayedAutoStartUp = 1
	}
	return windows.ChangeServiceConfig2(handle,
		windows.SERVICE_CONFIG_DELAYED_AUTO_START_INFO, (*byte)(unsafe.Pointer(&d)))
}
