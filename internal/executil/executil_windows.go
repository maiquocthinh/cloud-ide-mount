//go:build windows

package executil

import "syscall"

// SysProcAttr returns platform-specific SysProcAttr.
// On Windows, hides the console window for background processes.
func SysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}
