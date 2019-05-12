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

const defaultBranchNameTemplate = "%s/%s"

func StartTicket(args []string) error {
	status, err := git.Status()
	if err != nil {
		return errors.New("error reading git status")
	}

	if status.HasChanges() {
		return errors.New("found files to be committed")
	}

	fmt.Println("Updating master branch")
	if err := git.Checkout("master"); err != nil {
		return errors.New("error changing to master branch")
	}

	if err := git.Fetch(); err != nil {
		return errors.New("error fetching remote changes")
	}

	if err := git.Reset("--hard", "origin/master"); err != nil {
		return errors.New("error reseting to origin/master")
	}

	newBranch := generateName(args)

	fmt.Printf("Creating new branch '%s' from 'master'\n", newBranch)
	if err := git.Checkout("-b", newBranch); err != nil {
		return errors.New("error creating new branch")
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
		return errors.New("error getting current branch")
	}

	if currentBranch == "master" {
		return errors.New("current branch is master")
	}

	if err := git.PushWithUpstream(currentBranch); err != nil {
		return errors.New("error pushing to remote")
	}

	fmt.Printf("Pushed to origin %s\n", currentBranch)

	return nil
}

var abc = regexp.MustCompile(`git@(.*):(.*)/(.*)\.git`)

func OpenPullRequest(args []string) error {
	remoteURL, err := git.RemoteURL("origin")
	if err != nil {
		return errors.Wrap(err, "error getting remote url")
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		return errors.Wrap(err, "error getting current branch")
	}

	result := abc.FindAllStringSubmatch(remoteURL, -1)
	if len(result) == 0 {
		return errors.New("can't parse remote url")
	}
	github, user, repo := result[0][1], result[0][2], result[0][3]

	return exec.Command("open", fmt.Sprintf("https://%s/%s/%s/pull/new/%s", github, user, repo, branch)).Run()

	return nil
}
