package process

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitForSignal blocks until SIGTERM or SIGINT is received.
// Returns the received signal.
func WaitForSignal() os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	return <-sigCh
}
