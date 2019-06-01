package git

import (
	"os/exec"
	"regexp"
	"strings"
)

var (
	changedFilesRegexp = regexp.MustCompile(`[ MADRCU]{2}\s+(.*)`)
	branchRegexp       = regexp.MustCompile(`[ *]{2}(.*)`)
)

type StatusOutput struct {
	changedFiles []string
}

func Status() (*StatusOutput, error) {
	out, err := exec.Command("git", "status", "--short").Output()
	if err != nil {
		return nil, err
	}

	status := &StatusOutput{}

	result := changedFilesRegexp.FindAllStringSubmatch(string(out), -1)

	for i := range result {
		status.changedFiles = append(status.changedFiles, result[i][1])
	}

	return status, nil
}

func (s *StatusOutput) HasChanges() bool {
	return len(s.changedFiles) > 0
}

func Branches(extra string) ([]*Branch, error) {
	out, err := exec.Command("git", "branch", extra).Output()
	if err != nil {
		return nil, err
	}

	result := branchRegexp.FindAllStringSubmatch(string(out), -1)

	var branches []*Branch

	for i := range result {
		current := string(result[i][0][0]) == "*"

		branch := &Branch{
			Name:    result[i][1],
			Current: current,
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

func DeleteBranch(name string) error {
	return exec.Command("git", "branch", "-d", name).Run()
}

type Branch struct {
	Name    string
	Current bool
}

func Checkout(options ...string) error {
	args := []string{"checkout"}
	return exec.Command("git", append(args, options...)...).Run()
}

func Fetch() error {
	return exec.Command("git", "fetch").Run()
}

func Reset(mode string, source string) error {
	return exec.Command("git", "reset", mode, source).Run()
}

func PushWithUpstream(remote string, branch string) error {
	return exec.Command("git", "push", "--set-upstream", remote, branch).Run()
}

func CurrentBranch() (string, error) {
	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	return strings.TrimSpace(string(branch)), err
}

func RemoteURL(remote string) (string, error) {
	url, err := exec.Command("git", "remote", "get-url", "--push", remote).Output()
	return strings.TrimSpace(string(url)), err
}

func Stash() error {
	return exec.Command("git", "stash").Run()
}

func StashPop() error {
	return exec.Command("git", "stash", "pop").Run()
}

func PruneRemote(remote string) error {
	return exec.Command("git", "remote", "prune", remote).Run()
}
