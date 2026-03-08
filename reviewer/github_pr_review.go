package reviewer

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/actions"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

const (
	// GitHub comment max size is 65536 characters
	// Reserve space for metadata HTML
	maxCommentSize = 65000
)

type GitHubReviewer struct {
	client     *gh.GitHubClient
	ctx        context.Context
	Repository repository.Repository
	Target     *github.PullRequest
}

type Comment struct {
	MetaData      *MetaData
	ReviewComment *github.PullRequestComment
	Comment       *github.IssueComment
}

type CommentOption struct {
	Update   bool
	Resolve  bool
	Delete   bool
	Truncate bool
}

type CommentTarget struct {
	Path      *string
	Line      *int
	StartLine *int
	Side      *string
	CommitSHA *string
}

type Comments []*Comment

func (c Comments) GetReviewComments() []*github.PullRequestComment {
	var result []*github.PullRequestComment
	for _, cm := range c {
		if cm.ReviewComment != nil {
			result = append(result, cm.ReviewComment)
		}
	}
	return result
}

func (c Comments) GetPullRequestComments() []*github.IssueComment {
	var result []*github.IssueComment
	for _, cm := range c {
		if cm.Comment != nil {
			result = append(result, cm.Comment)
		}
	}
	return result
}

func (c Comments) GetAllComments() []any {
	var result []any
	for _, cm := range c {
		if cm.Comment != nil {
			result = append(result, cm.Comment)
		}
		if cm.ReviewComment != nil {
			result = append(result, cm.ReviewComment)
		}
	}
	return result
}

func NewGitHubReviewer(repo repository.Repository, target string) (*GitHubReviewer, error) {
	client, err := gh.NewGitHubClientWithRepo(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}
	ctx := context.Background()
	pr, err := gh.FindPRByIdentifier(ctx, client, repo, target)
	if err != nil {
		return nil, fmt.Errorf("failed to find pull request %s: %w", target, err)
	}
	return &GitHubReviewer{
		client:     client,
		ctx:        ctx,
		Repository: repo,
		Target:     pr,
	}, nil
}

func (g *GitHubReviewer) ListPullRequestReviewComments() ([]*github.PullRequestComment, error) {
	comments, err := gh.ListPullRequestReviewComments(g.ctx, g.client, g.Repository, g.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull request review comments: %w", err)
	}
	return comments, nil
}

func (g *GitHubReviewer) CreateMetaData(source any, index int, group string) MetaData {
	var url string
	if actions.IsRunsOn() {
		url = actions.GetRunURL()
	}
	return CreateMetaData(source, index, group, url)
}

func (g *GitHubReviewer) Comment(body string, target *CommentTarget, meta MetaData, opt *CommentOption) (string, error) {
	comments, err := g.ListComments(meta)
	if err != nil {
		return "", fmt.Errorf("failed to list comments: %w", err)
	}
	last := g.GetLastComment(comments)
	if last != nil {
		meta.Index = last.MetaData.Index + 1
	}

	// Handle delete and resolve options for existing comments
	if opt != nil {
		for _, c := range comments {
			if opt.Update && c == last {
				continue
			}
			if opt.Delete {
				err = g.DeleteComment(c)
				if err != nil {
					return "", fmt.Errorf("failed to delete comment: %w", err)
				}
			} else if opt.Resolve && c.ReviewComment != nil {
				err = gh.ResolvePullRequestComment(g.ctx, g.client, g.Repository, g.Target, c.ReviewComment)
				if err != nil {
					return "", fmt.Errorf("failed to resolve comment: %w", err)
				}
			}
		}
	}

	// Check if body needs to be split or truncated
	metaHTML := meta.ToHTML()
	maxBodySize := maxCommentSize - len(metaHTML)

	// If truncate option is enabled, truncate instead of splitting
	if opt != nil && opt.Truncate && len(body) > maxBodySize {
		body = truncateComment(body, maxBodySize)
	}

	parts := splitComment(body, maxBodySize)

	if len(parts) == 1 {
		// Single comment (no split needed)
		commentBody := body + "\n" + metaHTML
		if opt != nil && opt.Update && last != nil {
			return g.editComment(last, commentBody)
		}
		return g.createComment(commentBody, target)
	}

	// Multiple parts - create split comments
	return g.createSplitComments(parts, target, meta, opt, last)
}

func (g *GitHubReviewer) createComment(commentBody string, target *CommentTarget) (string, error) {
	prComment, err := g.createPullRequestComment(commentBody, target)
	if err != nil {
		return "", fmt.Errorf("failed to create pull request comment object: %w", err)
	}
	if prComment != nil {
		c, err := gh.CreatePullRequestComment(g.ctx, g.client, g.Repository, g.Target, prComment)
		if err != nil {
			return "", fmt.Errorf("failed to create comment: %w", err)
		}
		return c.GetHTMLURL(), nil
	} else {
		c, err := gh.CreateIssueComment(g.ctx, g.client, g.Repository, g.Target, commentBody)
		if err != nil {
			return "", fmt.Errorf("failed to create comment: %w", err)
		}
		return c.GetHTMLURL(), nil
	}
}

func (g *GitHubReviewer) createPullRequestComment(body string, target *CommentTarget) (*github.PullRequestComment, error) {
	if target == nil {
		return nil, nil
	}
	commitID := g.Target.GetHead().GetSHA()
	if target.CommitSHA != nil && *target.CommitSHA != "" {
		commitID = *target.CommitSHA
	}
	if target.Path != nil && *target.Path != "" {
		if target.Line != nil && *target.Line > 0 {
			return &github.PullRequestComment{
				Body:        &body,
				Path:        target.Path,
				Line:        target.Line,
				StartLine:   target.StartLine,
				Side:        target.Side,
				SubjectType: github.Ptr("line"),
				CommitID:    &commitID,
			}, nil
		}
		return &github.PullRequestComment{
			Body:        &body,
			Path:        target.Path,
			SubjectType: github.Ptr("file"),
			CommitID:    &commitID,
		}, nil
	}
	return nil, nil
}

func (g *GitHubReviewer) editComment(comment *Comment, body string) (string, error) {
	if comment.Comment != nil {
		edited, err := gh.EditIssueComment(g.ctx, g.client, g.Repository, comment.Comment, body)
		if err != nil {
			return "", fmt.Errorf("failed to edit comment: %w", err)
		}
		comment.Comment = edited
		return edited.GetHTMLURL(), nil
	}
	if comment.ReviewComment != nil {
		edited, err := gh.EditPullRequestComment(g.ctx, g.client, g.Repository, comment.ReviewComment, body)
		if err != nil {
			return "", fmt.Errorf("failed to edit review comment: %w", err)
		}
		comment.ReviewComment = edited
		return edited.GetHTMLURL(), nil
	}
	return "", fmt.Errorf("no comment to edit")
}

func (g *GitHubReviewer) DeleteComment(comment *Comment) error {
	if comment.Comment != nil {
		return gh.DeleteIssueComment(g.ctx, g.client, g.Repository, comment.Comment)
	}
	if comment.ReviewComment != nil {
		return gh.DeletePullRequestComment(g.ctx, g.client, g.Repository, comment.ReviewComment)
	}
	return fmt.Errorf("no comment to delete")
}

func (g *GitHubReviewer) ListComments(meta MetaData) (Comments, error) {
	var result Comments
	prComments, err := g.ListPullRequestComments(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull request comments: %w", err)
	}
	result = append(result, prComments...)
	reviewComments, err := g.ListReviewComments(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to list review comments: %w", err)
	}
	result = append(result, reviewComments...)
	return result, nil
}

func (g *GitHubReviewer) ListReviewComments(meta MetaData) (Comments, error) {
	comments, err := gh.ListPullRequestReviewComments(g.ctx, g.client, g.Repository, g.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	var result Comments
	for _, c := range comments {
		m, err := g.extractMetaData(*c.Body, meta)
		if err != nil {
			continue
		}
		result = append(result, &Comment{
			MetaData:      m,
			ReviewComment: c,
		})
	}
	return result, nil
}

func (g *GitHubReviewer) ListPullRequestComments(meta MetaData) (Comments, error) {
	comments, err := gh.ListIssueComments(g.ctx, g.client, g.Repository, g.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	var result Comments
	for _, c := range comments {
		m, err := g.extractMetaData(*c.Body, meta)
		if err != nil {
			continue
		}
		result = append(result, &Comment{
			MetaData: m,
			Comment:  c,
		})
	}
	return result, nil
}

func (g *GitHubReviewer) extractMetaData(commentBody string, meta MetaData) (*MetaData, error) {
	m, _, err := ParseMetaData(commentBody)
	if err != nil {
		return nil, err
	}
	// If meta.Group is empty, we accept all groups.
	// This is useful to find comments across all groups.
	if meta.Group != "" && m.Group != meta.Group {
		return nil, fmt.Errorf("comment does not belong to group %q", meta.Group)
	}
	return m, nil
}

func (g *GitHubReviewer) LastComment(meta MetaData) (*Comment, error) {
	comments, err := g.ListComments(meta)
	if err != nil {
		return nil, err
	}
	return g.GetLastComment(comments), nil
}

func (g *GitHubReviewer) GetLastComment(comments []*Comment) *Comment {
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

func (g *GitHubReviewer) GetTargetURL() (string, error) {
	issue, err := gh.GetIssue(g.ctx, g.client, g.Repository, g.Target)
	if err != nil {
		return "", fmt.Errorf("failed to get issue: %w", err)
	}
	return issue.GetHTMLURL(), nil
}

func (g *GitHubReviewer) createSplitComments(parts []string, target *CommentTarget, meta MetaData, opt *CommentOption, last *Comment) (string, error) {
	totalParts := len(parts)
	var firstCommentURL string

	// Delete old split comments if updating
	if opt != nil && opt.Update {
		oldSplitComments, err := g.findSplitComments(meta)
		if err != nil {
			return "", fmt.Errorf("failed to find old split comments: %w", err)
		}
		for _, c := range oldSplitComments {
			if err := g.DeleteComment(c); err != nil {
				return "", fmt.Errorf("failed to delete old split comment: %w", err)
			}
		}
	}

	// Create first comment (PR Review Comment if target specified)
	meta.TotalParts = totalParts
	meta.PartNumber = 1
	firstBody := parts[0] + "\n" + meta.ToHTML()
	url, err := g.createComment(firstBody, target)
	if err != nil {
		return "", fmt.Errorf("failed to create first part comment: %w", err)
	}
	firstCommentURL = url

	// Create subsequent parts as issue comments (replies)
	for i := 1; i < totalParts; i++ {
		meta.PartNumber = i + 1
		partBody := parts[i] + "\n" + meta.ToHTML()

		// Reply as issue comment
		_, err := gh.CreateIssueComment(g.ctx, g.client, g.Repository, g.Target, partBody)
		if err != nil {
			return "", fmt.Errorf("failed to create part %d/%d comment: %w", i+1, totalParts, err)
		}
	}

	return firstCommentURL, nil
}

func (g *GitHubReviewer) findSplitComments(meta MetaData) (Comments, error) {
	comments, err := g.ListComments(meta)
	if err != nil {
		return nil, err
	}

	var splitComments Comments
	for _, c := range comments {
		if c.MetaData.TotalParts > 0 {
			splitComments = append(splitComments, c)
		}
	}
	return splitComments, nil
}
