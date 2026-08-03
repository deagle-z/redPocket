package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const domainVisitEventName = "page_view"

type domainVisitAggregate struct {
	PageViews      int64 `gorm:"column:page_views"`
	UniqueVisitors int64 `gorm:"column:unique_visitors"`
}

type domainVisitScope struct {
	condition string
	args      []any
}

func GetAdminDomainVisitStats(db *gorm.DB) (pojo.DomainVisitStatsBack, error) {
	start, end := domainVisitTodayRange(time.Now())
	return getDomainVisitStats(db, start, end, domainVisitScope{})
}

func GetAdminDomainVisits(db *gorm.DB, search pojo.DomainVisitDetailSearch) (pojo.DomainVisitDetailResp, error) {
	var scope domainVisitScope
	if rawDomain := strings.TrimSpace(search.Domain); rawDomain != "" {
		domain := utils.NormalizeDomain(rawDomain)
		if domain == "" {
			return pojo.DomainVisitDetailResp{}, errors.New("invalid_domain")
		}
		scope = domainVisitScope{condition: "domain = ?", args: []any{domain}}
	}
	start, end := domainVisitTodayRange(time.Now())
	return getDomainVisitDetails(db, search.PageInfo, start, end, scope)
}

func GetTenantDomainVisitStats(db *gorm.DB, tenantID int64) (pojo.DomainVisitStatsBack, error) {
	scope, bound, err := getTenantDomainVisitScope(db, tenantID)
	if err != nil {
		return pojo.DomainVisitStatsBack{}, err
	}
	start, end := domainVisitTodayRange(time.Now())
	if !bound {
		return pojo.DomainVisitStatsBack{Date: start.Format("2006-01-02")}, nil
	}
	return getDomainVisitStats(db, start, end, scope)
}

func GetTenantDomainVisits(db *gorm.DB, tenantID int64, pageInfo pojo.PageInfo) (pojo.DomainVisitDetailResp, error) {
	scope, bound, err := getTenantDomainVisitScope(db, tenantID)
	if err != nil {
		return pojo.DomainVisitDetailResp{}, err
	}
	start, end := domainVisitTodayRange(time.Now())
	if !bound {
		return newDomainVisitDetailResponse(pageInfo, start), nil
	}
	return getDomainVisitDetails(db, pageInfo, start, end, scope)
}

func getTenantDomainVisitScope(db *gorm.DB, tenantID int64) (domainVisitScope, bool, error) {
	if db == nil || tenantID <= 0 {
		return domainVisitScope{}, false, errors.New("invalid_params")
	}

	var tenant pojo.SysTenant
	err := db.Select("bind_domain").Where("id = ? AND deleted_at IS NULL", tenantID).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainVisitScope{}, false, errors.New("tenant_not_found")
		}
		return domainVisitScope{}, false, err
	}
	if tenant.BindDomain == nil || strings.TrimSpace(*tenant.BindDomain) == "" {
		return domainVisitScope{}, false, nil
	}

	scope, err := domainVisitScopeForBinding(*tenant.BindDomain)
	return scope, err == nil, err
}

func domainVisitScopeForBinding(binding string) (domainVisitScope, error) {
	binding = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(binding), "."))
	if strings.HasPrefix(binding, "*.") {
		suffix := strings.TrimPrefix(binding, "*.")
		if utils.NormalizeDomain(suffix) != suffix {
			return domainVisitScope{}, errors.New("invalid_bind_domain")
		}
		return domainVisitScope{condition: "domain LIKE ?", args: []any{"%." + suffix}}, nil
	}

	domain := utils.NormalizeDomain(binding)
	if domain == "" || domain != binding {
		return domainVisitScope{}, errors.New("invalid_bind_domain")
	}
	return domainVisitScope{condition: "domain = ?", args: []any{domain}}, nil
}

func getDomainVisitStats(db *gorm.DB, start time.Time, end time.Time, scope domainVisitScope) (pojo.DomainVisitStatsBack, error) {
	result := pojo.DomainVisitStatsBack{Date: start.Format("2006-01-02")}
	var aggregate domainVisitAggregate
	err := applyDomainVisitScope(domainVisitBaseQuery(db, start, end), scope).
		Select("COUNT(*) AS page_views, COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_visitors").
		Scan(&aggregate).Error
	if err != nil {
		return result, err
	}
	result.PageViews = aggregate.PageViews
	result.UniqueVisitors = aggregate.UniqueVisitors
	return result, nil
}

func getDomainVisitDetails(db *gorm.DB, pageInfo pojo.PageInfo, start time.Time, end time.Time, scope domainVisitScope) (pojo.DomainVisitDetailResp, error) {
	result := newDomainVisitDetailResponse(pageInfo, start)

	stats, err := getDomainVisitStats(db, start, end, scope)
	if err != nil {
		return result, err
	}
	result.PageViews = stats.PageViews
	result.UniqueVisitors = stats.UniqueVisitors

	if err = applyDomainVisitScope(domainVisitBaseQuery(db, start, end), scope).
		Distinct("domain").
		Count(&result.Total).Error; err != nil {
		return result, err
	}

	if err = applyDomainVisitScope(domainVisitBaseQuery(db, start, end), scope).
		Select("domain, COUNT(*) AS page_views, COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_visitors").
		Group("domain").
		Order("page_views DESC, domain ASC").
		Limit(pageInfo.PageSize).
		Offset(pageInfo.PageSize * pageInfo.CurrentPage).
		Scan(&result.List).Error; err != nil {
		return result, err
	}

	var hourlyRows []pojo.DomainVisitHourlyBack
	if err = applyDomainVisitScope(domainVisitBaseQuery(db, start, end), scope).
		Select("HOUR(created_at) AS hour, COUNT(*) AS page_views, COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_visitors").
		Group("HOUR(created_at)").
		Order("hour ASC").
		Scan(&hourlyRows).Error; err != nil {
		return result, err
	}
	for _, row := range hourlyRows {
		if row.Hour >= 0 && row.Hour < len(result.Hourly) {
			result.Hourly[row.Hour] = row
		}
	}
	return result, nil
}

func newDomainVisitDetailResponse(pageInfo pojo.PageInfo, start time.Time) pojo.DomainVisitDetailResp {
	result := pojo.DomainVisitDetailResp{
		BasePageResponse: pojo.BasePageResponse[pojo.DomainVisitRowBack]{
			List:        make([]pojo.DomainVisitRowBack, 0),
			PageSize:    pageInfo.PageSize,
			CurrentPage: pageInfo.CurrentPage,
		},
		Date:   start.Format("2006-01-02"),
		Hourly: make([]pojo.DomainVisitHourlyBack, 24),
	}
	for hour := range result.Hourly {
		result.Hourly[hour].Hour = hour
	}
	return result
}

func domainVisitBaseQuery(db *gorm.DB, start time.Time, end time.Time) *gorm.DB {
	return db.Model(&pojo.AttributionEvent{}).
		Where("event_name = ? AND domain <> ? AND created_at >= ? AND created_at < ?", domainVisitEventName, "", start, end)
}

func applyDomainVisitScope(query *gorm.DB, scope domainVisitScope) *gorm.DB {
	if scope.condition == "" {
		return query
	}
	return query.Where(scope.condition, scope.args...)
}

func domainVisitTodayRange(now time.Time) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 0, 1)
}
