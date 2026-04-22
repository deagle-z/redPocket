# BaseGoUni 项目测试方案与测试用例

更新时间：2026-04-22  
输出依据：基于当前仓库代码扫描结果整理，包含 `core/`、`tenant/`、`RedPocketH5/`、`pure-admin-thin/`、`RedTenantAdmin/`。部分页面行为或业务口径若代码未完全闭环，文中会标注“基于代码推断”。

## 项目测试分析摘要

### 1. 系统组成

- Go 后端主路由集中在 `core/common/web_routes.go`
- C 端 H5 页面集中在 `RedPocketH5/src/pages`
- 超级管理员后台集中在 `pure-admin-thin/src/views`
- 租户后台集中在 `RedTenantAdmin/src/views`
- 核心业务实体集中在 `core/pojo`
- 核心业务逻辑集中在 `core/repository`、`core/services`

### 2. 角色划分

| 角色 | 鉴权入口 | 主要路由前缀 | 说明 |
| --- | --- | --- | --- |
| 超级管理员 | `/api/v1/user/login` | `/api/v1/admin` | 平台级管理、配置、奖池、用户、订单、活动 |
| 租户用户 | `/api/v1/tenant/login` | `/api/v1/tenant` | 租户内红包、用户、订单、群组、提现 |
| App 用户 | `/api/v1/app/tg/*` | `/api/v1/app` | H5 注册、登录、充值、提现、红包、抽奖、VIP |
| 外部/公共接口 | 无或特殊鉴权 | `/api/v1/outside`、部分 `/api/v1/app`、`/api/v1/pay` | 红包列表、Banner、配置、支付回调 |

### 3. 核心业务链路

1. 注册/登录
   App 用户支持 Telegram 登录、邮箱登录、手机号登录、邮箱注册、手机号注册、找回密码。
2. 充值
   创建充值订单 -> 三方支付/手动回调 -> 充值到账 -> 首充固定赠送 `recharge_gift_amount` -> 活动赠送 `activity_type` -> 邀请首充奖励 -> VIP 升级检查。
3. 红包玩法
   发红包 -> 生成红包明细 -> 用户/机器人抢包 -> 中雷或奇偶结算 -> 佣金/奖池/邀请返佣 -> 红包完成或过期退回。
4. 提现
   维护提现账户 -> 提现申请 -> 冻结/扣减余额 -> 后台审核/打款 -> 失败/取消/退回返还。
5. 抽奖/VIP
   根据流水换算抽奖次数 -> 抽奖 -> 奖金入账 -> VIP 条件达成 -> 待领取奖励 -> 用户领取。
6. 后台核对
   管理后台查看充值、提现、红包、账变、平台利润、用户、奖池、配置，租户后台查看租户范围数据。

### 4. 核心接口与页面

#### App 核心接口

- 登录注册
  - `POST /api/v1/app/tg/login`
  - `POST /api/v1/app/tg/loginByEmail`
  - `POST /api/v1/app/tg/phoneLogin`
  - `POST /api/v1/app/tg/registerByEmail`
  - `POST /api/v1/app/tg/registerByPhone`
  - `POST /api/v1/app/tg/forgotPasswordByEmail`
  - `POST /api/v1/app/tg/forgotPasswordByPhone`
- 充值
  - `POST /api/v1/app/rechargeOrder`
  - `GET /api/v1/app/recharge/isFirst`
  - `GET /api/v1/app/config/:key`
  - `GET /api/v1/app/countries`
  - `GET /api/v1/app/country/:code/recharge`
  - `GET /api/v1/app/country/:code/rechargeFields`
- 红包
  - `POST /api/v1/app/lucky/send`
  - `POST /api/v1/app/lucky/grab`
  - `POST /api/v1/app/lucky/list`
  - `POST /api/v1/app/lucky/detail`
  - `POST /api/v1/app/lucky/history`
  - `POST /api/v1/app/lucky/recentWinners`
- 账户
  - `GET /api/v1/app/tg/currentUserInfo`
  - `POST /api/v1/app/tg/rebate/transfer`
  - `POST /api/v1/app/cashHistory/list`
  - `GET /api/v1/app/withdrawAccount/list`
  - `POST /api/v1/app/withdrawAccount`
  - `POST /api/v1/app/withdrawAccount/:id/update`
  - `DELETE /api/v1/app/withdrawAccount/:id`
  - `POST /api/v1/app/withdrawAccount/:id/setDefault`
- VIP/抽奖
  - `GET /api/v1/app/vip/progress`
  - `GET /api/v1/app/vip/rewards`
  - `POST /api/v1/app/vip/rewards/:id/claim`
  - `GET /api/v1/app/lottery/chances`
  - `POST /api/v1/app/lottery/draw`
  - `GET /api/v1/app/lottery/history`

#### H5 核心页面

- 登录注册找回
  - `RedPocketH5/src/pages/login/index.vue`
  - `RedPocketH5/src/pages/register/index.vue`
  - `RedPocketH5/src/pages/resetpwd/index.vue`
- 充值提现
  - `RedPocketH5/src/pages/recharge/index.vue`
  - `RedPocketH5/src/pages/withdraw/index.vue`
  - `RedPocketH5/src/pages/withdrawAccount/index.vue`
  - `RedPocketH5/src/pages/transform/index.vue`
- 红包与钱包
  - `RedPocketH5/src/pages/packetList/index.vue`
  - `RedPocketH5/src/pages/luckyDetail/index.vue`
  - `RedPocketH5/src/pages/sendPacket/index.vue`
  - `RedPocketH5/src/pages/history/index.vue`
  - `RedPocketH5/src/pages/wallet/index.vue`
- 其他
  - `RedPocketH5/src/pages/profile/index.vue`
  - `RedPocketH5/src/pages/team/index.vue`
  - `RedPocketH5/src/pages/invite/index.vue`
  - `RedPocketH5/src/pages/prize/index.vue`

#### 后台页面

- 超级管理员后台
  - 用户：`pure-admin-thin/src/views/system/tg_user/index.vue`
  - 充值订单：`pure-admin-thin/src/views/system/recharge_order/index.vue`
  - 提现订单：`pure-admin-thin/src/views/system/withdraw_order_br/index.vue`
  - 系统配置：`pure-admin-thin/src/views/system/sysConfig/index.vue`
  - 国家配置：`pure-admin-thin/src/views/system/country/index.vue`
  - 支付通道/支付方式：`pure-admin-thin/src/views/system/sysPayChannel/index.vue`、`sysPayMethod/index.vue`
  - Banner：`pure-admin-thin/src/views/system/banner/index.vue`
  - VIP：`pure-admin-thin/src/views/system/sysVipLevel/index.vue`
  - 奖池配置：`pure-admin-thin/src/views/system/prizeConfig/index.vue`
  - 红包、红包历史、账变：`pure-admin-thin/src/views/lucky/*`
- 租户后台
  - 红包：`RedTenantAdmin/src/views/lucky/lucky_money/index.vue`
  - 红包历史：`RedTenantAdmin/src/views/lucky/lucky_history/index.vue`
  - 充值订单：`RedTenantAdmin/src/views/lucky/recharge_order/index.vue`
  - 提现订单：`RedTenantAdmin/src/views/lucky/withdraw_order_br/index.vue`
  - 用户/返佣/授权群：`RedTenantAdmin/src/views/lucky/tg_user`、`rebate_record`、`auth_group`

### 5. 关键实体与状态

| 实体 | 路径 | 关键状态/字段 |
| --- | --- | --- |
| `TgUser` | `core/pojo/tg_user.go` | `status` 1正常 0禁用 -1删除；`balance`、`gift_amount`、`gift_total`、`recharge_amount`、`rebate_amount` |
| `RechargeOrder` | `core/pojo/recharge_order.go` | `status` 默认 0 待支付，成功流程写 1；`activity_type` 0无 1首充 2今日首充 |
| `WithdrawOrderBr` | `core/pojo/withdraw_order_br.go` | `status` 0待审核 1待打款 2打款中 3成功 4失败 5取消 6退回 |
| `LuckyMoney` | `core/pojo/lucky_money.go` | `game_mode` 0雷号 1奇偶；`status` 1进行中 2已完成 |
| `LuckyHistory` | `core/pojo/lucky_history.go` | `is_thunder`、`guess`、`grab_type`、`amount`、`lose_money` |
| `CashHistory` | `core/pojo/cash_history.go` | 唯一键 `user_id + award_uni`；类型覆盖充值、红包、提现、抽奖等 |
| `SysConfig` | `core/pojo/sys_config.go` | 配置项统一入口，多个业务通过 key 驱动 |
| `SysTenantPrizePool` | `core/pojo/sys_tenant_prize_pool.go` | `pool_code`、`balance` |
| `SysTenantPrizePoolConfig` | `core/pojo/sys_tenant_prize_pool_config.go` | `probabilities`、`amounts`、`peer_amount`、`count` |

### 6. 异步与定时任务

| 类型 | 路径 | 说明 |
| --- | --- | --- |
| 定时任务 | `core/common/common_scheduler.go` | 每 1 分钟扫描过期红包 |
| Asynq 任务 | `core/services/lucky_expire_task.go` | 红包过期退回、机器人抢包、首充分段赠送 worker 注册 |
| 延迟任务 | `core/services/recharge_first_gift_task.go` | 首充活动第 2、3 天赠送 |
| 机器人任务 | `core/services/lucky_bot_grab_task.go` | 随机秒数自动抢包，配置 `random_grab_second` |

## 测试范围与策略

### 测试范围

- App 用户注册、登录、找回密码、登出、当前用户资料
- 充值下单、支付回调、首充固定赠送、首充活动赠送、今日首充赠送、邀请首充奖励
- 红包发包、抢包、雷号玩法、奇偶玩法、机器人参与、过期退回、邀请返佣
- 钱包、账变、佣金转余额、团队/邀请数据
- 提现账户管理、提现订单、后台审核/退回
- 抽奖次数计算、抽奖发奖、奖池配置、VIP 进度与奖励领取
- 系统配置、国家配置、支付配置、Banner、用户管理、租户管理、租户后台核对
- WebSocket 广播事件与异步任务执行结果

### 不在本次范围

- 第三方支付通道真实生产环境清算链路
- 第三方短信/邮件平台真实到达率
- 复杂压测容量评估报告
- CDN、R2、部署脚本、容器编排层面专项测试
- 未在当前仓库发现完整后端实现的能力，示例：H5 调用 `/api/v1/app/withdraw` 的实际后端路由

### 测试策略

- 接口优先：先验证 API 状态流、金额流、幂等，再验证页面展示
- 金额闭环优先：对充值、红包、提现、抽奖、返佣、VIP 奖励逐步核对 `tg_user`、`cash_history`、`platform_profit_ledger`
- 配置驱动优先：对 `sys_config` 类逻辑采用“默认值/合法值/非法值/边界值/变更后生效”全覆盖
- 异步任务单列：对 Asynq 与 Cron 任务进行单独设计，避免仅凭页面观察
- 回归按主链路执行：注册 -> 登录 -> 充值 -> 首充活动 -> 发/抢红包 -> 提现 -> 后台核对

### 测试环境建议

| 环境项 | 建议 |
| --- | --- |
| 应用实例 | 1 套后端 + 1 套 Redis + 1 套 MySQL |
| 数据库 | 独立测试库，支持多 `table_prefix` 租户数据 |
| Redis | 需启用，用于登录态、验证码、抽奖池、异步任务、单点/限流 |
| Asynq Worker | 必须启动，否则无法验证红包过期、机器人抢包、首充分天赠送 |
| Cron Scheduler | 必须开启 `RunScheduler=true` 才能验证红包过期扫描 |
| 前端 | H5、超级管理员后台、租户后台均需可访问 |
| 第三方支付 | 建议保留 MANUAL/dev 回调模式与模拟回调 |
| WebSocket | 需可连接 `/ws` 或 `/api/v1/ws` |

### 测试数据准备方案

| 数据项 | 建议样本 |
| --- | --- |
| 租户 | 至少 2 个租户，1 个默认租户，1 个非默认租户 |
| 来源渠道 | 至少 2 个 `sys_source_channel`，用于核对来源渠道透传 |
| 国家 | `BR`、`MX`、`ID`，并配置不同充值/提现字段 |
| 用户 | 普通用户 5 个、机器人用户 3 个、邀请上下级用户 2 组、禁用用户 1 个 |
| 充值订单 | 成功、待支付、重复回调、活动类型 0/1/2 各至少 1 单 |
| 提现订单 | 状态 0/1/2/3/4/5/6 各至少 1 单 |
| 红包 | 雷号红包、奇偶红包、已完成红包、进行中红包、即将过期红包各至少 1 个 |
| 配置项 | 需提前准备 `first_recharge_gift_config`、`today_first_recharge_gift`、`recharge_gift_amount`、`invite_first_recharge_reward`、`lucky_*`、`random_grab_second`、`send_min_max` |
| 奖池 | `pool_code=lucky` 的 `sys_tenant_prize_pool` 与一套启用中的 `sys_tenant_prize_pool_config` |

### 测试角色与权限矩阵

| 角色 | 登录方式 | 可访问能力 | 重点验证 |
| --- | --- | --- | --- |
| App 未登录用户 | 无 | 公共红包列表、红包详情、Banner、配置、奖池余额 | 公共接口、无 token 场景 |
| App 登录用户 | Telegram/邮箱/手机号 | 充值、红包、提现账户、账变、邀请、抽奖、VIP | 资金链路、页面展示 |
| 禁用 App 用户 | 登录后状态改为 0 | 应被拒绝访问受保护接口 | 鉴权与状态校验 |
| 超级管理员 | `/api/v1/user/login` | 所有 `/api/v1/admin/*` | 配置、审核、人工回调、奖池 |
| 租户用户 | `/api/v1/tenant/login` | 所有 `/api/v1/tenant/*` | 数据隔离、租户可见性 |

## 详细测试用例

### 模块一：用户注册、登录、找回密码

#### 测试点

- 手机号注册不校验短信码，邮箱注册校验邮箱验证码
- 手机号登录当前只按 `phone` 查询，不按 `country` 过滤
- 登录后 token、当前用户信息、登出、单点/登录态过期
- 禁用用户登录拦截
- 忘记密码验证码校验、限流与密码更新

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| AUTH-001 | 用户鉴权 | 手机号注册 | 正常手机号注册成功 | 测试手机号未注册 | 1. 调用 `POST /api/v1/app/tg/registerByPhone` 2. 传 `phone/country/password` | 返回成功；`tg_user.phone` 创建成功；密码为哈希值；`invite_code`、`uid` 已生成 | P0 | 功能 |
| AUTH-002 | 用户鉴权 | 手机号注册 | 手机号已注册拦截 | 已存在相同手机号用户 | 重复调用注册接口 | 返回“手机号已注册”；数据库不新增用户 | P0 | 异常 |
| AUTH-003 | 用户鉴权 | 邮箱注册 | 邮箱验证码错误拦截 | Redis 中验证码与入参不一致 | 调用 `POST /api/v1/app/tg/registerByEmail` | 返回验证码错误；不新增用户 | P0 | 异常 |
| AUTH-004 | 用户鉴权 | 手机号登录 | 登录不依赖国家字段 | 已存在手机号用户且密码正确 | 1. 调用 `POST /api/v1/app/tg/phoneLogin` 2. 只传 `phone/password` | 登录成功；返回 `accessToken`；当前用户信息可查询 | P0 | 接口 |
| AUTH-005 | 用户鉴权 | 手机号登录 | 手机号重复用户拦截 | 制造同手机号多记录 | 调用手机号登录 | 返回“手机号重复，请联系管理员” | P1 | 异常 |
| AUTH-006 | 用户鉴权 | 用户状态 | 禁用用户无法登录 | 用户 `status=0` | 分别走邮箱登录、手机号登录、当前用户信息接口 | 登录失败或 token 鉴权失败，返回禁用信息 | P0 | 安全 |
| AUTH-007 | 用户鉴权 | 忘记密码 | 手机号找回密码成功 | Redis 有正确短信验证码 | 调用 `POST /api/v1/app/tg/forgotPasswordByPhone` | 密码更新成功；旧密码失效；新密码可登录 | P0 | 功能 |
| AUTH-008 | 用户鉴权 | 验证码发送 | 短信/邮箱发送限流 | 同一 IP 1 分钟内已发送过 | 重复调用 `sendSMSCode/sendEmailCode` | 返回限流提示；验证码缓存不重复覆盖或受控 | P1 | 异常 |

### 模块二：充值、活动赠送、邀请首充、支付回调

#### 测试点

- 充值订单创建参数校验、国家字段校验、自定义字段校验
- `RechargeOrder.status`、`credit_amount`、`bonus_amount` 更新
- 首充固定赠送 `recharge_gift_amount`
- 活动类型 `activity_type=1/2`
- `first_recharge_gift_config` 分 3 天赠送
- `today_first_recharge_gift` 一次性赠送
- 邀请首充奖励 `invite_first_recharge_reward`
- 重复回调幂等、手动回调、平台利润流水

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| RCG-001 | 充值 | 创建订单 | 国家充值字段必填校验 | 国家已配置 `rechargeFields` 且包含必填项 | 调用 `POST /api/v1/app/rechargeOrder`，缺少必填字段 | 返回字段提示；订单不创建 | P0 | 功能 |
| RCG-002 | 充值 | 创建订单 | 正常创建充值订单 | 用户已登录，有可用通道和方式 | 提交 `amount/channel/payMethod/countryCode/extraFields/activityType` | 返回 `orderNo` 和 `payUrl` 或 `devCallback`；数据库生成待支付订单 | P0 | 接口 |
| RCG-003 | 充值 | 支付回调 | 首次充值固定赠送 | `recharge_gift_amount` 已配置，用户历史充值额为 0 | 将订单回调为成功 | `tg_user.balance` 增加充值金额+固定赠送；`cash_history` 有 `recharge_*` 和 `recharge_gift_*` 两条记录 | P0 | 功能 |
| RCG-004 | 充值 | 首充活动 | `activity_type=1` 首充活动首日赠送 | `first_recharge_gift_config=12(30|30|40)`，用户首充，订单 `activityType=1` | 支付成功回调 | 首日额外赠送充值金额的 3.6%；写 `cash_history.award_uni=first_recharge_gift_<orderNo>_1`；`platform_profit_ledger` 有对应支出 | P0 | 功能 |
| RCG-005 | 充值 | 首充活动 | 分天赠送第 2、3 段延迟执行 | 已完成 RCg-004，Asynq worker 正常 | 人工等待或直接触发任务 `recharge:first_gift_installment` | 第 2 天、第 3 天分别到账 3.6%、4.8%；三段合计等于 12%；不重复发放 | P0 | 异步 |
| RCG-006 | 充值 | 今日首充活动 | `activity_type=2` 一次性赠送 | `today_first_recharge_gift` 已配置 | 创建今日首充订单并回调成功 | 一次性赠送；不创建分天任务；`award_uni=today_first_recharge_gift_<orderNo>` | P0 | 功能 |
| RCG-007 | 充值 | 支付幂等 | 重复支付回调不重复入账 | 已存在成功订单 | 重复调用支付成功回调或管理后台手动回调 | 用户余额、`cash_history`、`platform_profit_ledger` 不重复增加 | P0 | 并发 |
| RCG-008 | 充值 | 邀请奖励 | 首充邀请奖励只发一次 | 用户有 `parent_id`，配置 `invite_first_recharge_reward` | 完成首充后再次进行二次充值 | 仅首次成功充值触发邀请奖励；再次充值不再发 | P1 | 功能 |
| RCG-009 | 充值 | 配置容错 | `first_recharge_gift_config` 配置非法不影响充值主流程 | 将配置改为空、非法格式、比例和非 10/100 | 完成首充回调 | 充值到账正常；活动赠送跳过；日志记录解析失败 | P0 | 异常 |
| RCG-010 | 充值 | 精度 | 小数金额首充分段合计精确 | 配置 `12(30|30|40)`，充值金额含小数如 99.99 | 完成回调并执行 3 段赠送 | 前两段按规则计算，最后一段兜底；三段总和严格等于总赠送额 | P0 | 功能 |

### 模块三：提现账户与提现订单

#### 测试点

- 提现账户 CRUD、默认账户切换、国家字段校验
- 提现申请扣款、状态流转、失败/取消/退回返还
- 余额不足拦截
- 后台修改订单状态触发扣款/退款逻辑
- 账变类型 `CashHistoryTypeWithdrawApply`、`CashHistoryTypeWithdrawRefund`

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WTD-001 | 提现 | 提现账户 | 新增提现账户成功 | 已登录用户，国家已配置 `withdrawFields` | 调用 `POST /api/v1/app/withdrawAccount` | 账户保存成功；列表可见；字段 JSON 正确落库 | P0 | 功能 |
| WTD-002 | 提现 | 提现账户 | 设置默认提现账户 | 至少存在 2 个账户 | 调用 `POST /api/v1/app/withdrawAccount/:id/setDefault` | 仅一个账户 `isDefault=1` | P1 | 功能 |
| WTD-003 | 提现 | 提现申请 | 前端提现申请链路验证 | H5 `withdraw/index.vue` 可用 | 在 H5 发起提现 | 若后端 `/api/v1/app/withdraw` 未实现，应记录为阻塞风险；若实现则生成订单 | P0 | 回归 |
| WTD-004 | 提现 | 余额冻结 | 创建待审核提现订单自动扣减余额 | 用户余额充足 | 后台创建状态为 0 的 `withdraw_order_br` | 用户余额减少；`cash_history` 写入 `withdraw_apply_<orderNo>` | P0 | 功能 |
| WTD-005 | 提现 | 退款 | 提现失败后返还余额 | 已有状态 0/1/2/3 的订单 | 将状态改为 4/5/6 | 用户余额返还；写 `withdraw_refund_<orderNo>`；不重复返还 | P0 | 功能 |
| WTD-006 | 提现 | 余额校验 | 用户余额不足拦截 | 用户余额小于提现金额 | 创建提现订单 | 返回“用户余额不足”；订单不创建或事务回滚 | P0 | 异常 |
| WTD-007 | 提现 | 状态回放 | 已退款订单再次改失败不重复退款 | 已执行一次退款 | 再次保存失败/退回状态 | 余额与账变不重复增加 | P0 | 并发 |
| WTD-008 | 提现 | 核对 | 后台提现订单与账变一致性 | 准备多状态订单 | 核对 `withdraw_order_br`、`tg_user.balance`、`cash_history` | 金额闭环一致，无漏账、重账 | P0 | 数据一致性 |

### 模块四：红包玩法、机器人、过期退回、WebSocket

#### 测试点

- 发红包金额、数量、玩法模式、雷号/奇偶参数校验
- 抢包余额校验、并发锁 `lucky_grab:<luckyId>`
- 雷号模式与奇偶模式金额结算
- 抽成、奖池注入、邀请返佣
- 机器人抢包配置 `random_grab_second`
- 红包过期退回定时扫描与异步处理
- WebSocket 事件 `lucky_sent`、`lucky_grabbed`、`lucky_finished`

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| LKY-001 | 红包 | 发包 | 雷号模式发包成功 | 用户已登录且满足发包条件 | 调用 `POST /api/v1/app/lucky/send`，`gameMode=0`、带 `thunder` | 生成 `lucky_money` 与 `lucky_money_item`；广播 `lucky_sent` | P0 | 功能 |
| LKY-002 | 红包 | 发包 | 奇偶模式发包成功 | 用户已登录 | 调用发包接口，`gameMode=1` 且不传雷号 | 红包创建成功；详情页显示奇偶模式 | P0 | 功能 |
| LKY-003 | 红包 | 抢包 | 并发抢包幂等 | 同一红包，多个用户并发抢 | 并发调用 `POST /api/v1/app/lucky/grab` | 每个包仅被一个用户抢到；余额与领取记录无重复 | P0 | 并发 |
| LKY-004 | 红包 | 雷号结算 | 抢包中雷损失与发包收益正确 | 准备雷号红包，命中雷号 | 抢包并核对发包方与抢包方余额 | 抢包方产生 `GrabRedPacketThunder`；发包方产生 `RedPacketThunderIncome`；抽成和平台利润正确 | P0 | 功能 |
| LKY-005 | 红包 | 奇偶结算 | 奇偶猜对/猜错结算正确 | 准备奇偶红包 | 分别用 odd/even 参与 | 猜对按中奖逻辑，猜错按损失逻辑；`guess` 字段正确 | P0 | 功能 |
| LKY-006 | 红包 | 邀请返佣 | 上级返佣只发一次且金额正确 | 抢包/中雷用户存在上级，返佣配置生效 | 完成红包结算 | `tg_user_rebate_record`、`rebate_amount`、`cash_history` 一致；同一场景不重复返佣 | P1 | 功能 |
| LKY-007 | 红包 | 机器人抢包 | 机器人随机抢包任务正常 | 有机器人用户，`random_grab_second` 有效 | 发送红包并等待异步任务 | 机器人在设定时间窗口抢包；无余额的机器人会补充余额或跳过 | P1 | 异步 |
| LKY-008 | 红包 | 过期退回 | 红包过期退回成功 | 准备未抢完且过期红包 | 等待 Cron 或直接执行过期逻辑 | `lucky_money.status=2`；发送者余额返还未抢金额；写 `lucky_expire_refund_<id>` | P0 | 异步 |
| LKY-009 | 红包 | WebSocket | 页面实时更新广播正确 | H5 列表页和详情页在线 | 发包、抢包、完成红包 | 前端收到 `lucky_sent/lucky_grabbed/lucky_finished`，列表和详情实时刷新 | P1 | 兼容 |

### 模块五：钱包、账变、返佣、团队

#### 测试点

- 当前用户信息、余额、赠金、返水显示
- 账变筛选、账变类型与金额方向正确
- 返水转余额
- 邀请统计与团队数据
- 流水排行榜 `historyUserFlow`

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WAL-001 | 钱包 | 当前用户 | 当前用户信息与数据库一致 | 用户已登录 | 调用 `GET /api/v1/app/tg/currentUserInfo` | 返回 `balance/gift_amount/rebate_amount/vip_level/audio_open` 与库一致 | P0 | 接口 |
| WAL-002 | 钱包 | 账变列表 | 账变分页与筛选正确 | 账户存在多类账变 | 调用 `POST /api/v1/app/cashHistory/list` | 分页、排序、`cashMark` 过滤正确 | P1 | 功能 |
| WAL-003 | 钱包 | 返佣转余额 | 返水余额转主余额成功 | 用户有 `rebate_amount>0` | 调用 `POST /api/v1/app/tg/rebate/transfer` | 主余额增加、返水减少、写 `CashHistoryTypeRebateTransfer` | P0 | 功能 |
| WAL-004 | 钱包 | 返佣转余额 | 返水为 0 时拦截 | `rebate_amount=0` | 调用转余额接口 | 返回失败，不写账变 | P1 | 异常 |
| WAL-005 | 团队 | 邀请统计 | 邀请数据与用户关系一致 | 有上下级和充值数据 | 调用 `GET /api/v1/app/tg/inviteStats` | 邀请人数、充值人数、佣金统计正确 | P1 | 数据一致性 |
| WAL-006 | 团队 | 邀请规则 | 邀请规则配置展示正确 | 配置 `lucky_send_commission` 等存在 | 调用 `GET /api/v1/app/tg/inviteRuleConfig` | 返回值与 `sys_config` 一致 | P1 | 配置 |
| WAL-007 | 流水排行 | historyUserFlow | 返回头像、脱敏名、金额倒序前 20 | 存在多名参与用户 | 调用 `POST /api/v1/admin/lucky/historyUserFlow` | 返回 20 条内，按 `flowAmount desc`；名称已脱敏；带头像 | P1 | 接口 |

### 模块六：抽奖、奖池、VIP

#### 测试点

- 抽奖次数 = `floor(totalFlow / peerAmount) - usedCount`
- Redis 抽奖池生成与弹出
- 抽奖中奖金额入账
- 奖池概率配置、奖池余额修改
- VIP 升级条件、待领取奖励、奖励发放

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| LTY-001 | 抽奖 | 次数计算 | 抽奖次数计算正确 | `sys_tenant_prize_pool_config.peer_amount` 已配置 | 调用 `GET /api/v1/app/lottery/chances` | `earnedCount=floor(totalFlow/peerAmount)`，`availableCount=earned-used` | P0 | 功能 |
| LTY-002 | 抽奖 | 抽奖执行 | 正常抽奖中奖/未中奖 | 用户有可用次数 | 调用 `POST /api/v1/app/lottery/draw` | 生成 `user_lottery_record`；中奖时增加余额并写 `lottery_award_<id>` | P0 | 功能 |
| LTY-003 | 抽奖 | 并发抽奖 | 重复点击只消耗 1 次 | 有 1 次可用机会 | 并发调用抽奖接口 | 因 `lottery_draw:<userId>` 锁保护，仅成功一次 | P0 | 并发 |
| LTY-004 | 奖池 | 概率配置 | 概率列表与金额列表数量不一致拦截 | 后台打开奖池配置页 | 保存不等长的 `probabilities`、`amounts` | 前端提示失败；后端不落库 | P1 | 功能 |
| LTY-005 | 奖池 | lucky 奖池余额 | 后台单独修改 `pool_code=lucky` 余额 | 后台进入 `prizeConfig/index.vue` | 修改余额并保存 | `sys_tenant_prize_pool.balance` 更新成功 | P1 | 功能 |
| LTY-006 | VIP | 升级判定 | 充值/下注满足新等级自动升级 | 用户接近阈值 | 充值成功后触发 `CheckAndUpgradeVipLevel` | `tg_user.vip_level` 升级；写待领取奖励记录 | P1 | 功能 |
| LTY-007 | VIP | 奖励领取 | 领取指定 VIP 奖励成功 | 存在 `pending` 奖励 | 调用 `POST /api/v1/app/vip/rewards/:id/claim` | 状态改为已领取；余额增加；写 `vip_reward_<id>` 流水 | P1 | 功能 |
| LTY-008 | VIP | 重复领取 | 已领取奖励不可重复领取 | 奖励状态已 done | 再次调用领取 | 返回失败或无可领取奖励；余额不重复增加 | P0 | 并发 |

### 模块七：后台配置、国家、支付、Banner、租户与权限

#### 测试点

- `sys_config` CRUD 与关键配置可读性
- `sys_country` 国家、充值字段、提现字段
- `sys_pay_channel`、`sys_pay_method`、`sys_pay_channel_method`
- `sys_banner` 多位置、多语言、多平台
- 租户、租户用户、授权群、来源渠道、菜单权限
- 租户后台只可见本租户数据

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| ADM-001 | 后台配置 | sys_config | 新增并读取配置项成功 | 超级管理员已登录 | 在后台新增 `first_recharge_gift_config`，再用 App 接口读取 | `GET /api/v1/app/config/first_recharge_gift_config` 返回新值 | P0 | 功能 |
| ADM-002 | 后台配置 | 国家字段 | 国家充值字段配置生效 | 已配置国家 `recharge_fields` | H5 切换国家进入充值页 | 页面动态展示字段；后端创建订单按字段校验 | P0 | 功能 |
| ADM-003 | 后台配置 | 支付通道 | 通道与支付方式绑定生效 | 配置通道、方式及绑定关系 | 调用 `GET /api/v1/app/country/:code/recharge` | 页面仅展示已绑定方式 | P1 | 功能 |
| ADM-004 | 后台配置 | Banner | 不同 position 返回正确分组 | 已配置首页、弹窗、活动 Banner | 调用 `POST /api/v1/app/banners` | 返回 `home/popup/activity` 分组数据正确 | P1 | 功能 |
| ADM-005 | 后台配置 | 用户管理 | 封禁用户后 App 鉴权失效 | 普通用户已登录 | 后台调用 `/api/v1/admin/tgUser/status` 设置禁用 | 用户后续访问受保护接口被拒绝 | P0 | 安全 |
| ADM-006 | 租户后台 | 数据隔离 | 租户后台不可见其他租户数据 | 准备 2 个租户数据 | 分别登录租户后台查询用户/订单/红包 | 仅返回本租户数据 | P0 | 安全 |
| ADM-007 | 权限 | 管理路由 | 超级管理员接口无 token 拦截 | 无 token 或错误 token | 调用 `/api/v1/admin/*` | 返回未授权；仅明确公开的接口例外 | P0 | 安全 |
| ADM-008 | 权限 | 特殊公开接口 | 公开接口权限符合预期 | 无 token | 调用 `/api/v1/admin/lucky/historyUserFlow` | 当前代码允许匿名访问，需确认是否符合预期 | P0 | 风险 |

## 专项测试

### 一、首充、今日首充、分天赠送专项测试

#### 重点配置

- `recharge_gift_amount`
- `first_recharge_gift_config`
- `today_first_recharge_gift`
- `invite_first_recharge_reward`

#### 专项场景

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| FRG-001 | 首充专项 | 配置兼容 | `12(3|3|4)` 与 `12(30|30|40)` 均可解析 | 分别配置 2 种值 | 完成首充回调并观察 H5 充值页 | 页面均显示 `+12%`；后端均正常发放 3 段奖励 | P0 | 配置 |
| FRG-002 | 首充专项 | 页面展示 | 充值页首充活动展示完整 | H5 登录后进入充值页 | 查看 `RedPocketH5/src/pages/recharge/index.vue` 页面效果 | 显示总赠送百分比、分几天到账、第一天到账多少 | P0 | 回归 |
| FRG-003 | 首充专项 | 首日发放 | 首日奖励在支付成功事务内发放 | 活动类型为 1 | 完成回调 | 当天立即到账；`award_uni` 后缀 `_1` | P0 | 功能 |
| FRG-004 | 首充专项 | 二三日发放 | 第 2/3 天任务重复执行不重复发奖 | 已有延迟任务 | 重复投递相同 payload | 因 `cash_history.award_uni` 幂等，不重复赠送 | P0 | 并发 |
| FRG-005 | 首充专项 | 互斥关系 | 首充固定赠送、活动赠送、今日首充赠送口径正确 | 分别准备 `activityType=0/1/2` | 逐一支付成功 | 固定赠送仅受“是否历史首充”影响；活动 1 分天；活动 2 一次性 | P0 | 功能 |

### 二、异步任务与延迟任务专项测试

#### 覆盖任务

- `lucky:expire`
- `lucky:bot_grab`
- `recharge:first_gift_installment`
- Cron `扫描过期红包`

| 用例编号 | 模块 | 子模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 | 类型 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| ASY-001 | 异步任务 | Worker 启动 | Asynq worker 未启动时主链路行为 | 停止 worker | 完成首充回调、发送红包 | 首日发放/发包主流程不阻塞；延迟任务或机器人任务仅记录日志失败 | P1 | 异常 |
| ASY-002 | 异步任务 | 过期扫描 | Cron 每分钟扫描过期红包 | 开启调度器 | 等待红包过期 | 过期红包被扫描并退款，且不会重复退款 | P1 | 异步 |
| ASY-003 | 异步任务 | 机器人链式抢包 | 一个红包可连续触发多次机器人任务 | 机器人池可用 | 发包并观察 remainingCount 变化 | 随剩余数量继续入队，直到红包抢完或结束 | P2 | 异步 |
| ASY-004 | 异步任务 | 重试机制 | 任务失败后可重试且最终幂等 | 人为制造 DB 临时失败 | 重放异步任务 | 最终仅一笔有效账变，无重复奖励/退款 | P1 | 并发 |

### 三、数据一致性与资金核对专项测试

#### 核对表

- `tg_user`
- `cash_history`
- `platform_profit_ledger`
- `recharge_order`
- `withdraw_order_br`
- `lucky_money`
- `lucky_history`
- `user_lottery_record`
- `sys_vip_reward_log`

#### 核对规则

1. 充值到账后：`tg_user.balance` 增量 = `credit_amount` + 活动当日到账奖励
2. 提现申请后：`tg_user.balance` 减量 = `withdraw_order_br.amount`
3. 提现退回后：返还金额与订单金额一致
4. 红包发送、抢包、中雷、退回，各自账变方向必须与业务一致
5. 抽奖中奖、VIP 奖励、活动赠送均需有唯一 `award_uni`

### 四、安全测试

#### 重点项

- `/api/v1/admin/*`、`/api/v1/tenant/*`、`/api/v1/app` 受保护接口鉴权
- Token 过期、伪造、跨 host 使用、禁用用户访问
- 文件上传接口类型限制
- 登录接口暴力请求与验证码接口限流
- 越权读取其他租户数据
- 公开接口是否暴露敏感后台数据

### 五、兼容性测试

#### 重点项

- H5 登录、充值、红包、提现页在移动端 Chrome、Android WebView、iOS Safari 表现
- WebSocket 断线重连后列表页和详情页刷新逻辑
- `first_recharge_gift_config` 格式兼容
  - `12(3|3|4)`
  - `12(30|30|40)`
  - 带 `%`
  - 中文括号/竖线
- 金额显示组件 `CoinAmount`、本地货币换算显示

## 测试风险与关注点

1. `POST /api/v1/admin/lucky/historyUserFlow` 当前在 `core/common/web_routes.go` 中被放到公共 `/api/v1` 分组，无 token 即可访问。
2. H5 前端 `RedPocketH5/src/api/user.ts` 存在 `createWithdrawOrder -> /api/v1/app/withdraw` 调用，但当前主路由未扫描到对应后端路由，提现 C 端链路可能阻塞。
3. `TgPhoneLoginReq` 结构仍保留 `country` 字段，但后端登录逻辑已只按 `phone` 查询，需要验证前后端约定是否一致。
4. 首充配置格式近期发生变化，需重点验证页面展示与后端发放都兼容 `3|3|4` 和 `30|30|40`。
5. H5 `getLuckyHistoryUserFlow` 仍通过 `/api/v1/admin/lucky/historyUserFlow` 调用，且请求类型中仍保留 `luckyId`，需确认接口契约是否与最新后端一致。
6. `go.mod` 当前 Go 版本格式在本地工具链下可能导致 `go test` 无法直接执行，这会影响自动化测试落地。

## 回归测试清单

### P0 回归

- 手机号注册、手机号登录、邮箱登录、找回密码
- 充值下单、支付成功回调、手动回调
- 首充固定赠送、首充活动首日赠送、今日首充赠送
- 红包发包、抢包、雷号/奇偶结算
- 红包过期退回
- 提现账户维护、提现订单扣款/退款
- 钱包账变列表、返水转余额
- 抽奖次数查询、抽奖发奖、VIP 奖励领取
- 超级管理员登录、租户登录、用户封禁、充值/提现/红包列表查看

### P1 回归

- Banner、国家/支付方式动态联动
- 团队数据、邀请统计、流水排行
- 机器人抢包链路
- WebSocket 页面实时刷新
- 奖池余额与概率配置页面
- 租户后台数据隔离

## 上线验收清单

| 检查项 | 验收标准 |
| --- | --- |
| 数据库初始化 | `sys_config`、国家、支付通道、奖池配置、Banner、VIP 等基础数据齐全 |
| Redis | 登录态、验证码、抽奖池、异步任务均可正常使用 |
| Asynq Worker | `lucky:expire`、`lucky:bot_grab`、`recharge:first_gift_installment` 正常消费 |
| Scheduler | 过期红包扫描任务已启动 |
| 支付回调 | 测试支付与手动回调均成功入账，幂等生效 |
| 首充活动 | `first_recharge_gift_config` 页面展示正确，首日到账正确，延迟任务可执行 |
| 提现 | 若 `/api/v1/app/withdraw` 尚未实现，必须明确不上线或改为后台流程 |
| WebSocket | 红包列表和详情页可实时接收广播 |
| 后台核对 | 管理后台与租户后台均能查询到对应测试数据 |
| 日志与风控 | 管理日志、异常日志、任务失败日志可检索 |

## 附：建议优先落地的自动化测试

- 接口自动化
  - 登录注册
  - 充值订单创建
  - 支付回调幂等
  - 首充分段发放
  - 红包抢包并发
  - 提现扣款/退款
  - 抽奖与 VIP 奖励领取
- 数据核对自动化
  - `tg_user.balance` 与 `cash_history` 增减额核对
  - `platform_profit_ledger` 与奖励支出核对
  - `recharge_order` 与 `withdraw_order_br` 状态流转核对
- 前端自动化
  - H5 登录、充值、红包列表、首充活动展示
  - 管理后台系统配置、充值订单、奖池配置页面
