package main

import (
	"fmt"
	"os"

	"github.com/brunograsselli/wf/cmd"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "wf",
		Short: "Automated workflow tasks",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start [feature id] [description]",
		Short: "Start a new ticket",
		Args:  cobra.MinimumNArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			if err := cmd.StartTicket(args); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "push",
		Short: "Push to remote (setup upstream)",
		Run: func(_ *cobra.Command, args []string) {
			if err := cmd.Push(args); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "pr",
		Short: "Open a new pull request",
		Run: func(_ *cobra.Command, args []string) {
			if err := cmd.OpenPullRequest(args); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
