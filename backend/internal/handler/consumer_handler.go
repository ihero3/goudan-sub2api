package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ConsumerHandler handles consumer-related requests
type ConsumerHandler struct {
	consumerService service.ConsumerService
}

// NewConsumerHandler creates a new ConsumerHandler
func NewConsumerHandler(consumerService service.ConsumerService) *ConsumerHandler {
	return &ConsumerHandler{
		consumerService: consumerService,
	}
}

// CreateConsumerRequest represents the create consumer request payload
type CreateConsumerRequest struct {
	Name   string `json:"name" binding:"required"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	DeptID *int64 `json:"dept_id"`
}

// UpdateConsumerRequest represents the update consumer request payload
type UpdateConsumerRequest struct {
	Name        string  `json:"name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	Title       *string `json:"title"`
	Type        *string `json:"type"`
	Description *string `json:"description"`
	DeptID      *int64  `json:"dept_id"`
	Status      *string `json:"status"`
}

// List handles listing consumers for a team
// GET /api/v1/teams/:id/consumers
func (h *ConsumerHandler) List(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	var deptID *int64
	if didStr := c.Query("dept_id"); didStr != "" {
		did, err := strconv.ParseInt(didStr, 10, 64)
		if err == nil {
			deptID = &did
		}
	}

	consumers, err := h.consumerService.ListConsumers(c.Request.Context(), teamID, deptID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, consumers, int64(len(consumers)), 1, len(consumers))
}

// Get handles getting a single consumer
// GET /api/v1/teams/:team_id/consumers/:consumer_id
func (h *ConsumerHandler) Get(c *gin.Context) {
	consumerID, err := strconv.ParseInt(c.Param("consumer_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid consumer ID")
		return
	}

	consumer, err := h.consumerService.GetConsumer(c.Request.Context(), consumerID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, consumer)
}

// Create handles creating a new consumer
// POST /api/v1/teams/:id/consumers
func (h *ConsumerHandler) Create(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	var req CreateConsumerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "default"
	}

	consumer, err := h.consumerService.CreateConsumer(c.Request.Context(), teamID, req.DeptID, req.Name, req.Email, req.Type)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, consumer)
}

// Update handles updating a consumer
// PUT /api/v1/teams/:team_id/consumers/:consumer_id
func (h *ConsumerHandler) Update(c *gin.Context) {
	consumerID, err := strconv.ParseInt(c.Param("consumer_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid consumer ID")
		return
	}

	var req UpdateConsumerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	consumer, err := h.consumerService.UpdateConsumer(c.Request.Context(), consumerID, req.Name, req.Email, req.Phone, req.Title, req.Type, req.Description, req.DeptID, req.Status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, consumer)
}

// Delete handles deleting a consumer
// DELETE /api/v1/teams/:team_id/consumers/:consumer_id
func (h *ConsumerHandler) Delete(c *gin.Context) {
	consumerID, err := strconv.ParseInt(c.Param("consumer_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid consumer ID")
		return
	}

	if err := h.consumerService.DeleteConsumer(c.Request.Context(), consumerID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Consumer deleted successfully"})
}
