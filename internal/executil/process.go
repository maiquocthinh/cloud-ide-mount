package executil

import (
	"fmt"
	"os"
	"time"
)

// KillProcess kills the process with the given PID and waits for it to exit.
// If the process does not exit within the given timeout, an error is returned.
func KillProcess(pid int, timeout time.Duration) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process %d: %w", pid, err)
	}

	if err := p.Kill(); err != nil {
		return fmt.Errorf("killing process %d: %w", pid, err)
	}

	done := make(chan error, 1)
	go func() {
		_, waitErr := p.Wait()
		done <- waitErr
	}()

	select {
	case <-time.After(timeout):
		return fmt.Errorf("process %d did not exit within %v", pid, timeout)
	case err := <-done:
		return err
	}
}
