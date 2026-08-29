package handler

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// BlogHandler handles public blog operations
type BlogHandler struct {
	blogService *service.BlogService
}

func NewBlogHandler(blogService *service.BlogService) *BlogHandler {
	return &BlogHandler{blogService: blogService}
}

// List handles listing published blogs
// GET /api/v1/blogs
func (h *BlogHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	sortBy := c.DefaultQuery("sort_by", "published_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	tag := strings.TrimSpace(c.Query("tag"))

	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	var (
		items           []service.Blog
		paginationResult *pagination.PaginationResult
		err             error
	)
	if tag != "" {
		items, paginationResult, err = h.blogService.ListPublishedByTag(c.Request.Context(), tag, params)
	} else {
		items, paginationResult, err = h.blogService.ListPublished(c.Request.Context(), params)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UserBlog, 0, len(items))
	for i := range items {
		out = append(out, *dto.UserBlogFromService(&items[i]))
	}
	response.Paginated(c, out, paginationResult.Total, page, pageSize)
}

// GetByID handles getting a published blog by ID
// GET /api/v1/blogs/:id
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
	if item.Status != service.BlogStatusPublished {
		response.NotFound(c, "Blog not found")
		return
	}

	response.Success(c, dto.UserBlogFromService(item))
}
