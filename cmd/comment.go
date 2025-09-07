package cmd

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-commentator/commentator"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

type ToOptions struct {
	Exporter cmdutil.Exporter
}

func NewCommentCmd() *cobra.Command {
	opts := &ToOptions{}
	var repo string
	var name string
	cmd := &cobra.Command{
		Use:     "comment <target>",
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := parser.Repository(parser.RepositoryInput(repo))
			if err != nil {
				return fmt.Errorf("failed to resolve repository: %w", err)
			}
			c, err := commentator.NewGitHubCommentator(repository, args[0])
			if err != nil {
				return fmt.Errorf("failed to create commentator: %w", err)
			}
			meta := c.CreateMetaData("", 0, "")
			err = c.Comment(args[1], name, meta)
			if err != nil {
				return fmt.Errorf("failed to post comment: %w", err)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&name, "name", "n", "", "comment author name")
	f.StringVarP(&repo, "repo", "R", "", "Repository in the format 'owner/repo'")
	cmdutil.AddFormatFlags(cmd, &opts.Exporter)
	return cmd
}

func init() {
	rootCmd.AddCommand(NewCommentCmd())
}
