package commentator

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/actions"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

type GitHubCommentator struct {
	client     *gh.GitHubClient
	ctx        context.Context
	Repository repository.Repository
	Target     any
}

type Comment struct {
	MetaData *MetaData
	Comment  *github.IssueComment
}

func NewGitHubCommentator(repo repository.Repository, target any) (*GitHubCommentator, error) {
	client, err := gh.NewGitHubClientWithRepo(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}
	ctx := context.Background()
	return &GitHubCommentator{
		client:     client,
		ctx:        ctx,
		Repository: repo,
		Target:     target,
	}, nil
}

func (g *GitHubCommentator) CreateMetaData(source any, index int, group string) MetaData {
	var url string
	if actions.IsRunsOn() {
		url = actions.GetRunURL()
	}
	return CreateMetaData(source, index, group, url)
}

func (g *GitHubCommentator) Comment(body string, user string, meta MetaData) (string, error) {
	comments, err := g.ListComments(meta)
	if err != nil {
		return "", fmt.Errorf("failed to list comments: %w", err)
	}
	last := g.GetLastComment(comments)
	if last != nil {
		meta.Index = last.MetaData.Index + 1
	}
	c, err := gh.CreateIssueComment(g.ctx, g.client, g.Repository, g.Target, body+"\n"+meta.ToHTML(), user)
	if err != nil {
		return "", fmt.Errorf("failed to create comment: %w", err)
	}
	return c.GetHTMLURL(), nil
}

func (g *GitHubCommentator) ListComments(meta MetaData) ([]*Comment, error) {
	comments, err := gh.ListIssueComments(g.ctx, g.client, g.Repository, g.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	var result []*Comment
	for _, c := range comments {
		m, err := ParseMetaData(*c.Body)
		if err != nil {
			continue
		}
		if m.Group != meta.Group {
			continue
		}
		result = append(result, &Comment{
			MetaData: m,
			Comment:  c,
		})
	}
	return result, nil
}

func (g *GitHubCommentator) LastComment(meta MetaData) (*Comment, error) {
	comments, err := g.ListComments(meta)
	if err != nil {
		return nil, err
	}
	return g.GetLastComment(comments), nil
}

func (g *GitHubCommentator) GetLastComment(comments []*Comment) *Comment {
	var last *Comment
	var lastIndex int
	for _, c := range comments {
		if last == nil || lastIndex < c.MetaData.Index {
			last = c
			lastIndex = c.MetaData.Index
		}
	}
	return last
}

func (g *GitHubCommentator) GetTargetURL() (string, error) {
	issue, err := gh.GetIssue(g.ctx, g.client, g.Repository, g.Target)
	if err != nil {
		return "", fmt.Errorf("failed to get issue: %w", err)
	}
	return issue.GetHTMLURL(), nil
}
