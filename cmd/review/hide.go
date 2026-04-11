package review

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-comment-kit/reviewer"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

func NewHideCmd() *cobra.Command {
	var repo string
	var group string
	var reason string
	cmd := &cobra.Command{
		Use:     "hide <target>",
		Aliases: []string{"h"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "Hide gh-comment-kit comments on the pull request",
		Long: `Hide (minimize) pull request comments that contain gh-comment-kit metadata.

Use --group to hide comments for a specific group. If --group is omitted, comments from all groups are matched.
The --reason flag sets the classifier used when hiding comments.
Valid values: ABUSE, DUPLICATE, OFF_TOPIC, OUTDATED, RESOLVED, SPAM (default: OUTDATED).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			repository, err := parser.Repository(parser.RepositoryInput(repo), parser.RepositoryFromURL(target))
			if err != nil {
				return fmt.Errorf("failed to resolve repository: %w", err)
			}
			r, err := reviewer.NewGitHubReviewer(cmd.Context(), repository, target)
			if err != nil {
				return fmt.Errorf("failed to create reviewer: %w", err)
			}
			meta := r.CreateMetaData("", 0, group)
			comments, err := r.ListComments(meta)
			if err != nil {
				return fmt.Errorf("failed to list comments: %w", err)
			}
			for _, c := range comments {
				if err := r.HideComment(c, reason); err != nil {
					return fmt.Errorf("failed to hide comment: %w", err)
				}
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&group, "group", "g", "", "comment group to hide")
	cmdutil.StringEnumFlag(cmd, &reason, "reason", "r", gh.HideClassifierOutdated, gh.HideClassifiers, "reason for hiding")
	f.StringVarP(&repo, "repo", "R", "", "Repository in the format 'owner/repo'")
	return cmd
}
