package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-comment-kit/version"
	"github.com/srz-zumix/go-gh-extension/pkg/actions"
	"github.com/srz-zumix/go-gh-extension/pkg/cmdflags"
)

var rootCmd = &cobra.Command{
	Use:     "gh-comment-kit",
	Short:   "A tool to manage GitHub comments",
	Long:    `A tool for posting trackable comments and providing the latest updates.`,
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
	cmdflags.AddPersistentFlags(rootCmd)
}
