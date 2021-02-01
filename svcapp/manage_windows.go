package svcapp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func IsInteractive() (bool, error) {
	is, err := svc.IsAnInteractiveSession()
	if err != nil {
		return false, err
	}
	return !is, nil
}

func Install(params InstallParams) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(params.Name)
	if err == nil {
		s.Close()
		return errors.New("service is already installed")
	}

	args := strings.Split(params.Args, " ")
	cfg := mgr.Config{
		DisplayName: params.DispName,
		Description: params.Description,
		StartType:   mgr.StartAutomatic,
	}

	s, err = m.CreateService(params.Name, params.Executable, cfg, args...)
	if err != nil {
		fmt.Println("CreateService failed")
		return err
	}
	defer s.Close()

	// set recovery action for service
	// restart after 5 seconds for the first 3 times
	// restart after 1 minute, otherwise
	r := []mgr.RecoveryAction{
		mgr.RecoveryAction{
			Type:  mgr.ServiceRestart,
			Delay: 5000 * time.Millisecond,
		},
		mgr.RecoveryAction{
			Type:  mgr.ServiceRestart,
			Delay: 5000 * time.Millisecond,
		},
		mgr.RecoveryAction{
			Type:  mgr.ServiceRestart,
			Delay: 5000 * time.Millisecond,
		},
		mgr.RecoveryAction{
			Type:  mgr.ServiceRestart,
			Delay: 60000 * time.Millisecond,
		},
	}
	// set reset period as a day
	s.SetRecoveryActions(r, uint32(86400))

	return nil
}

func Start(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	err = s.Start()
	return err
}

func Stop(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}
	timeDuration := time.Millisecond * 50

	timeout := time.After(getStopTimeout() + (timeDuration * 2))
	tick := time.NewTicker(timeDuration)
	defer tick.Stop()

	for status.State != svc.Stopped {
		select {
		case <-tick.C:
			status, err = s.Query()
			if err != nil {
				return err
			}
		case <-timeout:
			break
		}
	}
	return nil
}

func Status(name string) (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		if syserr, ok := err.(syscall.Errno); ok && uintptr(syserr) == 1060 {
			return "uninstalled", nil
		}
		return "", err
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return "", err
	}
	switch status.State {
	case svc.Stopped:
		return "stopped", nil
	case svc.StartPending:
		return "start pending", nil
	case svc.StopPending:
		return "stop pending", nil
	case svc.Running:
		return "running", nil
	case svc.ContinuePending:
		return "continue pending", nil
	case svc.PausePending:
		return "pause pending", nil
	case svc.Paused:
		return "paused", nil
	default:
		return "unknown", nil
	}
}

func Uninstall(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	return nil
}

func getStopTimeout() time.Duration {
	// For default and paths see https://support.microsoft.com/en-us/kb/146092
	defaultTimeout := time.Millisecond * 20000
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control`, registry.READ)
	if err != nil {
		return defaultTimeout
	}
	sv, _, err := key.GetStringValue("WaitToKillServiceTimeout")
	if err != nil {
		return defaultTimeout
	}
	v, err := strconv.Atoi(sv)
	if err != nil {
		return defaultTimeout
	}
	return time.Millisecond * time.Duration(v)
}
