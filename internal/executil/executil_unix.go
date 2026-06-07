//go:build !windows

package executil

import "syscall"

// SysProcAttr returns platform-specific SysProcAttr.
// On Unix platforms, no special attributes are needed.
func SysProcAttr() *syscall.SysProcAttr {
	return nil
}
