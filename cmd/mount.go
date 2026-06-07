package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/executil"
	"cloud-ide-mount/internal/logging"
	"cloud-ide-mount/internal/rclone"
	"cloud-ide-mount/internal/state"
	"cloud-ide-mount/internal/tunnel"
	"cloud-ide-mount/internal/ui"

	"github.com/spf13/cobra"
)

var startFlag bool

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Interactive: pick codespaces, mode, drive(s)",
	RunE:  mountRunE,
}

func mountRunE(_ *cobra.Command, _ []string) error {
	if err := checkDeps(); err != nil {
		return err
	}
	if _, err := os.Stat(KeyFile); err != nil {
		return fmt.Errorf("key file not found: %s", KeyFile)
	}

	all, err := codespace.List()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return fmt.Errorf("no codespaces found")
	}
	ui.ShowCsList(all)

	selected := ui.ReadSelection(all, "")

	if startFlag {
		if err := startCodespaces(selected); err != nil {
			return err
		}
	}

	available := filterAvailable(selected)
	if len(available) == 0 {
		return fmt.Errorf("no available codespaces in selection")
	}

	mode := ui.ReadMountMode(len(available))
	assignments := ui.ReadDriveAssignments(available, mode)

	if !Force && !showPlan(mode, assignments, available) {
		return nil
	}

	fmt.Println()

	s, err := state.Load()
	if err != nil {
		return err
	}
	if s == nil {
		s = &state.State{}
	}
	initState(s)

	csPathMap := codespace.BuildUpstreamPaths(available)

	upstreams, err := orchestrateTunnels(available, s, StartPort, KeyFile, csPathMap)
	if err != nil {
		return err
	}
	if len(upstreams) == 0 {
		return fmt.Errorf("no remotes created")
	}

	if err := mountDrives(mode, assignments, upstreams, s); err != nil {
		return err
	}

	if err := state.Save(s); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}
	fmt.Println()
	logging.Info("Done. Run status to verify.")
	return nil
}

func init() {
	rootCmd.AddCommand(mountCmd)
	mountCmd.Flags().BoolVarP(&startFlag, "start", "s", false, "Start codespace(s) before mounting")
}

// ─── Helper: codespace management ──────────────────────────────────

// startCodespaces starts codespaces that are not yet available and waits for them.
func startCodespaces(selected []codespace.Codespace) error {
	for _, cs := range selected {
		if cs.State == "Available" {
			continue
		}
		logging.Info(fmt.Sprintf("Starting codespace: %s...", cs.Name), "codespace", cs.Name)
		if _, err := execCmdOutput("gh", "cs", "start", "-c", cs.Name); err != nil {
			return fmt.Errorf("starting codespace %s: %w", cs.Name, err)
		}
		for {
			out, err := execCmdOutput("gh", "cs", "view", "-c", cs.Name, "--json", "state")
			if err == nil {
				stateStr := strings.TrimSpace(out)
				stateStr = strings.Trim(stateStr, `"`)
				if stateStr == "Available" {
					logging.Info(fmt.Sprintf("%s is now Available.", cs.Name), "codespace", cs.Name)
					break
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
	return nil
}

// filterAvailable returns only codespaces in "Available" state.
func filterAvailable(selected []codespace.Codespace) []codespace.Codespace {
	var available []codespace.Codespace
	for _, cs := range selected {
		if cs.State == "Available" {
			available = append(available, cs)
		} else {
			logging.Warn(fmt.Sprintf("Skipping %s (%s)", cs.Name, cs.State), "codespace", cs.Name, "state", cs.State)
		}
	}
	return available
}

// showPlan displays the mount plan and asks for user confirmation.
// Returns false if the user cancels.
func showPlan(mode string, assignments []ui.DriveAssignment, available []codespace.Codespace) bool {
	fmt.Println()
	fmt.Println("  Ready to mount:")
	if mode == "combined" {
		di := assignments[0]
		action := fmt.Sprintf("new → %s", di.Drive)
		if di.Extend {
			action = fmt.Sprintf("extend %s", di.Drive)
		}
		fmt.Printf("    Mode : combined (%s)\n", action)
		paths := codespace.BuildUpstreamPaths(available)
		for _, p := range paths {
			fmt.Printf("      %s\\%s\n", di.Drive, p.FolderPath)
		}
	} else {
		fmt.Println("    Mode : separate")
		for _, a := range assignments {
			fmt.Printf("      %s → %s\n", a.Codespace, a.Drive)
		}
	}
	fmt.Println()
	return ui.Confirm("Proceed? [y/N]")
}

// initState initializes nil fields in a state object.
func initState(s *state.State) {
	if s.Remotes == nil {
		s.Remotes = []state.Remote{}
	}
	if s.Mounts == nil {
		s.Mounts = []state.Mount{}
	}
}

// ─── Tunnel orchestration ──────────────────────────────────────────

// orchestrateTunnels sets up SSH tunnels for each available codespace that doesn't
// already have an active tunnel, creates rclone SFTP remotes, and returns the list of upstreams.
func orchestrateTunnels(available []codespace.Codespace, s *state.State, startPort int, keyFile string, csPathMap []codespace.UpstreamPath) ([]rclone.Upstream, error) {
	port := startPort
	var upstreams []rclone.Upstream

	for _, cs := range available {
		name := cs.Name
		safeName := codespace.SafeName(name)
		remoteName := "cs-" + safeName

		var existingRemote *state.Remote
		for i := range s.Remotes {
			if s.Remotes[i].Codespace == name && tunnel.PortOpen(s.Remotes[i].Port) {
				existingRemote = &s.Remotes[i]
				break
			}
		}
		if existingRemote != nil {
			logging.Info(fmt.Sprintf("Tunnel already active: %s (port %d)", name, existingRemote.Port), "codespace", name, "port", existingRemote.Port)
			folderPath := findFolderPath(csPathMap, name)
			upstreams = append(upstreams, rclone.Upstream{FolderPath: folderPath, Remote: remoteName})
			continue
		}

		ap, err := tunnel.AllocatePort(port)
		if err != nil {
			return nil, fmt.Errorf("allocating port for %s: %w", name, err)
		}
		allocPort := ap.Port

		logging.Info(fmt.Sprintf("Starting sshd: %s...", name), "codespace", name)
		if _, err := execCmdOutput("gh", "cs", "ssh", "-c", name, "--", "sudo", "service", "ssh", "start"); err != nil {
			logging.Warn(fmt.Sprintf("sshd start may have failed: %v (continuing)", err), "codespace", name, "error", err)
		}

		sshPort := tunnel.DetectSSHPort(name)
		logging.Info(fmt.Sprintf("SSH port: %d", sshPort), "codespace", name, "sshPort", sshPort)

		logging.Info(fmt.Sprintf("Starting tunnel: %s -> local %d", name, allocPort), "codespace", name, "port", allocPort)
		if err := ap.Close(); err != nil {
			logging.Warn(fmt.Sprintf("error releasing temporary port: %v", err), "port", allocPort, "error", err)
		}
		tunCmd, err := tunnel.StartTunnel(name, allocPort, sshPort)
		if err != nil {
			return nil, fmt.Errorf("tunnel start failed for %s: %w", name, err)
		}

		ready := tunnel.WaitPort(allocPort, 30*time.Second)
		if !ready {
			logging.Error(fmt.Sprintf("Tunnel failed: %s", name), "codespace", name)
			if tunCmd.Process != nil {
				if err := tunCmd.Process.Kill(); err != nil {
					logging.Warn(fmt.Sprintf("error killing failed tunnel: %v", err), "error", err)
				}
			}
			if err := ap.Close(); err != nil {
				logging.Warn(fmt.Sprintf("error releasing failed port: %v", err), "error", err)
			}
			port++
			continue
		}

		logging.Info(fmt.Sprintf("Tunnel ready: %s", name), "codespace", name)

		if err := rclone.NewSFTPRemote(remoteName, allocPort, keyFile); err != nil {
			return nil, fmt.Errorf("creating rclone remote for %s: %w", name, err)
		}

		folderPath := findFolderPath(csPathMap, name)
		s.Remotes = append(s.Remotes, state.Remote{
			Name:       remoteName,
			Codespace:  name,
			Port:       allocPort,
			TunnelPid:  tunCmd.Process.Pid,
			FolderPath: folderPath,
		})

		upstreams = append(upstreams, rclone.Upstream{FolderPath: folderPath, Remote: remoteName})

		if err := ap.Close(); err != nil {
			logging.Warn(fmt.Sprintf("error releasing port: %v", err), "port", allocPort, "error", err)
		}
		port = allocPort + 1
	}

	return upstreams, nil
}

// ─── Rclone config building ────────────────────────────────────────

// buildConfig creates the rclone combine remote for the given upstreams.
func buildConfig(remoteName string, upstreams []rclone.Upstream) error {
	return rclone.SetCombineRemote(remoteName, upstreams)
}

// ─── Mount orchestration ───────────────────────────────────────────

// mountDrives performs the mounting for combined or separate mode based on the user's selection.
func mountDrives(mode string, assignments []ui.DriveAssignment, upstreams []rclone.Upstream, s *state.State) error {
	if mode == "combined" {
		return mountCombined(assignments[0], upstreams, s)
	}
	mountSeparate(assignments, upstreams, s)
	return nil
}

// mountCombined mounts all upstreams under a single drive in combined mode.
func mountCombined(di ui.DriveAssignment, upstreams []rclone.Upstream, s *state.State) error {
	drive := di.Drive

	if di.Extend {
		stopExistingMount(drive, s)
	}

	logging.Info("Building combine remote...")
	if err := buildConfig(CombineRemote, upstreams); err != nil {
		return fmt.Errorf("building combine remote: %w", err)
	}

	logging.Info(fmt.Sprintf("Mounting %s -> %s", CombineRemote, drive), "remote", CombineRemote, "drive", drive)
	mp, err := rclone.Mount(CombineRemote, drive, "Codespaces")
	if err != nil {
		return fmt.Errorf("mounting %s: %w", drive, err)
	}

	s.Mounts = append(s.Mounts, state.Mount{
		Drive:     drive,
		RclonePid: mp.Cmd.Process.Pid,
		Remote:    CombineRemote,
		Mode:      "combined",
	})

	if waitForMount(drive, mp) {
		logging.Info(fmt.Sprintf("Mounted: %s", drive), "drive", drive)
	} else {
		logging.Warn(fmt.Sprintf("%s not visible after 15s. Run status to check.", drive), "drive", drive)
	}

	return nil
}

// mountSeparate mounts each codespace to its own drive in separate mode.
func mountSeparate(assignments []ui.DriveAssignment, upstreams []rclone.Upstream, s *state.State) {
	driveGroups := groupByDrive(assignments)

	for drive, csNames := range driveGroups {
		if len(csNames) > 1 {
			mountMultipleOnDrive(drive, csNames, upstreams, s)
		} else {
			mountSingleOnDrive(drive, csNames[0], upstreams, s)
		}
	}
}

// groupByDrive groups drive assignments by drive letter.
func groupByDrive(assignments []ui.DriveAssignment) map[string][]string {
	driveGroups := map[string][]string{}
	for _, a := range assignments {
		driveGroups[a.Drive] = append(driveGroups[a.Drive], a.Codespace)
	}
	return driveGroups
}

// mountMultipleOnDrive mounts multiple codespaces on a shared drive using combine.
func mountMultipleOnDrive(drive string, csNames []string, upstreams []rclone.Upstream, s *state.State) {
	var sharedUpstreams []rclone.Upstream
	for _, cn := range csNames {
		for _, u := range upstreams {
			if u.Remote == "cs-"+codespace.SafeName(cn) {
				sharedUpstreams = append(sharedUpstreams, u)
				break
			}
		}
	}
	if len(sharedUpstreams) == 0 {
		return
	}

	stopExistingMount(drive, s)

	logging.Info(fmt.Sprintf("Building combine remote for %s...", drive), "drive", drive)
	if err := buildConfig(CombineRemote, sharedUpstreams); err != nil {
		logging.Error(fmt.Sprintf("Combine remote error: %v (skipping %s)", err, drive), "error", err, "drive", drive)
		return
	}

	mp, err := rclone.Mount(CombineRemote, drive, "Codespaces")
	if err != nil {
		logging.Error(fmt.Sprintf("Mount error for %s: %v", drive, err), "drive", drive, "error", err)
		return
	}
	s.Mounts = append(s.Mounts, state.Mount{
		Drive:     drive,
		RclonePid: mp.Cmd.Process.Pid,
		Remote:    CombineRemote,
		Mode:      "combined",
	})

	if waitForMount(drive, mp) {
		logging.Info(fmt.Sprintf("Mounted: %s (combined)", drive), "drive", drive)
	} else {
		logging.Warn(fmt.Sprintf("%s not visible after 15s.", drive), "drive", drive)
	}
}

// mountSingleOnDrive mounts a single codespace on its dedicated drive.
func mountSingleOnDrive(drive, csName string, upstreams []rclone.Upstream, s *state.State) {
	var allUpstreams []rclone.Upstream

	for _, u := range upstreams {
		if u.Remote == "cs-"+codespace.SafeName(csName) {
			allUpstreams = append(allUpstreams, u)
			break
		}
	}

	for _, m := range s.Mounts {
		if m.Drive == drive && m.Codespace != "" && m.Codespace != csName {
			for _, r := range s.Remotes {
				if r.Codespace == m.Codespace {
					fp := r.FolderPath
					if fp == "" {
						fp = codespace.SafeName(m.Codespace)
					}
					allUpstreams = append(allUpstreams, rclone.Upstream{
						FolderPath: fp,
						Remote:     r.Name,
					})
					break
				}
			}
		}
	}

	if len(allUpstreams) == 0 {
		logging.Warn(fmt.Sprintf("No remote for %s, skipping.", csName), "codespace", csName)
		return
	}

	stopExistingMount(drive, s)

	logging.Info(fmt.Sprintf("Building combine remote for %s...", drive), "drive", drive)
	if err := buildConfig(CombineRemote, allUpstreams); err != nil {
		logging.Error(fmt.Sprintf("Combine remote error: %v (skipping %s)", err, drive), "error", err, "drive", drive)
		return
	}

	logging.Info(fmt.Sprintf("Mounting %s -> %s", csName, drive), "codespace", csName, "drive", drive)
	volname := strings.TrimRight(drive, ":")
	mp, err := rclone.Mount(CombineRemote, drive, volname)
	if err != nil {
		logging.Error(fmt.Sprintf("Mount error for %s: %v", drive, err), "drive", drive, "error", err)
		return
	}

	mode := "separate"
	if len(allUpstreams) > 1 {
		mode = "combined"
	}
	s.Mounts = append(s.Mounts, state.Mount{
		Drive:     drive,
		RclonePid: mp.Cmd.Process.Pid,
		Remote:    CombineRemote,
		Codespace: csName,
		Mode:      mode,
	})

	if waitForMount(drive, mp) {
		logging.Info(fmt.Sprintf("Mounted: %s (%s)", drive, csName), "drive", drive, "codespace", csName)
	} else {
		logging.Warn(fmt.Sprintf("rclone running but %s not visible yet.", drive), "drive", drive)
	}
}

// stopExistingMount kills the rclone process for a drive and removes it from state.
func stopExistingMount(drive string, s *state.State) {
	for i, m := range s.Mounts {
		if m.Drive == drive {
			logging.Info(fmt.Sprintf("Stopping rclone on %s to rebuild combine...", drive), "drive", drive)
			if err := executil.KillProcess(m.RclonePid, 5*time.Second); err != nil {
				logging.Warn(fmt.Sprintf("error stopping rclone PID %d: %v", m.RclonePid, err), "pid", m.RclonePid, "error", err)
			}
			s.Mounts = append(s.Mounts[:i], s.Mounts[i+1:]...)
			break
		}
	}
}

// ─── Existing helpers ──────────────────────────────────────────────

func checkDeps() error {
	if _, err := execLook("gh"); err != nil {
		return fmt.Errorf("gh not found: %w", err)
	}
	if _, err := execLook("rclone"); err != nil {
		return fmt.Errorf("rclone not found: %w", err)
	}
	return nil
}

var execLook = func(name string) (string, error) {
	return exec.LookPath(name)
}

//nolint:unparam
func execCmdOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func findFolderPath(paths []codespace.UpstreamPath, csName string) string {
	for _, p := range paths {
		if p.Cs.Name == csName {
			return p.FolderPath
		}
	}
	return codespace.SafeName(csName)
}

func waitForMount(drive string, mp *rclone.MountProcess) bool {
	for i := 0; i < 15; i++ {
		select {
		case err := <-mp.Done:
			var stderr string
			select {
			case s := <-mp.Stderr:
				stderr = s
			default:
			}
			if stderr != "" {
				logging.Warn(fmt.Sprintf("rclone exited: %v (stderr: %s)", err, stderr), "error", err)
			} else if err != nil {
				logging.Warn(fmt.Sprintf("rclone exited: %v", err), "error", err)
			}
			return false
		default:
		}
		time.Sleep(time.Second)
		if _, err := os.Stat(drive + "\\"); err == nil {
			return true
		}
	}
	select {
	case err := <-mp.Done:
		var stderr string
		select {
		case s := <-mp.Stderr:
			stderr = s
		default:
		}
		if stderr != "" {
			logging.Warn(fmt.Sprintf("rclone exited: %v (stderr: %s)", err, stderr), "error", err)
		} else if err != nil {
			logging.Warn(fmt.Sprintf("rclone exited: %v", err), "error", err)
		}
	default:
	}
	return false
}
