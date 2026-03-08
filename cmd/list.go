package cmd

import (
	"context"
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
	"github.com/srz-zumix/go-gh-extension/pkg/render"
)

type ListOptions struct {
	Exporter cmdutil.Exporter
}

func NewListCmd() *cobra.Command {
	opts := &ListOptions{}
	var repo string
	cmd := &cobra.Command{
		Use:     "list <target>",
		Aliases: []string{"ls"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "List comments for the target (issue or pull request)",
		Long:    `List comments for the target (issue or pull request).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			repository, err := parser.Repository(parser.RepositoryInput(repo), parser.RepositoryFromURL(target))
			if err != nil {
				return fmt.Errorf("failed to resolve repository: %w", err)
			}
			client, err := gh.NewGitHubClientWithRepo(repository)
			if err != nil {
				return fmt.Errorf("failed to create GitHub client: %w", err)
			}
			ctx := context.Background()

			// Get issue
			issue, err := gh.FindIssueByIdentifier(ctx, client, repository, target)
			if err != nil {
				return fmt.Errorf("failed to get issue %s: %w", target, err)
			}
			comments, err := gh.ListIssueComments(ctx, client, repository, issue.GetNumber())
			if err != nil {
				return fmt.Errorf("failed to list comments: %w", err)
			}

			renderer := render.NewRenderer(opts.Exporter)
			renderer.RenderCommentDefault(comments)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&repo, "repo", "R", "", "Repository in the format 'owner/repo'")
	cmdutil.AddFormatFlags(cmd, &opts.Exporter)
	return cmd
}

func init() {
	rootCmd.AddCommand(NewListCmd())
}
