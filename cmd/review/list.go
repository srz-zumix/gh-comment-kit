package review

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-comment-kit/reviewer"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
	"github.com/srz-zumix/go-gh-extension/pkg/render"
)

type ListOptions struct {
	Exporter cmdutil.Exporter
}

func NewListCmd() *cobra.Command {
	opts := &ListOptions{}
	var repo string
	var group string
	cmd := &cobra.Command{
		Use:     "list <target>",
		Aliases: []string{"ls"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "List comments for the pull request",
		Long:    `List comments for the pull request.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			repository, err := parser.Repository(parser.RepositoryInput(repo), parser.RepositoryFromURL(target))
			if err != nil {
				return fmt.Errorf("failed to resolve repository: %w", err)
			}
			r, err := reviewer.NewGitHubReviewer(repository, target)
			if err != nil {
				return fmt.Errorf("failed to create reviewer: %w", err)
			}
			meta := r.CreateMetaData("", 0, group)
			comments, err := r.ListComments(meta)
			if err != nil {
				return fmt.Errorf("failed to list comments: %w", err)
			}

			renderer := render.NewRenderer(opts.Exporter)
			renderer.RenderCommentDefault(comments.GetAllComments())
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&group, "group", "g", "", "comment group")
	f.StringVarP(&repo, "repo", "R", "", "Repository in the format 'owner/repo'")
	cmdutil.AddFormatFlags(cmd, &opts.Exporter)
	return cmd
}
