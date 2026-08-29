package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	BlogStatusDraft     = domain.BlogStatusDraft
	BlogStatusPublished = domain.BlogStatusPublished
)

var (
	ErrBlogNotFound        = domain.ErrBlogNotFound
	ErrBlogInvalidStatus   = domain.ErrBlogInvalidStatus
	ErrBlogNilInput        = infraerrors.BadRequest("BLOG_INPUT_REQUIRED", "blog input is required")
	ErrBlogInvalidTitle    = infraerrors.BadRequest("BLOG_TITLE_INVALID", "blog title is invalid")
	ErrBlogContentRequired = infraerrors.BadRequest("BLOG_CONTENT_REQUIRED", "blog content is required")
)

// NormalizeBlogStatus 校验并规范化博客状态
func NormalizeBlogStatus(status string) (string, error) {
	return domain.NormalizeBlogStatus(status)
}

type Blog = domain.Blog

type BlogListFilters struct {
	Status string
	Search string
}

type BlogRepository interface {
	Create(ctx context.Context, b *Blog) error
	GetByID(ctx context.Context, id int64) (*Blog, error)
	Update(ctx context.Context, b *Blog) error
	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, params pagination.PaginationParams, filters BlogListFilters) ([]Blog, *pagination.PaginationResult, error)
	ListPublished(ctx context.Context, params pagination.PaginationParams) ([]Blog, *pagination.PaginationResult, error)
	ListPublishedByTag(ctx context.Context, tag string, params pagination.PaginationParams) ([]Blog, *pagination.PaginationResult, error)
}
