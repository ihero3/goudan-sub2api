package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrConsumerNotFound      = infraerrors.NotFound("CONSUMER_NOT_FOUND", "consumer not found")
	ErrConsumerNameEmpty     = infraerrors.BadRequest("CONSUMER_NAME_EMPTY", "consumer name cannot be empty")
	ErrInvalidConsumerID     = infraerrors.BadRequest("INVALID_CONSUMER_ID", "invalid consumer id")
	ErrInvalidTeamIDForConsumer = infraerrors.BadRequest("INVALID_TEAM_ID", "invalid team id for consumer")
	ErrInvalidDeptIDForConsumer = infraerrors.BadRequest("INVALID_DEPT_ID", "invalid department id for consumer")
)

// ConsumerRepository 消费者数据访问接口（在 service 包内定义以避免循环依赖）
type ConsumerRepository interface {
	Create(ctx context.Context, c *Consumer) error
	Update(ctx context.Context, c *Consumer) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Consumer, error)
	ListByTeam(ctx context.Context, teamID int64) ([]*Consumer, error)
	ListByDepartment(ctx context.Context, departmentID int64) ([]*Consumer, error)
}

// ConsumerService 消费者服务接口
type ConsumerService interface {
	CreateConsumer(ctx context.Context, teamID int64, deptID *int64, name, email, phone, title, consumerType, description string) (*Consumer, error)
	UpdateConsumer(ctx context.Context, consumerID int64, name string, email, phone, title, consumerType, description *string, deptID *int64, status *string) (*Consumer, error)
	DeleteConsumer(ctx context.Context, consumerID int64) error
	GetConsumer(ctx context.Context, consumerID int64) (*Consumer, error)
	ListConsumers(ctx context.Context, teamID int64, deptID *int64) ([]*Consumer, error)
	GetConsumersByDepartment(ctx context.Context, deptID int64) ([]*Consumer, error)
}

// consumerService 消费者服务实现
type consumerService struct {
	consumerRepo ConsumerRepository
}

// NewConsumerService 创建消费者服务实例
func NewConsumerService(consumerRepo ConsumerRepository) ConsumerService {
	return &consumerService{
		consumerRepo: consumerRepo,
	}
}

// CreateConsumer 创建消费者
func (s *consumerService) CreateConsumer(ctx context.Context, teamID int64, deptID *int64, name, email, phone, title, consumerType, description string) (*Consumer, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForConsumer
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrConsumerNameEmpty
	}

	consumer := &Consumer{
		TeamID:   teamID,
		Name:     strings.TrimSpace(name),
		Type:     consumerType,
		Status:   "active",
		Source:   "manual",
		Settings: map[string]interface{}{},
	}

	if deptID != nil && *deptID > 0 {
		consumer.DepartmentID = deptID
	}
	if email != "" {
		e := strings.TrimSpace(email)
		consumer.Email = &e
	}
	if phone != "" {
		p := strings.TrimSpace(phone)
		consumer.Phone = &p
	}
	if title != "" {
		t := strings.TrimSpace(title)
		consumer.Title = &t
	}
	if description != "" {
		d := strings.TrimSpace(description)
		consumer.AppDescription = &d
	}

	if err := s.consumerRepo.Create(ctx, consumer); err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}

	return consumer, nil
}

// UpdateConsumer 更新消费者信息
func (s *consumerService) UpdateConsumer(ctx context.Context, consumerID int64, name string, email, phone, title, consumerType, description *string, deptID *int64, status *string) (*Consumer, error) {
	if consumerID <= 0 {
		return nil, ErrInvalidConsumerID
	}

	consumer, err := s.consumerRepo.GetByID(ctx, consumerID)
	if err != nil {
		return nil, fmt.Errorf("get consumer: %w", err)
	}
	if consumer == nil {
		return nil, ErrConsumerNotFound
	}

	if name != "" {
		consumer.Name = strings.TrimSpace(name)
	}
	if email != nil {
		if *email == "" {
			consumer.Email = nil
		} else {
			e := strings.TrimSpace(*email)
			consumer.Email = &e
		}
	}
	if description != nil {
		if *description == "" {
			consumer.AppDescription = nil
		} else {
			desc := strings.TrimSpace(*description)
			consumer.AppDescription = &desc
		}
	}
	if phone != nil {
		if *phone == "" {
			consumer.Phone = nil
		} else {
			p := strings.TrimSpace(*phone)
			consumer.Phone = &p
		}
	}
	if title != nil {
		if *title == "" {
			consumer.Title = nil
		} else {
			ti := strings.TrimSpace(*title)
			consumer.Title = &ti
		}
	}
	if consumerType != nil && *consumerType != "" {
		consumer.Type = strings.TrimSpace(*consumerType)
	}
	if deptID != nil {
		consumer.DepartmentID = deptID
	}
	if status != nil && *status != "" {
		newStatus := strings.TrimSpace(*status)
		if newStatus != consumer.Status {
			consumer.Status = newStatus
			if newStatus == "inactive" {
				now := time.Now()
				consumer.DeactivatedAt = &now
			} else if newStatus == "active" {
				consumer.DeactivatedAt = nil
			}
		}
	}

	if err := s.consumerRepo.Update(ctx, consumer); err != nil {
		return nil, fmt.Errorf("update consumer: %w", err)
	}

	return consumer, nil
}

// DeleteConsumer 删除消费者（软删除）
func (s *consumerService) DeleteConsumer(ctx context.Context, consumerID int64) error {
	if consumerID <= 0 {
		return ErrInvalidConsumerID
	}

	consumer, err := s.consumerRepo.GetByID(ctx, consumerID)
	if err != nil {
		return fmt.Errorf("get consumer: %w", err)
	}
	if consumer == nil {
		return ErrConsumerNotFound
	}

	if err := s.consumerRepo.Delete(ctx, consumerID); err != nil {
		return fmt.Errorf("delete consumer: %w", err)
	}

	return nil
}

// GetConsumer 获取消费者详情
func (s *consumerService) GetConsumer(ctx context.Context, consumerID int64) (*Consumer, error) {
	if consumerID <= 0 {
		return nil, ErrInvalidConsumerID
	}

	consumer, err := s.consumerRepo.GetByID(ctx, consumerID)
	if err != nil {
		return nil, fmt.Errorf("get consumer: %w", err)
	}
	if consumer == nil {
		return nil, ErrConsumerNotFound
	}

	return consumer, nil
}

// ListConsumers 列出团队下的消费者
func (s *consumerService) ListConsumers(ctx context.Context, teamID int64, deptID *int64) ([]*Consumer, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForConsumer
	}

	var consumers []*Consumer
	var err error

	if deptID != nil && *deptID > 0 {
		consumers, err = s.consumerRepo.ListByDepartment(ctx, *deptID)
	} else {
		consumers, err = s.consumerRepo.ListByTeam(ctx, teamID)
	}

	if err != nil {
		return nil, fmt.Errorf("list consumers: %w", err)
	}

	return consumers, nil
}

// GetConsumersByDepartment 获取指定部门下的所有消费者
func (s *consumerService) GetConsumersByDepartment(ctx context.Context, deptID int64) ([]*Consumer, error) {
	if deptID <= 0 {
		return nil, ErrInvalidDeptIDForConsumer
	}

	consumers, err := s.consumerRepo.ListByDepartment(ctx, deptID)
	if err != nil {
		return nil, fmt.Errorf("list consumers by department: %w", err)
	}

	return consumers, nil
}
