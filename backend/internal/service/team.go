package service

import (
	"encoding/json"
	"time"
)

// Team 团队领域模型
type Team struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Timezone     string                 `json:"timezone"`
	Language     string                 `json:"language"`
	OwnerID      int64                  `json:"owner_user_id"`
	BillingEmail *string                `json:"billing_email,omitempty"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	Status       string                 `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeletedAt    *time.Time             `json:"deleted_at,omitempty"`
}

// TeamMember 团队成员领域模型
type TeamMember struct {
	ID           int64
	TeamID       int64
	UserID       int64
	Role         string
	Status       string
	DisplayName  string
	DepartmentID *int64
	ConsumerID   *int64
	JoinedAt     time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Department 部门领域模型
type Department struct {
	ID             int64      `json:"id"`
	TeamID         int64      `json:"team_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CostCenterCode *string    `json:"cost_center_code"`
	ParentID       *int64     `json:"parent_id"`
	Level          int        `json:"level"`
	Path           string     `json:"path"`
	ExternalID     *string    `json:"external_id"`
	Source         string     `json:"source"`
	SortOrder      int        `json:"sort_order"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// Consumer 消费者领域模型
type Consumer struct {
	ID             int64                  `json:"id"`
	TeamID         int64                  `json:"team_id"`
	DepartmentID   *int64                 `json:"department_id"`
	Type           string                 `json:"type"`
	Name           string                 `json:"name"`
	Email          *string                `json:"email"`
	Phone          *string                `json:"phone"`
	Title          *string                `json:"title"`
	AppID          *string                `json:"app_id"`
	AppDescription *string                `json:"app_description"`
	ExternalID     *string                `json:"external_id"`
	Source         string                 `json:"source"`
	DeactivatedAt  *time.Time             `json:"deactivated_at"`
	Status         string                 `json:"status"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
}

// Description returns app_description for backward compatibility.
func (c Consumer) MarshalJSON() ([]byte, error) {
	type Alias Consumer
	return json.Marshal(&struct {
		Alias
		Description *string `json:"description"`
	}{
		Alias:       Alias(c),
		Description: c.AppDescription,
	})
}

// TeamAuditLog 团队审计日志领域模型
type TeamAuditLog struct {
	ID            int64
	TeamID        int64
	UserID        *int64
	Action        string
	OperationType *string
	ResourceType  string
	ResourceID    *int64
	Changes       map[string]interface{}
	IP            *string
	UserAgent     *string
	CreatedAt     time.Time
}

// TeamUsageTeamDaily 团队每日聚合数据
type TeamUsageTeamDaily struct {
	ID                  int64
	TeamID              int64
	BucketDate          time.Time
	TotalRequests       int64
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
	TotalCost           float64
	ActualCost          float64
	ComputedAt          time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TeamUsageDeptDaily 部门每日聚合数据
type TeamUsageDeptDaily struct {
	ID                  int64
	TeamID              int64
	DepartmentID        int64
	DepartmentName      *string
	CostCenterCode      *string
	BucketDate          time.Time
	TotalRequests       int64
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
	TotalCost           float64
	ActualCost          float64
	ComputedAt          time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TeamUsageConsumerDaily 消费者每日聚合数据
type TeamUsageConsumerDaily struct {
	ID                  int64
	TeamID              int64
	ConsumerID          int64
	ConsumerName        *string
	ConsumerType        *string
	BucketDate          time.Time
	TotalRequests       int64
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
	TotalCost           float64
	ActualCost          float64
	ComputedAt          time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TeamUsageModelDaily 模型每日聚合数据
type TeamUsageModelDaily struct {
	ID                  int64
	TeamID              int64
	DepartmentID        *int64
	ConsumerID          *int64
	BucketDate          time.Time
	ModelName           string
	TotalRequests       int64
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
	TotalCost           float64
	ActualCost          float64
	ComputedAt          time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TeamOverview 团队概览统计
type TeamOverview struct {
	TeamID        int64   `json:"team_id"`
	TotalRequests int64   `json:"total_requests"`
	InputTokens   int64   `json:"input_tokens"`
	OutputTokens  int64   `json:"output_tokens"`
	TotalCost     float64 `json:"total_cost"`
	ActualCost    float64 `json:"actual_cost"`
}

// DeptRankingItem 部门排名项
type DeptRankingItem struct {
	DepartmentID   int64   `json:"department_id"`
	DepartmentName string  `json:"department_name"`
	TotalRequests  int64   `json:"total_requests"`
	InputTokens    int64   `json:"input_tokens"`
	OutputTokens   int64   `json:"output_tokens"`
	TotalCost      float64 `json:"total_cost"`
	ActualCost     float64 `json:"actual_cost"`
}

// ConsumerRankingItem 消费者排名项
type ConsumerRankingItem struct {
	ConsumerID    int64   `json:"consumer_id"`
	ConsumerName  string  `json:"consumer_name"`
	ConsumerType  string  `json:"consumer_type"`
	TotalRequests int64   `json:"total_requests"`
	InputTokens   int64   `json:"input_tokens"`
	OutputTokens  int64   `json:"output_tokens"`
	TotalCost     float64 `json:"total_cost"`
	ActualCost    float64 `json:"actual_cost"`
}
