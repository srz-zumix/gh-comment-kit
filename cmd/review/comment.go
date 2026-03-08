package review

import (
	"fmt"
	"os"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-comment-kit/reviewer"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

type ToOptions struct {
	Exporter cmdutil.Exporter
}

func NewCommentCmd() *cobra.Command {
	opts := &ToOptions{}
	var repo string
	var body string
	var bodyFile string
	var dryrun bool
	var group string
	var path string
	var line int
	var commentOpts reviewer.CommentOption
	cmd := &cobra.Command{
		Use:     "comment <target>",
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "Post a review comment to the pull request",
		Long:    `Post a review comment to the pull request.`,
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
			if bodyFile != "" {
				if bodyFile == "-" {
					bodyFile = "/dev/stdin"
				}
				data, err := os.ReadFile(bodyFile)
				if err != nil {
					return fmt.Errorf("failed to read body file: %w", err)
				}
				body = string(data)
			}

			meta := r.CreateMetaData("", 0, group)
			if dryrun {
				url, err := r.GetTargetURL()
				if err != nil {
					return fmt.Errorf("failed to get target URL: %w", err)
				}
				fmt.Printf("Dry run: would post comment to %s on %s\n", url, path)
				fmt.Println("-----")
				fmt.Println(body)
				fmt.Println("-----")
				fmt.Printf("MetaData: %s\n", meta.ToHTML())
			} else {
				target := &reviewer.CommentTarget{
					Path: &path,
				}
				if line > 0 {
					target.Line = &line
				}
				url, err := r.Comment(body, target, meta, &commentOpts)
				if err != nil {
					return fmt.Errorf("failed to post comment: %w", err)
				}
				fmt.Printf("Comment posted: %s\n", url)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&body, "body", "b", "", "comment body")
	f.StringVarP(&bodyFile, "body-file", "F", "", "comment body file")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	f.StringVarP(&path, "path", "p", "", "file path to comment on")
	f.IntVarP(&line, "line", "l", 0, "line number to comment on")
	f.BoolVarP(&dryrun, "dryrun", "n", false, "Dry run: do not post comment, just print what would be sent")
	f.StringVarP(&group, "group", "g", "gh-comment-kit", "comment group")
	f.BoolVar(&commentOpts.Update, "update", false, "update the last comment")
	f.BoolVar(&commentOpts.Resolve, "resolve", false, "resolve previous review comments in the same group")
	f.BoolVar(&commentOpts.Delete, "delete", false, "delete previous comments in the same group")
	f.BoolVar(&commentOpts.Truncate, "truncate", false, "truncate comment if it exceeds size limit instead of splitting")
	f.StringVarP(&repo, "repo", "R", "", "Repository in the format 'owner/repo'")
	cmdutil.AddFormatFlags(cmd, &opts.Exporter)
	return cmd
}
