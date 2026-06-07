package rclone

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"cloud-ide-mount/internal/executil"
)

type Upstream struct {
	FolderPath string
	Remote     string
}

type MountProcess struct {
	Cmd    *exec.Cmd
	Done   chan error
	Stderr chan string
}

func createConfig(args ...string) error {
	cmd := exec.Command("rclone", args...)
	return cmd.Run()
}

func DeleteRemote(name string) {
	_ = createConfig("config", "delete", name)
}

func NewSFTPRemote(name string, port int, keyFile string) error {
	DeleteRemote(name)
	return createConfig(
		"config", "create", name, "sftp",
		"host", "127.0.0.1",
		"user", "codespace",
		"port", fmt.Sprintf("%d", port),
		"key_file", keyFile,
		"shell_type", "unix",
	)
}

func NewAliasRemote(name, target string) error {
	DeleteRemote(name)
	return createConfig("config", "create", name, "alias", "remote", target)
}

func SetCombineRemote(name string, newUpstreams []Upstream) error {
	// Create alias for each upstream, group by org
	type orgEntry struct {
		repos []string // "repo=alias:" pairs
	}
	orgs := map[string]*orgEntry{}
	orgOrder := []string{}

	for _, u := range newUpstreams {
		// FolderPath = "org/repo"
		parts := strings.SplitN(u.FolderPath, "/", 2)
		org := parts[0]
		repo := parts[1]

		safePath := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(u.FolderPath, "-")
		aliasName := "cs-alias-" + safePath
		DeleteRemote(aliasName)
		_ = createConfig("config", "create", aliasName, "alias", "remote", u.Remote+":/")

		if _, ok := orgs[org]; !ok {
			orgs[org] = &orgEntry{}
			orgOrder = append(orgOrder, org)
		}
		orgs[org].repos = append(orgs[org].repos, repo+"="+aliasName+":")
	}

	// Create sub-combine for each org
	for _, org := range orgOrder {
		subName := "cs-combine-" + org
		DeleteRemote(subName)
		upstreamsText := strings.Join(orgs[org].repos, " ")
		_ = createConfig("config", "create", subName, "combine", "upstreams", upstreamsText)
	}

	// Create top-level combine
	var topParts []string
	for _, org := range orgOrder {
		topParts = append(topParts, org+"=cs-combine-"+org+":")
	}

	DeleteRemote(name)
	return createConfig("config", "create", name, "combine", "upstreams", strings.Join(topParts, " "))
}

func Mount(remote, drive, volname string) (*MountProcess, error) {
	args := []string{
		"mount", remote + ":", drive,
		"--vfs-cache-mode", "writes",
		"--dir-cache-time", "10s",
		"--poll-interval", "10s",
		"--volname", volname,
	}
	cmd := exec.Command("rclone", args...)
	cmd.SysProcAttr = executil.SysProcAttr()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mounting %s -> %s: %w", remote, drive, err)
	}

	mp := &MountProcess{
		Cmd:    cmd,
		Done:   make(chan error, 1),
		Stderr: make(chan string, 1),
	}

	go func() {
		errBuf, _ := io.ReadAll(stderr)
		waitErr := cmd.Wait()
		if waitErr != nil || len(errBuf) > 0 {
			mp.Stderr <- strings.TrimSpace(string(errBuf))
		}
		mp.Done <- waitErr
	}()

	return mp, nil
}
