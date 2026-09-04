package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

// mockVideoTaskRepo 最小化实现 VideoTaskRepo 接口，用于单元测试。
type mockVideoTaskRepo struct {
	mu        sync.Mutex
	nextID    int64
	byLocal   map[string]*VideoTaskRecord
	byID      map[int64]*VideoTaskRecord
	byUser    map[int64][]*VideoTaskRecord
	createErr error
	updateErr error
	getErr    error
}

func newMockVideoTaskRepo() *mockVideoTaskRepo {
	return &mockVideoTaskRepo{
		nextID:  1,
		byLocal: make(map[string]*VideoTaskRecord),
		byID:    make(map[int64]*VideoTaskRecord),
		byUser:  make(map[int64][]*VideoTaskRecord),
	}
}

func (m *mockVideoTaskRepo) Create(ctx context.Context, t *VideoTaskRecord) (*VideoTaskRecord, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	t.ID = m.nextID
	m.nextID++
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = time.Now()
	cp := *t
	m.byLocal[t.LocalID] = &cp
	m.byID[t.ID] = &cp
	m.byUser[t.UserID] = append(m.byUser[t.UserID], &cp)
	return &cp, nil
}

func (m *mockVideoTaskRepo) GetByLocalID(ctx context.Context, localID string) (*VideoTaskRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.byLocal[localID]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *v
	return &cp, nil
}

func (m *mockVideoTaskRepo) GetByID(ctx context.Context, id int64) (*VideoTaskRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *v
	return &cp, nil
}

func (m *mockVideoTaskRepo) UpdateStatus(ctx context.Context, id int64, status, errorMsg string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.byID[id]
	if !ok {
		return errors.New("not found")
	}
	v.Status = status
	v.ErrorMessage = errorMsg
	v.UpdatedAt = time.Now()
	if status == "succeeded" || status == "failed" || status == "cancelled" {
		now := time.Now()
		v.FinishedAt = &now
	}
	return nil
}

func (m *mockVideoTaskRepo) UpdateResult(ctx context.Context, id int64, status, videoURL, thumbnailURL string, durationSec int, costUSD float64) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.byID[id]
	if !ok {
		return errors.New("not found")
	}
	v.Status = status
	v.VideoURL = videoURL
	v.ThumbnailURL = thumbnailURL
	v.DurationSec = durationSec
	v.CostUSD = costUSD
	v.UpdatedAt = time.Now()
	if status == "succeeded" || status == "failed" || status == "cancelled" {
		now := time.Now()
		v.FinishedAt = &now
	}
	return nil
}

func (m *mockVideoTaskRepo) UpdateUpstreamTaskID(ctx context.Context, id int64, upstreamTaskID string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.byID[id]
	if !ok {
		return errors.New("not found")
	}
	v.UpstreamTaskID = upstreamTaskID
	v.UpdatedAt = time.Now()
	return nil
}

func (m *mockVideoTaskRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*VideoTaskRecord, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := m.byUser[userID]
	total := len(items)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	out := make([]*VideoTaskRecord, 0, end-start)
	for _, t := range items[start:end] {
		cp := *t
		out = append(out, &cp)
	}
	return out, total, nil
}

func (m *mockVideoTaskRepo) ListProcessingTasks(ctx context.Context, before time.Time, limit int) ([]*VideoTaskRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*VideoTaskRecord
	for _, t := range m.byID {
		if t.Status == "processing" {
			cp := *t
			out = append(out, &cp)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
