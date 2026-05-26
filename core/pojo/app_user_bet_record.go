package pojo

import (
	"fmt"
	"time"
)

type AppUserBetRecord struct {
	ID           int64      `json:"id" gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	UID          *int64     `json:"uid" gorm:"column:uid;type:bigint;index"`
	UserID       *int64     `json:"userId" gorm:"column:user_id;type:bigint;index"`
	GameID       *string    `json:"gameId" gorm:"column:game_id;type:varchar(255);index"`
	GameName     *string    `json:"gameName" gorm:"column:game_name;type:varchar(255)"`
	PlatformCode *string    `json:"platformCode" gorm:"column:platform_code;type:varchar(255);index"`
	BetAmount    *float64   `json:"betAmount" gorm:"column:bet_amount;type:decimal(20,2)"`
	WinAmount    *float64   `json:"winAmount" gorm:"column:win_amount;type:decimal(20,2)"`
	RoundID      *string    `json:"roundId" gorm:"column:round_id;type:varchar(255);index"`
	TraceID      *string    `json:"traceId" gorm:"column:trace_id;type:varchar(255);index"`
	RoundEnd     *int       `json:"roundEnd" gorm:"column:round_end;type:int"`
	Date         *time.Time `json:"date" gorm:"column:date;index"`
	Sort         *int       `json:"sort" gorm:"column:sort;type:int"`
	DisabledFlag *int       `json:"disabledFlag" gorm:"column:disabled_flag;type:int"`
	DeletedFlag  *int       `json:"deletedFlag" gorm:"column:deleted_flag;type:int"`
	Remark       *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`
	UpdateTime   *time.Time `json:"updateTime" gorm:"column:update_time"`
	CreateTime   *time.Time `json:"createTime" gorm:"column:create_time"`
}

type AppUserBetRecordSearch struct {
	PageInfo
	ID           int64  `json:"id"`
	UID          *int64 `json:"uid"`
	UserID       *int64 `json:"userId"`
	GameID       string `json:"gameId"`
	GameName     string `json:"gameName"`
	PlatformCode string `json:"platformCode"`
	RoundID      string `json:"roundId"`
	TraceID      string `json:"traceId"`
	RoundEnd     *int   `json:"roundEnd"`
	StartTime    int64  `json:"startTime"`
	EndTime      int64  `json:"endTime"`
	DisabledFlag *int   `json:"disabledFlag"`
	DeletedFlag  *int   `json:"deletedFlag"`
}

type AppUserBetRecordSet struct {
	ID           int64      `json:"id"`
	UID          *int64     `json:"uid"`
	UserID       *int64     `json:"userId"`
	GameID       *string    `json:"gameId"`
	GameName     *string    `json:"gameName"`
	PlatformCode *string    `json:"platformCode"`
	BetAmount    *float64   `json:"betAmount"`
	WinAmount    *float64   `json:"winAmount"`
	RoundID      *string    `json:"roundId"`
	TraceID      *string    `json:"traceId"`
	RoundEnd     *int       `json:"roundEnd"`
	Date         *time.Time `json:"date"`
	Sort         *int       `json:"sort"`
	DisabledFlag *int       `json:"disabledFlag"`
	DeletedFlag  *int       `json:"deletedFlag"`
	Remark       *string    `json:"remark"`
}

type AppUserBetRecordResp struct {
	BasePageResponse[AppUserBetRecord]
}

var AppUserBetRecordTableName = "app_user_bet_record"
var AppUserBetRecordShards = 16

func (AppUserBetRecord) TableName() string {
	return AppUserBetRecordTableName
}

func AppUserBetRecordShardIndex(userID int64) int {
	index := userID % int64(AppUserBetRecordShards)
	if index < 0 {
		index = -index
	}
	return int(index)
}

func AppUserBetRecordShardTableName(index int) string {
	return fmt.Sprintf("%s_%d", AppUserBetRecordTableName, index)
}

func AppUserBetRecordTableNameByUserID(userID int64) string {
	return AppUserBetRecordShardTableName(AppUserBetRecordShardIndex(userID))
}
