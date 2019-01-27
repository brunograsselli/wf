package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/brunograsselli/wf/git"
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

	return fmt.Sprintf(template, args[1], strings.Join(args[2:], "-"))
}
