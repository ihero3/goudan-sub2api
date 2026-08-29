package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// BlogHandler handles admin blog management
type BlogHandler struct {
	blogService *service.BlogService
}

func NewBlogHandler(blogService *service.BlogService) *BlogHandler {
	return &BlogHandler{blogService: blogService}
}

type CreateBlogRequest struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Summary     string `json:"summary"`
	CoverImage  string `json:"cover_image"`
	Status      string `json:"status" binding:"omitempty,oneof=draft published"`
	Tags        string `json:"tags"`
	PublishedAt *int64 `json:"published_at"` // Unix seconds, 0 = auto set on publish
}

type UpdateBlogRequest struct {
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	Summary     *string `json:"summary"`
	CoverImage  *string `json:"cover_image"`
	Status      *string `json:"status" binding:"omitempty,oneof=draft published"`
	Tags        *string `json:"tags"`
	PublishedAt *int64  `json:"published_at"` // Unix seconds, 0 = clear
}

// List handles listing blogs with filters
// GET /api/v1/admin/blogs
func (h *BlogHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	if len(search) > 200 {
		search = search[:200]
	}

	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	items, paginationResult, err := h.blogService.List(
		c.Request.Context(),
		params,
		service.BlogListFilters{Status: status, Search: search},
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.Blog, 0, len(items))
	for i := range items {
		out = append(out, *dto.BlogFromService(&items[i]))
	}
	response.Paginated(c, out, paginationResult.Total, page, pageSize)
}

// GetByID handles getting a blog by ID
// GET /api/v1/admin/blogs/:id
func (h *BlogHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid blog ID")
		return
	}

	item, err := h.blogService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.BlogFromService(item))
}

// Create handles creating a new blog
// POST /api/v1/admin/blogs
func (h *BlogHandler) Create(c *gin.Context) {
	var req CreateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	input := &service.CreateBlogInput{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Status:     req.Status,
		Tags:       req.Tags,
		ActorID:    &subject.UserID,
	}

	if req.PublishedAt != nil && *req.PublishedAt > 0 {
		t := time.Unix(*req.PublishedAt, 0)
		input.PublishedAt = &t
	}

	created, err := h.blogService.Create(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.BlogFromService(created))
}

// Update handles updating a blog
// PUT /api/v1/admin/blogs/:id
func (h *BlogHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid blog ID")
		return
	}

	var req UpdateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	input := &service.UpdateBlogInput{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Status:     req.Status,
		Tags:       req.Tags,
		ActorID:    &subject.UserID,
	}

	if req.PublishedAt != nil {
		if *req.PublishedAt == 0 {
			var cleared *time.Time
			input.PublishedAt = &cleared
		} else {
			t := time.Unix(*req.PublishedAt, 0)
			ptr := &t
			input.PublishedAt = &ptr
		}
	}

	updated, err := h.blogService.Update(c.Request.Context(), id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.BlogFromService(updated))
}

// Delete handles deleting a blog
// DELETE /api/v1/admin/blogs/:id
func (h *BlogHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid blog ID")
		return
	}

	if err := h.blogService.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Blog deleted successfully"})
}
