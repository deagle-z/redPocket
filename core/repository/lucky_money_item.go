package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetLuckyMoneyItems 红包明细列表（分页）
func GetLuckyMoneyItems(db *gorm.DB, search pojo.LuckyMoneyItemSearch) (result pojo.LuckyMoneyItemResp) {
	var items []pojo.LuckyMoneyItem
	query := db.Model(&pojo.LuckyMoneyItem{})

	if search.RedPacketID > 0 {
		query = query.Where("red_packet_id = ?", search.RedPacketID)
	}
	if search.SeqNo > 0 {
		query = query.Where("seq_no = ?", search.SeqNo)
	}
	if search.IsGrabbed != nil {
		query = query.Where("is_grabbed = ?", *search.IsGrabbed)
	}
	if search.GrabbedUid != nil {
		query = query.Where("grabbed_uid = ?", *search.GrabbedUid)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&items)

	for _, item := range items {
		var temp pojo.LuckyMoneyItemBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetLuckyMoneyItem 创建或更新红包明细
func SetLuckyMoneyItem(db *gorm.DB, req pojo.LuckyMoneyItemSet) (result pojo.LuckyMoneyItemBack, err error) {
	var dbItem pojo.LuckyMoneyItem
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbItem)
		if dbItem.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbItem, &req)
		err = db.Save(&dbItem).Error
	} else {
		_ = copier.Copy(&dbItem, &req)
		err = db.Create(&dbItem).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbItem)
	return result, nil
}

// DelLuckyMoneyItem 删除红包明细
func DelLuckyMoneyItem(db *gorm.DB, id uint64) (result string, err error) {
	var dbItem pojo.LuckyMoneyItem
	db.Where("id = ?", id).First(&dbItem)
	if dbItem.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&dbItem).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetLuckyMoneyItemById 根据ID获取红包明细
func GetLuckyMoneyItemById(db *gorm.DB, id uint64) (result pojo.LuckyMoneyItemBack, err error) {
	var dbItem pojo.LuckyMoneyItem
	db.Where("id = ?", id).First(&dbItem)
	if dbItem.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbItem)
	return result, nil
}

// MarkLuckyMoneyItemGrabbed 标记第N个子红包已被抢（兼容旧数据：若无明细记录则跳过）
func MarkLuckyMoneyItemGrabbed(db *gorm.DB, luckyID int64, seqNo int, userID int64, isThunder int8, grabbedAt time.Time) error {
	var itemCount int64
	if err := db.Model(&pojo.LuckyMoneyItem{}).Where("red_packet_id = ?", luckyID).Count(&itemCount).Error; err != nil {
		return err
	}
	// 兼容历史数据：早期可能没有 lucky_money_item 记录
	if itemCount == 0 {
		return nil
	}

	uid := uint64(userID)
	result := db.Model(&pojo.LuckyMoneyItem{}).
		Where("red_packet_id = ? and seq_no = ? and is_grabbed = 0", luckyID, seqNo).
		Updates(map[string]any{
			"is_grabbed":  1,
			"thunder":     isThunder,
			"grabbed_uid": uid,
			"grabbed_at":  grabbedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New(fmt.Sprintf("红包第%d个已被抢", seqNo))
	}
	return nil
}
