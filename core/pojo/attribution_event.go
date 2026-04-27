package pojo

import (
	"encoding/json"
	"time"
)

// AttributionEvent 通用事件归因明细。
type AttributionEvent struct {
	ID                int64     `json:"id" gorm:"primaryKey;"`
	CreatedAt         time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime;index:idx_attribution_event_name_time,priority:2;index:idx_attribution_channel_event_time,priority:3;index"`
	UpdatedAt         time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
	TenantID          int64     `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;default:0;index"`
	UserID            *int64    `json:"userId" gorm:"column:user_id;type:bigint;index"`
	VisitorID         string    `json:"visitorId" gorm:"column:visitor_id;type:varchar(128);index"`
	SessionID         string    `json:"sessionId" gorm:"column:session_id;type:varchar(128);index"`
	EventName         string    `json:"eventName" gorm:"column:event_name;type:varchar(64);not null;index:idx_attribution_event_name_time,priority:1;index:idx_attribution_channel_event_time,priority:2"`
	SourceChannelID   *int64    `json:"sourceChannelId" gorm:"column:source_channel_id;type:bigint;index:idx_attribution_channel_event_time,priority:1"`
	SourceChannelCode *string   `json:"sourceChannelCode" gorm:"column:source_channel_code;type:varchar(64);index"`
	PageURL           *string   `json:"pageUrl" gorm:"column:page_url;type:varchar(1024)"`
	Referrer          *string   `json:"referrer" gorm:"column:referrer;type:varchar(1024)"`
	IP                *string   `json:"ip" gorm:"column:ip;type:varchar(64);index"`
	UserAgent         *string   `json:"userAgent" gorm:"column:user_agent;type:varchar(512)"`
	Metadata          *string   `json:"metadata" gorm:"column:metadata;type:json"`
}

func (AttributionEvent) TableName() string {
	return "attribution_event"
}

type AttributionEventCreateReq struct {
	EventName         string          `json:"eventName"`
	SourceChannelCode string          `json:"sourceChannelCode"`
	ChannelCode       string          `json:"channelCode"`
	VisitorID         string          `json:"visitorId"`
	SessionID         string          `json:"sessionId"`
	PageURL           string          `json:"pageUrl"`
	Referrer          string          `json:"referrer"`
	Metadata          json.RawMessage `json:"metadata"`
}

type AttributionEventSearch struct {
	PageInfo
	EventName         string `json:"eventName"`
	SourceChannelID   int64  `json:"sourceChannelId"`
	SourceChannelCode string `json:"sourceChannelCode"`
	UserID            int64  `json:"userId"`
	VisitorID         string `json:"visitorId"`
	StartTime         int64  `json:"startTime"`
	EndTime           int64  `json:"endTime"`
}

type AttributionEventSummarySearch struct {
	EventName         string `json:"eventName"`
	SourceChannelID   int64  `json:"sourceChannelId"`
	SourceChannelCode string `json:"sourceChannelCode"`
	UserID            int64  `json:"userId"`
	VisitorID         string `json:"visitorId"`
	StartTime         int64  `json:"startTime"`
	EndTime           int64  `json:"endTime"`
}

type AttributionEventBack struct {
	ID                int64     `json:"id"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	TenantID          int64     `json:"tenantId"`
	UserID            *int64    `json:"userId"`
	VisitorID         string    `json:"visitorId"`
	SessionID         string    `json:"sessionId"`
	EventName         string    `json:"eventName"`
	SourceChannelID   *int64    `json:"sourceChannelId"`
	SourceChannelCode *string   `json:"sourceChannelCode"`
	PageURL           *string   `json:"pageUrl"`
	Referrer          *string   `json:"referrer"`
	IP                *string   `json:"ip"`
	UserAgent         *string   `json:"userAgent"`
	Metadata          *string   `json:"metadata"`
}

type AttributionEventResp struct {
	BasePageResponse[AttributionEventBack]
}

type AttributionEventSummaryBack struct {
	SourceChannelID   *int64  `json:"sourceChannelId"`
	SourceChannelCode *string `json:"sourceChannelCode"`
	EventName         string  `json:"eventName"`
	EventCount        int64   `json:"eventCount"`
	UniqueVisitors    int64   `json:"uniqueVisitors"`
	UniqueUsers       int64   `json:"uniqueUsers"`
}
