package service

import (
	"context"
	"testing"
	"time"
)

// TestVideoTaskRecord_CRUD 验证仓储接口 mock 满足 VideoTaskRepo 契约。
func TestVideoTaskRecord_CRUD(t *testing.T) {
	repo := newMockVideoTaskRepo()
	ctx := context.Background()

	// Create
	task := &VideoTaskRecord{
		LocalID:       "vid_test_001",
		UserID:        100,
		APIKeyID:      200,
		PublicModel:   "seedance-2.5",
		UpstreamModel: "doubao-seedance-2-5",
		AccountID:     50,
		Status:        "processing",
	}
	created, err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt after create")
	}

	// GetByLocalID
	got, err := repo.GetByLocalID(ctx, "vid_test_001")
	if err != nil {
		t.Fatalf("get by local_id failed: %v", err)
	}
	if got.PublicModel != "seedance-2.5" {
		t.Fatalf("unexpected public_model: %s", got.PublicModel)
	}

	// GetByID
	got2, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id failed: %v", err)
	}
	if got2.LocalID != "vid_test_001" {
		t.Fatalf("unexpected local_id: %s", got2.LocalID)
	}

	// UpdateStatus
	if err := repo.UpdateStatus(ctx, created.ID, "succeeded", ""); err != nil {
		t.Fatalf("update status failed: %v", err)
	}
	got3, _ := repo.GetByID(ctx, created.ID)
	if got3.Status != "succeeded" {
		t.Fatalf("expected status=succeeded, got %s", got3.Status)
	}
	if got3.FinishedAt == nil {
		t.Fatal("expected FinishedAt set after succeeded")
	}

	// UpdateResult
	if err := repo.UpdateResult(ctx, created.ID, "succeeded", "https://cdn.example.com/v.mp4", "https://cdn.example.com/v.jpg", 5, 0.05); err != nil {
		t.Fatalf("update result failed: %v", err)
	}
	got4, _ := repo.GetByID(ctx, created.ID)
	if got4.VideoURL != "https://cdn.example.com/v.mp4" {
		t.Fatalf("unexpected video_url: %s", got4.VideoURL)
	}
	if got4.CostUSD != 0.05 {
		t.Fatalf("unexpected cost_usd: %f", got4.CostUSD)
	}

	// UpdateUpstreamTaskID
	if err := repo.UpdateUpstreamTaskID(ctx, created.ID, "upstream_xxx"); err != nil {
		t.Fatalf("update upstream task id failed: %v", err)
	}
	got5, _ := repo.GetByID(ctx, created.ID)
	if got5.UpstreamTaskID != "upstream_xxx" {
		t.Fatalf("unexpected upstream_task_id: %s", got5.UpstreamTaskID)
	}

	// ListByUserID
	// Add more tasks for pagination
	for i := 0; i < 3; i++ {
		repo.Create(ctx, &VideoTaskRecord{
			LocalID:    "vid_test_pg_" + string(rune('a'+i)),
			UserID:     100,
			PublicModel: "seedance-2.5",
			Status:     "processing",
		})
	}
	items, total, err := repo.ListByUserID(ctx, 100, 2, 0)
	if err != nil {
		t.Fatalf("list by user_id failed: %v", err)
	}
	if total != 4 {
		t.Fatalf("expected total=4, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items in page 1, got %d", len(items))
	}

	// ListProcessingTasks
	processing, err := repo.ListProcessingTasks(ctx, time.Now(), 10)
	if err != nil {
		t.Fatalf("list processing failed: %v", err)
	}
	// 1 originally processing (now succeeded) + 3 new processing = 3
	if len(processing) != 3 {
		t.Fatalf("expected 3 processing tasks, got %d", len(processing))
	}
}

// TestVideoTaskRecord_CancelTask 测试取消任务的状态转换。
func TestVideoTaskRecord_CancelTask(t *testing.T) {
	repo := newMockVideoTaskRepo()
	ctx := context.Background()

	created, _ := repo.Create(ctx, &VideoTaskRecord{
		LocalID:    "vid_to_cancel",
		UserID:     1,
		PublicModel: "wan3.0-video",
		Status:     "processing",
	})
	if err := repo.UpdateStatus(ctx, created.ID, "cancelled", "cancelled by admin"); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	got, _ := repo.GetByID(ctx, created.ID)
	if got.Status != "cancelled" {
		t.Fatalf("expected status=cancelled, got %s", got.Status)
	}
	if got.FinishedAt == nil {
		t.Fatal("expected FinishedAt set after cancel")
	}
	if got.ErrorMessage != "cancelled by admin" {
		t.Fatalf("unexpected error message: %s", got.ErrorMessage)
	}
}
