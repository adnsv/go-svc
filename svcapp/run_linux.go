package svcapp

import (
	"os"
	"os/signal"
	"syscall"
)

func Run(name string) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-interrupt
	return nil
}
