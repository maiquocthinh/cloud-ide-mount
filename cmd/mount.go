package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"codespaces-mount/internal/codespace"
	"codespaces-mount/internal/rclone"
	"codespaces-mount/internal/state"
	"codespaces-mount/internal/tunnel"
	"codespaces-mount/internal/ui"

	"github.com/spf13/cobra"
)

var startFlag bool

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Interactive: pick codespaces, mode, drive(s)",
	RunE: func(_ *cobra.Command, _ []string) error {
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
			for _, cs := range selected {
				if cs.State == "Available" {
					continue
				}
				fmt.Printf("  Starting codespace: %s...\n", cs.Name)
				execCmdOutput("gh", "cs", "start", "-c", cs.Name)
				for {
					out, err := execCmdOutput("gh", "cs", "view", "-c", cs.Name, "--json", "state")
					if err == nil {
						stateStr := strings.TrimSpace(out)
						stateStr = strings.Trim(stateStr, `"`)
						if stateStr == "Available" {
							fmt.Printf("  %s is now Available.\n", cs.Name)
							break
						}
					}
					time.Sleep(2 * time.Second)
				}
			}
		}

		var available []codespace.Codespace
		for _, cs := range selected {
			if cs.State == "Available" {
				available = append(available, cs)
			} else {
				fmt.Printf("  Skipping %s (%s)\n", cs.Name, cs.State)
			}
		}
		if len(available) == 0 {
			return fmt.Errorf("no available codespaces in selection")
		}

		mode := ui.ReadMountMode(len(available))
		assignments := ui.ReadDriveAssignments(available, mode)

		if !Force {
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
			if !ui.Confirm("Proceed? [y/N]") {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		fmt.Println()

		s, err := state.Load()
		if err != nil {
			return err
		}
		if s == nil {
			s = &state.State{}
		}
		if s.Remotes == nil {
			s.Remotes = []state.Remote{}
		}
		if s.Mounts == nil {
			s.Mounts = []state.Mount{}
		}

		port := StartPort
		var upstreams []rclone.Upstream
		csPathMap := codespace.BuildUpstreamPaths(available)

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
				fmt.Printf("  Tunnel already active: %s (port %d)\n", name, existingRemote.Port)
				folderPath := findFolderPath(csPathMap, name)
				upstreams = append(upstreams, rclone.Upstream{FolderPath: folderPath, Remote: remoteName})
				continue
			}

			port = tunnel.NextFreePort(port)

			// Start sshd
			fmt.Printf("  Starting sshd: %s...\n", name)
			execCmdOutput("gh", "cs", "ssh", "-c", name, "--", "sudo", "service", "ssh", "start")

			// Detect SSH port first, then tunnel with correct port
			sshPort, err := detectSSHPort(name)
			if err != nil {
				return fmt.Errorf("SSH port detection failed for %s: %w", name, err)
			}
			fmt.Printf("  SSH port: %d\n", sshPort)

			fmt.Printf("  Starting tunnel: %s → local %d\n", name, port)
			tunCmd, err := tunnel.StartTunnel(name, port, sshPort)
			if err != nil {
				return fmt.Errorf("tunnel start failed for %s: %w", name, err)
			}

			ready := tunnel.WaitPort(port, 30*time.Second)
			if !ready {
				fmt.Printf("  Tunnel failed: %s\n", name)
				if tunCmd.Process != nil {
					tunCmd.Process.Kill()
				}
				port++
				continue
			}

			fmt.Printf("  Tunnel ready: %s\n", name)

			if err := rclone.NewSFTPRemote(remoteName, port, KeyFile); err != nil {
				return fmt.Errorf("creating rclone remote for %s: %w", name, err)
			}

			folderPath := findFolderPath(csPathMap, name)
			s.Remotes = append(s.Remotes, state.Remote{
				Name:       remoteName,
				Codespace:  name,
				Port:       port,
				TunnelPid:  tunCmd.Process.Pid,
				FolderPath: folderPath,
			})

			upstreams = append(upstreams, rclone.Upstream{FolderPath: folderPath, Remote: remoteName})

			port++
		}

		if len(upstreams) == 0 {
			return fmt.Errorf("no remotes created")
		}

		if mode == "combined" {
			di := assignments[0]
			drive := di.Drive

			if di.Extend {
				for i, m := range s.Mounts {
					if m.Drive == drive {
						fmt.Printf("  Stopping rclone on %s to rebuild combine...\n", drive)
						killPid(m.RclonePid)
						time.Sleep(time.Second)
						s.Mounts = append(s.Mounts[:i], s.Mounts[i+1:]...)
						break
					}
				}
			}

			fmt.Println("  Building combine remote...")
			if err := rclone.SetCombineRemote(CombineRemote, upstreams); err != nil {
				return fmt.Errorf("building combine remote: %w", err)
			}

			fmt.Printf("  Mounting %s: → %s\n", CombineRemote, drive)
			mp, err := rclone.Mount(CombineRemote, drive, "Codespaces")
			if err != nil {
				return err
			}

			s.Mounts = append(s.Mounts, state.Mount{
				Drive:     drive,
				RclonePid: mp.Cmd.Process.Pid,
				Remote:    CombineRemote,
				Mode:      "combined",
			})

			if waitForMount(drive, mp) {
				fmt.Printf("  Mounted: %s\n", drive)
			} else {
				fmt.Printf("  %s not visible after 15s. Run status to check.\n", drive)
			}

		} else {
			driveGroups := map[string][]string{}
			for _, a := range assignments {
				driveGroups[a.Drive] = append(driveGroups[a.Drive], a.Codespace)
			}

			for drive, csNames := range driveGroups {
				if len(csNames) > 1 {
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
						continue
					}

					for i, m := range s.Mounts {
						if m.Drive == drive {
							killPid(m.RclonePid)
							time.Sleep(time.Second)
							s.Mounts = append(s.Mounts[:i], s.Mounts[i+1:]...)
							break
						}
					}

					fmt.Printf("  Building combine remote for %s...\n", drive)
					if err := rclone.SetCombineRemote(CombineRemote, sharedUpstreams); err != nil {
						fmt.Printf("  Combine remote error: %v\n", err)
					}

					mp, err := rclone.Mount(CombineRemote, drive, "Codespaces")
					if err != nil {
						fmt.Printf("  Mount error for %s: %v\n", drive, err)
						continue
					}
					s.Mounts = append(s.Mounts, state.Mount{
						Drive:     drive,
						RclonePid: mp.Cmd.Process.Pid,
						Remote:    CombineRemote,
						Mode:      "combined",
					})

					if waitForMount(drive, mp) {
						fmt.Printf("  Mounted: %s  (combined)\n", drive)
					} else {
						fmt.Printf("  %s not visible after 15s.\n", drive)
					}

				} else {
					csName := csNames[0]

					// Collect upstreams: new + existing on this drive
					var allUpstreams []rclone.Upstream

					// Add new upstream
					for _, u := range upstreams {
						if u.Remote == "cs-"+codespace.SafeName(csName) {
							allUpstreams = append(allUpstreams, u)
							break
						}
					}

					// Add existing upstreams from state (drive already mounted)
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
						fmt.Printf("  No remote for %s, skipping.\n", csName)
						continue
					}

					for i, m := range s.Mounts {
						if m.Drive == drive {
							killPid(m.RclonePid)
							time.Sleep(time.Second)
							s.Mounts = append(s.Mounts[:i], s.Mounts[i+1:]...)
							break
						}
					}

					fmt.Printf("  Building combine remote for %s...\n", drive)
					if err := rclone.SetCombineRemote(CombineRemote, allUpstreams); err != nil {
						fmt.Printf("  Combine remote error: %v\n", err)
					}

					fmt.Printf("  Mounting %s → %s\n", csName, drive)
					mp, err := rclone.Mount(CombineRemote, drive, strings.TrimRight(drive, ":"))
					if err != nil {
						fmt.Printf("  Mount error for %s: %v\n", drive, err)
						continue
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
						fmt.Printf("  Mounted: %s  (%s)\n", drive, csName)
					} else {
						fmt.Printf("  rclone running but %s not visible yet.\n", drive)
					}
				}
			}
		}

		state.Save(s)
		fmt.Println()
		fmt.Println("  Done. Run status to verify.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mountCmd)
	mountCmd.Flags().BoolVarP(&startFlag, "start", "s", false, "Start codespace(s) before mounting")
}

func checkDeps() error {
	if _, err := execLook("gh"); err != nil {
		return fmt.Errorf("gh not found: %w", err)
	}
	if _, err := execLook("rclone"); err != nil {
		return fmt.Errorf("rclone not found: %w", err)
	}
	return nil
}

func execLook(name string) (string, error) {
	return exec.LookPath(name)
}

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
				fmt.Printf("  rclone exited: %v\n  stderr: %s\n", err, stderr)
			} else if err != nil {
				fmt.Printf("  rclone exited: %v\n", err)
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
			fmt.Printf("  rclone exited: %v\n  stderr: %s\n", err, stderr)
		} else if err != nil {
			fmt.Printf("  rclone exited: %v\n", err)
		}
	default:
	}
	return false
}

func detectSSHPort(csName string) (int, error) {
	out, err := execCmdOutput("gh", "cs", "ssh", "-c", csName, "--",
		"sudo", "grep", "^Port", "/etc/ssh/sshd_config")
	if err != nil {
		// grep exit 1 = no Port line, default 22
		return 22, nil
	}
	fields := strings.Fields(out)
	if len(fields) >= 2 {
		port, err := strconv.Atoi(strings.TrimSpace(fields[1]))
		if err == nil {
			return port, nil
		}
	}
	return 22, nil
}
