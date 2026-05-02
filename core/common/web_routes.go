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
	router.GET("/api/v1/ws", WsHandler)
	_ = mime.AddExtensionType(".js", "application/javascript")
	router.Use(static.ServeRoot("/", "dist"))
	apiGroup := router.Group("/api/v1")
	{
		apiGroup.GET("/heath/check", heathCheck)
		apiGroup.POST("/user/award", api.AwardUser)            // 内部用户余额变动
		apiGroup.POST("/user/login", api.UserLogin)            // 管理员登录
		apiGroup.POST("/tenant/login", api.SysTenantUserLogin) // 租户用户登录
		apiGroup.POST("/admin/lucky/historyUserFlow", api.GetLuckyHistoryUserFlowListAdmin)
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
		adminGroup.GET("/onlineUsers/stats", api.GetAdminOnlineUserStats)
		adminGroup.GET("/dashboard/stats", api.GetAdminDashboardStats)
		adminGroup.POST("/dashboard/onlineUsers", api.GetAdminDashboardOnlineUsers)
		adminGroup.POST("/dashboard/rechargeUsers", api.GetAdminDashboardRechargeUsers)
		adminGroup.POST("/withdrawalTask", api.SendWithdrawalTask)
		adminGroup.POST("/verifyCodeTask", api.SendVerifyCodeTask)
		adminGroup.POST("/tgUser/list", api.GetTgUsers)                    // 获取Telegram用户列表
		adminGroup.GET("/tgUser/:id", api.GetTgUserById)                   // 获取Telegram用户详情
		adminGroup.POST("/tgUserRebate/list", api.GetTgUserRebateRecords)  // 获取Telegram反水记录列表
		adminGroup.GET("/tgUserRebate/:id", api.GetTgUserRebateRecordById) // 获取Telegram反水记录详情
		adminGroup.POST("/lucky/list", api.GetLuckyMoneyListAdmin)         // 管理员获取红包列表
		adminGroup.POST("/lucky/history", api.GetLuckyHistoryListAdmin)    // 管理员获取领取历史
		adminGroup.GET("/lucky/:id", api.GetLuckyMoneyDetailAdmin)         // 管理员获取红包详情
		adminGroup.POST("/luckyItem/list", api.GetLuckyMoneyItems)         // 管理员获取红包明细列表
		adminGroup.GET("/luckyItem/:id", api.GetLuckyMoneyItemById)        // 管理员获取红包明细详情
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
		adminGroup.POST("/payChannel/list", api.GetPayChannels)            // 获取支付通道列表
		adminGroup.GET("/payChannel/:id", api.GetPayChannelById)           // 获取支付通道详情
		adminGroup.POST("/sysCountry/list", api.GetSysCountries)
		adminGroup.GET("/sysCountry/:id", api.GetSysCountryById)
		adminGroup.POST("/sysBanner/list", api.GetSysBanners)                          // 获取轮播图列表
		adminGroup.GET("/sysBanner/:id", api.GetSysBannerById)                         // 获取轮播图详情
		adminGroup.POST("/sysConfig/list", api.GetSysConfigs)                          // 获取系统配置列表
		adminGroup.GET("/sysConfig/:id", api.GetAppSysConfig)                          // 根据ID获取系统配置（复用key接口占位）
		adminGroup.POST("/sysPayChannel/list", api.GetSysPayChannels)                  // 获取支付通道列表
		adminGroup.GET("/sysPayChannel/:id", api.GetSysPayChannelById)                 // 获取支付通道详情
		adminGroup.POST("/sysSourceChannel/list", api.GetSysSourceChannels)            // 获取投流来源渠道列表
		adminGroup.GET("/sysSourceChannel/:id/stats", api.GetSysSourceChannelStats)    // 获取投流来源渠道统计
		adminGroup.GET("/sysSourceChannel/:id", api.GetSysSourceChannelById)           // 获取投流来源渠道详情
		adminGroup.POST("/attributionEvent/list", api.GetAttributionEvents)            // 获取事件归因明细列表
		adminGroup.POST("/attributionEvent/summary", api.GetAttributionEventSummary)   // 获取事件归因汇总
		adminGroup.POST("/sysPayMethod/list", api.GetSysPayMethods)                    // 获取支付方式列表
		adminGroup.GET("/sysPayMethod/:id", api.GetSysPayMethodById)                   // 获取支付方式详情
		adminGroup.GET("/sysPayChannelMethod/:channelId", api.GetSysPayChannelMethods) // 获取通道绑定的支付方式
		adminGroup.POST("/sysVipLevel/list", api.GetSysVipLevels)                      // 获取VIP等级列表
		adminGroup.GET("/sysVipLevel/:id", api.GetSysVipLevelById)                     // 获取VIP等级详情
		adminGroup.GET("/prizePoolConfig/:poolId", api.GetPrizePoolConfig)             // 获取奖池概率配置
		adminGroup.GET("/prizePoolBalance/:poolCode", api.GetPrizePoolByCodeAdmin)     // 获取奖池余额
		adminGroup.POST("/userLotteryRecord/list", api.GetUserLotteryRecords)          // 抽奖记录列表
		adminGroup.GET("/userLotteryRecord/:id", api.GetUserLotteryRecordById)         // 抽奖记录详情
		adminGroup.POST("/sysCustomField/list", api.GetSysCustomFields)
		adminGroup.GET("/sysCustomField/:id", api.GetSysCustomFieldById)
		adminGroup.POST("/platformProfitLedger/list", api.GetPlatformProfitLedgers)
		adminGroup.GET("/platformProfitLedger/:id", api.GetPlatformProfitLedgerById)
		adminGroup.POST("/tgUser/listWithSubStats", api.GetTgUsersWithSubStats)
		adminGroup.POST("/tgUser/subStatsSummary", api.GetTgUsersWithSubStatsSummary)
		adminGroup.POST("/userWithdrawAccount/list", api.GetSysUserWithdrawAccounts)  // 获取用户提现账户列表
		adminGroup.GET("/userWithdrawAccount/:id", api.GetSysUserWithdrawAccountById) // 获取用户提现账户详情
	}
	adminGroupLog := router.Group("/api/v1/admin")
	adminGroupLog.Use(authMiddleware([]int{1}, false, true), manageLog())
	{
		adminGroupLog.PUT("/menus", api.SetMenus)
		adminGroupLog.PUT("/role", api.SetRole)
		adminGroupLog.DELETE("/role/:id", api.DelRole)
		adminGroupLog.PUT("/user", api.SetUser)
		adminGroupLog.POST("/upload", api.AdminUpload) // 文件上传（R2）
		adminGroupLog.POST("/tgUser", api.SetTgUser)   // 创建或更新Telegram用户
		adminGroupLog.POST("/tgUser/batchCreateBot", api.BatchCreateBotTgUsers)
		adminGroupLog.POST("/tgUser/batchUpdateBot", api.BatchUpdateBotTgUsers)
		adminGroupLog.POST("/tgUser/status", api.SetTgUserStatus)            // 封禁/解封Telegram用户
		adminGroupLog.DELETE("/tgUser/:id", api.DelTgUser)                   // 删除Telegram用户
		adminGroupLog.POST("/tgUserRebate", api.SetTgUserRebateRecord)       // 创建或更新Telegram反水记录
		adminGroupLog.DELETE("/tgUserRebate/:id", api.DelTgUserRebateRecord) // 删除Telegram反水记录
		adminGroupLog.POST("/luckyItem", api.SetLuckyMoneyItem)              // 创建或更新红包明细
		adminGroupLog.DELETE("/luckyItem/:id", api.DelLuckyMoneyItem)        // 删除红包明细
		adminGroupLog.POST("/tenant", api.SetSysTenant)                      // 创建或更新租户
		adminGroupLog.POST("/tenant/resetPassword", api.ResetSysTenantPassword)
		adminGroupLog.DELETE("/tenant/:id", api.DelSysTenant)                             // 删除租户
		adminGroupLog.POST("/tenantUser", api.SetSysTenantUser)                           // 创建或更新租户用户
		adminGroupLog.DELETE("/tenantUser/:id", api.DelSysTenantUser)                     // 删除租户用户
		adminGroupLog.POST("/rechargeOrder", api.SetRechargeOrder)                        // 创建或更新充值订单
		adminGroupLog.DELETE("/rechargeOrder/:id", api.DelRechargeOrder)                  // 删除充值订单
		adminGroupLog.POST("/rechargeOrder/:id/callback", api.AdminRechargeOrderCallback) // 手动回调充值订单
		adminGroupLog.POST("/withdrawOrderBr", api.SetWithdrawOrderBr)                    // 创建或更新巴西提现订单
		adminGroupLog.DELETE("/withdrawOrderBr/:id", api.DelWithdrawOrderBr)              // 删除巴西提现订单
		adminGroupLog.POST("/payChannel", api.SetPayChannel)                              // 创建或更新支付通道
		adminGroupLog.DELETE("/payChannel/:id", api.DelPayChannel)                        // 删除支付通道
		adminGroupLog.POST("/sysPayChannel", api.SetSysPayChannel)                        // 创建或更新支付通道
		adminGroupLog.DELETE("/sysPayChannel/:id", api.DelSysPayChannel)                  // 删除支付通道
		adminGroupLog.POST("/sysSourceChannel", api.SetSysSourceChannel)                  // 创建或更新投流来源渠道
		adminGroupLog.DELETE("/sysSourceChannel/:id", api.DelSysSourceChannel)            // 删除投流来源渠道
		adminGroupLog.POST("/sysPayMethod", api.SetSysPayMethod)                          // 创建或更新支付方式
		adminGroupLog.DELETE("/sysPayMethod/:id", api.DelSysPayMethod)                    // 删除支付方式
		adminGroupLog.POST("/sysPayChannelMethod", api.SetSysPayChannelMethods)           // 设置通道支付方式
		adminGroupLog.DELETE("/sysPayChannelMethod/:id", api.DelSysPayChannelMethod)      // 删除通道-方式绑定
		adminGroupLog.POST("/sysVipLevel", api.SetSysVipLevel)                            // 创建或更新VIP等级
		adminGroupLog.DELETE("/sysVipLevel/:id", api.DelSysVipLevel)                      // 删除VIP等级
		adminGroupLog.POST("/prizePoolConfig", api.SetPrizePoolConfig)                    // 创建或更新奖池概率配置
		adminGroupLog.POST("/prizePoolBalance", api.SetPrizePoolBalanceAdmin)             // 设置奖池余额
		adminGroupLog.DELETE("/prizePoolConfig/:id", api.DelPrizePoolConfig)              // 删除奖池概率配置
		adminGroupLog.POST("/sysCountry", api.SetSysCountry)
		adminGroupLog.DELETE("/sysCountry/:id", api.DelSysCountry)
		adminGroupLog.POST("/sysBanner", api.SetSysBanner)
		adminGroupLog.DELETE("/sysBanner/:id", api.DelSysBanner)
		adminGroupLog.POST("/sysConfig", api.SetSysConfig)
		adminGroupLog.DELETE("/sysConfig/:id", api.DelSysConfig)
		adminGroupLog.POST("/sysCustomField", api.SetSysCustomField)
		adminGroupLog.DELETE("/sysCustomField/:id", api.DelSysCustomField)
		adminGroupLog.POST("/platformProfitLedger", api.SetPlatformProfitLedger)
		adminGroupLog.DELETE("/platformProfitLedger/:id", api.DelPlatformProfitLedger)
		adminGroupLog.POST("/userWithdrawAccount", api.AdminSetSysUserWithdrawAccount)       // 创建或更新用户提现账户
		adminGroupLog.DELETE("/userWithdrawAccount/:id", api.AdminDelSysUserWithdrawAccount) // 删除用户提现账户
		//adminGroupLog.PUT("/host_info", api2.SetHostInfo)
		//adminGroupLog.DELETE("/host_info/:id", api2.DelHostInfo)
	}

	tenantGroup := router.Group("/api/v1/tenant")
	tenantGroup.Use(tenantAuthMiddleware(false), manageLog())
	{
		tenantGroup.GET("/dashboard/stats", tenantApi.GetDashboardStats)
		tenantGroup.POST("/dashboard/onlineUsers", tenantApi.GetDashboardOnlineUsers)
		tenantGroup.POST("/dashboard/rechargeUsers", tenantApi.GetDashboardRechargeUsers)

		tenantGroup.POST("/authGroup/list", tenantApi.GetAuthGroups)
		tenantGroup.POST("/authGroup", tenantApi.SetAuthGroup)
		tenantGroup.DELETE("/authGroup/:id", tenantApi.DelAuthGroup)

		tenantGroup.POST("/lucky/list", tenantApi.GetLuckyMoneyList)
		tenantGroup.POST("/lucky/history", tenantApi.GetLuckyHistoryList)
		tenantGroup.POST("/lucky/historyUserFlow", tenantApi.GetLuckyHistoryUserFlowList)
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

		tenantGroup.POST("/cashHistory/list", tenantApi.GetCashHistoryList)
		tenantGroup.GET("/onlineUsers/stats", api.GetTenantOnlineUserStats)

		tenantGroup.POST("/withdrawOrderBr/list", tenantApi.GetWithdrawOrderBrs)
		tenantGroup.GET("/withdrawOrderBr/:id", tenantApi.GetWithdrawOrderBrById)
		tenantGroup.POST("/withdrawOrderBr", tenantApi.SetWithdrawOrderBr)
		tenantGroup.DELETE("/withdrawOrderBr/:id", tenantApi.DelWithdrawOrderBr)
	}

	// 支付回调（公开，三方主动调用，无 token）
	payCallbackRouter := router.Group("/api/v1/pay")
	{
		payCallbackRouter.POST("/gctpk/notify", api.GctpkPayinCallback)       // GCTPK 代收回调（兼容自动识别）
		payCallbackRouter.POST("/gctpkmxn/notify", api.GctpkMxnPayinCallback) // GCTPK MXN 代收回调
		payCallbackRouter.POST("/gctpkbrl/notify", api.GctpkBrlPayinCallback) // GCTPK BRL 代收回调
	}

	appRouter := router.Group("/api/v1/app")
	{
		//appRouter.GET("/getPkByName/:name", api.GetPkManagerUrlByName) // 获取PK管理器列表
		appRouter.POST("/tg/login", api.TgAuthLogin)
		appRouter.POST("/tg/loginByEmail", api.LoginTgByEmail)
		appRouter.POST("/tg/phoneLogin", api.LoginTgByPhone)
		appRouter.POST("/tg/sendEmailCode", api.SendTgEmailCode)
		appRouter.POST("/tg/sendSMSCode", api.SendTgSMSCode)
		appRouter.POST("/tg/registerByEmail", api.RegisterTgByEmail)
		appRouter.POST("/tg/registerByPhone", api.RegisterTgByPhone)
		appRouter.POST("/attribution/event", api.CreateAttributionEvent)
		appRouter.POST("/tg/forgotPasswordByEmail", api.ForgotPasswordByEmail)
		appRouter.POST("/tg/forgotPasswordByPhone", api.ForgotPasswordByPhone)
		appRouter.POST("/lucky/list", api.GetRedPacketListApp)          // 不校验token
		appRouter.POST("/lucky/detail", api.GetLuckyDetailApp)          // 不校验token
		appRouter.GET("/prizePool/balance", api.GetPrizePoolBalanceApp) // 不校验token
		appRouter.POST("/banners", api.GetAppBanners)                   // 轮播图按position分组
		appRouter.GET("/config/:key", api.GetAppSysConfig)              // 根据key获取系统配置
	}

	appAuthRouter := router.Group("/api/v1/app")
	appAuthRouter.Use(appAuthMiddle(true))
	{
		appAuthRouter.POST("/rechargeOrder", api.AppCreateRechargeOrder)
		appAuthRouter.GET("/recharge/isFirst", api.CheckIsFirstRecharge)
		appAuthRouter.POST("/lucky/send", api.SendRedPacketApp)
		appAuthRouter.POST("/lucky/history", api.GetLuckyAppHistory)
		appRouter.POST("/lucky/recentWinners", api.GetRecentLuckyWinnersApp) // 不校验token
		appAuthRouter.POST("/tg/logout", api.TgLogout)
		appAuthRouter.GET("/tg/currentUserInfo", api.GetCurrentTgUserInfo)
		appAuthRouter.GET("/tg/withdrawSummary", api.GetCurrentTgWithdrawSummary)
		appAuthRouter.POST("/tg/avatar", api.UpdateCurrentTgUserAvatar)
		appAuthRouter.POST("/tg/name", api.UpdateCurrentTgUserName)
		appAuthRouter.POST("/tg/bindEmail", api.BindCurrentTgEmail)
		appAuthRouter.POST("/tg/audioOpen", api.SetAudioOpen)
		appAuthRouter.GET("/tenant/serviceLinks", api.GetAppTenantServiceLinks)
		appAuthRouter.GET("/tg/inviteStats", api.GetCurrentTgInviteStats)
		appAuthRouter.GET("/tg/inviteRuleConfig", api.GetCurrentTgInviteRuleConfig)
		appAuthRouter.POST("/tg/rebate/transfer", api.TransferRebateToBalance)
		appAuthRouter.POST("/tg/rebate/list", api.GetCurrentTgUserRebateRecords)
		appAuthRouter.POST("/cashHistory/list", api.GetCurrentTgCashHistory)
		appAuthRouter.POST("/upload", api.AppUpload)
		appAuthRouter.POST("/lucky/grab", api.GrabRedPacketApp)
		appAuthRouter.GET("/countries", api.GetAppCountries)                     // App端获取可用国家列表（IP置顶）
		appAuthRouter.GET("/country/:code/recharge", api.GetCountryRechargeInfo) // App端获取国家充值信息（字段+通道+方式）

		appAuthRouter.GET("/country/:code/withdrawFields", api.GetCountryWithdrawFields)        // App端获取国家提现字段配置
		appAuthRouter.GET("/country/:code/rechargeFields", api.GetCountryRechargeFields)        // App端获取国家充值字段配置
		appAuthRouter.POST("/withdraw", api.AppCreateWithdrawOrder)                             // App端创建提现订单
		appAuthRouter.GET("/withdrawAccount/list", api.GetAppWithdrawAccounts)                  // App端获取当前用户提现账户列表
		appAuthRouter.POST("/withdrawAccount", api.AppAddWithdrawAccount)                       // App端新增提现账户
		appAuthRouter.POST("/withdrawAccount/:id/update", api.AppUpdateWithdrawAccount)         // App端修改提现账户
		appAuthRouter.DELETE("/withdrawAccount/:id", api.AppDelWithdrawAccount)                 // App端删除提现账户
		appAuthRouter.POST("/withdrawAccount/:id/setDefault", api.AppSetDefaultWithdrawAccount) // App端设置默认提现账户
		appAuthRouter.GET("/vip/progress", api.AppGetVipProgress)                               // App端获取当前用户VIP进度
		appAuthRouter.GET("/vip/rewards", api.AppGetClaimableVipRewards)                        // App端查询可领取VIP奖励列表
		appAuthRouter.POST("/vip/rewards/:id/claim", api.AppClaimVipReward)                     // App端领取VIP奖励（id=0领取全部）
		appAuthRouter.GET("/lottery/chances", api.GetLotteryChances)                            // App端查询抽奖次数
		appAuthRouter.POST("/lottery/draw", api.DrawLottery)                                    // App端消耗一次抽奖机会
		appAuthRouter.GET("/lottery/history", api.GetLotteryHistory)                            // App端查询抽奖历史
		appAuthRouter.GET("/prizePool/outRecords", api.GetPrizePoolOutRecordsApp)               // App端查询奖池消耗流水
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
		if c.Request.URL.Path == "/ws" || c.Request.URL.Path == "/api/v1/ws" {
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
		utils.TouchAdminOnlineUser(userId)
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
		utils.TouchTenantOnlineUser(user.TenantId, userId)
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

func appAuthMiddle(singleLogin bool) gin.HandlerFunc {
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
		userId, hostName, tenantId, parseErr := utils.ParseAppToken(hostInfo.AccessSecret, authHeader)
		if parseErr != nil {
			utils.UnauthorizedBack(c, "token is invalid")
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
		if singleLogin {
			key := utils.KeyRdTgOnline + utils.MD5(authHeader)
			data := utils.RD.Get(context.Background(), key)
			if data == nil || data.Err() != nil || data.Val() == "" {
				utils.UnauthorizedBack(c, "token is passed")
				c.Abort()
				return
			}
		}

		utils.TouchTgOnlineUser(tenantId, userId)
		c.Set("userId", userId)
		c.Set("userType", 5)
		c.Set("token", authHeader)
		c.Set("tenantId", tenantId)
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
