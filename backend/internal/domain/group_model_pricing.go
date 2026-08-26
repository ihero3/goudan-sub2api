package domain

import "time"

// BillingMode 计费模式
type BillingMode string

// ChannelModelPricing 渠道模型定价条目（JSON 镜像，供 ent schema 与 DB JSONB 列使用）
type ChannelModelPricing struct {
	ID               int64               `json:"id,omitempty"`
	ChannelID        int64               `json:"channel_id,omitempty"`
	Platform         string              `json:"platform"`
	Models           []string            `json:"models"`
	BillingMode      BillingMode         `json:"billing_mode"`
	InputPrice       *float64            `json:"input_price"`
	OutputPrice      *float64            `json:"output_price"`
	CacheWritePrice  *float64            `json:"cache_write_price"`
	CacheReadPrice   *float64            `json:"cache_read_price"`
	ImageInputPrice  *float64            `json:"image_input_price"`
	ImageOutputPrice *float64            `json:"image_output_price"`
	PerRequestPrice  *float64            `json:"per_request_price"`
	Intervals        []PricingInterval   `json:"intervals"`
	TimePricing      *ChannelTimePricing `json:"time_pricing,omitempty"`
	CreatedAt        time.Time           `json:"created_at,omitempty"`
	UpdatedAt        time.Time           `json:"updated_at,omitempty"`
}

// ChannelTimePricing 渠道模型定价的分时倍率配置。
type ChannelTimePricing struct {
	Timezone string                     `json:"timezone"`
	Periods  []ChannelTimePricingPeriod `json:"periods"`
}

// ChannelTimePricingPeriod 是秒级的左闭右开分时倍率区间，并兼容历史 HH:mm 数据。
type ChannelTimePricingPeriod struct {
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Multiplier float64 `json:"multiplier"`
}

// PricingInterval 定价区间（token 区间 / 按次分层 / 图片分辨率分层）
type PricingInterval struct {
	ID              int64     `json:"id,omitempty"`
	PricingID       int64     `json:"pricing_id,omitempty"`
	MinTokens       int       `json:"min_tokens"`
	MaxTokens       *int      `json:"max_tokens"`
	TierLabel       string    `json:"tier_label"`
	InputPrice      *float64  `json:"input_price"`
	OutputPrice     *float64  `json:"output_price"`
	CacheWritePrice *float64  `json:"cache_write_price"`
	CacheReadPrice  *float64  `json:"cache_read_price"`
	PerRequestPrice *float64  `json:"per_request_price"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}
