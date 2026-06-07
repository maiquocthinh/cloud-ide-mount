package codespace

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Codespace struct {
	Name       string `json:"name"`
	State      string `json:"state"`
	Repository string `json:"repository"`
}

type UpstreamPath struct {
	Cs         Codespace
	FolderPath string
}

var safeRe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func SafeName(name string) string {
	return safeRe.ReplaceAllString(name, "-")
}

func List() ([]Codespace, error) {
	out, err := exec.Command("gh", "cs", "list", "--limit", "100", "--json", "name,state,repository").Output()
	if err != nil {
		return nil, fmt.Errorf("cannot fetch codespaces (check: gh auth status): %w", err)
	}
	var cs []Codespace
	if err := json.Unmarshal(out, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// BuildUpstreamPaths builds org/repo folder paths for each codespace.
// Always uses org/repo format, deduplicating with _2, _3 suffixes.
func BuildUpstreamPaths(codespaces []Codespace) []UpstreamPath {
	used := map[string]bool{}
	var result []UpstreamPath

	for _, cs := range codespaces {
		parts := strings.SplitN(cs.Repository, "/", 2)
		org := "unknown"
		repo := SafeName(cs.Repository)
		if len(parts) == 2 {
			org = SafeName(parts[0])
			repo = SafeName(parts[1])
		}

		base := org + "/" + repo
		path := base
		n := 2
		for used[path] {
			path = fmt.Sprintf("%s_%d", base, n)
			n++
		}
		used[path] = true

		result = append(result, UpstreamPath{Cs: cs, FolderPath: path})
	}

	return result
}
