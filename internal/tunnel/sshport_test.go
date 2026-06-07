package tunnel

import (
	"testing"
)

func TestParseSSHPort_Default(t *testing.T) {
	config := `#	$OpenBSD: sshd_config,v 1.104 2021/07/02 11:18:09 dtucker Exp $

# This is the sshd server system-wide configuration file.  See
# sshd_config(5) for more information.

Port 22
`
	port := parseSSHPort(config)
	if port != 22 {
		t.Errorf("parseSSHPort() = %d, want 22", port)
	}
}

func TestParseSSHPort_CustomPort(t *testing.T) {
	config := `Port 2222
`
	port := parseSSHPort(config)
	if port != 2222 {
		t.Errorf("parseSSHPort() = %d, want 2222", port)
	}
}

func TestParseSSHPort_CommentedPort(t *testing.T) {
	config := `#Port 2222
Port 22
`
	port := parseSSHPort(config)
	if port != 22 {
		t.Errorf("parseSSHPort() = %d, want 22 (skip commented line)", port)
	}
}

func TestParseSSHPort_OnlyCommentedPort(t *testing.T) {
	config := `#Port 2222
#Port 22
`
	port := parseSSHPort(config)
	if port != 0 {
		t.Errorf("parseSSHPort() = %d, want 0 (no uncommented Port)", port)
	}
}

func TestParseSSHPort_EmptyConfig(t *testing.T) {
	port := parseSSHPort("")
	if port != 0 {
		t.Errorf("parseSSHPort('') = %d, want 0", port)
	}
}

func TestParseSSHPort_InvalidPort(t *testing.T) {
	config := `Port abc
`
	port := parseSSHPort(config)
	if port != 0 {
		t.Errorf("parseSSHPort() = %d, want 0 (invalid port)", port)
	}
}

func TestParseSSHPort_OutOfRangePort(t *testing.T) {
	config := `Port 99999
`
	port := parseSSHPort(config)
	if port != 0 {
		t.Errorf("parseSSHPort() = %d, want 0 (out of range)", port)
	}
}

func TestParseSSHPort_MultiplePorts(t *testing.T) {
	config := `Port 2222
Port 2223
`
	port := parseSSHPort(config)
	if port != 2222 {
		t.Errorf("parseSSHPort() = %d, want 2222 (first valid port)", port)
	}
}

func TestParseSSHPort_CaseInsensitive(t *testing.T) {
	config := `PORT 2222
`
	port := parseSSHPort(config)
	if port != 2222 {
		t.Errorf("parseSSHPort() = %d, want 2222 (case insensitive)", port)
	}
}

func TestParseSSHPort_ExtraWhitespace(t *testing.T) {
	config := `   Port   2222
`
	port := parseSSHPort(config)
	if port != 2222 {
		t.Errorf("parseSSHPort() = %d, want 2222 (extra whitespace)", port)
	}
}

func TestDetectSSHPort_SuccessFirstTry(t *testing.T) {
	saved := execSSHCommand
	defer func() { execSSHCommand = saved }()

	execSSHCommand = func(name string, args ...string) (string, error) {
		return "Port 2222\n", nil
	}

	port := DetectSSHPort("test-cs")
	if port != 2222 {
		t.Errorf("DetectSSHPort() = %d, want 2222", port)
	}
}

func TestDetectSSHPort_RetryThenSuccess(t *testing.T) {
	saved := execSSHCommand
	defer func() { execSSHCommand = saved }()

	attempts := 0
	execSSHCommand = func(name string, args ...string) (string, error) {
		attempts++
		if attempts <= 1 {
			return "", &mockExecError{msg: "first attempt failed"}
		}
		return "Port 2222\n", nil
	}

	port := DetectSSHPort("test-cs")
	if port != 2222 {
		t.Errorf("DetectSSHPort() = %d, want 2222", port)
	}
	if attempts < 2 {
		t.Errorf("expected at least 2 attempts, got %d", attempts)
	}
}

func TestDetectSSHPort_AllFail(t *testing.T) {
	saved := execSSHCommand
	defer func() { execSSHCommand = saved }()

	execSSHCommand = func(name string, args ...string) (string, error) {
		return "", &mockExecError{msg: "always fail"}
	}

	port := DetectSSHPort("test-cs")
	if port != 22 {
		t.Errorf("DetectSSHPort() = %d, want 22 (default on all failures)", port)
	}
}

func TestDetectSSHPort_FallbackSudo(t *testing.T) {
	saved := execSSHCommand
	defer func() { execSSHCommand = saved }()

	attempts := 0
	execSSHCommand = func(name string, args ...string) (string, error) {
		attempts++
		// First call (cat) fails, second call (sudo cat) succeeds
		if attempts == 1 {
			return "", &mockExecError{msg: "cat permission denied"}
		}
		return "Port 2222\n", nil
	}

	port := DetectSSHPort("test-cs")
	if port != 2222 {
		t.Errorf("DetectSSHPort() = %d, want 2222 (sudo fallback)", port)
	}
}

// mockExecError implements the error interface for testing.
type mockExecError struct {
	msg string
}

func (e *mockExecError) Error() string {
	return e.msg
}
