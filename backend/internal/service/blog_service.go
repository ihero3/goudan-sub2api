package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type BlogService struct {
	blogRepo BlogRepository
}

func NewBlogService(blogRepo BlogRepository) *BlogService {
	return &BlogService{blogRepo: blogRepo}
}

type CreateBlogInput struct {
	Title       string
	Content     string
	Summary     string
	CoverImage  string
	Status      string
	Tags        string
	PublishedAt *time.Time
	ActorID     *int64
}

type UpdateBlogInput struct {
	Title       *string
	Content     *string
	Summary     *string
	CoverImage  *string
	Status      *string
	Tags        *string
	PublishedAt **time.Time
	ActorID     *int64
}

func (s *BlogService) Create(ctx context.Context, input *CreateBlogInput) (*Blog, error) {
	if input == nil {
		return nil, ErrBlogNilInput
	}

	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || len(title) > 500 {
		return nil, ErrBlogInvalidTitle
	}
	if content == "" {
		return nil, ErrBlogContentRequired
	}

	status, err := NormalizeBlogStatus(input.Status)
	if err != nil {
		return nil, err
	}

	summary := strings.TrimSpace(input.Summary)
	if len(summary) > 1000 {
		return nil, ErrBlogInvalidStatus
	}
	coverImage := strings.TrimSpace(input.CoverImage)
	if len(coverImage) > 1000 {
		return nil, ErrBlogInvalidStatus
	}
	tags := strings.TrimSpace(input.Tags)

	publishedAt := input.PublishedAt
	if status == BlogStatusPublished && publishedAt == nil {
		now := time.Now()
		publishedAt = &now
	}
	if status == BlogStatusDraft {
		publishedAt = nil
	}

	b := &Blog{
		Title:       title,
		Content:     content,
		Summary:     summary,
		CoverImage:  coverImage,
		Status:      status,
		Tags:        tags,
		PublishedAt: publishedAt,
	}
	if input.ActorID != nil && *input.ActorID > 0 {
		b.CreatedBy = input.ActorID
	}

	if err := s.blogRepo.Create(ctx, b); err != nil {
		return nil, fmt.Errorf("create blog: %w", err)
	}
	return b, nil
}

func (s *BlogService) Update(ctx context.Context, id int64, input *UpdateBlogInput) (*Blog, error) {
	if input == nil {
		return nil, ErrBlogNilInput
	}

	b, err := s.blogRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" || len(title) > 500 {
			return nil, ErrBlogInvalidTitle
		}
		b.Title = title
	}
	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, ErrBlogContentRequired
		}
		b.Content = content
	}
	if input.Summary != nil {
		summary := strings.TrimSpace(*input.Summary)
		if len(summary) > 1000 {
			return nil, ErrBlogInvalidStatus
		}
		b.Summary = summary
	}
	if input.CoverImage != nil {
		coverImage := strings.TrimSpace(*input.CoverImage)
		if len(coverImage) > 1000 {
			return nil, ErrBlogInvalidStatus
		}
		b.CoverImage = coverImage
	}
	if input.Tags != nil {
		b.Tags = strings.TrimSpace(*input.Tags)
	}

	if input.Status != nil {
		status, err := NormalizeBlogStatus(*input.Status)
		if err != nil {
			return nil, err
		}
		b.Status = status
	}

	// 处理发布时间
	if input.PublishedAt != nil {
		// 显式传入指针（包含 nil 表示清空）
		b.PublishedAt = *input.PublishedAt
	}

	// 状态与发布时间联动
	if b.Status == BlogStatusPublished && b.PublishedAt == nil {
		now := time.Now()
		b.PublishedAt = &now
	}
	if b.Status == BlogStatusDraft {
		b.PublishedAt = nil
	}

	if err := s.blogRepo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("update blog: %w", err)
	}
	return b, nil
}

func (s *BlogService) Delete(ctx context.Context, id int64) error {
	if err := s.blogRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete blog: %w", err)
	}
	return nil
}

func (s *BlogService) GetByID(ctx context.Context, id int64) (*Blog, error) {
	return s.blogRepo.GetByID(ctx, id)
}

func (s *BlogService) List(ctx context.Context, params pagination.PaginationParams, filters BlogListFilters) ([]Blog, *pagination.PaginationResult, error) {
	return s.blogRepo.List(ctx, params, filters)
}

func (s *BlogService) ListPublished(ctx context.Context, params pagination.PaginationParams) ([]Blog, *pagination.PaginationResult, error) {
	return s.blogRepo.ListPublished(ctx, params)
}

func (s *BlogService) ListPublishedByTag(ctx context.Context, tag string, params pagination.PaginationParams) ([]Blog, *pagination.PaginationResult, error) {
	return s.blogRepo.ListPublishedByTag(ctx, strings.TrimSpace(tag), params)
}
