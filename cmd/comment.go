package cmd

import (
	"fmt"
	"os"

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
	var body string
	var bodyFile string
	var dryrun bool
	var group string
	cmd := &cobra.Command{
		Use:     "comment <target>",
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "Post a comment to the target (issue or pull request)",
		Long:    `Post a comment to the target (issue or pull request).`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if body == "" && bodyFile == "" {
				return fmt.Errorf("either --body or --body-file must be specified")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			repository, err := parser.Repository(parser.RepositoryInput(repo))
			if err != nil {
				return fmt.Errorf("failed to resolve repository: %w", err)
			}
			c, err := commentator.NewGitHubCommentator(repository, target)
			if err != nil {
				return fmt.Errorf("failed to create commentator: %w", err)
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

			meta := c.CreateMetaData("", 0, group)
			if dryrun {
				url, err := c.GetTargetURL()
				if err != nil {
					return fmt.Errorf("failed to get target URL: %w", err)
				}
				fmt.Printf("Dry run: would post comment to %s\n", url)
				fmt.Println("-----")
				fmt.Println(body)
				fmt.Println("-----")
				fmt.Printf("MetaData: %s\n", meta.ToHTML())
			} else {
				url, err := c.Comment(body, name, meta)
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
	f.BoolVarP(&dryrun, "dryrun", "n", false, "Dry run: do not actually set labels")
	f.StringVarP(&group, "group", "g", "gh-commentator", "comment group")
	f.StringVar(&name, "name", "", "comment author name")
	f.StringVarP(&repo, "repo", "R", "", "Repository in the format 'owner/repo'")
	cmdutil.AddFormatFlags(cmd, &opts.Exporter)
	return cmd
}

func init() {
	rootCmd.AddCommand(NewCommentCmd())
}
