package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/brunograsselli/wf/config"
	"github.com/brunograsselli/wf/git"
)

const (
	defaultRemote = "origin"
)

var remoteURLPattern = regexp.MustCompile(`git@(.*):(.*)/(.*)\.git`)

func StartTicket(args []string, config *config.Config) error {
	mainBranch := config.MainBranch

	status, err := git.Status()
	if err != nil {
		return fmt.Errorf("error reading git status: %w", err)
	}

	hasChanges := status.HasChanges()
	if hasChanges {
		c, err := askForConfirmation("Found changes to be committed, would like to continue and move the changes?")
		if err != nil {
			return fmt.Errorf("error getting confirmation: %w", err)
		}

		if !c {
			fmt.Println("Aborting...")
			return nil
		}

		fmt.Println("Stashing changes")
		if err := git.Stash(); err != nil {
			return fmt.Errorf("error stashing changes: %w", err)
		}
	}

	fmt.Printf("Updating %s branch\n", mainBranch)
	if err := git.Checkout(mainBranch); err != nil {
		return fmt.Errorf("error changing to %s branch: %w", mainBranch, err)
	}

	if err := git.Fetch(); err != nil {
		return fmt.Errorf("error fetching remote changes: %w", err)
	}

	if err := git.Reset("--hard", remoteAndBranch(mainBranch)); err != nil {
		return fmt.Errorf("error reseting to %s: %w", remoteAndBranch(mainBranch), err)
	}

	newBranch := newBranchName(args, config.BranchNameTemplate)

	fmt.Printf("Creating new branch '%s' from '%s'\n", newBranch, mainBranch)
	if err := git.Checkout("-b", newBranch); err != nil {
		return fmt.Errorf("error creating new branch '%s': %w", newBranch, err)
	}

	if hasChanges {
		fmt.Println("Applying changes")
		if err := git.StashPop(); err != nil {
			return fmt.Errorf("error applying changes: %w", err)
		}
	}

	return nil
}

func Push(args []string, config *config.Config) error {
	mainBranch := config.MainBranch

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting current branch: %w", err)
	}

	if currentBranch == mainBranch {
		return fmt.Errorf("current branch is %s", mainBranch)
	}

	if err := git.PushWithUpstream(defaultRemote, currentBranch); err != nil {
		return fmt.Errorf("error pushing to remote: %w", err)
	}

	fmt.Printf("Pushed to %s/%s\n", defaultRemote, currentBranch)

	return nil
}

func OpenPullRequest(args []string, config *config.Config) error {
	remoteURL, err := git.RemoteURL(defaultRemote)
	if err != nil {
		return fmt.Errorf("error getting remote url: %w", err)
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting current branch: %w", err)
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
		return fmt.Errorf("error reading git status: %w", err)
	}

	if status.HasChanges() {
		return errors.New("your current branch has uncommited changes, aborting")

	}

	previousBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting current branch: %w", err)
	}

	fmt.Printf("Updating %s branch\n", mainBranch)

	if err := git.Fetch(); err != nil {
		return fmt.Errorf("error fetching remote changes: %w", err)
	}

	if previousBranch != mainBranch {
		if err := git.Checkout(mainBranch); err != nil {
			return fmt.Errorf("error changing to %s branch: %w", mainBranch, err)
		}
	}

	if err := git.Reset("--hard", remoteAndBranch(mainBranch)); err != nil {
		return fmt.Errorf("error reseting to %s: %w", remoteAndBranch(mainBranch), err)
	}

	mergedBranches, err := git.Branches("--merged")
	if err != nil {
		return fmt.Errorf("error listing branches: %w", err)
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
			return fmt.Errorf("error changing back to previous branch: %w", err)
		}
	}

	if err := git.PruneRemote(defaultRemote); err != nil {
		return fmt.Errorf("error pruning remote: %w", err)
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
