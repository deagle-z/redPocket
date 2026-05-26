package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetAppUserBetRecords(db *gorm.DB, search pojo.AppUserBetRecordSearch) (result pojo.AppUserBetRecordResp) {
	if search.UserID != nil {
		return getAppUserBetRecordsFromTable(db.Table(pojo.AppUserBetRecordTableNameByUserID(*search.UserID)), search)
	}
	return getAppUserBetRecordsFromShards(db, search)
}

func getAppUserBetRecordsFromShards(db *gorm.DB, search pojo.AppUserBetRecordSearch) (result pojo.AppUserBetRecordResp) {
	offset := search.PageSize * search.CurrentPage
	limit := offset + search.PageSize
	if limit <= 0 {
		limit = search.PageSize
	}
	var merged []pojo.AppUserBetRecord
	for i := 0; i < pojo.AppUserBetRecordShards; i++ {
		table := pojo.AppUserBetRecordShardTableName(i)
		var shardTotal int64
		countQuery := buildAppUserBetRecordSearchQuery(db.Table(table), search)
		countQuery.Count(&shardTotal)
		result.Total += shardTotal
		if limit <= 0 {
			continue
		}

		var shardList []pojo.AppUserBetRecord
		buildAppUserBetRecordSearchQuery(db.Table(table), search).
			Order("date desc, id desc").
			Limit(limit).
			Find(&shardList)
		merged = append(merged, shardList...)
	}
	sort.SliceStable(merged, func(i, j int) bool {
		leftDate := appUserBetRecordTimeValue(merged[i].Date)
		rightDate := appUserBetRecordTimeValue(merged[j].Date)
		if leftDate.Equal(rightDate) {
			return merged[i].ID > merged[j].ID
		}
		return leftDate.After(rightDate)
	})
	if offset > len(merged) {
		result.List = []pojo.AppUserBetRecord{}
	} else {
		end := offset + search.PageSize
		if end > len(merged) {
			end = len(merged)
		}
		result.List = merged[offset:end]
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func getAppUserBetRecordsFromTable(query *gorm.DB, search pojo.AppUserBetRecordSearch) (result pojo.AppUserBetRecordResp) {
	var list []pojo.AppUserBetRecord
	query = buildAppUserBetRecordSearchQuery(query, search)

	query.Count(&result.Total)
	query.Order("date desc, id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&list)

	result.List = list
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func buildAppUserBetRecordSearchQuery(query *gorm.DB, search pojo.AppUserBetRecordSearch) *gorm.DB {
	if search.DeletedFlag == nil {
		query = query.Where("COALESCE(deleted_flag, 0) = 0")
	} else {
		query = query.Where("deleted_flag = ?", *search.DeletedFlag)
	}
	if search.ID > 0 {
		query = query.Where("id = ?", search.ID)
	}
	if search.UID != nil {
		query = query.Where("uid = ?", *search.UID)
	}
	if search.UserID != nil {
		query = query.Where("user_id = ?", *search.UserID)
	}
	if gameID := strings.TrimSpace(search.GameID); gameID != "" {
		query = query.Where("game_id = ?", gameID)
	}
	if gameName := strings.TrimSpace(search.GameName); gameName != "" {
		query = query.Where("game_name LIKE ?", "%"+gameName+"%")
	}
	if platformCode := strings.TrimSpace(search.PlatformCode); platformCode != "" {
		query = query.Where("platform_code = ?", platformCode)
	}
	if roundID := strings.TrimSpace(search.RoundID); roundID != "" {
		query = query.Where("round_id = ?", roundID)
	}
	if traceID := strings.TrimSpace(search.TraceID); traceID != "" {
		query = query.Where("trace_id = ?", traceID)
	}
	if search.RoundEnd != nil {
		query = query.Where("round_end = ?", *search.RoundEnd)
	}
	if search.StartTime > 0 {
		query = query.Where("date >= ?", time.Unix(search.StartTime, 0))
	}
	if search.EndTime > 0 {
		query = query.Where("date <= ?", time.Unix(search.EndTime, 0))
	}
	if search.DisabledFlag != nil {
		query = query.Where("disabled_flag = ?", *search.DisabledFlag)
	}
	return query
}

func appUserBetRecordTimeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func GetAppUserBetRecordByID(db *gorm.DB, id int64) (pojo.AppUserBetRecord, error) {
	var entity pojo.AppUserBetRecord
	err := db.Where("id = ? AND COALESCE(deleted_flag, 0) = 0", id).First(&entity).Error
	if err != nil {
		return pojo.AppUserBetRecord{}, err
	}
	return entity, nil
}

func SetAppUserBetRecord(db *gorm.DB, req pojo.AppUserBetRecordSet) (pojo.AppUserBetRecord, error) {
	now := time.Now()
	var entity pojo.AppUserBetRecord
	table := pojo.AppUserBetRecordTableName
	if req.UserID != nil {
		table = pojo.AppUserBetRecordTableNameByUserID(*req.UserID)
	}
	if req.ID > 0 {
		if err := db.Table(table).Where("id = ? AND COALESCE(deleted_flag, 0) = 0", req.ID).First(&entity).Error; err != nil {
			return pojo.AppUserBetRecord{}, err
		}
		_ = copier.Copy(&entity, &req)
		entity.UpdateTime = &now
		return entity, db.Table(table).Save(&entity).Error
	}

	_ = copier.Copy(&entity, &req)
	entity.CreateTime = &now
	entity.UpdateTime = &now
	return entity, db.Table(table).Create(&entity).Error
}

func DelAppUserBetRecord(db *gorm.DB, id int64) (string, error) {
	now := time.Now()
	result := db.Model(&pojo.AppUserBetRecord{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"deleted_flag": 1,
			"update_time":  now,
		})
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", errors.New("record_not_found_delete")
	}
	return "success", nil
}
