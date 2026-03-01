package repository

import (
	"BaseGoUni/core/pojo"
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

	for _, lucky := range luckyMoneyList {
		var tempLucky pojo.LuckyMoneyBack
		_ = copier.Copy(&tempLucky, &lucky)
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

	remainingWindow := 3 * time.Minute

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

		var sender pojo.TgUser
		if err := db.Select("avatar").Where("id = ?", lucky.SenderID).First(&sender).Error; err == nil && sender.ID > 0 {
			itemBack.SenderAvatar = sender.Avatar
		}

		// 已抢数量
		_ = db.Model(&pojo.LuckyHistory{}).Where("lucky_id = ?", lucky.ID).Count(&itemBack.GrabbedCount).Error
		// 中雷次数
		_ = db.Model(&pojo.LuckyHistory{}).Where("lucky_id = ? and is_thunder = 1", lucky.ID).Count(&itemBack.HitCount).Error

		// 剩余时间（默认3分钟窗口）
		remain := int64(time.Until(lucky.CreatedAt.Add(remainingWindow)).Seconds())
		if remain < 0 {
			remain = 0
		}
		itemBack.RemainingSeconds = remain
		itemBack.RemainingText = fmt.Sprintf("%02d:%02d", remain/60, remain%60)

		var items []pojo.LuckyMoneyItem
		_ = db.Where("red_packet_id = ?", lucky.ID).Order("seq_no asc").Find(&items).Error
		for _, it := range items {
			isMine := int8(0)
			if it.GrabbedUid != nil && uint64(currentUserID) == *it.GrabbedUid {
				isMine = 1
			}
			itemBack.Items = append(itemBack.Items, pojo.LuckyMoneyAppItemBack{
				SeqNo:      it.SeqNo,
				Amount:     it.Amount,
				IsGrabbed:  it.IsGrabbed,
				Thunder:    it.Thunder,
				IsGrabMine: isMine,
			})
		}

		result.List = append(result.List, itemBack)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}
