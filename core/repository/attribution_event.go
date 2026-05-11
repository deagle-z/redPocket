package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

const attributionMetadataMaxLength = 4096

var attributionEventNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

func ValidateAttributionEventName(eventName string) (string, error) {
	eventName = strings.TrimSpace(eventName)
	if !attributionEventNameRegexp.MatchString(eventName) {
		return "", errors.New("invalid_event_name")
	}
	return eventName, nil
}

func CreateAttributionEvent(db *gorm.DB, req pojo.AttributionEventCreateReq, userID int64, tenantID int64, ip string, userAgent string) (pojo.AttributionEventBack, error) {
	var result pojo.AttributionEventBack
	if db == nil {
		return result, errors.New("invalid_params")
	}

	eventName, err := ValidateAttributionEventName(req.EventName)
	if err != nil {
		return result, err
	}

	var sourceChannelID *int64
	var sourceChannelCode *string
	channelCode := FirstSourceChannelCode(req.SourceChannelCode, req.ChannelCode)
	if channelCode != "" {
		channel, err := ResolveSourceChannelByCode(db, 0, channelCode)
		if err != nil {
			return result, err
		}
		if channel != nil {
			id := channel.ID
			code := channel.ChannelCode
			sourceChannelID = &id
			sourceChannelCode = &code
		}
	}

	var metadata *string
	if len(req.Metadata) > 0 && string(req.Metadata) != "null" {
		raw := strings.TrimSpace(string(req.Metadata))
		if len(raw) > attributionMetadataMaxLength {
			return result, errors.New("metadata_too_long")
		}
		metadata = &raw
	}

	var userIDPtr *int64
	if userID > 0 {
		userIDPtr = &userID
	}

	entity := pojo.AttributionEvent{
		TenantID:          tenantID,
		UserID:            userIDPtr,
		VisitorID:         trimMax(req.VisitorID, 128),
		SessionID:         trimMax(req.SessionID, 128),
		EventName:         eventName,
		ThirdPartyEventID: nullableTrimMax(req.ThirdPartyEventID, 128),
		PixelID:           nullableTrimMax(req.PixelID, 128),
		SourceChannelID:   sourceChannelID,
		SourceChannelCode: sourceChannelCode,
		PageURL:           nullableTrimMax(req.PageURL, 1024),
		Referrer:          nullableTrimMax(req.Referrer, 1024),
		IP:                nullableTrimMax(ip, 64),
		UserAgent:         nullableTrimMax(userAgent, 512),
		Metadata:          metadata,
	}
	if err := db.Create(&entity).Error; err != nil {
		return result, err
	}

	_ = copier.Copy(&result, &entity)
	return result, nil
}

func GetAttributionEvents(db *gorm.DB, search pojo.AttributionEventSearch) pojo.AttributionEventResp {
	var result pojo.AttributionEventResp
	var list []pojo.AttributionEvent
	query := applyAttributionEventFilters(db.Model(&pojo.AttributionEvent{}), attributionEventFilter{
		EventName:         search.EventName,
		ThirdPartyEventID: search.ThirdPartyEventID,
		PixelID:           search.PixelID,
		SourceChannelID:   search.SourceChannelID,
		SourceChannelCode: search.SourceChannelCode,
		UserID:            search.UserID,
		VisitorID:         search.VisitorID,
		StartTime:         search.StartTime,
		EndTime:           search.EndTime,
	})

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.AttributionEventBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetAttributionEventSummary(db *gorm.DB, search pojo.AttributionEventSummarySearch) ([]pojo.AttributionEventSummaryBack, error) {
	var result []pojo.AttributionEventSummaryBack
	query := applyAttributionEventFilters(db.Model(&pojo.AttributionEvent{}), attributionEventFilter{
		EventName:         search.EventName,
		ThirdPartyEventID: search.ThirdPartyEventID,
		PixelID:           search.PixelID,
		SourceChannelID:   search.SourceChannelID,
		SourceChannelCode: search.SourceChannelCode,
		UserID:            search.UserID,
		VisitorID:         search.VisitorID,
		StartTime:         search.StartTime,
		EndTime:           search.EndTime,
	})

	err := query.
		Select(`source_channel_id, source_channel_code, event_name,
			COUNT(*) AS event_count,
			COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_visitors,
			COUNT(DISTINCT user_id) AS unique_users`).
		Group("source_channel_id, source_channel_code, event_name").
		Order("event_count desc, event_name asc").
		Scan(&result).Error
	return result, err
}

type attributionEventFilter struct {
	EventName         string
	ThirdPartyEventID string
	PixelID           string
	SourceChannelID   int64
	SourceChannelCode string
	UserID            int64
	VisitorID         string
	StartTime         int64
	EndTime           int64
}

func applyAttributionEventFilters(query *gorm.DB, filter attributionEventFilter) *gorm.DB {
	if strings.TrimSpace(filter.EventName) != "" {
		query = query.Where("event_name = ?", strings.TrimSpace(filter.EventName))
	}
	if strings.TrimSpace(filter.ThirdPartyEventID) != "" {
		query = query.Where("third_party_event_id = ?", strings.TrimSpace(filter.ThirdPartyEventID))
	}
	if strings.TrimSpace(filter.PixelID) != "" {
		query = query.Where("pixel_id = ?", strings.TrimSpace(filter.PixelID))
	}
	if filter.SourceChannelID > 0 {
		query = query.Where("source_channel_id = ?", filter.SourceChannelID)
	}
	if strings.TrimSpace(filter.SourceChannelCode) != "" {
		query = query.Where("source_channel_code = ?", NormalizeSourceChannelCode(filter.SourceChannelCode))
	}
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if strings.TrimSpace(filter.VisitorID) != "" {
		query = query.Where("visitor_id = ?", strings.TrimSpace(filter.VisitorID))
	}
	if filter.StartTime > 0 {
		query = query.Where("created_at >= ?", time.Unix(filter.StartTime, 0))
	}
	if filter.EndTime > 0 {
		query = query.Where("created_at < ?", time.Unix(filter.EndTime, 0))
	}
	return query
}

func trimMax(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func nullableTrimMax(value string, max int) *string {
	value = trimMax(value, max)
	if value == "" {
		return nil
	}
	return &value
}
