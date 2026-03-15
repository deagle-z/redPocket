package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"gorm.io/gorm"
)

// GetSysPayChannelMethods 获取通道绑定的支付方式列表
func GetSysPayChannelMethods(db *gorm.DB, channelId int64) (result []pojo.SysPayChannelMethod, err error) {
	err = db.Where("channel_id = ?", channelId).Find(&result).Error
	return result, err
}

// SetSysPayChannelMethods 设置通道支持的支付方式（全量覆盖）
func SetSysPayChannelMethods(db *gorm.DB, channelId int64, methodIds []int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("channel_id = ?", channelId).Delete(&pojo.SysPayChannelMethod{}).Error; err != nil {
			return err
		}
		if len(methodIds) == 0 {
			return nil
		}
		var records []pojo.SysPayChannelMethod
		for _, mid := range methodIds {
			records = append(records, pojo.SysPayChannelMethod{
				ChannelID: channelId,
				MethodID:  mid,
			})
		}
		return tx.Create(&records).Error
	})
}

// DelSysPayChannelMethod 删除单条通道-方式绑定
func DelSysPayChannelMethod(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysPayChannelMethod
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&entity).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}
