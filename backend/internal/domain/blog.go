package domain

import (
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	BlogStatusDraft     = "draft"
	BlogStatusPublished = "published"
)

var (
	ErrBlogNotFound      = infraerrors.NotFound("BLOG_NOT_FOUND", "blog not found")
	ErrBlogInvalidStatus = infraerrors.BadRequest("BLOG_INVALID_STATUS", "invalid blog status")
)

func NormalizeBlogStatus(status string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "", BlogStatusDraft:
		return BlogStatusDraft, nil
	case BlogStatusPublished:
		return BlogStatusPublished, nil
	default:
		return "", ErrBlogInvalidStatus
	}
}

type Blog struct {
	ID          int64
	Title       string
	Content     string
	Summary     string
	CoverImage  string
	Status      string
	Tags        string
	PublishedAt *time.Time
	CreatedBy   *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
