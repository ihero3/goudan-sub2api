package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type Blog struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	CoverImage  string    `json:"cover_image"`
	Status      string    `json:"status"`
	Tags        string    `json:"tags"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedBy   *int64    `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserBlog struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	CoverImage  string    `json:"cover_image"`
	Tags        string    `json:"tags"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func BlogFromService(b *service.Blog) *Blog {
	if b == nil {
		return nil
	}
	return &Blog{
		ID:          b.ID,
		Title:       b.Title,
		Content:     b.Content,
		Summary:     b.Summary,
		CoverImage:  b.CoverImage,
		Status:      b.Status,
		Tags:        b.Tags,
		PublishedAt: b.PublishedAt,
		CreatedBy:   b.CreatedBy,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func UserBlogFromService(b *service.Blog) *UserBlog {
	if b == nil {
		return nil
	}
	return &UserBlog{
		ID:          b.ID,
		Title:       b.Title,
		Content:     b.Content,
		Summary:     b.Summary,
		CoverImage:  b.CoverImage,
		Tags:        b.Tags,
		PublishedAt: b.PublishedAt,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}
