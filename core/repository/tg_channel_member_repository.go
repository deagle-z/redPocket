package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TgRequiredChannelIDConfigKey    = "tg_required_channel_id"
	TgBindFreeLotteryCountConfigKey = "register_free_lottery_count"
	defaultBindFreeLotteryCount     = 1
	tgChannelMemberStatusInChannel  = int8(1)
	tgChannelMemberStatusOutChannel = int8(0)
)

type TgChannelMemberSnapshot struct {
	ChannelID string
	TgID      int64
	TgName    string
	FirstName string
	Status    int8
}

type TgChannelMembershipChecker func(ctx context.Context, channelID string, tgID int64) (TgChannelMemberSnapshot, error)

var (
	tgChannelMembershipCheckerMu sync.RWMutex
	tgChannelMembershipChecker   TgChannelMembershipChecker
)

func RegisterTgChannelMembershipChecker(fn TgChannelMembershipChecker) {
	tgChannelMembershipCheckerMu.Lock()
	defer tgChannelMembershipCheckerMu.Unlock()
	tgChannelMembershipChecker = fn
}

func NormalizeTgChannelName(tgName string) (string, error) {
	ptr := formatTelegramAtName(tgName)
	if ptr == nil {
		return "", errors.New("tg_name_required")
	}
	return *ptr, nil
}

func UpsertTgChannelMember(db *gorm.DB, snapshot TgChannelMemberSnapshot) error {
	snapshot.ChannelID = normalizeTgRequiredChannelID(snapshot.ChannelID)
	if snapshot.ChannelID == "" || snapshot.TgID == 0 {
		return errors.New("invalid_tg_channel_member")
	}
	normalizedName := ""
	if snapshot.TgName != "" {
		if name, err := NormalizeTgChannelName(snapshot.TgName); err == nil {
			normalizedName = name
		}
	}

	now := time.Now()
	member := pojo.TgChannelMember{
		ChannelID: snapshot.ChannelID,
		TgID:      snapshot.TgID,
		TgName:    normalizedName,
		FirstName: truncateRunes(snapshot.FirstName, 128),
		Status:    snapshot.Status,
	}
	if member.Status == tgChannelMemberStatusInChannel {
		member.JoinedAt = &now
	} else {
		member.LeftAt = &now
	}

	updates := map[string]any{
		"tg_name":    member.TgName,
		"first_name": member.FirstName,
		"status":     member.Status,
		"updated_at": now,
	}
	if member.Status == tgChannelMemberStatusInChannel {
		updates["joined_at"] = now
		updates["left_at"] = nil
	} else {
		updates["left_at"] = now
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel_id"}, {Name: "tg_id"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(&member).Error
}

func BindCurrentTgChannelName(ctx context.Context, db *gorm.DB, tablePrefix string, userID int64, tgName string) (pojo.TgBindChannelNameBack, error) {
	start := time.Now()
	rawTgName := strings.TrimSpace(tgName)
	log.Printf("[tg-channel-bind] start prefix=%q user_id=%d raw_tg_name=%q", tablePrefix, userID, rawTgName)
	if userID <= 0 {
		log.Printf("[tg-channel-bind] reject invalid user_id prefix=%q user_id=%d", tablePrefix, userID)
		return pojo.TgBindChannelNameBack{}, errors.New("token_invalid")
	}
	normalizedName, err := NormalizeTgChannelName(tgName)
	if err != nil {
		log.Printf("[tg-channel-bind] reject invalid tg_name prefix=%q user_id=%d raw_tg_name=%q err=%v", tablePrefix, userID, rawTgName, err)
		return pojo.TgBindChannelNameBack{}, err
	}
	channelID := getRequiredTgChannelID(tablePrefix)
	log.Printf("[tg-channel-bind] config loaded prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q", tablePrefix, userID, channelID, normalizedName)
	if channelID == "" {
		log.Printf("[tg-channel-bind] reject channel not configured prefix=%q user_id=%d normalized_tg_name=%q", tablePrefix, userID, normalizedName)
		return pojo.TgBindChannelNameBack{}, errors.New("tg_channel_not_configured")
	}

	var current pojo.TgUser
	if err = db.Where("id = ? AND status = ?", userID, 1).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[tg-channel-bind] reject user not found prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q", tablePrefix, userID, channelID, normalizedName)
			return pojo.TgBindChannelNameBack{}, errors.New("user_not_found")
		}
		log.Printf("[tg-channel-bind] user query failed prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q err=%v", tablePrefix, userID, channelID, normalizedName, err)
		return pojo.TgBindChannelNameBack{}, err
	}
	currentTgName := ""
	if current.TgName != nil {
		currentTgName = *current.TgName
	}
	log.Printf("[tg-channel-bind] user loaded prefix=%q user_id=%d channel_id=%q user_tg_id=%d current_tg_name=%q free_lottery_count=%d", tablePrefix, userID, channelID, current.TgID, currentTgName, current.FreeLotteryCount)

	member, err := findActiveTgChannelMember(db, channelID, normalizedName)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[tg-channel-bind] member query failed prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q err=%v", tablePrefix, userID, channelID, normalizedName, err)
			return pojo.TgBindChannelNameBack{}, err
		}
		log.Printf("[tg-channel-bind] member cache miss prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q user_tg_id=%d fallback=true", tablePrefix, userID, channelID, normalizedName, current.TgID)
		member, err = fallbackCheckAndStoreTgChannelMember(ctx, db, current, channelID, normalizedName)
		if err != nil {
			log.Printf("[tg-channel-bind] fallback failed prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q user_tg_id=%d err=%v", tablePrefix, userID, channelID, normalizedName, current.TgID, err)
			return pojo.TgBindChannelNameBack{}, err
		}
		log.Printf("[tg-channel-bind] fallback stored member prefix=%q user_id=%d channel_id=%q member_id=%d member_tg_id=%d member_tg_name=%q member_bind_user_id=%v", tablePrefix, userID, channelID, member.ID, member.TgID, member.TgName, member.BindUserID)
	} else {
		log.Printf("[tg-channel-bind] member cache hit prefix=%q user_id=%d channel_id=%q member_id=%d member_tg_id=%d member_tg_name=%q member_bind_user_id=%v", tablePrefix, userID, channelID, member.ID, member.TgID, member.TgName, member.BindUserID)
	}

	if member.BindUserID != nil && *member.BindUserID != userID {
		log.Printf("[tg-channel-bind] reject member already bound prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q member_id=%d bind_user_id=%d", tablePrefix, userID, channelID, normalizedName, member.ID, *member.BindUserID)
		return pojo.TgBindChannelNameBack{}, errors.New("tg_name_already_bound")
	}

	var result pojo.TgBindChannelNameBack
	err = db.Transaction(func(tx *gorm.DB) error {
		log.Printf("[tg-channel-bind] tx start prefix=%q user_id=%d channel_id=%q member_id=%d normalized_tg_name=%q", tablePrefix, userID, channelID, member.ID, normalizedName)
		var exists pojo.TgUser
		queryErr := tx.Where("tg_name = ? AND id <> ? AND status <> ?", normalizedName, userID, -1).First(&exists).Error
		if queryErr == nil && exists.ID > 0 {
			log.Printf("[tg-channel-bind] tx reject tg_name already used prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q exists_user_id=%d", tablePrefix, userID, channelID, normalizedName, exists.ID)
			return errors.New("tg_name_already_bound")
		}
		if queryErr != nil && !errors.Is(queryErr, gorm.ErrRecordNotFound) {
			log.Printf("[tg-channel-bind] tx tg_name uniqueness query failed prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q err=%v", tablePrefix, userID, channelID, normalizedName, queryErr)
			return queryErr
		}

		var lockedUser pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", userID, 1).First(&lockedUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("[tg-channel-bind] tx reject locked user not found prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q", tablePrefix, userID, channelID, normalizedName)
				return errors.New("user_not_found")
			}
			log.Printf("[tg-channel-bind] tx lock user failed prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q err=%v", tablePrefix, userID, channelID, normalizedName, err)
			return err
		}
		lockedUserTgName := ""
		if lockedUser.TgName != nil {
			lockedUserTgName = *lockedUser.TgName
		}
		log.Printf("[tg-channel-bind] tx user locked prefix=%q user_id=%d channel_id=%q user_tg_id=%d current_tg_name=%q free_lottery_count=%d", tablePrefix, userID, channelID, lockedUser.TgID, lockedUserTgName, lockedUser.FreeLotteryCount)

		var lockedMember pojo.TgChannelMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", member.ID, tgChannelMemberStatusInChannel).
			First(&lockedMember).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("[tg-channel-bind] tx reject active member not found prefix=%q user_id=%d channel_id=%q member_id=%d normalized_tg_name=%q", tablePrefix, userID, channelID, member.ID, normalizedName)
				return errors.New("tg_channel_member_not_found")
			}
			log.Printf("[tg-channel-bind] tx lock member failed prefix=%q user_id=%d channel_id=%q member_id=%d normalized_tg_name=%q err=%v", tablePrefix, userID, channelID, member.ID, normalizedName, err)
			return err
		}
		if lockedMember.BindUserID != nil && *lockedMember.BindUserID != userID {
			log.Printf("[tg-channel-bind] tx reject locked member already bound prefix=%q user_id=%d channel_id=%q member_id=%d bind_user_id=%d", tablePrefix, userID, channelID, lockedMember.ID, *lockedMember.BindUserID)
			return errors.New("tg_name_already_bound")
		}
		log.Printf("[tg-channel-bind] tx member locked prefix=%q user_id=%d channel_id=%q member_id=%d member_tg_id=%d member_tg_name=%q member_bind_user_id=%v", tablePrefix, userID, channelID, lockedMember.ID, lockedMember.TgID, lockedMember.TgName, lockedMember.BindUserID)

		var existingBoundCount int64
		if err := tx.Model(&pojo.TgChannelMember{}).
			Where("bind_user_id = ?", userID).
			Count(&existingBoundCount).Error; err != nil {
			log.Printf("[tg-channel-bind] tx count existing bind failed prefix=%q user_id=%d channel_id=%q err=%v", tablePrefix, userID, channelID, err)
			return err
		}
		awardCount := 0
		if existingBoundCount == 0 {
			awardCount = getBindFreeLotteryCount(tx)
		}
		log.Printf("[tg-channel-bind] tx award decided prefix=%q user_id=%d channel_id=%q existing_bound_count=%d award_count=%d", tablePrefix, userID, channelID, existingBoundCount, awardCount)
		if existingBoundCount > 0 && !stringPtrValueEquals(lockedUser.TgName, normalizedName) {
			if err := tx.Model(&pojo.TgChannelMember{}).
				Where("bind_user_id = ? AND id <> ?", userID, lockedMember.ID).
				Update("bind_user_id", nil).Error; err != nil {
				log.Printf("[tg-channel-bind] tx unbind old member failed prefix=%q user_id=%d channel_id=%q member_id=%d err=%v", tablePrefix, userID, channelID, lockedMember.ID, err)
				return err
			}
			log.Printf("[tg-channel-bind] tx old member unbound prefix=%q user_id=%d channel_id=%q keep_member_id=%d", tablePrefix, userID, channelID, lockedMember.ID)
		}

		if err := tx.Model(&pojo.TgChannelMember{}).
			Where("id = ?", lockedMember.ID).
			Update("bind_user_id", userID).Error; err != nil {
			log.Printf("[tg-channel-bind] tx bind member failed prefix=%q user_id=%d channel_id=%q member_id=%d err=%v", tablePrefix, userID, channelID, lockedMember.ID, err)
			return err
		}
		log.Printf("[tg-channel-bind] tx member bound prefix=%q user_id=%d channel_id=%q member_id=%d", tablePrefix, userID, channelID, lockedMember.ID)

		updates := map[string]any{"tg_name": normalizedName}
		if awardCount > 0 {
			updates["free_lottery_count"] = gorm.Expr("free_lottery_count + ?", awardCount)
		}
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", lockedUser.ID).Updates(updates).Error; err != nil {
			log.Printf("[tg-channel-bind] tx update user failed prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q award_count=%d err=%v", tablePrefix, userID, channelID, normalizedName, awardCount, err)
			return err
		}
		lockedUser.TgName = &normalizedName
		lockedUser.FreeLotteryCount += awardCount
		result = pojo.TgBindChannelNameBack{
			TgName:           normalizedName,
			FreeLotteryCount: lockedUser.FreeLotteryCount,
			AwardedCount:     awardCount,
		}
		log.Printf("[tg-channel-bind] tx success prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q award_count=%d free_lottery_count=%d", tablePrefix, userID, channelID, normalizedName, awardCount, lockedUser.FreeLotteryCount)
		return nil
	})
	if err != nil {
		log.Printf("[tg-channel-bind] finish failed prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q elapsed=%s err=%v", tablePrefix, userID, channelID, normalizedName, time.Since(start), err)
		return result, err
	}
	log.Printf("[tg-channel-bind] finish success prefix=%q user_id=%d channel_id=%q normalized_tg_name=%q awarded_count=%d free_lottery_count=%d elapsed=%s", tablePrefix, userID, channelID, result.TgName, result.AwardedCount, result.FreeLotteryCount, time.Since(start))
	return result, err
}

func findActiveTgChannelMember(db *gorm.DB, channelID string, tgName string) (pojo.TgChannelMember, error) {
	var member pojo.TgChannelMember
	err := db.Where("channel_id = ? AND tg_name = ? AND status = ?", channelID, tgName, tgChannelMemberStatusInChannel).
		Order("id desc").
		First(&member).Error
	return member, err
}

func fallbackCheckAndStoreTgChannelMember(ctx context.Context, db *gorm.DB, user pojo.TgUser, channelID string, tgName string) (pojo.TgChannelMember, error) {
	if user.TgID == 0 {
		log.Printf("[tg-channel-bind] fallback skipped no tg_id user_id=%d channel_id=%q tg_name=%q", user.ID, channelID, tgName)
		return pojo.TgChannelMember{}, errors.New("tg_channel_member_not_found")
	}
	tgChannelMembershipCheckerMu.RLock()
	checker := tgChannelMembershipChecker
	tgChannelMembershipCheckerMu.RUnlock()
	if checker == nil {
		log.Printf("[tg-channel-bind] fallback skipped checker nil user_id=%d channel_id=%q tg_id=%d tg_name=%q", user.ID, channelID, user.TgID, tgName)
		return pojo.TgChannelMember{}, errors.New("tg_channel_member_not_found")
	}

	log.Printf("[tg-channel-bind] fallback check telegram start user_id=%d channel_id=%q tg_id=%d tg_name=%q", user.ID, channelID, user.TgID, tgName)
	snapshot, err := checker(ctx, channelID, user.TgID)
	if err != nil {
		log.Printf("[tg-channel-bind] fallback check telegram failed user_id=%d channel_id=%q tg_id=%d tg_name=%q err=%v", user.ID, channelID, user.TgID, tgName, err)
		return pojo.TgChannelMember{}, errors.New("tg_channel_member_not_found")
	}
	log.Printf("[tg-channel-bind] fallback check telegram result user_id=%d channel_id=%q tg_id=%d snapshot_channel_id=%q snapshot_tg_id=%d snapshot_tg_name=%q snapshot_status=%d", user.ID, channelID, user.TgID, snapshot.ChannelID, snapshot.TgID, snapshot.TgName, snapshot.Status)
	if snapshot.Status != tgChannelMemberStatusInChannel {
		log.Printf("[tg-channel-bind] fallback reject not in channel user_id=%d channel_id=%q tg_id=%d status=%d", user.ID, channelID, user.TgID, snapshot.Status)
		return pojo.TgChannelMember{}, errors.New("tg_channel_member_not_found")
	}
	if snapshot.TgName != "" {
		checkedName, err := NormalizeTgChannelName(snapshot.TgName)
		if err != nil || checkedName != tgName {
			log.Printf("[tg-channel-bind] fallback reject tg_name mismatch user_id=%d channel_id=%q tg_id=%d input_tg_name=%q telegram_tg_name=%q checked_tg_name=%q err=%v", user.ID, channelID, user.TgID, tgName, snapshot.TgName, checkedName, err)
			return pojo.TgChannelMember{}, errors.New("tg_name_not_in_required_channel")
		}
	} else {
		log.Printf("[tg-channel-bind] fallback telegram username empty use input user_id=%d channel_id=%q tg_id=%d input_tg_name=%q", user.ID, channelID, user.TgID, tgName)
		snapshot.TgName = tgName
	}
	snapshot.ChannelID = channelID
	snapshot.TgID = user.TgID
	if err := UpsertTgChannelMember(db, snapshot); err != nil {
		log.Printf("[tg-channel-bind] fallback upsert member failed user_id=%d channel_id=%q tg_id=%d tg_name=%q err=%v", user.ID, channelID, user.TgID, tgName, err)
		return pojo.TgChannelMember{}, err
	}
	log.Printf("[tg-channel-bind] fallback upsert member success user_id=%d channel_id=%q tg_id=%d tg_name=%q", user.ID, channelID, user.TgID, tgName)
	return findActiveTgChannelMember(db, channelID, tgName)
}

func getRequiredTgChannelID(tablePrefix string) string {
	defaultValue := ""
	value := utils.GetStringCache(tablePrefix, TgRequiredChannelIDConfigKey, &defaultValue)
	if value == nil {
		return ""
	}
	return normalizeTgRequiredChannelID(*value)
}

func normalizeTgRequiredChannelID(channelID string) string {
	return truncateRunes(strings.TrimSpace(channelID), 64)
}

func getBindFreeLotteryCount(db *gorm.DB) int {
	var config pojo.SysConfig
	if err := db.Where("config_key = ?", TgBindFreeLotteryCountConfigKey).First(&config).Error; err != nil || strings.TrimSpace(config.ConfigValue) == "" {
		return defaultBindFreeLotteryCount
	}
	count, err := strconv.Atoi(strings.TrimSpace(config.ConfigValue))
	if err != nil || count < 0 {
		return defaultBindFreeLotteryCount
	}
	return count
}

func stringPtrValueEquals(ptr *string, value string) bool {
	return ptr != nil && strings.TrimSpace(*ptr) == value
}
