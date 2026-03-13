package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"time"
)

// CreateLuckyMoney 创建红包
func CreateLuckyMoney(db *gorm.DB, luckyMoney *pojo.LuckyMoney) error {
	return db.Create(luckyMoney).Error
}

// CreateLuckyMoneyItems 创建红包明细
func CreateLuckyMoneyItems(db *gorm.DB, redPacketID int64, redList []float64) error {
	if len(redList) == 0 {
		return nil
	}
	items := make([]pojo.LuckyMoneyItem, 0, len(redList))
	for i, amount := range redList {
		items = append(items, pojo.LuckyMoneyItem{
			RedPacketID: uint64(redPacketID),
			SeqNo:       uint(i + 1),
			Amount:      amount,
			IsGrabbed:   0,
		})
	}
	return db.Create(&items).Error
}

// GetLuckyMoney 获取红包详情
func GetLuckyMoney(db *gorm.DB, luckyID int64) (pojo.LuckyMoney, error) {
	var luckyMoney pojo.LuckyMoney
	err := db.Where("id = ?", luckyID).First(&luckyMoney).Error
	return luckyMoney, err
}

// UpdateLuckyMoney 更新红包状态
func UpdateLuckyMoney(db *gorm.DB, luckyID int64, updates map[string]interface{}) error {
	return db.Model(&pojo.LuckyMoney{}).Where("id = ?", luckyID).Updates(updates).Error
}

// GetLuckyMoneyList 红包列表查询（分页）
func GetLuckyMoneyList(db *gorm.DB, search pojo.LuckyMoneySearch) (result pojo.LuckyMoneyResp) {
	var luckyMoneyList []pojo.LuckyMoney
	query := db.Model(&pojo.LuckyMoney{})

	if search.SenderID > 0 {
		query = query.Where("sender_id = ?", search.SenderID)
	}
	if search.ChatID > 0 {
		query = query.Where("chat_id = ?", search.ChatID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&luckyMoneyList)

	senderAvatarMap := map[int64]*string{}
	if len(luckyMoneyList) > 0 {
		senderIDs := make([]int64, 0, len(luckyMoneyList))
		seen := map[int64]struct{}{}
		for _, lucky := range luckyMoneyList {
			if lucky.SenderID <= 0 {
				continue
			}
			if _, ok := seen[lucky.SenderID]; ok {
				continue
			}
			seen[lucky.SenderID] = struct{}{}
			senderIDs = append(senderIDs, lucky.SenderID)
		}
		if len(senderIDs) > 0 {
			type senderAvatarRow struct {
				ID     int64   `gorm:"column:id"`
				Avatar *string `gorm:"column:avatar"`
			}
			var avatarRows []senderAvatarRow
			_ = db.Model(&pojo.TgUser{}).Select("id, avatar").Where("id IN ?", senderIDs).Scan(&avatarRows).Error
			for _, row := range avatarRows {
				senderAvatarMap[row.ID] = row.Avatar
			}
		}
	}

	for _, lucky := range luckyMoneyList {
		var tempLucky pojo.LuckyMoneyBack
		_ = copier.Copy(&tempLucky, &lucky)
		tempLucky.SenderAvatar = senderAvatarMap[lucky.SenderID]
		result.List = append(result.List, tempLucky)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// GetLuckyMoneyRedList 获取红包金额列表
func GetLuckyMoneyRedList(db *gorm.DB, luckyID int64) ([]float64, error) {
	// 优先读取红包明细表
	var items []pojo.LuckyMoneyItem
	err := db.Where("red_packet_id = ?", luckyID).Order("seq_no asc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		redList := make([]float64, 0, len(items))
		for _, item := range items {
			redList = append(redList, item.Amount)
		}
		return redList, nil
	}

	// 兼容旧数据：回退到主表 red_list
	var luckyMoney pojo.LuckyMoney
	err = db.Where("id = ?", luckyID).Select("red_list").First(&luckyMoney).Error
	if err != nil {
		return nil, err
	}
	var redList []float64
	err = json.Unmarshal([]byte(luckyMoney.RedList), &redList)
	if err != nil {
		return nil, err
	}

	return redList, nil
}

// CheckLuckyMoneyStatus 检查红包状态
func CheckLuckyMoneyStatus(db *gorm.DB, luckyID int64) (bool, error) {
	var luckyMoney pojo.LuckyMoney
	err := db.Where("id = ?", luckyID).First(&luckyMoney).Error
	if err != nil {
		return false, err
	}
	if luckyMoney.Status != 1 {
		return false, errors.New("红包已结束")
	}
	return true, nil
}

// GetLuckyMoneyAppList app端红包大厅列表
func GetLuckyMoneyAppList(db *gorm.DB, search pojo.LuckyMoneyAppListSearch, currentUserID int64) (result pojo.LuckyMoneyAppResp) {
	var luckyMoneyList []pojo.LuckyMoney
	query := db.Model(&pojo.LuckyMoney{})
	if search.LuckyID > 0 {
		query = query.Where("id = ?", search.LuckyID)
	}
	if search.ChatID > 0 {
		query = query.Where("chat_id = ?", search.ChatID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	// 未完结(status=1)优先，其次按最新ID排序
	query = query.Order("CASE WHEN status = 1 THEN 0 ELSE 1 END ASC, id DESC").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&luckyMoneyList)

	if len(luckyMoneyList) == 0 {
		result.PageSize = search.PageSize
		result.CurrentPage = search.CurrentPage
		return result
	}

	luckyIDs := make([]int64, 0, len(luckyMoneyList))
	senderIDs := make([]int64, 0, len(luckyMoneyList))
	senderSeen := map[int64]struct{}{}
	for _, lucky := range luckyMoneyList {
		luckyIDs = append(luckyIDs, lucky.ID)
		if lucky.SenderID > 0 {
			if _, ok := senderSeen[lucky.SenderID]; !ok {
				senderSeen[lucky.SenderID] = struct{}{}
				senderIDs = append(senderIDs, lucky.SenderID)
			}
		}
	}

	senderAvatarMap := map[int64]*string{}
	if len(senderIDs) > 0 {
		type senderAvatarRow struct {
			ID     int64   `gorm:"column:id"`
			Avatar *string `gorm:"column:avatar"`
		}
		var avatarRows []senderAvatarRow
		_ = db.Model(&pojo.TgUser{}).Select("id, avatar").Where("id IN ?", senderIDs).Scan(&avatarRows).Error
		for _, row := range avatarRows {
			senderAvatarMap[row.ID] = row.Avatar
		}
	}

	type luckyCountRow struct {
		LuckyID      int64 `gorm:"column:lucky_id"`
		GrabbedCount int64 `gorm:"column:grabbed_count"`
		HitCount     int64 `gorm:"column:hit_count"`
	}
	luckyCountMap := map[int64]luckyCountRow{}
	grabbedCountMap := map[int64]int64{}
	var countRows []luckyCountRow
	_ = db.Model(&pojo.LuckyHistory{}).
		Select("lucky_id, COUNT(1) AS grabbed_count, SUM(CASE WHEN is_thunder = 1 THEN 1 ELSE 0 END) AS hit_count").
		Where("lucky_id IN ?", luckyIDs).
		Group("lucky_id").
		Scan(&countRows).Error
	for _, row := range countRows {
		luckyCountMap[row.LuckyID] = row
		grabbedCountMap[row.LuckyID] = row.GrabbedCount
	}

	itemMap := map[int64][]pojo.LuckyMoneyItem{}
	var allItems []pojo.LuckyMoneyItem
	_ = db.Where("red_packet_id IN ?", luckyIDs).Order("red_packet_id asc, seq_no asc").Find(&allItems).Error
	for _, item := range allItems {
		id := int64(item.RedPacketID)
		itemMap[id] = append(itemMap[id], item)
	}
	hideSeqMap := buildPendingSecondLastHideSeqMap(luckyMoneyList, grabbedCountMap, itemMap, time.Now())

	for _, lucky := range luckyMoneyList {
		var itemBack pojo.LuckyMoneyAppBack
		itemBack.ID = lucky.ID
		itemBack.SenderID = lucky.SenderID
		itemBack.SenderName = lucky.SenderName
		itemBack.Amount = lucky.Amount
		itemBack.Received = lucky.Received
		itemBack.Number = lucky.Number
		itemBack.Thunder = lucky.Thunder
		itemBack.LoseRate = lucky.LoseRate
		itemBack.Status = lucky.Status
		itemBack.CreatedAt = lucky.CreatedAt
		itemBack.SenderAvatar = senderAvatarMap[lucky.SenderID]
		if c, ok := luckyCountMap[lucky.ID]; ok {
			itemBack.GrabbedCount = c.GrabbedCount
			itemBack.HitCount = c.HitCount
		}

		// 剩余时间（优先使用 expire_time，兼容旧数据回退到默认3分钟）
		expireAt := lucky.ExpireTime
		if expireAt.IsZero() {
			expireAt = lucky.CreatedAt.Add(3 * time.Minute)
		}
		itemBack.ExpireTime = expireAt
		remain := int64(time.Until(expireAt).Seconds())
		if remain < 0 {
			remain = 0
		}
		itemBack.RemainingSeconds = remain
		itemBack.RemainingText = fmt.Sprintf("%02d:%02d", remain/60, remain%60)

		for _, it := range itemMap[lucky.ID] {
			isMine := int8(0)
			if it.GrabbedUid != nil && uint64(currentUserID) == *it.GrabbedUid {
				isMine = 1
			}
			amount := it.Amount
			thunderAmount := it.ThunderAmount
			if hideSeq, ok := hideSeqMap[lucky.ID]; ok && hideSeq == it.SeqNo {
				amount = 0
				thunderAmount = 0
			}
			itemBack.Items = append(itemBack.Items, pojo.LuckyMoneyAppItemBack{
				SeqNo:         it.SeqNo,
				Amount:        amount,
				ThunderAmount: thunderAmount,
				IsGrabbed:     it.IsGrabbed,
				Thunder:       it.Thunder,
				IsGrabMine:    isMine,
			})
		}

		result.List = append(result.List, itemBack)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// GetLuckyMoneyAppDetail app端红包详情
func GetLuckyMoneyAppDetail(db *gorm.DB, luckyID int64, currentUserID int64) (pojo.LuckyMoneyAppDetailResp, error) {
	var result pojo.LuckyMoneyAppDetailResp
	_ = currentUserID
	cacheKey := fmt.Sprintf("bgu_lucky_detail_%d", luckyID)
	if utils.RD != nil {
		if cacheStr, err := utils.RD.Get(context.Background(), cacheKey).Result(); err == nil && cacheStr != "" {
			if e := json.Unmarshal([]byte(cacheStr), &result); e == nil {
				return result, nil
			}
		}
	}

	type luckyBaseRow struct {
		ID           int64      `gorm:"column:id"`
		SenderID     int64      `gorm:"column:sender_id"`
		SenderName   string     `gorm:"column:sender_name"`
		SenderAvatar *string    `gorm:"column:sender_avatar"`
		Amount       float64    `gorm:"column:amount"`
		Number       int        `gorm:"column:number"`
		Thunder      int        `gorm:"column:thunder"`
		LoseRate     float64    `gorm:"column:lose_rate"`
		Status       int        `gorm:"column:status"`
		CreatedAt    time.Time  `gorm:"column:created_at"`
		ExpireTime   *time.Time `gorm:"column:expire_time"`
	}
	var base luckyBaseRow
	err := db.Table("lucky_money l").
		Select("l.id, l.sender_id, l.sender_name, u.avatar as sender_avatar, l.amount, l.number, l.thunder, l.lose_rate, l.status, l.created_at, l.expire_time").
		Joins("left join tg_user u on u.id = l.sender_id").
		Where("l.id = ?", luckyID).
		Take(&base).Error
	if err != nil || base.ID == 0 {
		return result, fmt.Errorf("红包不存在")
	}

	expireAt := time.Time{}
	if base.ExpireTime != nil {
		expireAt = *base.ExpireTime
	}
	if expireAt.IsZero() {
		expireAt = base.CreatedAt.Add(3 * time.Minute)
	}
	statusText := "进行中"
	if base.Status != 1 {
		statusText = "已结束"
	}

	result.Summary = pojo.LuckyMoneyAppDetailSummary{
		ID:           base.ID,
		Status:       base.Status,
		StatusText:   statusText,
		Amount:       base.Amount,
		Thunder:      base.Thunder,
		LoseRate:     base.LoseRate,
		ExpireTime:   expireAt,
		GrabbedCount: 0,
		Number:       base.Number,
		GameText:     "Game",
		RoomText:     "PUBLIC",
		UnitAmount:   "Random",
	}
	result.Sender = pojo.LuckyMoneyAppDetailSender{
		SenderID:     base.SenderID,
		SenderName:   base.SenderName,
		SenderAvatar: base.SenderAvatar,
		SendTime:     base.CreatedAt,
	}

	// 参与记录（按子红包序号）
	type participantRow struct {
		SeqNo         uint      `gorm:"column:seq_no"`
		UserID        int64     `gorm:"column:user_id"`
		FirstName     string    `gorm:"column:first_name"`
		Avatar        *string   `gorm:"column:avatar"`
		Amount        float64   `gorm:"column:amount"`
		ThunderAmount float64   `gorm:"column:thunder_amount"`
		IsThunder     int       `gorm:"column:is_thunder"`
		CreatedAt     time.Time `gorm:"column:created_at"`
	}
	var rows []participantRow
	_ = db.Table("lucky_money_item i").
		Select("i.seq_no, COALESCE(i.grabbed_uid, 0) as user_id, COALESCE(u.first_name, u.username, '') as first_name, u.avatar, i.amount, i.thunder_amount, i.thunder as is_thunder, i.grabbed_at as created_at").
		Joins("left join tg_user u on u.id = i.grabbed_uid").
		Where("i.red_packet_id = ? AND i.is_grabbed = 1", base.ID).
		Order("i.seq_no asc").
		Scan(&rows).Error
	hideSeqNo := uint(0)
	hiddenAmount := 0.0
	hiddenThunderAmount := 0.0
	hiddenIsThunder := 0
	now := time.Now()
	if shouldHidePendingSecondLastAmount(base.Status, expireAt, base.Number, len(rows), now) {
		var (
			latestRow participantRow
			found     bool
		)
		for _, row := range rows {
			if !found || row.CreatedAt.After(latestRow.CreatedAt) {
				latestRow = row
				found = true
			}
		}
		if found {
			hideSeqNo = latestRow.SeqNo
			hiddenAmount = latestRow.Amount
			hiddenThunderAmount = latestRow.ThunderAmount
			hiddenIsThunder = latestRow.IsThunder
		}
	}
	for _, row := range rows {
		amount := row.Amount
		thunderAmount := row.ThunderAmount
		if hideSeqNo > 0 && hideSeqNo == row.SeqNo {
			amount = 0
			thunderAmount = 0
		}
		result.Participants = append(result.Participants, pojo.LuckyMoneyAppDetailParticipant{
			SeqNo:         row.SeqNo,
			UserID:        row.UserID,
			FirstName:     row.FirstName,
			Avatar:        row.Avatar,
			Amount:        amount,
			ThunderAmount: thunderAmount,
			IsThunder:     row.IsThunder,
			CreatedAt:     row.CreatedAt,
		})
	}

	// 统计统一从子红包聚合
	var agg struct {
		GrabbedCount   int64   `gorm:"column:grabbed_count"`
		HitCount       int64   `gorm:"column:hit_count"`
		ReceivedAmount float64 `gorm:"column:received_amount"`
		ThunderIncome  float64 `gorm:"column:thunder_income"`
	}
	_ = db.Table("lucky_money_item").
		Select("COUNT(1) as grabbed_count, COALESCE(SUM(CASE WHEN thunder = 1 THEN 1 ELSE 0 END), 0) as hit_count, COALESCE(SUM(amount), 0) as received_amount, COALESCE(SUM(thunder_amount), 0) as thunder_income").
		Where("red_packet_id = ? AND is_grabbed = 1", base.ID).
		Scan(&agg).Error
	if hideSeqNo > 0 {
		agg.ReceivedAmount -= hiddenAmount
		if agg.ReceivedAmount < 0 {
			agg.ReceivedAmount = 0
		}
		agg.ThunderIncome -= hiddenThunderAmount
		if agg.ThunderIncome < 0 {
			agg.ThunderIncome = 0
		}
		if hiddenIsThunder == 1 && agg.HitCount > 0 {
			agg.HitCount--
		}
	}

	var refundRow struct {
		RefundAmount float64 `gorm:"column:refund_amount"`
	}
	_ = db.Table("cash_history").
		Select("COALESCE(SUM(amount), 0) as refund_amount").
		Where("user_id = ? AND type = ? AND award_uni = ?", base.SenderID, pojo.CashHistoryTypeLuckyExpireRefund, fmt.Sprintf("lucky_expire_refund_%d", base.ID)).
		Scan(&refundRow).Error

	result.ParticipantCount = agg.GrabbedCount
	result.Summary.GrabbedCount = agg.GrabbedCount
	finalProfit := agg.ThunderIncome + refundRow.RefundAmount - base.Amount
	result.Finance = pojo.LuckyMoneyAppDetailFinance{
		SendAmount:     base.Amount,
		ReceivedAmount: agg.ReceivedAmount,
		RefundAmount:   refundRow.RefundAmount,
		ThunderIncome:  agg.ThunderIncome,
		HitCount:       agg.HitCount,
		FinalProfit:    finalProfit,
	}

	// 已结束红包缓存短时间，减少重复读取压力
	if base.Status != 1 && utils.RD != nil {
		if b, e := json.Marshal(result); e == nil {
			_ = utils.RD.SetEX(context.Background(), cacheKey, b, 30*time.Second).Err()
		}
	}
	return result, nil
}

func buildPendingSecondLastHideSeqMap(
	luckyMoneyList []pojo.LuckyMoney,
	grabbedCountMap map[int64]int64,
	itemMap map[int64][]pojo.LuckyMoneyItem,
	now time.Time,
) map[int64]uint {
	result := make(map[int64]uint)
	for _, lucky := range luckyMoneyList {
		grabbedCount, ok := grabbedCountMap[lucky.ID]
		if !ok {
			continue
		}
		expireAt := lucky.ExpireTime
		if expireAt.IsZero() {
			expireAt = lucky.CreatedAt.Add(3 * time.Minute)
		}
		if !shouldHidePendingSecondLastAmount(lucky.Status, expireAt, lucky.Number, int(grabbedCount), now) {
			continue
		}
		latest, found := findLatestGrabbedItem(itemMap[lucky.ID])
		if !found || latest.GrabbedUid == nil {
			continue
		}
		result[lucky.ID] = latest.SeqNo
	}
	return result
}

func shouldHidePendingSecondLastAmount(status int, expireAt time.Time, totalCount int, grabbedCount int, now time.Time) bool {
	if status != 1 {
		return false
	}
	if totalCount <= 1 {
		return false
	}
	if grabbedCount != totalCount-1 {
		return false
	}
	return now.Before(expireAt)
}

func findLatestGrabbedItem(items []pojo.LuckyMoneyItem) (pojo.LuckyMoneyItem, bool) {
	var latest pojo.LuckyMoneyItem
	found := false
	for _, it := range items {
		if it.IsGrabbed != 1 || it.GrabbedUid == nil {
			continue
		}
		if !found || itemGrabTime(it).After(itemGrabTime(latest)) {
			latest = it
			found = true
		}
	}
	return latest, found
}

func itemGrabTime(item pojo.LuckyMoneyItem) time.Time {
	if item.GrabbedAt != nil {
		return *item.GrabbedAt
	}
	return item.UpdatedAt
}
