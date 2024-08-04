package main

import (
	"fmt"
	"os"

	"github.com/brunograsselli/wf/cmd"
	"github.com/brunograsselli/wf/config"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "wf",
		Short: "Automated workflow tasks",
	}

	ticketCommand := &cobra.Command{
		Use:     "ticket [command]",
		Aliases: []string{"t"},
		Short:   "[t] Commands related to tickets",
	}

	ticketCommand.AddCommand(&cobra.Command{
		Use:     "start [ticket id] [description]",
		Short:   "[s] Start a new ticket",
		Aliases: []string{"s", "st"},
		Args:    cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.StartTicket(args, config.Init())
		},
	})

	ticketCommand.AddCommand(&cobra.Command{
		Use:     "push",
		Short:   "[p] Push to remote (setup upstream)",
		Aliases: []string{"p"},
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.Push(args, config.Init())
		},
	})

	ticketCommand.AddCommand(&cobra.Command{
		Use:     "open-pull-request",
		Short:   "[pr] Open a new pull request",
		Aliases: []string{"open-pr", "pr"},
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.OpenPullRequest(args, config.Init())
		},
	})

	repoCommand := &cobra.Command{
		Use:     "repository [command]",
		Aliases: []string{"r", "repo"},
		Short:   "[r] Commands related to the git repository",
	}

	repoCommand.AddCommand(&cobra.Command{
		Use:     "prune",
		Short:   "[p] Prune merged local branches and deleted remote ones",
		Aliases: []string{"p"},
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.PruneBranches(args, config.Init())
		},
	})

	rootCmd.AddCommand(ticketCommand)
	rootCmd.AddCommand(repoCommand)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
