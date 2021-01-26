package svcapp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func servicePath(name string) string {
	return "/etc/systemd/system/" + name + ".service"
}

func isInstalled(name string) bool {
	if _, err := os.Stat(servicePath(name)); err == nil {
		return true
	}
	return false
}

func isRunning(name string) bool {
	output, err := exec.Command("systemctl", "--lines=0", "status", name+".service").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Active: active")
}

var ErrAlreadyInstalled = errors.New("service is already installed")
var ErrNotInstalled = errors.New("service is not installed")
var ErrInsufficientPrivileges = errors.New("insufficient privileges")
var ErrUnsupportedSystem = errors.New("unsupported system")
var ErrInvalidSystemResponse = errors.New("invalid system response")
var ErrAlreadyRunning = errors.New("service is already ranning")
var ErrAlreadyStopped = errors.New("service had already been stopped")

func Install(params InstallParams) error {
	if isInstalled(params.Name) {
		return ErrAlreadyInstalled
	}
	if ok, err := checkPrivileges(); !ok {
		return err
	}
	f, err := os.Create(servicePath(params.Name))
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	cmd := params.Executable
	if params.Args != "" {
		cmd += " " + params.Args
	}

	fmt.Fprintf(f, "[Unit]\n")
	fmt.Fprintf(f, "Description=%s\n\n", params.Description)
	fmt.Fprintf(f, "[Service]\n")
	fmt.Fprintf(f, "ExecStart=%s\n", cmd)
	fmt.Fprintf(f, "Restart=on-failure\n\n")
	fmt.Fprintf(f, "[Install]\n")
	fmt.Fprintf(f, "WantedBy=multi-user.target\n")
	f.Close()
	f = nil

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return err
	}

	if err := exec.Command("systemctl", "enable", params.Name+".service").Run(); err != nil {
		return err
	}
	return nil
}

func Start(name string) error {
	if !isInstalled(name) {
		return ErrNotInstalled
	}
	if ok, err := checkPrivileges(); !ok {
		return err
	}
	if isRunning(name) {
		return ErrAlreadyRunning
	}
	if err := exec.Command("systemctl", "start", name+".service").Run(); err != nil {
		return err
	}
	return nil
}

func Stop(name string) error {
	if !isInstalled(name) {
		return ErrNotInstalled
	}
	if ok, err := checkPrivileges(); !ok {
		return err
	}
	if !isRunning(name) {
		return ErrAlreadyStopped
	}
	if err := exec.Command("systemctl", "stop", name+".service").Run(); err != nil {
		return err
	}
	return nil
}

func Status(name string) (string, error) {
	if !isInstalled(name) {
		return "", ErrNotInstalled
	}
	if ok, err := checkPrivileges(); !ok {
		return "", err
	}
	if isRunning(name) {
		return "running", nil
	}
	return "stopped", nil
}

func Uninstall(name string) error {
	if !isInstalled(name) {
		return ErrNotInstalled
	}
	if ok, err := checkPrivileges(); !ok {
		return err
	}
	if err := exec.Command("systemctl", "disable", name+".service").Run(); err != nil {
		return err
	}
	if err := os.Remove(servicePath(name)); err != nil {
		return err
	}
	return nil
}

func checkPrivileges() (bool, error) {

	if output, err := exec.Command("id", "-g").Output(); err == nil {
		if gid, parseErr := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 32); parseErr == nil {
			if gid == 0 {
				return true, nil
			}
			return false, ErrInsufficientPrivileges
		}
		return false, ErrInvalidSystemResponse
	}
	return false, ErrUnsupportedSystem
}
