package ide

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type IDE struct {
	Name    string
	Command string
}

type SSHInfo struct {
	Host       string
	Port       int
	User       string
	KeyFile    string
	Alias      string
	RemotePath string
}

var All = []IDE{
	{Name: "VS Code", Command: "code"},
	{Name: "IntelliJ IDEA", Command: "idea64"},
	{Name: "Zed", Command: "zed"},
}

func Available() []IDE {
	var avail []IDE
	for _, ide := range All {
		if _, err := exec.LookPath(ide.Command); err == nil {
			avail = append(avail, ide)
		}
	}
	return avail
}

func ensureSSHConfig(info SSHInfo) error {
	sshDir := filepath.Join(os.Getenv("USERPROFILE"), ".ssh")
	configPath := filepath.Join(sshDir, "config")
	os.MkdirAll(sshDir, 0700)

	data, _ := os.ReadFile(configPath)
	if strings.Contains(string(data), "Host "+info.Alias) {
		return nil
	}

	entry := fmt.Sprintf("\nHost %s\n  HostName %s\n  Port %d\n  User %s\n  IdentityFile %s\n  StrictHostKeyChecking no\n  UserKnownHostsFile NUL\n",
		info.Alias, info.Host, info.Port, info.User, info.KeyFile)

	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(entry)
	return err
}

func Open(ide IDE, info SSHInfo) error {
	switch ide.Command {
	case "code":
		return openVSCode(info)
	case "zed":
		return openZed(info)
	case "idea64":
		return openIntelliJ(info)
	default:
		return fmt.Errorf("unknown IDE: %s", ide.Name)
	}
}

func openVSCode(info SSHInfo) error {
	if err := ensureSSHConfig(info); err != nil {
		return fmt.Errorf("SSH config: %w", err)
	}
	uri := fmt.Sprintf("vscode-remote://ssh-remote+%s%s", info.Alias, info.RemotePath)
	return exec.Command("code", "--folder-uri", uri).Start()
}

func openZed(info SSHInfo) error {
	if err := ensureSSHConfig(info); err != nil {
		return fmt.Errorf("SSH config: %w", err)
	}
	addr := fmt.Sprintf("ssh://%s%s", info.Alias, info.RemotePath)
	return exec.Command("zed", addr).Start()
}

func openIntelliJ(info SSHInfo) error {
	addr := fmt.Sprintf("ssh://%s@%s:%d", info.User, info.Host, info.Port)
	return exec.Command("idea64", "remote-dev", addr).Start()
}
