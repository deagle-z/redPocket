package pojo

import "time"

type TrialBotUser struct {
	BaseModel
	Username  *string `json:"username" gorm:"type:varchar(64);index"`
	FirstName *string `json:"firstName" gorm:"type:varchar(64);index"`
	Avatar    *string `json:"avatar" gorm:"type:varchar(512)"`
	Balance   float64 `json:"balance" gorm:"type:decimal(20,2);not null;default:1000000.00"`
	Status    int8    `json:"status" gorm:"type:tinyint;not null;default:1;index"`
	TenantId  int64   `json:"tenantId" gorm:"type:bigint;index"`
}

type TrialBotUserSearch struct {
	PageInfo
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	Status    *int8  `json:"status"`
	TenantId  int64  `json:"tenantId"`
}

type TrialBotUserBack struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Username  *string   `json:"username"`
	FirstName *string   `json:"firstName"`
	Avatar    *string   `json:"avatar"`
	Balance   float64   `json:"balance"`
	Status    int8      `json:"status"`
	TenantId  int64     `json:"tenantId"`
}

type TrialBotUserResp struct {
	BasePageResponse[TrialBotUserBack]
}

type TrialBotBatchCreateReq struct {
	Num         int      `json:"num"`
	RandomName  bool     `json:"randomName"`
	NameFile    string   `json:"nameFile"`
	AvatarLinks []string `json:"avatarLinks"`
	Balance     float64  `json:"balance"`
}

type TrialBotBatchUpdateReq struct {
	IDs         []int64  `json:"ids"`
	RandomName  bool     `json:"randomName"`
	NameFile    string   `json:"nameFile"`
	AvatarLinks []string `json:"avatarLinks"`
	Status      *int8    `json:"status"`
	Balance     *float64 `json:"balance"`
}

type TrialBotStatusSet struct {
	ID     int64 `json:"id"`
	Status int8  `json:"status"`
}

type TrialBotBatchResp struct {
	Count int                `json:"count"`
	List  []TrialBotUserBack `json:"list"`
}

var TrialBotUserTableName = "trial_bot_user"

func (TrialBotUser) TableName() string {
	return TrialBotUserTableName
}
