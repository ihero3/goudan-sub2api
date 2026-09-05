package service

import "context"

// settleMediaReservedQuota 完成时结算：reserved 预估预扣 vs actual 实际费用（视频/媒体共用）。
// diff = reserved - actual；diff>0 退回差额，diff<0 补扣差额。
func settleMediaReservedQuota(apiKeySvc *APIKeyService, ctx context.Context, apiKeyID int64, reserved, actual float64) {
	if apiKeySvc == nil {
		return
	}
	if reserved <= 0 {
		// 无预扣：退化为完成时直接扣实际费用（兜底）。
		if actual > 0 {
			_ = apiKeySvc.UpdateQuotaUsed(ctx, apiKeyID, actual)
		}
		return
	}
	diff := reserved - actual
	if diff > 0 {
		_ = apiKeySvc.ReleaseQuotaUsed(ctx, apiKeyID, diff)
	} else if diff < 0 {
		_ = apiKeySvc.UpdateQuotaUsed(ctx, apiKeyID, -diff)
	}
}

// releaseMediaReservedQuota 失败/取消时全额退回预扣费用。
func releaseMediaReservedQuota(apiKeySvc *APIKeyService, ctx context.Context, apiKeyID int64, reserved float64) {
	if apiKeySvc == nil || reserved <= 0 {
		return
	}
	_ = apiKeySvc.ReleaseQuotaUsed(ctx, apiKeyID, reserved)
}
