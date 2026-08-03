package pojo

type DomainVisitStatsBack struct {
	Date           string `json:"date"`
	PageViews      int64  `json:"pageViews"`
	UniqueVisitors int64  `json:"uniqueVisitors"`
}

type DomainVisitDetailSearch struct {
	PageInfo
	Domain string `json:"domain"`
}

type DomainVisitRowBack struct {
	Domain         string `json:"domain"`
	PageViews      int64  `json:"pageViews"`
	UniqueVisitors int64  `json:"uniqueVisitors"`
}

type DomainVisitHourlyBack struct {
	Hour           int   `json:"hour"`
	PageViews      int64 `json:"pageViews"`
	UniqueVisitors int64 `json:"uniqueVisitors"`
}

type DomainVisitDetailResp struct {
	BasePageResponse[DomainVisitRowBack]
	Date           string                  `json:"date"`
	PageViews      int64                   `json:"pageViews"`
	UniqueVisitors int64                   `json:"uniqueVisitors"`
	Hourly         []DomainVisitHourlyBack `json:"hourly"`
}
