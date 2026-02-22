package common

import (
	//api2 "BaseGoUni/app/api"
	"BaseGoUni/core/api"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantApi "BaseGoUni/tenant/api"
	//"BaseGoUni/docs"
	//_ "BaseGoUni/docs" // 导入生成的docs
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"
)

func InitGin() {
	gin.DefaultWriter = ioutil.Discard
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	startWsHub()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Content-Type", "X-XSRF-TOKEN", "Accept", "X-Requested-With", "Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	router.Use(hostInfoMiddleware())
	router.Use(cors.New(corsConfig))
	//docs.SwaggerInfo.Title = "rcs服务api"
	//docs.SwaggerInfo.Description = "rcs服务api"
	//docs.SwaggerInfo.Version = "1.0"
	//docs.SwaggerInfo.Host = "localhost:8080"
	//docs.SwaggerInfo.BasePath = "{{host}}"
	//docs.SwaggerInfo.Schemes = []string{"http", "https"}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/ws", WsHandler)
	_ = mime.AddExtensionType(".js", "application/javascript")
	router.Use(static.ServeRoot("/", "dist"))
	apiGroup := router.Group("/api/v1")
	{
		apiGroup.GET("/heath/check", heathCheck)
		apiGroup.POST("/user/award", api.AwardUser)            // 内部用户余额变动
		apiGroup.POST("/user/login", api.UserLogin)            // 管理员登录
		apiGroup.POST("/tenant/login", api.SysTenantUserLogin) // 租户用户登录
	}
	// 通用接口
	commonGroup := router.Group("/api/v1/outside")
	commonGroup.Use(authMiddleware([]int{1, 2, 3, 4}, false, true))
	{
		commonGroup.POST("/menus", api.GetMenus)
		commonGroup.GET("/routes", api.GetRoutes)
		commonGroup.POST("/user", api.GetUsers)
		commonGroup.POST("/roles", api.GetRoles)
		commonGroup.POST("/userInfo", api.CurrentUserInfo)
		commonGroup.GET("/roleIds/:userId", api.GetRoleIds)
		commonGroup.POST("/role-menu", api.GetRoleMenus)
		commonGroup.GET("/role-menu-ids/:roleId", api.GetRoleMenuIds)
		commonGroup.POST("/pkManagers", api.GetPkManagers)                 // 获取PK管理器列表
		commonGroup.GET("/pkManager/url/:name", api.GetPkManagerUrlByName) // 根据名称获取URL
		commonGroup.GET("/pkManager/:id", api.GetPkManagerById)            // 根据ID获取PK管理器
		//commonGroup.POST("/lucky/send", api.SendRedPacket)                 // 发送红包
		//commonGroup.POST("/lucky/grab", api.GrabRedPacket)                 // 抢红包
		commonGroup.POST("/lucky/list", api.GetRedPacketList)         // 获取红包列表
		commonGroup.GET("/lucky/:id", api.GetRedPacketDetail)         // 获取红包详情
		commonGroup.GET("/lucky/status/:id", api.GetRedPacketStatus)  // 获取红包状态
		commonGroup.POST("/lucky/checkBalance", api.CheckGrabBalance) // 检查抢包余额
	}
	commonGroupLog := router.Group("/api/v1/outside")
	commonGroupLog.Use(authMiddleware([]int{1, 2, 3}, false, true), manageLog())
	{
		commonGroupLog.GET("/user/unbind/gauth", api.UnBindGAuth) // 解绑自己谷歌验证
		commonGroupLog.PUT("/user/pass", api.ChangePass)          // 修改密码
		commonGroupLog.POST("/resetPwd", api.ResetPassword)
	}
	// 管理员接口
	manageGroup := router.Group("/api/v1/manager")
	manageGroup.Use(authMiddleware([]int{1, 2}, false, true))
	{
		manageGroup.POST("/cashHistory", api.UserCashHistory)
	}
	manageGroupLog := router.Group("/api/v1/manager")
	manageGroupLog.Use(authMiddleware([]int{1, 2}, false, true), manageLog())
	{
		manageGroupLog.GET("/user/unbind/gauth/:userId", api.UnBindGAuth) // 解绑用户谷歌验证
		manageGroupLog.POST("/setRole", api.SetRole)
		manageGroupLog.DELETE("/role/:id", api.DelRole)
		manageGroupLog.POST("/setUser", api.SetUser)
		manageGroupLog.POST("/delUsers", api.DelUsers)
		manageGroupLog.PUT("/setMenus", api.SetMenus)
		manageGroupLog.DELETE("/menu/:id", api.DelMenu)
		manageGroupLog.POST("/user/award", api.AdminAwardUser)    // 用户余额管理(送钱/扣钱)
		manageGroupLog.POST("/pkManager", api.SetPkManager)       // 创建或更新PK管理器
		manageGroupLog.DELETE("/pkManager/:id", api.DelPkManager) // 删除PK管理器
	}
	// 超级管理员接口
	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(authMiddleware([]int{1}, false, true))
	{
		adminGroup.POST("/manage/logs", api.GetManageLogs) // 获取管理员操作日志
		//adminGroup.POST("/host_infos", api2.GetHostInfos)
		adminGroup.GET("/onlineDevices", api.GetOnlineDevices)
		adminGroup.POST("/withdrawalTask", api.SendWithdrawalTask)
		adminGroup.POST("/verifyCodeTask", api.SendVerifyCodeTask)
		adminGroup.POST("/tgUser/list", api.GetTgUsers)                    // 获取Telegram用户列表
		adminGroup.GET("/tgUser/:id", api.GetTgUserById)                   // 获取Telegram用户详情
		adminGroup.POST("/tgUserRebate/list", api.GetTgUserRebateRecords)  // 获取Telegram反水记录列表
		adminGroup.GET("/tgUserRebate/:id", api.GetTgUserRebateRecordById) // 获取Telegram反水记录详情
		adminGroup.POST("/lucky/list", api.GetLuckyMoneyListAdmin)         // 管理员获取红包列表
		adminGroup.POST("/lucky/history", api.GetLuckyHistoryListAdmin)    // 管理员获取领取历史
		adminGroup.GET("/lucky/:id", api.GetLuckyMoneyDetailAdmin)         // 管理员获取红包详情
		adminGroup.POST("/cashHistory/list", api.GetCashHistoryListAdmin)  // 管理员获取余额变动记录列表
		adminGroup.POST("/authGroup/list", api.GetAuthGroups)              // 获取授权群组列表
		adminGroup.POST("/authGroup", api.SetAuthGroup)                    // 创建或更新授权群组
		adminGroup.DELETE("/authGroup/:id", api.DelAuthGroup)              // 删除授权群组
		adminGroup.POST("/tenant/list", api.GetSysTenants)                 // 获取租户列表
		adminGroup.GET("/tenant/:id", api.GetSysTenantById)                // 获取租户详情
		adminGroup.POST("/tenantUser/list", api.GetSysTenantUsers)         // 获取租户用户列表
		adminGroup.GET("/tenantUser/:id", api.GetSysTenantUserById)        // 获取租户用户详情
		adminGroup.POST("/rechargeOrder/list", api.GetRechargeOrders)      // 获取充值订单列表
		adminGroup.GET("/rechargeOrder/:id", api.GetRechargeOrderById)     // 获取充值订单详情
		adminGroup.POST("/withdrawOrderBr/list", api.GetWithdrawOrderBrs)  // 获取巴西提现订单列表
		adminGroup.GET("/withdrawOrderBr/:id", api.GetWithdrawOrderBrById) // 获取巴西提现订单详情
		adminGroup.POST("/tgUser/listWithSubStats", api.GetTgUsersWithSubStats)
		adminGroup.POST("/tgUser/subStatsSummary", api.GetTgUsersWithSubStatsSummary)
	}
	adminGroupLog := router.Group("/api/v1/admin")
	adminGroupLog.Use(authMiddleware([]int{1}, false, true), manageLog())
	{
		adminGroupLog.PUT("/menus", api.SetMenus)
		adminGroupLog.PUT("/role", api.SetRole)
		adminGroupLog.DELETE("/role/:id", api.DelRole)
		adminGroupLog.PUT("/user", api.SetUser)
		adminGroupLog.POST("/upload", api.AdminUpload)                       // 文件上传（R2）
		adminGroupLog.POST("/tgUser", api.SetTgUser)                         // 创建或更新Telegram用户
		adminGroupLog.POST("/tgUser/status", api.SetTgUserStatus)            // 封禁/解封Telegram用户
		adminGroupLog.DELETE("/tgUser/:id", api.DelTgUser)                   // 删除Telegram用户
		adminGroupLog.POST("/tgUserRebate", api.SetTgUserRebateRecord)       // 创建或更新Telegram反水记录
		adminGroupLog.DELETE("/tgUserRebate/:id", api.DelTgUserRebateRecord) // 删除Telegram反水记录
		adminGroupLog.POST("/tenant", api.SetSysTenant)                      // 创建或更新租户
		adminGroupLog.POST("/tenant/resetPassword", api.ResetSysTenantPassword)
		adminGroupLog.DELETE("/tenant/:id", api.DelSysTenant)                // 删除租户
		adminGroupLog.POST("/tenantUser", api.SetSysTenantUser)              // 创建或更新租户用户
		adminGroupLog.DELETE("/tenantUser/:id", api.DelSysTenantUser)        // 删除租户用户
		adminGroupLog.POST("/rechargeOrder", api.SetRechargeOrder)           // 创建或更新充值订单
		adminGroupLog.DELETE("/rechargeOrder/:id", api.DelRechargeOrder)     // 删除充值订单
		adminGroupLog.POST("/withdrawOrderBr", api.SetWithdrawOrderBr)       // 创建或更新巴西提现订单
		adminGroupLog.DELETE("/withdrawOrderBr/:id", api.DelWithdrawOrderBr) // 删除巴西提现订单
		//adminGroupLog.PUT("/host_info", api2.SetHostInfo)
		//adminGroupLog.DELETE("/host_info/:id", api2.DelHostInfo)
	}

	tenantGroup := router.Group("/api/v1/tenant")
	tenantGroup.Use(tenantAuthMiddleware(false), manageLog())
	{
		tenantGroup.POST("/authGroup/list", tenantApi.GetAuthGroups)
		tenantGroup.POST("/authGroup", tenantApi.SetAuthGroup)
		tenantGroup.DELETE("/authGroup/:id", tenantApi.DelAuthGroup)

		tenantGroup.POST("/lucky/list", tenantApi.GetLuckyMoneyList)
		tenantGroup.POST("/lucky/history", tenantApi.GetLuckyHistoryList)
		tenantGroup.GET("/lucky/:id", tenantApi.GetLuckyMoneyDetail)

		tenantGroup.POST("/rechargeOrder/list", tenantApi.GetRechargeOrders)
		tenantGroup.GET("/rechargeOrder/:id", tenantApi.GetRechargeOrderById)
		tenantGroup.POST("/rechargeOrder", tenantApi.SetRechargeOrder)
		tenantGroup.DELETE("/rechargeOrder/:id", tenantApi.DelRechargeOrder)

		tenantGroup.POST("/tgUser/list", tenantApi.GetTgUsers)
		tenantGroup.POST("/tgUser/listWithSubStats", tenantApi.GetTgUsersWithSubStats)
		tenantGroup.POST("/tgUser/subStatsSummary", tenantApi.GetTgUsersWithSubStatsSummary)
		tenantGroup.GET("/tgUser/:id", tenantApi.GetTgUserById)
		tenantGroup.POST("/tgUser", tenantApi.SetTgUser)
		tenantGroup.POST("/tgUser/status", tenantApi.SetTgUserStatus)
		tenantGroup.DELETE("/tgUser/:id", tenantApi.DelTgUser)

		tenantGroup.POST("/tgUserRebate/list", tenantApi.GetTgUserRebateRecords)
		tenantGroup.GET("/tgUserRebate/:id", tenantApi.GetTgUserRebateRecordById)
		tenantGroup.POST("/tgUserRebate", tenantApi.SetTgUserRebateRecord)
		tenantGroup.DELETE("/tgUserRebate/:id", tenantApi.DelTgUserRebateRecord)

		tenantGroup.POST("/withdrawOrderBr/list", tenantApi.GetWithdrawOrderBrs)
		tenantGroup.GET("/withdrawOrderBr/:id", tenantApi.GetWithdrawOrderBrById)
		tenantGroup.POST("/withdrawOrderBr", tenantApi.SetWithdrawOrderBr)
		tenantGroup.DELETE("/withdrawOrderBr/:id", tenantApi.DelWithdrawOrderBr)
	}

	appRouter := router.Group("/api/v1/app")
	// appRouter.Use(authMiddleware([]int{1, 2, 3, 4}, false, true))
	{
		appRouter.GET("/getPkByName/:name", api.GetPkManagerUrlByName) // 获取PK管理器列表
	}

	log.Printf("Start server at %s:%d ", utils.GlobalConfig.Host, utils.GlobalConfig.Port)
	apiURL := fmt.Sprintf("%s:%d", utils.GlobalConfig.Host, utils.GlobalConfig.Port)
	err := router.Run(apiURL)
	if err != nil {
		log.Printf("Init gin error.err=%v\n", err)
		return
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(p []byte) (n int, err error) {
	rw.body.Write(p)
	return rw.ResponseWriter.Write(p)
}

func manageLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有文件上传
		_, fileErr := c.FormFile("file")
		if fileErr == nil {
			c.Next()
			return
		}
		method := c.Request.Method
		requestBody := ""
		if method == "GET" {
			queryParams := c.Request.URL.Query()
			queryParamsStr, _ := json.Marshal(queryParams)
			requestBody = string(queryParamsStr)
		} else {
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			requestBody = string(body)
		}
		path := c.Request.URL.Path
		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = rw
		c.Next()
		status := c.Writer.Status()
		responseBody := rw.body.String()
		tempHostInfo := c.MustGet("hostInfo").(pojo.HostInfo)
		currentUser, _ := utils.GetCurrentUser(c)
		ip := utils.GetIPAddress(c)
		manageLogData := pojo.ManageLog{
			Username:     currentUser.Username,
			RequestHost:  fmt.Sprintf("%s %d %s", method, status, path),
			RequestBody:  requestBody,
			ResponseBody: responseBody,
			Ip:           ip,
		}
		manageLogStr, _ := json.Marshal(manageLogData)
		//log.Printf("manageLogStr=%s", string(manageLogStr))
		_ = utils.PublishMQ(utils.MQMessage{
			MessageType: utils.KeyManageLogNotify,
			Data:        string(manageLogStr),
			DataMore:    tempHostInfo.TablePrefix,
		})
	}
}

func hostInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//log.Printf("%s %s", c.Request.Method, c.Request.RequestURI)
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}
		host := utils.GetRequestHost(c)
		hostInfo := utils.GetTempHostInfo(host)
		if hostInfo.ID == 0 {
			hostInfoStr, _ := json.Marshal(hostInfo)
			log.Printf("host:%s;hostInfo:%s", host, string(hostInfoStr))
			c.String(http.StatusNotFound, "")
			c.Abort()
			return
		}
		if isStaticFileRequest(c.Request.URL.Path) {
			if hostInfo.ShowAdmin {
				log.Printf("showAdmin:%s", hostInfo.HostName)
				c.Next()
				return
			}
			log.Printf("not showAdmin:%s", hostInfo.HostName)
			c.Status(404)
			c.Abort()
			return
		}
		db := utils.NewPrefixDb(hostInfo.TablePrefix)
		c.Set("hostInfo", hostInfo)
		c.Set("db", db)
		c.Next()
	}
}

func isStaticFileRequest(path string) bool {
	return (path == "/" || strings.Contains(path, "html") || strings.HasPrefix(path, "swagger") || strings.Contains(path, "#")) && (!strings.HasPrefix(path, "/library"))
}

// 通行用户类型 / 是否单点登录 / 是否过滤特殊token
func authMiddleware(types []int, singleLogin bool, passChild bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedBack(c, "Authorization header is missing")
			c.Abort()
			return
		}
		authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		hostInfo := utils.GetTempHostInfo(utils.GetRequestHost(c))
		userId, userType, hostName, childCode, _ := utils.ParseToken(utils.CsConfig.DefaultHost.AccessSecret, authHeader)
		if passChild && childCode != "" {
			utils.UnauthorizedBack(c, "token is invalid -1")
			c.Abort()
			return
		}
		if hostInfo.HostName != hostName {
			utils.UnauthorizedBack(c, "token is invalid 0")
			c.Abort()
			return
		}
		if userId == 0 {
			utils.UnauthorizedBack(c, "token is invalid 1")
			c.Abort()
			return
		}
		inType := false
		for _, tempUserType := range types {
			if userType == tempUserType {
				inType = true
				break
			}
		}
		if !inType {
			utils.UnauthorizedBack(c, "not support api")
			c.Abort()
			return
		}
		user := utils.GetTempUser(hostInfo.TablePrefix, userId)
		if !user.Enabled {
			utils.UnauthorizedBack(c, "token is invalid 2")
			c.Abort()
			return
		}
		if singleLogin {
			key := utils.KeySingle + utils.MD5(fmt.Sprintf("%d", userId))
			data := utils.RD.Get(context.Background(), key)
			if data == nil || data.Err() != nil {
				utils.UnauthorizedBack(c, "token is passed")
				c.Abort()
				return
			}
			if data.Val() != authHeader {
				utils.UnauthorizedBack(c, "already logout")
				c.Abort()
				return
			}
		}
		//requestKey := utils.MD5(fmt.Sprintf("%s_%s", c.Request.Method, c.Request.RequestURI))
		//lockKey := fmt.Sprintf(utils.KeyLockRequest, userId, requestKey)
		//lock, _ := utils.AcquireLock(lockKey, 1*time.Second)
		//if !lock {
		//	c.JSON(http.StatusBadRequest, gin.H{"error": "request too fast.Please try again later."})
		//	return
		//}
		c.Set("childCode", childCode)
		c.Set("userId", userId)
		c.Set("userType", userType)
		c.Set("token", authHeader)
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		if latencyTime > 1*time.Second {
			log.Printf("Request %s %s took %v", c.Request.Method, c.Request.URL, latencyTime)
		}
	}
}

func tenantAuthMiddleware(singleLogin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedBack(c, "Authorization header is missing")
			c.Abort()
			return
		}
		authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		hostInfo := utils.GetTempHostInfo(utils.GetRequestHost(c))
		userId, userType, hostName, _, _ := utils.ParseToken(utils.CsConfig.DefaultHost.AccessSecret, authHeader)
		if hostInfo.HostName != hostName {
			utils.UnauthorizedBack(c, "token is invalid 0")
			c.Abort()
			return
		}
		if userId == 0 {
			utils.UnauthorizedBack(c, "token is invalid 1")
			c.Abort()
			return
		}

		user := utils.GetTempTenantUser(hostInfo.TablePrefix, userId)
		if user.Status != 1 {
			utils.UnauthorizedBack(c, "User forbidden")
			c.Abort()
			return
		}
		if singleLogin {
			key := utils.KeyRdTenantOnline + utils.MD5(fmt.Sprintf("%d", userId))
			data := utils.RD.Get(context.Background(), key)
			if data == nil || data.Err() != nil {
				utils.UnauthorizedBack(c, "token is passed")
				c.Abort()
				return
			}
			if data.Val() != authHeader {
				utils.UnauthorizedBack(c, "already logout")
				c.Abort()
				return
			}
		}
		//requestKey := utils.MD5(fmt.Sprintf("%s_%s", c.Request.Method, c.Request.RequestURI))
		//lockKey := fmt.Sprintf(utils.KeyLockRequest, userId, requestKey)
		//lock, _ := utils.AcquireLock(lockKey, 1*time.Second)
		//if !lock {
		//	c.JSON(http.StatusBadRequest, gin.H{"error": "request too fast.Please try again later."})
		//	return
		//}
		c.Set("userId", userId)
		c.Set("userType", userType)
		c.Set("token", authHeader)
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		if latencyTime > 1*time.Second {
			log.Printf("Request %s %s took %v", c.Request.Method, c.Request.URL, latencyTime)
		}
	}
}

func heathCheck(c *gin.Context) {
	requestHost := utils.GetRequestHost(c)
	log.Printf("Request url: %s", requestHost)
	log.Printf("Request Host: %s", c.Request.Host)
	c.String(200, "ok")
}
