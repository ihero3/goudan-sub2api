package repository

import (
	"context"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/blog"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

type blogRepository struct {
	client *dbent.Client
}

func NewBlogRepository(client *dbent.Client) service.BlogRepository {
	return &blogRepository{client: client}
}

func (r *blogRepository) Create(ctx context.Context, b *service.Blog) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Blog.Create().
		SetTitle(b.Title).
		SetContent(b.Content).
		SetSummary(b.Summary).
		SetCoverImage(b.CoverImage).
		SetStatus(b.Status).
		SetTags(b.Tags)

	if b.PublishedAt != nil {
		builder.SetPublishedAt(*b.PublishedAt)
	}
	if b.CreatedBy != nil {
		builder.SetCreatedBy(*b.CreatedBy)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	applyBlogEntityToService(b, created)
	return nil
}

func (r *blogRepository) GetByID(ctx context.Context, id int64) (*service.Blog, error) {
	m, err := r.client.Blog.Query().
		Where(blog.IDEQ(id), blog.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrBlogNotFound, nil)
	}
	return blogEntityToService(m), nil
}

func (r *blogRepository) Update(ctx context.Context, b *service.Blog) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Blog.UpdateOneID(b.ID).
		SetTitle(b.Title).
		SetContent(b.Content).
		SetSummary(b.Summary).
		SetCoverImage(b.CoverImage).
		SetStatus(b.Status).
		SetTags(b.Tags)

	if b.PublishedAt != nil {
		builder.SetPublishedAt(*b.PublishedAt)
	} else {
		builder.ClearPublishedAt()
	}
	if b.CreatedBy != nil {
		builder.SetCreatedBy(*b.CreatedBy)
	} else {
		builder.ClearCreatedBy()
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrBlogNotFound, nil)
	}

	b.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *blogRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.Blog.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Save(ctx)
	return translatePersistenceError(err, service.ErrBlogNotFound, nil)
}

func (r *blogRepository) List(
	ctx context.Context,
	params pagination.PaginationParams,
	filters service.BlogListFilters,
) ([]service.Blog, *pagination.PaginationResult, error) {
	q := r.client.Blog.Query().
		Where(blog.DeletedAtIsNil())

	if filters.Status != "" {
		q = q.Where(blog.StatusEQ(filters.Status))
	}
	if filters.Search != "" {
		q = q.Where(
			blog.Or(
				blog.TitleContainsFold(filters.Search),
				blog.ContentContainsFold(filters.Search),
				blog.SummaryContainsFold(filters.Search),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	itemsQuery := q.
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range blogListOrders(params) {
		itemsQuery = itemsQuery.Order(order)
	}

	items, err := itemsQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := blogEntitiesToService(items)
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *blogRepository) ListPublished(
	ctx context.Context,
	params pagination.PaginationParams,
) ([]service.Blog, *pagination.PaginationResult, error) {
	q := r.client.Blog.Query().
		Where(blog.StatusEQ(service.BlogStatusPublished), blog.DeletedAtIsNil())

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	items, err := q.
		Order(dbent.Desc(blog.FieldPublishedAt), dbent.Desc(blog.FieldID)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := blogEntitiesToService(items)
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *blogRepository) ListPublishedByTag(
	ctx context.Context,
	tag string,
	params pagination.PaginationParams,
) ([]service.Blog, *pagination.PaginationResult, error) {
	q := r.client.Blog.Query().
		Where(
			blog.StatusEQ(service.BlogStatusPublished),
			blog.DeletedAtIsNil(),
			blog.TagsContains(tag),
		)

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	items, err := q.
		Order(dbent.Desc(blog.FieldPublishedAt), dbent.Desc(blog.FieldID)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := blogEntitiesToService(items)
	return out, paginationResultFromTotal(int64(total), params), nil
}

func blogListOrder(params pagination.PaginationParams) (string, string) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderDesc)

	switch sortBy {
	case "title":
		return blog.FieldTitle, sortOrder
	case "status":
		return blog.FieldStatus, sortOrder
	case "published_at":
		return blog.FieldPublishedAt, sortOrder
	case "id":
		return blog.FieldID, sortOrder
	case "", "created_at":
		return blog.FieldCreatedAt, sortOrder
	default:
		return blog.FieldCreatedAt, pagination.SortOrderDesc
	}
}

func blogListOrders(params pagination.PaginationParams) []func(*entsql.Selector) {
	field, sortOrder := blogListOrder(params)

	if sortOrder == pagination.SortOrderAsc {
		if field == blog.FieldID {
			return []func(*entsql.Selector){
				dbent.Asc(field),
			}
		}
		return []func(*entsql.Selector){
			dbent.Asc(field),
			dbent.Asc(blog.FieldID),
		}
	}

	if field == blog.FieldID {
		return []func(*entsql.Selector){
			dbent.Desc(field),
		}
	}
	return []func(*entsql.Selector){
		dbent.Desc(field),
		dbent.Desc(blog.FieldID),
	}
}

func applyBlogEntityToService(dst *service.Blog, src *dbent.Blog) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}

func blogEntityToService(m *dbent.Blog) *service.Blog {
	if m == nil {
		return nil
	}
	return &service.Blog{
		ID:          m.ID,
		Title:       m.Title,
		Content:     m.Content,
		Summary:     m.Summary,
		CoverImage:  m.CoverImage,
		Status:      m.Status,
		Tags:        m.Tags,
		PublishedAt: m.PublishedAt,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func blogEntitiesToService(models []*dbent.Blog) []service.Blog {
	out := make([]service.Blog, 0, len(models))
	for i := range models {
		if s := blogEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
