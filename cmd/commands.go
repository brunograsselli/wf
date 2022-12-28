package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/brunograsselli/wf/config"
	"github.com/brunograsselli/wf/git"
	"github.com/pkg/errors"
)

const (
	defaultRemote = "origin"
)

var remoteURLPattern = regexp.MustCompile(`git@(.*):(.*)/(.*)\.git`)

func StartTicket(args []string, config *config.Config) error {
	mainBranch := config.MainBranch

	status, err := git.Status()
	if err != nil {
		return errors.Wrap(err, "error reading git status")
	}

	hasChanges := status.HasChanges()
	if hasChanges {
		c, err := askForConfirmation("Found changes to be committed, would like to continue and move the changes?")
		if err != nil {
			return errors.Wrap(err, "error getting confirmation")
		}

		if !c {
			fmt.Println("Aborting...")
			return nil
		}

		fmt.Println("Stashing changes")
		if err := git.Stash(); err != nil {
			return errors.Wrap(err, "error stashing changes")
		}
	}

	fmt.Printf("Updating %s branch\n", mainBranch)
	if err := git.Checkout(mainBranch); err != nil {
		return errors.Wrapf(err, "error changing to %s branch", mainBranch)
	}

	if err := git.Fetch(); err != nil {
		return errors.Wrap(err, "error fetching remote changes")
	}

	if err := git.Reset("--hard", remoteAndBranch(mainBranch)); err != nil {
		return errors.Wrapf(err, "error reseting to %s", remoteAndBranch(mainBranch))
	}

	newBranch := newBranchName(args, config.BranchNameTemplate)

	fmt.Printf("Creating new branch '%s' from '%s'\n", newBranch, mainBranch)
	if err := git.Checkout("-b", newBranch); err != nil {
		return errors.Wrapf(err, "error creating new branch '%s'", newBranch)
	}

	if hasChanges {
		fmt.Println("Applying changes")
		if err := git.StashPop(); err != nil {
			return errors.Wrap(err, "error applying changes")
		}
	}

	return nil
}

func Push(args []string, config *config.Config) error {
	mainBranch := config.MainBranch

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return errors.Wrap(err, "error getting current branch")
	}

	if currentBranch == mainBranch {
		return fmt.Errorf("current branch is %s", mainBranch)
	}

	if err := git.PushWithUpstream(defaultRemote, currentBranch); err != nil {
		return errors.Wrap(err, "error pushing to remote")
	}

	fmt.Printf("Pushed to %s/%s\n", defaultRemote, currentBranch)

	return nil
}

func OpenPullRequest(args []string, config *config.Config) error {
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

func PruneBranches(args []string, config *config.Config) error {
	mainBranch := config.MainBranch

	status, err := git.Status()
	if err != nil {
		return errors.Wrap(err, "error reading git status")
	}

	if status.HasChanges() {
		return errors.New("your current branch has uncommited changes, aborting")

	}

	previousBranch, err := git.CurrentBranch()
	if err != nil {
		return errors.Wrap(err, "error getting current branch")
	}

	fmt.Printf("Updating %s branch\n", mainBranch)

	if err := git.Fetch(); err != nil {
		return errors.Wrap(err, "error fetching remote changes")
	}

	if previousBranch != mainBranch {
		if err := git.Checkout(mainBranch); err != nil {
			return errors.Wrapf(err, "error changing to %s branch", mainBranch)
		}
	}

	if err := git.Reset("--hard", remoteAndBranch(mainBranch)); err != nil {
		return errors.Wrapf(err, "error reseting to %s", remoteAndBranch(mainBranch))
	}

	mergedBranches, err := git.Branches("--merged")
	if err != nil {
		return errors.Wrap(err, "error listing branches")
	}

	deletedPreviousBranch := false

	for _, branch := range mergedBranches {
		if branch.Current || branch.Name == remoteAndBranch(mainBranch) {
			continue
		}

		if branch.Name == previousBranch {
			deletedPreviousBranch = true
		}

		fmt.Printf("* Deleting branch: %s\n", branch.Name)

		git.DeleteBranch(branch.Name)
	}

	if previousBranch != mainBranch && !deletedPreviousBranch {
		if err := git.Checkout("-"); err != nil {
			return errors.Wrap(err, "error changing back to previous branch")
		}
	}

	if err := git.PruneRemote(defaultRemote); err != nil {
		return errors.Wrap(err, "error pruning remote")
	}

	return nil
}

func newBranchName(args []string, template string) string {
	return fmt.Sprintf(template, args[0], strings.Join(args[1:], "-"))
}

func askForConfirmation(s string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/N]: ", s)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y", nil
}

func remoteAndBranch(b string) string {
	return fmt.Sprintf("%s/%s", defaultRemote, b)
}
