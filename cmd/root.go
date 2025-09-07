/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-announcement/version"
	"github.com/srz-zumix/go-gh-extension/pkg/actions"
)

var rootCmd = &cobra.Command{
	Use:     "gh-announcement",
	Short:   "A tool to manage GitHub comments",
	Long:    `A tool that lets you post trackable comments and announce updates to everyone.`,
	Version: version.Version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	if actions.IsRunsOn() {
		rootCmd.SetErrPrefix(actions.GetErrorPrefix())
	}
}
