package syscall

import (
	"os"
	"os/signal"
	"syscall"
)

// OnStop calling a function if you send a system event stop
func OnStop(callFunc func()) {
	quit := make(chan os.Signal, 4)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-quit

	callFunc()
}

// OnUp calling a function if you send a system event SIGHUP
func OnUp(callFunc func()) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP)
	<-quit

	callFunc()
}

func OnCustom(callFunc func(), sig ...os.Signal) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, sig...)
	<-quit

	callFunc()
}
