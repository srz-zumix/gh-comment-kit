package cmd

import (
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-comment-kit/cmd/review"
)

func NewReviewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "review",
		Short: "Pull request review comments",
		Long:  `Manage pull request review comments.`,
	}

	cmd.AddCommand(review.NewCommentCmd())
	cmd.AddCommand(review.NewListCmd())

	return cmd
}

func init() {
	rootCmd.AddCommand(NewReviewCmd())
}
