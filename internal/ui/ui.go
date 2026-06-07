package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/state"
)

var scanner = bufio.NewScanner(os.Stdin)

func readLine() string {
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

type DriveAssignment struct {
	Drive     string
	Extend    bool
	Codespace string // empty for combined
}

// ─── Helpers ─────────────────────────────────────────────────────

func driveExists(drive string) bool {
	_, err := os.Stat(drive + "\\")
	return err == nil
}

func nextFreeDrive(used []string) string {
	stateData, _ := state.Load()
	stateDrives := map[string]bool{}
	if stateData != nil {
		for _, m := range stateData.Mounts {
			if driveExists(m.Drive) {
				stateDrives[m.Drive] = true
			}
		}
	}

	usedSet := map[string]bool{}
	for _, d := range used {
		usedSet[d] = true
	}

	candidates := "DEFGHIJKLMNOPQRSTUVWXYZ"
	for _, ch := range candidates {
		drive := string(ch) + ":"
		if !usedSet[drive] && !stateDrives[drive] && !driveExists(drive) {
			return drive
		}
	}
	return "Z:"
}

// ─── Display ─────────────────────────────────────────────────────

func ShowCsList(codespaces []codespace.Codespace) {
	fmt.Println()
	fmt.Println("  #   STATE       NAME                            REPO")
	fmt.Println("  ─   ─────       ────                            ────")
	for i, cs := range codespaces {
		num := fmt.Sprintf("%2d", i+1)
		dot := "○"
		if cs.State == "Available" {
			dot = "●"
		}
		fmt.Printf("  %s  %s  %-10s  %-32s  %s\n", num, dot, cs.State, cs.Name, cs.Repository)
	}
	fmt.Println()
}

// ─── Selection ───────────────────────────────────────────────────

func ReadSelection(codespaces []codespace.Codespace, prompt string) []codespace.Codespace {
	if prompt == "" {
		prompt = "Select [all / 1 / 1,3 / 1-3]"
	}
	for {
		fmt.Printf("  %s → ", prompt)
		raw := readLine()
		if raw == "" || raw == "all" {
			return codespaces
		}

		var selected []codespace.Codespace
		valid := true

		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)

			if n, err := strconv.Atoi(part); err == nil {
				idx := n - 1
				if idx < 0 || idx >= len(codespaces) {
					fmt.Printf("  '%s' out of range.\n", part)
					valid = false
					break
				}
				selected = append(selected, codespaces[idx])
			} else if matches := parseRange(part); matches != nil {
				from, to := matches[0], matches[1]
				if from < 0 || to >= len(codespaces) || from > to {
					fmt.Printf("  Range '%s' invalid.\n", part)
					valid = false
					break
				}
				for i := from; i <= to; i++ {
					selected = append(selected, codespaces[i])
				}
			} else {
				fmt.Printf("  Cannot parse '%s'. Use: all / 1 / 1,3 / 1-3\n", part)
				valid = false
				break
			}
		}

		if valid && len(selected) > 0 {
			return selected
		}
	}
}

func parseRange(s string) []int {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return nil
	}
	from, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	to, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return nil
	}
	return []int{from - 1, to - 1}
}

// ─── Mount mode ──────────────────────────────────────────────────

func ReadMountMode(count int) string {
	if count == 1 {
		return "separate"
	}
	fmt.Println()
	fmt.Println("  Mount mode:")
	fmt.Println("    [1] Combined  — all under one drive  (X:\\repoA\\  X:\\repoB\\)")
	fmt.Println("    [2] Separate  — each gets its own drive letter")
	fmt.Println()
	for {
		fmt.Print("  Select [1/2] → ")
		raw := readLine()
		if raw == "1" {
			return "combined"
		}
		if raw == "2" {
			return "separate"
		}
		fmt.Println("  Enter 1 or 2.")
	}
}

// ─── Drive assignment ───────────────────────────────────────────

func ReadDriveForCombined() DriveAssignment {
	stateData, _ := state.Load()
	type existingEntry struct {
		index int
		mount state.Mount
	}
	var existingCombined []existingEntry
	if stateData != nil {
		for i, m := range stateData.Mounts {
			if m.Mode == "combined" {
				existingCombined = append(existingCombined, existingEntry{index: i, mount: m})
			}
		}
	}

	if len(existingCombined) > 0 {
		fmt.Println()
		fmt.Println("  Existing combined drives:")
		for i, e := range existingCombined {
			fmt.Printf("    [%d] %s  (extend)\n", i+1, e.mount.Drive)
		}
		fmt.Println("    [n] New drive letter")
		fmt.Println()
		fmt.Print("  Select → ")
		raw := readLine()

		if n, err := strconv.Atoi(raw); err == nil {
			idx := n - 1
			if idx >= 0 && idx < len(existingCombined) {
				return DriveAssignment{Drive: existingCombined[idx].mount.Drive, Extend: true}
			}
		}
	}

	suggestion := nextFreeDrive(nil)
	fmt.Printf("  Drive letter [%s] → ", suggestion)
	raw := readLine()
	drive := suggestion
	if raw != "" {
		drive = strings.ToUpper(strings.TrimRight(raw, ":")) + ":"
	}
	return DriveAssignment{Drive: drive, Extend: false}
}

func ReadDriveAssignments(selected []codespace.Codespace, mode string) []DriveAssignment {
	if mode == "combined" {
		info := ReadDriveForCombined()
		return []DriveAssignment{{Drive: info.Drive, Extend: info.Extend}}
	}

	fmt.Println()
	fmt.Println("  Assign drive letters:")

	var usedDrives []string
	var assignments []DriveAssignment

	stateData, _ := state.Load()
	stateDrives := map[string]bool{}
	if stateData != nil {
		for _, m := range stateData.Mounts {
			if driveExists(m.Drive) {
				stateDrives[m.Drive] = true
			}
		}
	}

	for _, cs := range selected {
		suggestion := nextFreeDrive(usedDrives)
		for {
			fmt.Printf("    %s [%s] → ", cs.Name, suggestion)
			raw := readLine()
			drive := suggestion
			if raw != "" {
				drive = strings.ToUpper(strings.TrimRight(raw, ":")) + ":"
			}

			if stateDrives[drive] && !contains(usedDrives, drive) {
				fmt.Printf("  ↳ %s already mounted — will combine\n", drive)
			} else if contains(usedDrives, drive) {
				fmt.Printf("  ↳ %s shared with above — will combine\n", drive)
			}
			usedDrives = append(usedDrives, drive)
			assignments = append(assignments, DriveAssignment{Drive: drive, Codespace: cs.Name})
			break
		}
	}
	return assignments
}

// ─── Confirmations ──────────────────────────────────────────────

func Confirm(prompt string) bool {
	if prompt == "" {
		prompt = "Proceed? [y/N]"
	}
	fmt.Printf("  %s → ", prompt)
	raw := readLine()
	return strings.EqualFold(raw, "y") || strings.EqualFold(raw, "yes")
}

func SelectMountsToUnmount(mounts []state.Mount) []state.Mount {
	fmt.Println()
	fmt.Println("  Which to unmount?")
	for i, m := range mounts {
		label := "combined"
		if m.Mode != "combined" {
			label = m.Codespace
		}
		fmt.Printf("    [%d] %s  %s\n", i+1, m.Drive, label)
	}
	fmt.Println()
	fmt.Printf("  Select [all / 1 / 1,2 / 1-3] → ")
	raw := readLine()

	if raw == "" || raw == "all" {
		return mounts
	}

	var result []state.Mount
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if n, err := strconv.Atoi(part); err == nil {
			idx := n - 1
			if idx >= 0 && idx < len(mounts) {
				result = append(result, mounts[idx])
			}
		} else if matches := parseRange(part); matches != nil {
			for i := matches[0]; i <= matches[1]; i++ {
				if i >= 0 && i < len(mounts) {
					result = append(result, mounts[i])
				}
			}
		}
	}
	return result
}

// ─── Misc ───────────────────────────────────────────────────────

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
