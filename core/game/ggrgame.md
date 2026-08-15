# GGR Game API 与 Seamless API 接入文档

本文档整理自 GGR Casino API 的 Game API 和 Seamless API 公开文档，并记录本项目已经实现的 GGR 运行约定。

- 整理日期：2026-08-15
- Game API：本项目调用 GGR，基础地址由 GGR 提供，本文统一记为 `https://{API_SERVER}`
- Seamless API：GGR 回调本项目，站点 Endpoint 为 `POST https://{YOUR_SITE}/gold_api`
- 请求方式：本文涉及的接口均使用 `POST`
- 请求格式：JSON；Seamless API、Game Launch 和 In-Game History 页面明确要求 `Content-Type: application/json`
- Game API 鉴权：`agent_code`、`agent_token` 放在 JSON 请求体中
- Seamless API 回调身份字段：请求体包含 `agent_code`、`agent_secret` 和 `user_token`；本项目将 `user_token` 映射为玩家 `tg_user.uid`，它不是独立密钥
- 通用状态：`status = 1` 表示成功，`status = 0` 表示失败

> 注意：本文只记录原始页面明确给出的字段、类型和行为。原文未说明的必填性、枚举、分页起始值、时间时区、金额精度等，不在本文中推断。

## 原始文档

- [Provider List](https://ggr.gitbook.io/docs/integration/game-api/provider-list)
- [Game List](https://ggr.gitbook.io/docs/integration/game-api/game-list)
- [Game Launch](https://ggr.gitbook.io/docs/integration/game-api/game-launch)
- [Agent & User Info](https://ggr.gitbook.io/docs/integration/game-api/agent-and-user-info)
- [Game Log](https://ggr.gitbook.io/docs/integration/game-api/game-log)
- [In-Game History](https://ggr.gitbook.io/docs/integration/game-api/in-game-history)
- [User Balance (Site Endpoint)](https://ggr.gitbook.io/docs/integration/seamless-api/user-balance-site-endpoint)
- [Transaction (Site Endpoint)](https://ggr.gitbook.io/docs/integration/seamless-api/transaction-site-endpoint)
- [Providers Appendix](https://ggr.gitbook.io/docs/appendix/providers)
- [Rate Limit](https://ggr.gitbook.io/docs/integration/rate-limit)

## 本项目已确认接入约定

| 项目 | 已确认约定 |
| --- | --- |
| `user_token` | 直接使用玩家 `tg_user.uid`；查询玩家时按 UID 精确匹配 |
| 钱包字段 | 所有 GGR 游戏类型统一读取和更新 `tg_user.balance` |
| Sportsbook 钱包 | 同样使用 `tg_user.balance`，不使用 `tg_user.sport_balance` |
| 玩家代码 | `user_code = user_token = tg_user.uid` |
| 金额精度 | 所有金额进入业务计算前使用 `utils.Truncate2` 截取两位小数 |
| 联合交易 | 原子计算 `E = S - bet_money + win_money`，只在最终余额小于 `0` 时返回余额不足 |
| 重复交易 | `txn_id` 在租户 `ggr_transaction` 表保留 24 小时；保留期内相同请求幂等，不同内容返回 `INTERNAL_ERROR` |
| 失败幂等 | 首次余额不足也写入交易表，24 小时保留期内相同 `txn_id` 固定返回余额不足 |

`user_token` 在本项目中是玩家标识而不是秘密凭证，不能用它代替 `agent_secret` 或后续确认的其他回调安全机制。

## 接口一览

| 功能 | `method` | 主要返回字段 |
| --- | --- | --- |
| 获取供应商列表 | `provider_list` | `providers` |
| 获取供应商游戏列表 | `game_list` | `games` |
| 启动游戏 | `game_launch` | `launch_url` |
| 查询代理/玩家余额 | `money_info` | `agent`、`user` 或 `user_list` |
| 查询游戏交易日志 | `get_game_log` | `slot` |
| 获取游戏内历史页面 | `get_game_history` | `history_url` |
| GGR 查询站点玩家余额 | `user_balance` | `user_balance` |
| GGR 通知站点处理下注/派奖 | `transaction` | 更新后的 `user_balance` |

---

## 1. Provider List

获取分配给当前代理的游戏供应商列表。

### Endpoint

```http
POST https://{API_SERVER}
```

### 请求字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `method` | `string` | 固定为 `provider_list` |
| `agent_code` | `string` | 代理代码 |
| `agent_token` | `string` | 代理鉴权 Token |

### 请求示例

```json
{
  "method": "provider_list",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>"
}
```

### 成功响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | `integer (32-bit)` | `1` 成功，`0` 失败 |
| `msg` | `string` | 响应消息 |
| `providers` | `array` | 可用供应商列表 |

`providers` 元素：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | `string` | 供应商代码 |
| `name` | `string` | 供应商名称 |
| `status` | `integer (32-bit)` | `1` 开放，`0` 维护中 |

### 成功响应示例

```json
{
  "status": 1,
  "msg": "SUCCESS",
  "providers": [
    {
      "code": "PRAGMATIC",
      "name": "Pragmatic Play",
      "status": 1
    }
  ]
}
```

### 失败响应示例

```json
{
  "status": 0,
  "msg": "INTERNAL_ERROR"
}
```

### Provider 本地分类映射

映射依据为官方 Providers Appendix 的 `Provider Code` 和 `Type`。本项目本地分类换算规则：`slot → slots`、`live → casino`、`SB → sports`、`MN → mini`。

| Provider Code | Provider Name | GGR Type | 本地 `category_code` |
| --- | --- | --- | --- |
| `EVOLUTION` | Evolution Live | `live` | `casino` |
| `PP_LIVE_PRO` | Pragmatic Play Live | `live` | `casino` |
| `SPORTSBOOK` | Nexustrike | `SB` | `sports` |
| `PRAGMATIC` | Pragmatic Slot | `slot` | `slots` |
| `PGSOFT` | PGSoft | `slot` | `slots` |
| `HABANERO` | Habanero | `slot` | `slots` |
| `BOOONGO` | Booongo | `slot` | `slots` |
| `PLAYSON` | Playson | `slot` | `slots` |
| `CQ9` | CQ9 | `slot` | `slots` |
| `EVOPLAY` | Evoplay | `slot` | `slots` |
| `TOPTREND` | TopTrend | `slot` | `slots` |
| `DREAMTECH` | DreamTech | `slot` | `slots` |
| `SPRIBE` | Spribe | `MN` | `mini` |
| `HACKSAW` | Hacksaw | `slot` | `slots` |
| `FACHAI` | FaChai | `slot` | `slots` |
| `PLAYNGO` | Play'n Go | `slot` | `slots` |
| `AMUSNET` | Amusnet | `slot` | `slots` |
| `EGT` | EGT Digital | `slot` | `slots` |
| `SPADEGAMING` | Spadegaming | `slot` | `slots` |
| `FASTSPIN` | FastSpin | `slot` | `slots` |
| `JOKERGAMING` | Joker Gaming | `slot` | `slots` |
| `RUBYPLAY` | RubyPlay | `slot` | `slots` |
| `AMATIC` | Amatic | `slot` | `slots` |
| `REELKINGDOM` | Reel Kingdom | `slot` | `slots` |
| `FATPANDA` | Fat Panda | `slot` | `slots` |
| `FISHHUNTER` | Fish Hunter | 待 GGR 确认 | `fishing` |

同步遇到不在映射表中的新 `provider_code` 时，会先在日志中输出 Provider List 返回的全部厂商代码、名称、状态以及所有缺失映射，再终止整次同步；不会写入部分游戏，也不能自动归入任意分类。

`FISHHUNTER` 使用项目已有的本地捕鱼分类，对应 `app_game.type=3`。GGR 当前公开的 Providers 文档尚未列出该厂商，其 Seamless Transaction 使用的 `game_type`（`slot` 或 `MN`）仍须由 GGR 确认；确认前不得根据本地 `fishing` 分类推断回调对象键。

---

## 2. Game List

根据供应商代码获取该供应商下的游戏列表。

### Endpoint

```http
POST https://{API_SERVER}
```

### 请求字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `method` | `string` | 固定为 `game_list` |
| `agent_code` | `string` | 代理代码 |
| `agent_token` | `string` | 代理鉴权 Token |
| `provider_code` | `string` | 供应商代码 |

### 请求示例

```json
{
  "method": "game_list",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>",
  "provider_code": "PRAGMATIC"
}
```

### 成功响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | `integer (32-bit)` | `1` 成功，`0` 失败 |
| `msg` | `string` | 响应消息 |
| `games` | `array` | 游戏列表 |

`games` 元素：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `game_code` | `string` | 游戏唯一代码 |
| `game_name` | `string` | 游戏名称 |
| `banner` | `string` | 游戏 Banner 图片 URL |
| `status` | `integer (32-bit)` | `1` 开放，`0` 维护中 |

### 成功响应示例

```json
{
  "status": 1,
  "msg": "SUCCESS",
  "games": [
    {
      "game_code": "vs20doghouse",
      "game_name": "The Dog House",
      "banner": "<BANNER_URL>",
      "status": 1
    },
    {
      "game_code": "vs243mwarrior",
      "game_name": "Monkey Warrior",
      "banner": "<BANNER_URL>",
      "status": 0
    }
  ]
}
```

### 失败响应示例

```json
{
  "status": 0,
  "msg": "INTERNAL_ERROR"
}
```

---

## 3. Game Launch

为玩家获取老虎机或真人游戏的启动 URL。

### Endpoint

```http
POST https://{API_SERVER}
Content-Type: application/json
```

### 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `method` | `string` | 是 | 固定为 `game_launch` |
| `agent_code` | `string` | 是 | 代理代码 |
| `agent_token` | `string` | 是 | 代理 API Token |
| `user_code` | `string` | 是 | 玩家标识 |
| `provider_code` | `string` | 是 | 供应商代码 |
| `game_code` | `string` | 可选 | 游戏代码；真人游戏大厅场景可以为空 |
| `lang` | `string` | 是 | 语言代码，例如 `en` |
| `rtp` | `number` | 否 | 本次启动的目标 RTP；省略时使用玩家已配置的 RTP |
| `lobby_url` | `string` | 否 | 玩家退出游戏后的跳转 URL；原文说明默认值为空字符串 |

> 对于真人游戏，`game_code` 可以不传或传空值；此时返回真人游戏大厅 URL。原文没有说明老虎机场景省略 `game_code` 的行为。

### Slot 请求示例

```json
{
  "method": "game_launch",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>",
  "user_code": "test",
  "provider_code": "PRAGMATIC",
  "game_code": "vs20doghouse",
  "lang": "en",
  "rtp": 92,
  "lobby_url": "https://your-site.com/lobby"
}
```

### Live 请求示例

```json
{
  "method": "game_launch",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>",
  "user_code": "test",
  "provider_code": "EVOLUTION",
  "game_code": "nxpkul2hgclallno",
  "lang": "en",
  "lobby_url": "https://your-site.com/lobby"
}
```

### 成功响应

```json
{
  "status": 1,
  "msg": "SUCCESS",
  "launch_url": "<LAUNCH_URL>"
}
```

### 失败响应

```json
{
  "status": 0,
  "msg": "INVALID_PROVIDER"
}
```

---

## 4. Agent & User Info

使用同一个 `money_info` 方法查询代理余额、指定玩家余额或全部玩家余额。

### Endpoint

```http
POST https://{API_SERVER}
```

### 4.1 Agent Info

查询当前代理余额。

#### 请求字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `method` | `string` | 固定为 `money_info` |
| `agent_code` | `string` | 代理代码 |
| `agent_token` | `string` | 代理鉴权 Token |

#### 请求示例

```json
{
  "method": "money_info",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>"
}
```

#### 成功响应示例

```json
{
  "status": 1,
  "msg": "SUCCESS",
  "agent": {
    "agent_code": "<AGENT_CODE>",
    "balance": 1000000
  }
}
```

`agent` 字段来自原文响应示例；该页面没有单独提供其正式字段表：

| 字段 | 示例类型 | 说明 |
| --- | --- | --- |
| `agent_code` | `string` | 代理代码 |
| `balance` | `number` | 代理余额 |

#### 失败响应示例

```json
{
  "status": 0,
  "msg": "INVALID_PARAMETER"
}
```

### 4.2 User Info

查询当前代理和指定玩家的余额。

#### 请求字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `method` | `string` | 固定为 `money_info` |
| `agent_code` | `string` | 代理代码 |
| `agent_token` | `string` | 代理鉴权 Token |
| `user_code` | `string` | 玩家代码 |

#### 请求示例

```json
{
  "method": "money_info",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>",
  "user_code": "testUser"
}
```

#### 成功响应示例

```json
{
  "status": 1,
  "msg": "SUCCESS",
  "agent": {
    "agent_code": "<AGENT_CODE>",
    "balance": 1000000
  },
  "user": {
    "user_code": "testUser",
    "balance": 100000
  }
}
```

`user` 字段来自原文响应示例；该页面没有单独提供其正式字段表：

| 字段 | 示例类型 | 说明 |
| --- | --- | --- |
| `user_code` | `string` | 玩家代码 |
| `balance` | `number` | 玩家余额 |

#### 失败响应示例

```json
{
  "status": 0,
  "msg": "INVALID_PARAMETER"
}
```

### 4.3 All Users Info

查询当前代理及其全部玩家余额。

#### 请求字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `method` | `string` | 请求示例使用 `money_info`；原文字段表误写为 `game_list`，联调前需要向 GGR 确认 |
| `agent_code` | `string` | 代理代码 |
| `agent_token` | `string` | 代理鉴权 Token |
| `all_users` | `boolean` | 固定传 `true` |

#### 请求示例

```json
{
  "method": "money_info",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>",
  "all_users": true
}
```

#### 成功响应示例

```json
{
  "status": 1,
  "msg": "SUCCESS",
  "agent": {
    "agent_code": "<AGENT_CODE>",
    "balance": 99550000
  },
  "user_list": [
    {
      "user_code": "testUser1",
      "balance": 450000
    },
    {
      "user_code": "testUser2",
      "balance": 20000
    }
  ]
}
```

#### 失败响应示例

```json
{
  "status": 0,
  "msg": "INVALID_PARAMETER"
}
```

---

## 5. Game Log

查询指定时间范围内某个玩家或代理的游戏交易日志。

### Endpoint

```http
POST https://{API_SERVER}
```

### 请求字段

原文没有提供 Required 列，因此下表不推断字段是否允许省略。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `method` | `string` | 固定为 `get_game_log` |
| `agent_code` | `string` | 代理代码 |
| `agent_token` | `string` | 代理鉴权 Token |
| `user_code` | `string` | 玩家代码 |
| `game_type` | `string` | 要查询的游戏类型；原文只给出 `slot` 示例，没有完整枚举 |
| `start` | `string` | 开始时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `end` | `string` | 结束时间，格式 `yyyy-MM-dd HH:mm:ss` |
| `page` | `integer` | 页码；原文示例为 `0`，没有明确页码是否固定从 0 开始 |
| `perPage` | `integer` | 每页记录数，最大 `1,000,000` |

### 请求示例

```json
{
  "method": "get_game_log",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>",
  "user_code": "<USER_CODE>",
  "game_type": "slot",
  "start": "2021-09-17 00:00:00",
  "end": "2021-09-17 23:59:00",
  "page": 0,
  "perPage": 1000
}
```

### 成功响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | `integer (32-bit)` | `1` 表示成功 |
| `total_count` | `integer (32-bit)` | 符合条件的日志总数 |
| `page` | `integer (32-bit)` | 当前页码 |
| `perPage` | `integer (32-bit)` | 本次请求的每页记录数 |
| `slot` | `array` | 游戏交易日志数组 |

`slot` 元素：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `history_id` | `integer (64-bit)` | 交易历史唯一 ID |
| `agent_code` | `string` | 处理该交易的代理代码 |
| `user_code` | `string` | 玩家代码 |
| `provider_code` | `string` | 游戏供应商代码 |
| `game_code` | `string` | 游戏唯一代码 |
| `type` | `string` | 游戏交易类型；原文示例为 `BASE`、`FREE` |
| `bet_money` | `double` | 玩家下注金额 |
| `win_money` | `double` | 玩家赢取金额 |
| `txn_id` | `string` | 唯一交易 ID |
| `txn_type` | `string` | 交易类型；原文示例为 `debit_credit` |
| `user_start_balance` | `double` | 玩家交易前余额 |
| `user_end_balance` | `double` | 玩家交易后余额 |
| `agent_start_balance` | `double` | 代理交易前余额 |
| `agent_end_balance` | `double` | 代理交易后余额 |
| `created_at` | `string` | 交易创建时间 |

### 成功响应示例

```json
{
  "status": 1,
  "total_count": 340,
  "page": 0,
  "perPage": 1000,
  "slot": [
    {
      "history_id": 777,
      "agent_code": "admin",
      "user_code": "test",
      "provider_code": "PRAGMATIC",
      "game_code": "vs20doghouse",
      "type": "BASE",
      "bet_money": 2000,
      "win_money": 0,
      "txn_id": "64a83f2fc597acc9004eec52c3f84c30",
      "txn_type": "debit_credit",
      "user_start_balance": 230500,
      "user_end_balance": 228500,
      "agent_start_balance": 22092000,
      "agent_end_balance": 22092000,
      "created_at": "2021-09-17T12:50:42.000Z"
    }
  ]
}
```

### 失败响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | `integer (32-bit)` | `0` 表示失败 |
| `msg` | `string` | 错误概要 |
| `detail` | `string` | 具体错误说明 |

### 失败响应示例

```json
{
  "status": 0,
  "msg": "Invalid Parameter.",
  "detail": "perPage must be less than or equal to 1000000"
}
```

---

## 6. In-Game History

获取由游戏供应商托管的玩家游戏历史页面 URL。

该接口与 `get_game_log` 不同：它不直接返回交易记录，而是返回一个可打开的供应商历史页面地址。

### 支持范围

当前仅支持 Pragmatic Play Slot：

```text
provider_code = PRAGMATIC
```

其他供应商会返回 `INVALID_PROVIDER`。

### Endpoint

```http
POST https://{API_SERVER}
Content-Type: application/json
```

### 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `method` | `string` | 是 | 固定为 `get_game_history` |
| `agent_code` | `string` | 是 | 代理代码 |
| `agent_token` | `string` | 是 | 代理 API Token |
| `user_code` | `string` | 是 | 玩家标识 |
| `provider_code` | `string` | 是 | 当前必须为 `PRAGMATIC` |
| `game_code` | `string` | 是 | Pragmatic Play Slot 游戏代码 |

### 请求示例

```json
{
  "method": "get_game_history",
  "agent_code": "<AGENT_CODE>",
  "agent_token": "<AGENT_TOKEN>",
  "user_code": "test",
  "provider_code": "PRAGMATIC",
  "game_code": "vs20doghouse"
}
```

### 成功响应

```json
{
  "status": 1,
  "msg": "SUCCESS",
  "history_url": "<HISTORY_URL>"
}
```

| 字段 | 说明 |
| --- | --- |
| `history_url` | 供应商托管的玩家游戏历史页面 URL |

### 失败响应

供应商不支持：

```json
{
  "status": 0,
  "msg": "INVALID_PROVIDER"
}
```

外部游戏供应商错误：

```json
{
  "status": 0,
  "msg": "EXTERNAL_ERROR : INVALID_GAME"
}
```

GGR 内部错误：

```json
{
  "status": 0,
  "msg": "INTERNAL_ERROR"
}
```

---

## 7. Seamless API：User Balance（站点 Endpoint）

GGR 在玩家进入游戏前或游戏过程中调用该接口，查询玩家当前余额。

### 调用方向

```text
GGR → 本项目站点
```

### Endpoint

```http
POST https://{YOUR_SITE}/gold_api
Content-Type: application/json
```

### 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `method` | `string` | 是 | 固定为 `user_balance` |
| `agent_code` | `string` | 是 | 代理代码 |
| `agent_secret` | `string` | 是 | 代理 Secret Key |
| `user_code` | `string` | 是 | 玩家标识 |
| `user_token` | `string` | 是 | GGR 定义为玩家会话 Token；本项目传递并按 `tg_user.uid` 解析 |
| `game_code` | `string` | 是 | 玩家当前游戏代码 |

### 请求示例

```json
{
  "method": "user_balance",
  "agent_code": "<AGENT_CODE>",
  "agent_secret": "<AGENT_SECRET>",
  "user_code": "<USER_CODE>",
  "user_token": "<TG_USER_UID>",
  "game_code": "vs20olympgate"
}
```

### 成功响应

即使玩家余额为 `0`，也必须返回余额字段：

本项目从 `tg_user.balance` 读取该余额。

```json
{
  "status": 1,
  "user_balance": 0
}
```

### 失败响应

```json
{
  "status": 0,
  "user_balance": 0,
  "msg": "INTERNAL_ERROR"
}
```

### 响应字段

原文没有提供单独的响应字段表，以下类型来自响应示例：

| 字段 | 示例类型 | 说明 |
| --- | --- | --- |
| `status` | `integer` | `1` 成功，`0` 失败 |
| `user_balance` | `number` | 玩家当前余额；成功且余额为 0 时仍需返回 |
| `msg` | `string` | 失败消息；成功示例中没有该字段 |

---

## 8. Seamless API：Transaction（站点 Endpoint）

GGR 在发生下注或派奖时调用该接口。站点后端必须处理玩家余额，并在成功响应中返回更新后的余额。

### 调用方向

```text
GGR → 本项目站点
```

### Endpoint

```http
POST https://{YOUR_SITE}/gold_api
Content-Type: application/json
```

### 顶层请求字段

原文没有为顶层字段提供正式字段表，以下字段和类型来自四种交易请求示例；除 `method = transaction` 外，原文没有逐项标注必填性。

| 字段 | 示例类型 | 说明 |
| --- | --- | --- |
| `method` | `string` | 固定为 `transaction` |
| `agent_code` | `string` | 代理代码 |
| `agent_secret` | `string` | 代理 Secret Key |
| `agent_balance` | `number` | 请求示例中的代理余额；原文未说明站点应如何使用该值 |
| `user_code` | `string` | 玩家标识 |
| `user_token` | `string` | GGR 定义为玩家会话 Token；本项目传递并按 `tg_user.uid` 解析 |
| `user_balance` | `number` | 请求示例中的玩家余额；原文未说明该值是否仅供参考 |
| `game_type` | `string` | 游戏类型，本文页面列出 `slot`、`live`、`SB`、`MN` |
| `info` | `string` | 真人和体育交易的顶层扩展信息，为序列化后的 JSON 字符串 |
| 动态游戏对象 | `object` | 对象键必须与 `game_type` 完全一致 |

### 游戏类型与对象键

| `game_type` | 必须使用的对象键 | 含义 |
| --- | --- | --- |
| `slot` | `slot` | 电子老虎机 |
| `live` | `live` | 真人游戏 |
| `SB` | `SB` | 体育游戏 |
| `MN` | `MN` | Mini Game |

例如，`"game_type": "SB"` 时，请求体必须包含 `"SB": { ... }`，不能改成其他大小写或通用字段名。

### 游戏对象字段

`slot`、`live`、`SB`、`MN` 对象内部使用相同字段结构：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `provider_code` | `string` | 是 | 游戏供应商代码 |
| `game_code` | `string` | 是 | 游戏标识 |
| `type` | `string` | 是 | 投注或局类型，例如 `BASE`、`baccarat` |
| `bet_money` | `number` | 是 | 下注金额 |
| `win_money` | `number` | 是 | 派奖金额 |
| `round_id` | `integer` | 否 | 单次 Spin/局的唯一 ID；同一局的多笔交易共享该值 |
| `txn_id` | `string` | 是 | 单笔交易唯一 ID |
| `txn_type` | `string` | 是 | `debit`、`credit` 或 `debit_credit` |

### `round_id` 与 `txn_id`

- `round_id`：一局或一次 Spin 的标识，类型为整数；同一局可以产生多笔交易。
- `txn_id`：每一笔独立扣款或加款的唯一标识，类型为字符串。
- 同一 `round_id` 下的不同交易必须具有不同的 `txn_id`。

### Slot 交易示例

```json
{
  "method": "transaction",
  "agent_code": "<AGENT_CODE>",
  "agent_secret": "<AGENT_SECRET>",
  "agent_balance": 10000000,
  "user_code": "<USER_CODE>",
  "user_token": "<TG_USER_UID>",
  "user_balance": 99200,
  "game_type": "slot",
  "slot": {
    "provider_code": "PRAGMATIC",
    "game_code": "vs20midas",
    "type": "BASE",
    "bet_money": 1000,
    "win_money": 200,
    "round_id": 63792613432127,
    "txn_id": "64a83f2fc597acc9004eec52c3f84c30",
    "txn_type": "debit_credit"
  }
}
```

### Live 交易示例

`info` 是序列化后的 JSON 字符串，内容为真人游戏的投注或结果详情。原文没有提供 Live `info` 的正式字段表。

```json
{
  "method": "transaction",
  "agent_code": "<AGENT_CODE>",
  "agent_secret": "<AGENT_SECRET>",
  "agent_balance": 10000000,
  "user_code": "<USER_CODE>",
  "user_token": "<TG_USER_UID>",
  "user_balance": 102000,
  "game_type": "live",
  "info": "<STRINGIFIED_JSON>",
  "live": {
    "provider_code": "EVOLUTION",
    "game_code": "100",
    "type": "baccarat",
    "bet_money": 2000,
    "win_money": 4000,
    "round_id": 15676627318,
    "txn_id": "64a83f2fc597acc9004eec52c3f84c30",
    "txn_type": "debit_credit"
  }
}
```

### Sportsbook 交易示例

```json
{
  "method": "transaction",
  "agent_code": "<AGENT_CODE>",
  "agent_secret": "<AGENT_SECRET>",
  "agent_balance": 10000000,
  "user_code": "<USER_CODE>",
  "user_token": "<TG_USER_UID>",
  "user_balance": 102000,
  "game_type": "SB",
  "info": "<STRINGIFIED_JSON>",
  "SB": {
    "provider_code": "SPORTSBOOK",
    "game_code": "SPORTSBOOK",
    "type": "accumulator",
    "bet_money": 5,
    "win_money": 0,
    "round_id": 561435,
    "txn_id": "44097c196a8b59b7371ee2ca6fe83999",
    "txn_type": "debit"
  }
}
```

### Mini Game 交易示例

```json
{
  "method": "transaction",
  "agent_code": "<AGENT_CODE>",
  "agent_secret": "<AGENT_SECRET>",
  "agent_balance": 10000000,
  "user_code": "<USER_CODE>",
  "user_token": "<TG_USER_UID>",
  "user_balance": 99200,
  "game_type": "MN",
  "MN": {
    "provider_code": "SPRIBE",
    "game_code": "minigame_01",
    "type": "BASE",
    "bet_money": 1000,
    "win_money": 200,
    "round_id": 123456789,
    "txn_id": "64a83f2fc597acc9004eec52c3f84c30",
    "txn_type": "debit_credit"
  }
}
```

### Sportsbook `info` 字段

Sportsbook 的 `info` 为序列化后的 JSON 字符串，其 JSON 对象包含：

| 字段 | 说明 |
| --- | --- |
| `couponCode` | 投注 Coupon 代码 |
| `status` | `pending`、`won`、`lost`、`cashedout` 或 `canceled` |
| `totalOdds` | 组合赔率 |
| `potentialWin` | 潜在派奖金额 |
| `stake` | 投注本金 |
| `payout` | 实际派奖金额 |
| `betslips` | 投注项数组 |
| `settlementData` | 结算详情对象 |

`betslips` 元素在原文中列出的字段：

```text
id
sportName
countryName
leagueName
matchName
homeTeam
awayTeam
marketName
oddName
oddRate
oddPoint
```

### `txn_type` 处理规则

| `txn_type` | 含义 | `bet_money` | `win_money` |
| --- | --- | --- | --- |
| `debit` | 仅下注扣款 | 使用 | 不使用 |
| `credit` | 仅派奖加款 | 不使用 | 使用 |
| `debit_credit` | 同一交易同时下注和派奖 | 使用 | 使用 |

> 本项目按已确认规则严格处理：两个金额字段都必须存在；`debit` 必须为 `bet_money > 0, win_money = 0`，`credit` 必须为 `bet_money = 0, win_money > 0`，`debit_credit` 两者必须非负且至少一项大于 `0`。

### 单局多交易示例

同一个 `round_id` 可以包含多笔不同 `txn_id` 的交易：

```json
{
  "txn_type": "debit",
  "bet_money": 2000,
  "win_money": 0,
  "round_id": 123456789,
  "txn_id": "64a83f2fc597acc9004eec52c3f84c30"
}
```

```json
{
  "txn_type": "credit",
  "bet_money": 0,
  "win_money": 1500,
  "round_id": 123456789,
  "txn_id": "<DIFFERENT_TXN_ID>"
}
```

```json
{
  "txn_type": "debit_credit",
  "bet_money": 2000,
  "win_money": 1500,
  "round_id": 123456790,
  "txn_id": "a1b2c3d4e5f6789012345678abcdef01"
}
```

### 成功响应

成功时必须返回余额变更后的玩家余额：

本项目对 Slot、Live、Sportsbook 和 Mini Game 均在同一数据库事务中更新 `tg_user.balance`，成功响应返回更新后的该字段值。

```json
{
  "status": 1,
  "user_balance": 1000
}
```

### 失败响应

原文给出的余额不足响应：

```json
{
  "status": 0,
  "msg": "INSUFFICIENT_USER_FUNDS"
}
```

---

## 9. 本项目运行链路

### 游戏同步

管理员调用：

```http
POST /api/v1/admin/appGame/sync
Content-Type: application/json

{"platformCode":"ggr","language":"en"}
```

同步流程为 Provider List → 校验全部 Provider 分类 → 按每秒一次顺序调用各 Provider 的 Game List → 单个数据库事务更新本地目录。任一步失败都不会产生部分同步结果。远端不存在或处于维护状态的游戏会被禁用；`hot`、`home_show`、`sort` 和 `show_index` 保留管理员设置。

本地游戏键为：

```text
(platform_code = ggr, third_game_category = provider_code, third_game_id = game_code)
```

### 游戏启动

App 继续调用现有 `POST /api/v1/app/appGame/launch`。当游戏 `platform_code = ggr` 时，服务端用 `tg_user.uid` 作为 `user_code`，从本地游戏记录读取 `provider_code` 和 `game_code`，并把当前租户 Origin 作为 `lobby_url`。未配置 RTP 时请求中不发送 `rtp`。

### Seamless 回调

- GGR Profile 的 Site Endpoint 配置为租户 HTTPS Origin，不包含 `/gold_api`；GGR 会追加该路径。
- `POST /gold_api` 不使用平台 JWT，以请求体中的 `agent_code`、`agent_secret` 鉴权。
- `user_balance` 成功时从 `tg_user.balance` 返回当前余额，包括余额为 `0` 的情况。
- `transaction` 在用户行锁保护的数据库事务内更新余额，并写入 `ggr_transaction`、`cash_history`、`app_user_bet_record` 和提现流水事件。
- 顶层 `agent_balance`、`user_balance` 只保存为审计快照，不参与本地余额计算。
- 所有 GGR 业务响应均使用 HTTP 200，并直接返回 GGR JSON，不套用项目通用响应结构。
- `runScheduler = true` 时，调度器每小时整点按 `received_at < 当前时间 - 24 小时` 清理所有启用租户的 `ggr_transaction`；记录删除后对应 `txn_id` 不再具备本地幂等保护。

### 部署步骤

1. 在 `core.yaml` 填写 GGR Profile 提供的 `apiUrl`、`agentCode`、`agentToken`、`agentSecret`。
2. 至少启动一次未设置 `BGU_SKIP_AUTO_MIGRATE=1` 的服务，为每个租户创建 `ggr_transaction`；跳过自动迁移的环境需人工创建同结构表。
3. 在 GGR Profile 配置租户 Site Endpoint。
4. 调用 GGR 游戏同步接口，确认游戏分类、图片和状态。
5. 联调零余额查询、`debit`、`credit`、`debit_credit`、余额不足、成功重复和冲突重复场景。

## 10. 仍需 GGR 确认的外部协议项

1. API Server 的正式地址、测试地址及当前代理的正式 Agent Code。
2. `agent_token` 是否存在请求体以外的额外保护要求。
3. 未在本项目运行链路实现的 Money Info、Game Log、In-Game History 字段歧义和分页/时区规则。
4. `rtp` 的合法范围、单位及不同供应商的支持情况。
5. `launch_url`、`history_url` 的有效期和来源域名/IP限制。
6. Seamless API 是否另有签名、时间戳、IP 白名单或防重放要求。
7. GGR 的回调超时、重试次数、失败补单机制及完整错误码集合。
8. Live `info` 的正式 JSON Schema，以及 Sportsbook `settlementData` 的完整结构。
