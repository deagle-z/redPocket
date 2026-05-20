package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
	"log"
	"sort"
	"strings"
	"time"
)

func AdminAwardInfo(currentUser pojo.SysUser, reqData pojo.AdminAwardInfo) (result string, err error) {
	reqData.Amount = utils.Truncate2(reqData.Amount)
	awardUni := utils.RandomString(6)
	checkKey := fmt.Sprintf(utils.KeyUserAwardCheck, reqData.UserId, awardUni)
	checkValue := utils.MD5(checkKey + "nj")
	awardInfo := pojo.AwardInfo{
		CheckKey:   checkKey,
		CheckValue: checkValue,
	}
	awardInfo.AwardUnis = append(awardInfo.AwardUnis, pojo.AwardUni{
		UserId:     reqData.UserId,
		Amount:     reqData.Amount,
		AwardUni:   awardUni,
		CashMark:   reqData.CashMark,
		CashDesc:   "",
		RefuseCash: false,
		FromUserId: currentUser.ID,
	})
	utils.RD.SetEX(context.Background(), checkKey, checkValue, 1*time.Minute)
	requestData, _ := json.Marshal(awardInfo)
	response, _, err := utils.ProxyPostRequest(utils.CsConfig.AwardUrl, utils.JsonHead, requestData, nil)
	log.Printf("AdminAwardInfo response = %s", string(response))
	if err == nil {
		var responseObj pojo.BaseResponse
		_ = json.Unmarshal(response, &responseObj)
		if responseObj.Success {
			return responseObj.Message, err
		}
	}
	return result, err
}

func LocalAwardInfo(currentUser pojo.SysUser, reqData pojo.AwardInfo) (result string, err error) {
	if len(reqData.AwardUnis) == 0 {
		return result, errors.New("data_error")
	}
	awardUni := reqData.AwardUnis[0].AwardUni
	reqData.CheckKey = fmt.Sprintf(utils.KeyUserAwardCheck, currentUser.ID, awardUni)
	reqData.CheckValue = utils.MD5(reqData.CheckKey + "nj")
	utils.RD.SetEX(context.Background(), reqData.CheckKey, reqData.CheckValue, 1*time.Minute)
	requestData, _ := json.Marshal(reqData)
	response, _, err := utils.ProxyPostRequest(utils.CsConfig.AwardUrl, utils.JsonHead, requestData, nil)
	log.Printf("LocalAwardInfo response = %s", string(response))
	if err == nil {
		var responseObj pojo.BaseResponse
		_ = json.Unmarshal(response, &responseObj)
		if responseObj.Success {
			return responseObj.Message, err
		}
	}
	return result, err
}

func AwardUser(db *gorm.DB, reqData pojo.AwardInfo) (result []int64, err error) {
	checkValue := utils.GetRdString(reqData.CheckKey, "")
	if reqData.CheckValue != checkValue {
		return result, errors.New("award_request_check_failed")
	}
	if len(reqData.AwardUnis) == 0 {
		return result, errors.New("data error")
	}
	lockKeys := make([]string, 0)
	updateUsers := make([]string, 0)
	tx := db.Begin()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Printf("db begin error.err=%v", p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
			for _, updateUser := range updateUsers {
				_ = utils.PublishMQ(utils.MQMessage{
					MessageType: utils.KeyMqUserUpdate,
					Data:        updateUser,
				})
			}
		}
		for _, lockKey := range lockKeys {
			_ = utils.ReleaseLock(lockKey)
			//log.Printf("释放锁:%s", lockKey)
		}
	}()
	for _, awardUni := range reqData.AwardUnis {
		awardUni.Amount = utils.Truncate2(awardUni.Amount)
		if awardUni.UserId == 0 || awardUni.Amount == 0 || awardUni.AwardUni == "" {
			return result, errors.New("award_data_check_failed")
		}
		prefix := utils.GetDbPrefix(db)
		awardUser := utils.GetTempUser(prefix, awardUni.UserId)
		blockKey := fmt.Sprintf(utils.KeyLockUserAward, awardUser.ID)
		acquired := false
		for i := 0; i < 100; i++ { // 重试10s
			acquired, err = utils.AcquireLock(blockKey, 20*time.Second)
			if acquired {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if !acquired {
			fmt.Printf("AwardUser error acquiring lock for user %s: %v\n", awardUser.Username, err)
			err = errors.New("block_user_error")
			return result, err
		}
		//log.Printf("加锁:%s", blockKey)
		lockKeys = append(lockKeys, blockKey)
		var currentUser pojo.SysUser
		tx.Where("id = ?", awardUser.ID).First(&currentUser)
		if awardUni.Amount < 0 && awardUser.Amount+awardUni.Amount < 0 {
			err = errors.New(utils.I18nMessage("amount_not_enough", map[string]interface{}{"amount": fmt.Sprintf("%.2f", awardUser.Amount), "required": fmt.Sprintf("%.2f", awardUni.Amount)})) // 余额不足
			return result, err
		}
		cashHistory := pojo.CashHistory{
			UserId:      currentUser.ID,
			AwardUni:    awardUni.AwardUni,
			Amount:      awardUni.Amount,
			StartAmount: currentUser.Amount,
			EndAmount:   utils.Truncate2(utils.ToMoney(currentUser.Amount).Add(utils.ToMoney(awardUni.Amount)).ToDollars()),
			CashMark:    awardUni.CashMark,
			CashDesc:    awardUni.CashDesc,
			FromUserId:  awardUni.FromUserId,
		}
		if awardUni.Amount >= 0 {
			cashHistory.Type = pojo.CashHistoryTypeAdminManualAward
		} else {
			cashHistory.Type = pojo.CashHistoryTypeAdminManualDeduct
		}
		err = tx.Create(&cashHistory).Error
		if err != nil {
			return result, err
		}
		update := make(map[string]any)
		update["amount"] = gorm.Expr(fmt.Sprintf("amount + %.2f", awardUni.Amount))
		if awardUni.Amount > 0 && !awardUni.RefuseCash {
			update["top_amount"] = gorm.Expr(fmt.Sprintf("top_amount + %.2f", awardUni.Amount))
		}
		err = tx.Model(&awardUser).Updates(update).Error
		if err != nil {
			return result, err
		}
		updateUsers = append(updateUsers, fmt.Sprintf("%d#%s", awardUser.ID, prefix))
		result = append(result, cashHistory.ID)
	}
	return result, err
}

func GetUsers(db *gorm.DB, userSearch pojo.UserSearch, currentUserName string, currentUserId int64) (result pojo.UserResp) {
	var users []pojo.SysUser
	//db = db.Model(&pojo.SysUser{}).Where("user_type != ?", 3)
	db = db.Model(&pojo.SysUser{})
	if userSearch.Username != "" {
		db = db.Where("username like ?", "%"+userSearch.Username+"%")
	}
	if currentUserName != "admin" {
		db.Where("parent_id = ?", currentUserId)
	}
	if userSearch.Enabled != nil {
		db = db.Where("enabled = ?", userSearch.Enabled)
	}
	db.Model(&pojo.SysUser{}).Count(&result.Total)
	db = db.Order("id desc").Limit(userSearch.PageSize).Offset(userSearch.PageSize * userSearch.CurrentPage)
	db.Find(&users)
	for _, user := range users {
		var tempUserBack pojo.UserBack
		_ = copier.Copy(&tempUserBack, &user)
		log.Printf("tempUserBack: %v", user.RoleStr)
		err := json.Unmarshal([]byte(user.RoleStr), &tempUserBack.Roles)
		if err != nil {
			log.Printf("err: %v", err)
		}
		log.Printf("tempUserBack.Roles:%v", tempUserBack.Roles)
		result.List = append(result.List, tempUserBack)
	}
	result.PageSize = userSearch.PageSize
	result.CurrentPage = userSearch.CurrentPage
	return result
}

func buildMenuTree(menus []pojo.SysMenu, parentID int64) []pojo.BackMenu {
	var result []pojo.BackMenu
	for _, menu := range menus {
		var tempMenu pojo.BackMenu
		_ = copier.Copy(&tempMenu, &menu)
		if tempMenu.ParentID == parentID {
			tempMenu.Children = buildMenuTree(menus, tempMenu.ID)
			_ = json.Unmarshal([]byte(menu.MetaStr), &tempMenu.Meta)
			result = append(result, tempMenu)
		}
	}
	return result
}

func UnBindGAuth(db *gorm.DB, currentUser pojo.SysUser) (result pojo.UserBack, err error) {
	currentUser.BindCode = false
	currentUser.GoogleCode = ""
	err = db.Save(&currentUser).Error
	_ = copier.Copy(&result, &currentUser)
	_ = json.Unmarshal([]byte(currentUser.RoleStr), &result.Roles)
	return result, err
}

func ChangePass(db *gorm.DB, hostInfo pojo.HostInfo, currentUser pojo.SysUser, userAdd pojo.UserPassChange) (result pojo.UserBack, err error) {
	isManager := currentUser.UserType == 1 || currentUser.UserType == 2
	if !isManager {
		userAdd.ID = currentUser.ID
	}
	dbUser := utils.GetTempUser(hostInfo.TablePrefix, userAdd.ID)
	if dbUser.ID == 0 {
		return result, errors.New("operator_data_not_found")
	}
	if len(userAdd.Password) < 6 || len(userAdd.Password) > 18 {
		return result, errors.New("password_length_6_18")
	}
	dbUser.Password = utils.EncodePass(hostInfo.Salt, userAdd.Password)
	err = db.Save(&dbUser).Error
	_ = copier.Copy(&result, &dbUser)
	_ = json.Unmarshal([]byte(dbUser.RoleStr), &result.Roles)
	return result, err
}

func SetUser(db *gorm.DB, hostInfo pojo.HostInfo, userAdd pojo.UserAdd, currentUserId int64) (result pojo.UserBack, err error) {
	dbUser := utils.GetTempUser(hostInfo.TablePrefix, userAdd.ID)
	if dbUser.ID == 0 {
		db.Where("username = ?", userAdd.Username).First(&dbUser)
		if dbUser.ID != 0 {
			return result, errors.New("username_duplicate")
		}
		_ = copier.Copy(&dbUser, &userAdd)
		roleStr, _ := json.Marshal(userAdd.Roles)
		dbUser.UniKey = utils.GetUserUniKey(hostInfo.TablePrefix)
		dbUser.SecurityKey = utils.RandomString(32)
		dbUser.RoleStr = string(roleStr)
		dbUser.Password = utils.EncodePass(hostInfo.Salt, userAdd.Password)
		err = db.Create(&dbUser).Error
	} else {
		if dbUser.Username != userAdd.Username {
			return result, errors.New("username_cannot_modify")
		}
		userAdd.Password = ""
		_ = copier.Copy(&dbUser, userAdd)
		roleStr, _ := json.Marshal(userAdd.Roles)
		dbUser.RoleStr = string(roleStr)
		dbUser.Enabled = userAdd.Enabled
		dbUser.Mark = userAdd.Mark
		err = db.Save(&dbUser).Error
	}
	_ = copier.Copy(&result, &dbUser)
	_ = json.Unmarshal([]byte(dbUser.RoleStr), &result.Roles)
	utils.UpdateTempUser(hostInfo.TablePrefix, dbUser)
	return result, err
}

func GetRoutes(db *gorm.DB, hostInfo pojo.HostInfo, currentUser pojo.SysUser) (result []pojo.BackMenu) {
	_ = json.Unmarshal([]byte(currentUser.RoleStr), &currentUser.Roles)
	var menus []pojo.SysMenu
	if currentUser.UserType == 1 {
		db.Find(&menus)
	} else {
		currentUserStr, _ := json.Marshal(currentUser)
		log.Printf("currentUser=%s", string(currentUserStr))
		var roles []pojo.SysRole
		db.Where("code in ?", currentUser.Roles).Find(&roles)
		menuIds := make([]int64, 0)
		for _, role := range roles {
			var tempIds []int64
			_ = json.Unmarshal([]byte(role.MenuIdStr), &tempIds)
			for _, tempId := range tempIds {
				if utils.InInt64s(menuIds, tempId) {
					continue
				}
				menuIds = append(menuIds, tempId)
			}
		}
		db.Where("id in ?", menuIds).Find(&menus)
	}
	//menusStr, _ := json.Marshal(menus)
	//log.Printf("menus=%s", string(menusStr))
	endMenus := make([]pojo.SysMenu, 0)
	for _, menu := range menus {
		if menu.Name == "SystemHostInfo" && hostInfo.HostName != utils.CsConfig.DefaultHost.HostName {
			continue
		}
		endMenus = append(endMenus, menu)
	}
	return buildMenuTree(endMenus, 0)
}

func WhiteUserLogin(db *gorm.DB, hostInfo pojo.HostInfo, reqUserLogin pojo.UserLogin, onlineUser pojo.OnlineUser) (data pojo.LoginBack, err error) {
	reqUserLoginStr, _ := json.Marshal(reqUserLogin)
	log.Printf("userLogin=%s;host=%s", string(reqUserLoginStr), hostInfo.HostName)
	var dbUser *pojo.SysUser
	db.Where("username = ?", reqUserLogin.Username).First(&dbUser)
	dbUserStr, _ := json.Marshal(dbUser)
	log.Printf("dbUser=%s", string(dbUserStr))
	if dbUser.ID == 0 {
		return data, errors.New("user_login_error")
	}
	if !utils.CheckPasswordHash(reqUserLogin.Password, dbUser.Password, hostInfo.Salt) {
		log.Printf("user login error.pass error userId = %d,pass=%s", dbUser.ID, reqUserLogin.Password)
		return data, errors.New("user_login_error")
	}
	if dbUser.UserType != 1 && dbUser.UserType != 2 && dbUser.UserType != 3 {
		return data, errors.New("user_error")
	}
	if !dbUser.Enabled {
		return data, errors.New("user_disabled_contact_admin")
	}
	data = GetUserInfo(*dbUser)
	token, err := utils.GetJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, dbUser.Username, dbUser.ID, dbUser.UserType, hostInfo.HostName)
	data.AccessToken = token
	key := utils.KeyRdOnline + utils.MD5(token)
	onlineUser.UserId = dbUser.ID
	onlineUser.Key = key
	//log.Printf("onlineUser=%v", onlineUser)
	userJSON, _ := json.Marshal(onlineUser)
	//log.Printf("userJSON=%v", string(userJSON))
	utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second)
	utils.TouchAdminOnlineUser(dbUser.ID)
	return data, err
}

func UserLogin(db *gorm.DB, hostInfo pojo.HostInfo, reqUserLogin pojo.UserLogin, onlineUser pojo.OnlineUser) (data pojo.LoginBack, err error) {
	reqUserLoginStr, _ := json.Marshal(reqUserLogin)
	log.Printf("userLogin=%s;host=%s", string(reqUserLoginStr), hostInfo.HostName)
	var dbUser *pojo.SysUser
	db.Where("username = ?", reqUserLogin.Username).First(&dbUser)
	dbUserStr, _ := json.Marshal(dbUser)
	log.Printf("dbUser=%s", string(dbUserStr))
	if dbUser.ID == 0 {
		return data, errors.New("user_login_error")
	}
	//needBind := false
	if !utils.CsConfig.PassGoogleAuth {
		if dbUser.GoogleCode != "" {
			//if reqUserLogin.Code == "" {
			//	return data, errors.New("请输入验证码")
			//}
			_, err2 := totp.Generate(totp.GenerateOpts{
				Issuer:      "gg",
				AccountName: dbUser.Username,
				Secret:      []byte(dbUser.GoogleCode),
			})
			if err2 != nil {
				return data, err2
			}
			valid := totp.Validate(reqUserLogin.Code, dbUser.GoogleCode)
			if !valid {
				if dbUser.BindCode {
					return data, errors.New("code_incorrect")
				}
				//needBind = true
			} else {
				db.Model(&dbUser).Update("bind_code", true)
				utils.UpdateTempUser(hostInfo.TablePrefix, *dbUser)
			}
		} else {
			//needBind = true
		}
	}
	//if needBind {
	//	key, err2 := totp.Generate(totp.GenerateOpts{
	//		Issuer:      "sg",
	//		AccountName: dbUser.Username,
	//	})
	//	if err2 != nil {
	//		return data, err2
	//	}
	//	db.Model(&dbUser).Update("google_code", key.Secret())
	//	utils.UpdateTempUser(hostInfo.TablePrefix, *dbUser)
	//	qrCode, err2 := qrcode.New(key.URL(), qrcode.Medium)
	//	if err2 != nil {
	//		return data, err2
	//	}
	//	pngData, err2 := qrCode.PNG(200)
	//	if err2 != nil {
	//		return data, err2
	//	}
	//	data.QrCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData)
	//	return data, errors.New("请先绑定二维码")
	//}
	//utils.EncodePass(hostInfo.Salt, reqUserLogin.Password)
	if !utils.CheckPasswordHash(reqUserLogin.Password, dbUser.Password, hostInfo.Salt) {
		log.Printf("user login error.pass error userId = %d,pass=%s", dbUser.ID, reqUserLogin.Password)
		return data, errors.New("user_login_error")
	}
	if dbUser.UserType != 1 && dbUser.UserType != 2 && dbUser.UserType != 3 {
		return data, errors.New("user_error")
	}
	if !dbUser.Enabled {
		return data, errors.New("user_disabled_contact_admin")
	}
	data = GetUserInfo(*dbUser)
	token, err := utils.GetJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, dbUser.Username, dbUser.ID, dbUser.UserType, hostInfo.HostName)
	data.AccessToken = token
	key := utils.KeyRdOnline + utils.MD5(token)
	onlineUser.UserId = dbUser.ID
	onlineUser.Key = key
	//log.Printf("onlineUser=%v", onlineUser)
	userJSON, _ := json.Marshal(onlineUser)
	//log.Printf("userJSON=%v", string(userJSON))
	utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second)
	utils.TouchAdminOnlineUser(dbUser.ID)
	return data, err
}

func GetUserInfo(dbUser pojo.SysUser) (data pojo.LoginBack) {
	userBak := pojo.UserBack{}
	_ = copier.Copy(&userBak, dbUser)
	data.Username = userBak.Username
	data.UserType = userBak.UserType
	data.Roles = dbUser.Roles
	return data
}

func DelUsers(db *gorm.DB, ids []int64) (result string, err error) {
	var users []pojo.SysUser
	if err := db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return result, err
	}
	for _, user := range users {
		if user.Username == "admin" {
			return result, errors.New("admin_user_cannot_delete")
		}
	}
	if err := db.Where("id IN ?", ids).Delete(&pojo.SysUser{}).Error; err != nil {
		return result, err
	}
	return result, nil
}

func UserAwardInfos(db *gorm.DB, search pojo.CashHistorySearch) (result pojo.CashHistoryPage, err error) {
	var cashHistoryList []pojo.CashHistory
	db = db.Model(&pojo.CashHistory{}).Where("user_id = ?", search.UserId)

	db.Model(&pojo.CashHistory{}).Count(&result.Total)
	db = db.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	db.Find(&cashHistoryList)

	for _, user := range cashHistoryList {
		var tempUserBack pojo.CashHistoryResp
		_ = copier.Copy(&tempUserBack, &user)
		result.List = append(result.List, tempUserBack)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage

	return result, err
}

// GetCashHistoryListAdmin 管理员获取余额变动记录列表（支持查询所有用户）
// 使用 all_cash_history 视图查询所有分表数据
func GetCashHistoryListAdmin(db *gorm.DB, search pojo.CashHistorySearch) (result pojo.CashHistoryPage) {
	var cashHistoryList []pojo.CashHistory
	start := time.Now()
	log.Printf("cashHistory admin list start userId=%d uid=%q cashMark=%q pageSize=%d currentPage=%d",
		search.UserId, search.Uid, search.CashMark, search.PageSize, search.CurrentPage)
	userIDs, userFiltered := resolveCashHistorySearchUserIDs(db, search)
	log.Printf("cashHistory admin list resolve users cost=%s userFiltered=%v userIDs=%v",
		time.Since(start), userFiltered, userIDs)
	if userFiltered && len(userIDs) == 0 {
		result.PageSize = search.PageSize
		result.CurrentPage = search.CurrentPage
		log.Printf("cashHistory admin list done cost=%s total=0 list=0 reason=no_user", time.Since(start))
		return result
	}
	if len(userIDs) != 1 {
		return getCashHistoryListAdminFromShards(db, search, userIDs, start)
	}

	queryStart := time.Now()
	query := db.Model(&pojo.CashHistory{}).Where("user_id = ?", userIDs[0])
	log.Printf("cashHistory admin list use sharding table userId=%d", userIDs[0])

	// 如果指定了余额备注，则按备注查询
	if search.CashMark != "" {
		query = query.Where("cash_mark LIKE ?", "%"+search.CashMark+"%")
	}
	log.Printf("cashHistory admin list build query cost=%s", time.Since(queryStart))

	// 统计总数
	countStart := time.Now()
	query.Count(&result.Total)
	log.Printf("cashHistory admin list count cost=%s total=%d", time.Since(countStart), result.Total)

	// 分页查询
	findStart := time.Now()
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&cashHistoryList)
	log.Printf("cashHistory admin list find cost=%s rows=%d", time.Since(findStart), len(cashHistoryList))

	// 转换为响应格式
	copyStart := time.Now()
	for _, history := range cashHistoryList {
		var tempResp pojo.CashHistoryResp
		_ = copier.Copy(&tempResp, &history)
		tempResp.Amount = utils.Truncate2(tempResp.Amount)
		tempResp.StartAmount = utils.Truncate2(tempResp.StartAmount)
		tempResp.EndAmount = utils.Truncate2(tempResp.EndAmount)
		result.List = append(result.List, tempResp)
	}
	log.Printf("cashHistory admin list copy cost=%s rows=%d", time.Since(copyStart), len(result.List))
	uidStart := time.Now()
	fillCashHistoryUIDs(db, result.List)
	log.Printf("cashHistory admin list fill uid cost=%s rows=%d", time.Since(uidStart), len(result.List))

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	log.Printf("cashHistory admin list done cost=%s total=%d list=%d", time.Since(start), result.Total, len(result.List))

	return result
}

func getCashHistoryListAdminFromShards(db *gorm.DB, search pojo.CashHistorySearch, userIDs []int64, start time.Time) (result pojo.CashHistoryPage) {
	var cashHistoryList []pojo.CashHistory
	log.Printf("cashHistory admin list use shard merge userIDsLen=%d", len(userIDs))

	countStart := time.Now()
	for i := 0; i < pojo.CashHistoryShards; i++ {
		var shardTotal int64
		query := applyCashHistoryAdminFilters(db.Table(cashHistoryShardTableName(i)), search, userIDs)
		query.Count(&shardTotal)
		result.Total += shardTotal
	}
	log.Printf("cashHistory admin list shard count cost=%s total=%d", time.Since(countStart), result.Total)

	findStart := time.Now()
	pageSize := search.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := search.PageSize * search.CurrentPage
	if offset < 0 {
		offset = 0
	}
	shardLimit := offset + pageSize
	if shardLimit <= 0 {
		shardLimit = pageSize
	}

	for i := 0; i < pojo.CashHistoryShards; i++ {
		var shardList []pojo.CashHistory
		query := applyCashHistoryAdminFilters(db.Table(cashHistoryShardTableName(i)), search, userIDs)
		query.Order("id desc").Limit(shardLimit).Find(&shardList)
		cashHistoryList = append(cashHistoryList, shardList...)
	}
	sort.Slice(cashHistoryList, func(i, j int) bool {
		return cashHistoryList[i].ID > cashHistoryList[j].ID
	})
	if offset < len(cashHistoryList) {
		end := offset + pageSize
		if end > len(cashHistoryList) {
			end = len(cashHistoryList)
		}
		cashHistoryList = cashHistoryList[offset:end]
	} else {
		cashHistoryList = nil
	}
	log.Printf("cashHistory admin list shard find cost=%s rows=%d", time.Since(findStart), len(cashHistoryList))

	copyStart := time.Now()
	for _, history := range cashHistoryList {
		var tempResp pojo.CashHistoryResp
		_ = copier.Copy(&tempResp, &history)
		tempResp.Amount = utils.Truncate2(tempResp.Amount)
		tempResp.StartAmount = utils.Truncate2(tempResp.StartAmount)
		tempResp.EndAmount = utils.Truncate2(tempResp.EndAmount)
		result.List = append(result.List, tempResp)
	}
	log.Printf("cashHistory admin list copy cost=%s rows=%d", time.Since(copyStart), len(result.List))
	uidStart := time.Now()
	fillCashHistoryUIDs(db, result.List)
	log.Printf("cashHistory admin list fill uid cost=%s rows=%d", time.Since(uidStart), len(result.List))

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	log.Printf("cashHistory admin list done cost=%s total=%d list=%d", time.Since(start), result.Total, len(result.List))
	return result
}

func applyCashHistoryAdminFilters(query *gorm.DB, search pojo.CashHistorySearch, userIDs []int64) *gorm.DB {
	if len(userIDs) > 0 {
		query = query.Where("user_id IN ?", userIDs)
	}
	if search.CashMark != "" {
		query = query.Where("cash_mark LIKE ?", "%"+search.CashMark+"%")
	}
	return query
}

func cashHistoryShardTableName(index int) string {
	return fmt.Sprintf("%s_%d", pojo.CashHistoryTableName, index)
}

func resolveCashHistorySearchUserIDs(db *gorm.DB, search pojo.CashHistorySearch) ([]int64, bool) {
	uid := strings.TrimSpace(search.Uid)
	start := time.Now()
	if search.UserId > 0 && uid == "" {
		log.Printf("cashHistory admin resolve users direct userId cost=%s", time.Since(start))
		return []int64{search.UserId}, true
	}
	if uid == "" {
		log.Printf("cashHistory admin resolve users no filter cost=%s", time.Since(start))
		return nil, false
	}

	var userIDs []int64
	err := db.Model(&pojo.TgUser{}).Where("uid = ?", uid).Pluck("id", &userIDs).Error
	log.Printf("cashHistory admin resolve users uid lookup cost=%s uid=%q userIDs=%v err=%v",
		time.Since(start), uid, userIDs, err)
	if search.UserId <= 0 {
		return userIDs, true
	}

	for _, userID := range userIDs {
		if userID == search.UserId {
			return []int64{search.UserId}, true
		}
	}
	return nil, true
}

func fillCashHistoryUIDs(db *gorm.DB, list []pojo.CashHistoryResp) {
	userIDs := make([]int64, 0, len(list))
	seen := make(map[int64]bool, len(list))
	for _, item := range list {
		if item.UserId <= 0 || seen[item.UserId] {
			continue
		}
		seen[item.UserId] = true
		userIDs = append(userIDs, item.UserId)
	}
	if len(userIDs) == 0 {
		return
	}

	var users []pojo.TgUser
	_ = db.Model(&pojo.TgUser{}).Select("id, uid").Where("id IN ?", userIDs).Find(&users).Error
	uidMap := make(map[int64]string, len(users))
	for _, user := range users {
		uidMap[user.ID] = user.Uid
	}
	for i := range list {
		list[i].Uid = uidMap[list[i].UserId]
	}
}

// GetCashHistoryListApp 当前TG用户流水列表（分页，排除抽成）
func GetCashHistoryListApp(db *gorm.DB, userID int64, search pojo.CashHistorySearch) (result pojo.CashHistoryPage) {
	var cashHistoryList []pojo.CashHistory
	query := db.Model(&pojo.CashHistory{}).
		Where("user_id = ?", userID).
		Where("type <> ?", pojo.CashHistoryTypeRedPacketCommission)

	if search.CashMark != "" {
		query = query.Where("cash_mark LIKE ?", "%"+search.CashMark+"%")
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&cashHistoryList)

	for _, history := range cashHistoryList {
		var tempResp pojo.CashHistoryResp
		_ = copier.Copy(&tempResp, &history)
		tempResp.Amount = utils.Truncate2(tempResp.Amount)
		tempResp.StartAmount = utils.Truncate2(tempResp.StartAmount)
		tempResp.EndAmount = utils.Truncate2(tempResp.EndAmount)
		result.List = append(result.List, tempResp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func ResetPwd(db *gorm.DB, salt string, newPassword string, userId int64) (err error) {
	result := db.Model(&pojo.SysUser{}).
		Where("id = ?", userId).
		Update("password", utils.EncodePass(salt, newPassword))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New(utils.I18nMessage("user_not_found_with_id", map[string]interface{}{"userId": userId}))
	}

	return nil
}
