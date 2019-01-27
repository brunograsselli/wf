package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/brunograsselli/workflow/git"
)

const defaultBranchNameTemplate = "%s/%s"

func main() {
	if len(os.Args) < 3 {
		printAndExit("not enough arguments", 1)
	}

	status, err := git.Status()
	if err != nil {
		printAndExit("error reading git status", 1)
	}

	if status.HasChanges() {
		printAndExit("found files to be committed", 1)
	}

	fmt.Println("Updating master branch")
	if err := git.Checkout("master"); err != nil {
		printAndExit("error changing to master branch", 1)
	}

	if err := git.Fetch(); err != nil {
		printAndExit("error fetching remote changes", 1)
	}

	if err := git.Reset("--hard", "origin/master"); err != nil {
		printAndExit("error reseting to origin/master", 1)
	}

	newBranch := generateName(os.Args)

	fmt.Printf("Creating new branch '%s' from 'master'\n", newBranch)
	if err := git.Checkout("-b", newBranch); err != nil {
		printAndExit("error creating new branch", 1)
	}
}

func printAndExit(msg string, status int) {
	fmt.Println(msg)
	os.Exit(status)
}

func generateName(args []string) string {
	template := os.Getenv("WF_BRANCH_NAME_TEMPLATE")
	if template == "" {
		template = defaultBranchNameTemplate
	}

	return fmt.Sprintf(template, args[1], strings.Join(args[2:], "-"))
}
