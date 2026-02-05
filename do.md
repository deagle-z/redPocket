
Telegram 红包功能迁移到 BaseGoUni 清单
一、数据库层迁移
1.1 创建数据表模型 (POJO)
[x] ✅ core/pojo/lucky_money.go - 红包表模型
字段：sender_id, amount, received, number, lucky, thunder, chat_id, red_list(JSON), sender_name, lose_rate, status
继承 BaseModel (ID, CreatedAt, UpdatedAt)
[x] ✅ core/pojo/lucky_history.go - 红包领取历史表模型
字段：user_id, lucky_id, is_thunder, amount, lose_money, first_name
继承 BaseModel
[ ] core/pojo/tg_user.go - Telegram用户表模型（可选，如果复用SysUser则不需要）- 已复用SysUser
字段：username, first_name, tg_id, balance, status, invite_user
继承 BaseModel
1.2 数据库迁移
[x] ✅ 修改 app/utils/db_utils.go 的 InitTables() 函数
添加 &pojo.LuckyMoney{} 到 AutoMigrate
添加 &pojo.LuckyHistory{} 到 AutoMigrate
如需独立表，添加 &pojo.TgUser{}
二、业务逻辑层 (Repository)
2.1 创建 Repository 文件
[x] ✅ core/repository/lucky_money.go
CreateLuckyMoney() - 创建红包
GetLuckyMoney() - 获取红包详情
UpdateLuckyMoney() - 更新红包状态
GetLuckyMoneyList() - 红包列表查询（分页）
[x] ✅ core/repository/lucky_history.go
CreateLuckyHistory() - 创建领取记录
GetLuckyHistoryByLuckyId() - 获取红包的所有领取记录
CheckUserGrabbed() - 检查用户是否已领取
GetLuckyHistoryList() - 领取历史列表（分页）
2.2 红包分配算法
[x] ✅ core/utils/lucky_money_utils.go
RedEnvelope() - 红包金额分配算法（从PHP迁移）
ValidateLuckyCount() - 验证红包数量
GetLuckyNumMin() - 获取最小红包数量
ParseLuckyNumConfig() - 解析红包数量配置
三、服务层 (Service/Business Logic)
3.1 创建 Service 文件
[x] ✅ core/services/lucky_money_service.go
SendRedPacket() - 发送红包业务逻辑
验证余额
生成红包金额数组
创建红包记录
扣除发送者余额
返回红包信息
GrabRedPacket() - 抢红包业务逻辑
验证用户状态
验证余额（需满足 lose_rate 倍数）
检查是否已领取
获取下一个红包金额
判断是否中雷
处理中雷/未中雷逻辑
更新余额和记录
返回结果
CheckGrabBalance() - 检查抢包余额
GetRedPacketStatus() - 获取红包状态
GetRedPacketDetails() - 获取红包详情（包含领取记录）
四、API 接口层
4.1 创建 API Handler
[x] ✅ core/api/lucky_money_api.go
SendRedPacket() - POST /api/v1/outside/lucky/send
权限：Member及以上
参数：amount, thunder, number(可选)
返回：红包ID和详情
GrabRedPacket() - POST /api/v1/outside/lucky/grab
权限：Member及以上
参数：lucky_id
返回：领取结果（金额、是否中雷）
GetRedPacketList() - POST /api/v1/outside/lucky/list
权限：Member及以上
参数：分页、状态筛选
返回：红包列表
GetRedPacketDetail() - GET /api/v1/outside/lucky/:id
权限：Member及以上
返回：红包详情和领取记录
4.2 管理员 API
[x] ✅ core/api/lucky_money_admin_api.go
GetLuckyMoneyList() - POST /api/v1/admin/lucky/list
权限：Admin
返回：所有红包列表（管理用）
GetLuckyHistoryList() - POST /api/v1/admin/lucky/history
权限：Admin/Manager
返回：领取历史记录
4.3 注册路由
[x] ✅ 修改 core/common/web_routes.go
添加红包相关路由到对应权限组
添加中间件（日志记录等）
五、Telegram Bot 集成
5.1 添加依赖
[x] ✅ 已集成 github.com/mymmrac/telego 库
修改 go.mod（运行 go get 后自动更新）
5.2 创建 Telegram Bot 服务
[x] ✅ app/services/telegram_bot_service.go
[x] ✅ InitTelegramBot() - 初始化Bot（完整实现）
[x] ✅ HandleRedPacketCommand() - 处理发包命令
[x] ✅ 解析命令格式：发10-1、10-1-3
[x] ✅ 调用 lucky_money_service.SendRedPacket()
[x] ✅ 发送Telegram消息和按钮
[x] ✅ HandleGrabCallback() - 处理抢包回调
[x] ✅ 调用 lucky_money_service.GrabRedPacket()
[x] ✅ 更新Telegram消息
[x] ✅ 显示结果
[x] ✅ HandleBalanceCommand() - 处理余额查询
[x] ✅ HandleRegisterCommand() - 处理注册命令
[x] ✅ HandleHelpCommand() - 处理帮助命令
[x] ✅ 实现完整的消息处理和更新逻辑
5.3 Bot 配置
[x] ✅ 修改 core/base/core_config.go
添加 Telegram 配置结构
BotToken
WebhookURL (可选)
SafeMode
[x] ✅ 修改 core.yaml 配置示例
添加 telegram 配置段
5.4 启动 Bot 服务
[x] ✅ 修改 main.go
初始化 Telegram Bot
启动 Bot 服务（goroutine）
5.5 授权群组管理
[x] ✅ 创建 core/pojo/auth_group.go
[x] ✅ 创建 core/api/auth_group_api.go
[x] ✅ 注册路由
六、配置管理
6.1 添加配置项
[x] ✅ 使用现有的 sys_config.go
lose_rate - 中雷倍数（默认1.8）
lucky_num - 红包数量配置（如 "3|9"）
default_balance - 新用户默认余额
valid_time - 红包有效期（秒）
6.2 配置服务
[x] ✅ core/services/lucky_money_service.go
GetLoseRate() - 获取中雷倍数
GetLuckyNumConfig() - 获取红包数量配置
GetDefaultBalance() - 获取默认余额
七、用户系统集成
7.1 用户余额集成
[x] ✅ 复用 SysUser 的 Amount 字段作为余额
[ ] 或创建 TgUser 表（如果Telegram用户独立管理）- 已复用SysUser
7.2 余额操作
[x] ✅ 复用现有的 AwardUser() 方法
[x] ✅ 在 Service 中直接操作余额和记录 CashHistory
八、工具函数
8.1 字符串处理
[x] ✅ core/utils/lucky_money_utils.go
FormatName() - 格式化用户名（截断）
从PHP迁移相关函数
8.2 金额处理
[x] ✅ 复用现有的 core/utils/money_utils.go
[x] ✅ 确保精度处理正确（使用 Money 类型）
九、权限控制
9.1 菜单权限
[x] ✅ 在 cs.yaml 中添加菜单配置
添加红包管理菜单项
设置权限码
9.2 角色权限
[x] ✅ API 已配置权限：Member及以上可发送和抢红包
[x] ✅ Admin 可以查看所有红包记录
十、前端适配
10.1 管理界面
[x] ✅ pure-admin-thin/src/views/lucky/lucky_money/ - 红包列表页面
[x] ✅ pure-admin-thin/src/views/lucky/lucky_money/utils/hook.tsx - 红包列表业务逻辑（包含详情弹窗）
[x] ✅ pure-admin-thin/src/views/lucky/lucky_history/ - 领取历史页面
[x] ✅ pure-admin-thin/src/views/lucky/auth_group/ - 授权群组管理页面
[x] ✅ pure-admin-thin/src/views/lucky/auth_group/form.vue - 授权群组表单
10.2 API 调用
[x] ✅ pure-admin-thin/src/api/luckyMoney.ts - 红包相关API
[x] ✅ pure-admin-thin/src/api/authGroup.ts - 授权群组API