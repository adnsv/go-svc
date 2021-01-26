package svcapp

import (
	"os"
	"os/signal"
	"syscall"
)

// MonitorSigterm creates a sigterm channel for stopping a service
func MonitorSigterm() chan bool {
	interrupt := make(chan os.Signal, 1)
	output := make(chan bool, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-interrupt
		if OnSigterm != nil {
			OnSigterm(sig)
		}
		output <- true
	}()
	return output
}

// OnSigterm is a callback envoked when the application that runs a service needs to stop
var OnSigterm func(signal os.Signal)
