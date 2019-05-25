package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/brunograsselli/wf/git"
	"github.com/pkg/errors"
)

const (
	defaultBranchNameTemplate = "%s/%s"
	defaultMasterBranch       = "master"
	defaultRemote             = "origin"
	defaultRemoteAndBranch    = "origin/master"
)

var remoteURLPattern = regexp.MustCompile(`git@(.*):(.*)/(.*)\.git`)

func StartTicket(args []string) error {
	status, err := git.Status()
	if err != nil {
		return errors.Wrap(err, "error reading git status")
	}

	if status.HasChanges() {
		return errors.New("found files to be committed")
	}

	fmt.Printf("Updating %s branch\n", defaultMasterBranch)
	if err := git.Checkout(defaultMasterBranch); err != nil {
		return errors.Wrapf(err, "error changing to %s branch", defaultMasterBranch)
	}

	if err := git.Fetch(); err != nil {
		return errors.Wrap(err, "error fetching remote changes")
	}

	if err := git.Reset("--hard", defaultRemoteAndBranch); err != nil {
		return errors.Wrapf(err, "error reseting to %s", defaultRemoteAndBranch)
	}

	newBranch := generateName(args)

	fmt.Printf("Creating new branch '%s' from '%s'\n", newBranch, defaultMasterBranch)
	if err := git.Checkout("-b", newBranch); err != nil {
		return errors.Wrapf(err, "error creating new branch '%s'", newBranch)
	}

	return nil
}

func generateName(args []string) string {
	template := os.Getenv("WF_BRANCH_NAME_TEMPLATE")
	if template == "" {
		template = defaultBranchNameTemplate
	}

	return fmt.Sprintf(template, args[0], strings.Join(args[1:], "-"))
}

func Push(args []string) error {
	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return errors.Wrap(err, "error getting current branch")
	}

	if currentBranch == defaultMasterBranch {
		return fmt.Errorf("current branch is %s", defaultMasterBranch)
	}

	if err := git.PushWithUpstream(defaultRemote, currentBranch); err != nil {
		return errors.Wrap(err, "error pushing to remote")
	}

	fmt.Printf("Pushed to %s/%s\n", defaultRemote, currentBranch)

	return nil
}

func OpenPullRequest(args []string) error {
	remoteURL, err := git.RemoteURL(defaultRemote)
	if err != nil {
		return errors.Wrap(err, "error getting remote url")
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		return errors.Wrap(err, "error getting current branch")
	}

	result := remoteURLPattern.FindAllStringSubmatch(remoteURL, -1)
	if len(result) == 0 {
		return errors.New("can't parse remote url")
	}
	github, user, repo := result[0][1], result[0][2], result[0][3]

	return exec.Command("open", fmt.Sprintf("https://%s/%s/%s/pull/new/%s", github, user, repo, branch)).Run()
}
