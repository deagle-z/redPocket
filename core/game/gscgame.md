# 2. 单一钱包

所有 API 请求都使用 `application/json` 内容类型头。

所有请求的时间格式为带有时区 GMT+8 的秒级时间戳。

## 2.1 Balance

提供商通过此 API 获取特定玩家余额。该接口由运营商侧的无缝钱包提供，用于让提供商查询玩家余额。

## Endpoint

```http
POST {{callback_url}}/v1/api/seamless/balance
```

## 请求参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `batch_requests` | `[]batch_requests` | 请求信息。 |
| `operator_code` | `string` | 操作员作为用户名登录后台的唯一标识符。 |
| `currency` | `string` | 货币代码。 |
| `sign` | `string` | 请求签名。 |
| `request_time` | `string` | 请求时间戳，单位秒。 |

## `batch_requests` 参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `member_account` | `string` | 操作员中成员的唯一标识符，限制 50 字符。 |
| `product_code` | `int` | 产品唯一标识符。 |

## 签名规则

```text
md5(operator_code + request_time + "getbalance" + secret_key)
```

示例：

```text
md5(ABCD + 1698219740 + getbalance + XXXX)
```

验签工具：

```text
https://testcase.gscplusmd.com
```

## 请求示例

```json
{
  "batch_requests": [
    {
      "member_account": "user1",
      "product_code": 1002
    },
    {
      "member_account": "user2",
      "product_code": 1020
    },
    {
      "member_account": "user3",
      "product_code": 1009
    }
  ],
  "operator_code": "ABCD",
  "currency": "CNY",
  "sign": "369af7416deef76a9cc4f019b8559f99",
  "request_time": "1694617425"
}
```

## 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data` | `[]data` | 响应数据。 |

## `data` 字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `member_account` | `string` | 操作员中成员的唯一标识符，限制 50 字符。 |
| `product_code` | `int` | 产品唯一标识符。 |
| `balance` | `float64` | 玩家余额，支持到小数第四位。 |
| `code` | `int` | 无缝钱包代码。 |
| `message` | `string` | 响应消息。 |

## 响应示例

```json
{
  "data": [
    {
      "member_account": "user1",
      "product_code": 1002,
      "balance": 12345,
      "code": 0,
      "message": ""
    },
    {
      "member_account": "user2",
      "product_code": 1020,
      "balance": 1000,
      "code": 0,
      "message": ""
    },
    {
      "member_account": "user3",
      "product_code": 1009,
      "balance": 1000,
      "code": 0,
      "message": ""
    }
  ]
}
```

## 2.2 Withdraw

运营商侧无缝钱包接口，用于玩家下注或类似扣款操作，例如小费。

## Endpoint

```http
POST {{callback_url}}/v1/api/seamless/withdraw
```

## 请求参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `batch_requests` | `[]batch_requests` | 请求信息。 |
| `operator_code` | `string` | 操作员作为用户名登录 BO 的唯一标识符。 |
| `game_type` | `string` | 可选，大多数情况为空值，部分情况会使用。 |
| `currency` | `string` | 货币代码。 |
| `sign` | `string` | 请求签名。 |
| `request_time` | `string` | 请求时间戳，单位秒。 |

## `batch_requests` 参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `member_account` | `string` | 操作员中成员的唯一标识符，限制 50 字符。 |
| `product_code` | `int` | 产品唯一标识符。 |
| `game_type` | `string` | 游戏类型。 |
| `transactions` | `[]Transaction` | 交易列表。 |

## `transactions` 常见字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `string` | 交易 ID。 |
| `action` | `string` | 交易动作，例如 `BET`。 |
| `wager_code` | `string` | 注单编号。 |
| `wager_status` | `string` | 注单状态，例如 `BET`。 |
| `amount` | `float64` | 扣款金额。 |
| `bet_amount` | `float64` | 下注金额。 |
| `valid_bet_amount` | `float64` | 有效下注金额。 |
| `prize_amount` | `float64` | 派彩金额。 |
| `tip_amount` | `float64` | 小费金额。 |
| `settled_at` | `int64` | 结算时间。 |
| `game_code` | `string` | 游戏代码。 |
| `round_id` | `string` | 回合 ID。 |
| `channel_code` | `string` | 渠道代码。 |
| `wager_type` | `string` | 注单类型，例如 `NORMAL`。 |

## 签名规则

```text
md5(operator_code + request_time + "withdraw" + secret_key)
```

示例：

```text
md5(ABCD + 1698219740 + withdraw + XXXX)
```

验签工具：

```text
https://testcase.gscplusmd.com
```

## 请求示例

```json
{
  "batch_requests": [
    {
      "member_account": "user1",
      "product_code": 1002,
      "game_type": "POKER",
      "transactions": [
        {
          "id": "23746",
          "action": "BET",
          "wager_code": "tZDwLV3ayzBeP4Nvwxhcti",
          "wager_status": "BET",
          "amount": 10,
          "bet_amount": 10,
          "valid_bet_amount": 10,
          "prize_amount": 0,
          "tip_amount": 0,
          "settled_at": 0,
          "game_code": "moreturkeyv10000",
          "round_id": "95978",
          "channel_code": "gscp",
          "wager_type": "NORMAL"
        }
      ]
    }
  ],
  "operator_code": "ABCD",
  "game_type": "",
  "currency": "CNY",
  "sign": "369af7416deef76a9cc4f019b8559f99",
  "request_time": "1694617425"
}
```

## 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data` | `[]data` | 响应数据。 |

## `data` 字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `member_account` | `string` | 操作员中成员的唯一标识符，限制 50 字符。 |
| `product_code` | `int` | 产品唯一标识符。 |
| `before_balance` | `float64` | 操作前玩家余额，支持到小数第四位。 |
| `balance` | `float64` | 操作后玩家余额，支持到小数第四位。 |
| `code` | `int` | 无缝钱包代码。 |
| `message` | `string` | 响应消息。 |

> 交易 ID 已存在时：如果 `tx_id` 存在于运营商系统中，并且之前已经退款，请返回重复交易。

## 响应示例

```json
{
  "data": [
    {
      "member_account": "user1",
      "product_code": 1002,
      "before_balance": 12345,
      "balance": 12340,
      "code": 0,
      "message": ""
    }
  ]
}
```

## 2.3 Deposit

运营商侧无缝钱包接口，用于玩家获得奖金或类似增加操作，例如根据玩家参与的活动给予信用。

注意：

- 没有 bet 也要接受 Deposit 加款。部分厂商会额外发活动奖金，如果因为没有 bet 就拒绝 settled，玩家会收不到奖金派奖。
- 产品 WBET 采用特殊机制，派奖不会通过 `/deposit` API 进行，而是需由运营商在接收到 `/push-bet-data` 通知后自行完成派奖流程。

## Endpoint

```http
POST {{callback_url}}/v1/api/seamless/deposit
```

## 请求参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `batch_requests` | `[]batch_requests` | 请求信息。 |
| `operator_code` | `string` | 操作员作为用户名登录 BO 的唯一标识符。 |
| `currency` | `string` | 货币代码。 |
| `sign` | `string` | 请求签名。 |
| `request_time` | `string` | 请求时间戳，单位秒。 |

## `batch_requests` 参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `member_account` | `string` | 操作员中成员的唯一标识符，限制 50 字符。 |
| `product_code` | `int` | 产品唯一标识符。 |
| `game_type` | `string` | 游戏类型。 |
| `transactions` | `[]Transaction` | 交易列表。 |

## `transactions` 常见字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `string` | 交易 ID。 |
| `action` | `string` | 交易动作，例如 `settled`。 |
| `wager_code` | `string` | 注单编号。 |
| `wager_status` | `string` | 注单状态，例如 `SETTLED`。 |
| `round_id` | `string` | 回合 ID。 |
| `channel_code` | `string` | 渠道代码。 |
| `amount` | `float64` | 加款金额。 |
| `bet_amount` | `float64` | 下注金额。 |
| `valid_bet_amount` | `float64` | 有效下注金额。 |
| `prize_amount` | `float64` | 派彩金额。 |
| `tip_amount` | `float64` | 小费金额。 |
| `settled_at` | `int64` | 结算时间。 |
| `game_code` | `string` | 游戏代码。 |
| `wager_type` | `string` | 注单类型，例如 `NORMAL`。 |

## 签名规则

```text
md5(operator_code + request_time + "deposit" + secret_key)
```

示例：

```text
md5(ABCD + 1698219740 + deposit + XXXX)
```

验签工具：

```text
https://testcase.gscplusmd.com
```

## 请求示例

```json
{
  "batch_requests": [
    {
      "member_account": "user1",
      "product_code": 1002,
      "game_type": "POKER",
      "transactions": [
        {
          "id": "23746",
          "action": "settled",
          "wager_code": "tZDwLV3ayzBeP4Nvwxhcti",
          "wager_status": "SETTLED",
          "round_id": "95978",
          "channel_code": "gscp",
          "amount": 10,
          "bet_amount": 10,
          "valid_bet_amount": 10,
          "prize_amount": 10,
          "tip_amount": 0,
          "settled_at": 1729134752372,
          "game_code": "moreturkeyv10000",
          "wager_type": "NORMAL"
        }
      ]
    }
  ],
  "operator_code": "ABCD",
  "currency": "CNY",
  "sign": "369af7416deef76a9cc4f019b8559f99",
  "request_time": "1694617425"
}
```

## 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data` | `[]data` | 响应数据。 |

## `data` 字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `member_account` | `string` | 操作员中成员的唯一标识符，限制 50 字符。 |
| `product_code` | `int` | 产品唯一标识符。 |
| `before_balance` | `float64` | 操作前玩家余额，支持到小数第四位。 |
| `balance` | `float64` | 操作后玩家余额，支持到小数第四位。 |
| `code` | `int` | 无缝钱包代码。 |
| `message` | `string` | 响应消息。 |

> 交易 ID 已存在时：如果 `id` 存在于运营商系统中，并且之前已经退款，请返回重复交易。

## 响应示例

```json
{
  "data": [
    {
      "member_account": "user1",
      "product_code": 1002,
      "before_balance": 12345,
      "balance": 12340,
      "code": 0,
      "message": ""
    }
  ]
}
```

## 2.4 Push Bet Data

运营商端无缝钱包 API，用于同步注单所有数据和状态。

## Endpoint

```http
POST {{callback_url}}/v1/api/seamless/pushbetdata
```

## 请求参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `operator_code` | `string` | 运营商唯一标识，作为后台登录用户名。 |
| `wagers` | `[]wagers` | 本次操作的注单列表。 |
| `sign` | `string` | 请求签名。 |
| `request_time` | `string` | 请求时间戳，单位秒。 |

## `wagers` 常见字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `member_account` | `string` | 操作员中成员的唯一标识符，限制 50 字符。 |
| `bet_amount` | `string` | 下注金额。 |
| `valid_bet_amount` | `string` | 有效下注金额。 |
| `prize_amount` | `string` | 派彩金额。 |
| `tip_amount` | `string` | 小费金额。 |
| `wager_type` | `string` | 注单类型，例如 `NORMAL`。 |
| `wager_code` | `string` | 注单编号。 |
| `wager_status` | `string` | 注单状态，例如 `SETTLED`。 |
| `round_id` | `string` | 回合 ID。 |
| `channel_code` | `string` | 渠道代码。 |
| `game_type` | `string` | 游戏类型。 |
| `settled_at` | `int64` | 结算时间。 |
| `created_at` | `int64` | 创建时间。 |
| `payload` | `object` | 提供商原始附加数据。 |
| `product_code` | `string` | 产品唯一标识符。 |
| `game_code` | `string` | 游戏代码。 |
| `currency` | `string` | 货币代码。 |

## 签名规则

```text
md5(operator_code + request_time + "pushbetdata" + secret_key)
```

示例：

```text
md5(ABCD + 1698219740 + pushbetdata + XXXX)
```

验签工具：

```text
https://testcase.gscplusmd.com
```

## 请求示例

```json
{
  "operator_code": "CMUT_V2",
  "wagers": [
    {
      "member_account": "s0350",
      "bet_amount": "10",
      "valid_bet_amount": "10",
      "prize_amount": "10",
      "tip_amount": "0",
      "wager_type": "NORMAL",
      "wager_code": "tZDwLV3ayzBeP4Nvwxhcti",
      "wager_status": "SETTLED",
      "round_id": "95978",
      "channel_code": "gscp",
      "game_type": "POKER",
      "settled_at": 1697439181000,
      "created_at": 1697435181000,
      "payload": {},
      "product_code": "1001",
      "game_code": "1001",
      "currency": "CNY"
    },
    {
      "member_account": "s0351",
      "bet_amount": "100",
      "valid_bet_amount": "100",
      "prize_amount": "10",
      "tip_amount": "0",
      "wager_type": "NORMAL",
      "wager_code": "txDwLV5aazBeP4evwxhcti",
      "wager_status": "SETTLED",
      "round_id": "95785",
      "game_type": "POKER",
      "settled_at": 1697439181000,
      "created_at": 1697438182000,
      "payload": {},
      "product_code": "1001",
      "game_code": "1001",
      "currency": "CNY"
    }
  ],
  "sign": "369af7416deef76a9cc4f019b8559f99",
  "request_time": "1694617425"
}
```

## 响应字段

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `code` | `int` | 无缝钱包代码。 |
| `message` | `string` | 响应消息。 |

## 响应示例

```json
{
  "code": 0,
  "message": ""
}
```

# 3.1 开启游戏 (Launch Game)

用于为指定会员启动游戏，返回游戏 URL 或特定提供商需要展示的 HTML 内容。

## Endpoint

```http
POST {{operator_url}}/api/operators/launch-game
```

## 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operator_code` | `string` | 是 | 运营商唯一标识，作为后台登录用户名。 |
| `member_account` | `string` | 是 | 运营商中会员的唯一标识，限制 50 字符。 |
| `password` | `string` | 是 | 会员在运营商系统中的密码，用于验证身份。 |
| `nickname` | `string` | 否 | 游戏中显示的会员昵称。 |
| `currency` | `string` | 是 | 会员在游戏中使用的货币，需确保提供商支持。 |
| `game_code` | `string` | 可选 | 游戏列表 API 中提供商给定的唯一标识符。若提供商支持直接打开游戏则必填，否则可不填。 |
| `product_code` | `int` | 是 | 产品唯一标识符。 |
| `game_type` | `string` | 是 | 游戏类型。 |
| `language_code` | `string` | 否 | 会员语言代码，默认 `0`。 |
| `ip` | `string` | 是 | 会员 IP 地址。 |
| `platform` | `string` | 是 | 平台类型。枚举：`WEB`、`DESKTOP`、`MOBILE`、`Widget`、`Streaming`。SABA 体育极速投注组件使用 `Widget`，直播链接使用 `Streaming`。 |
| `widget_id` | `string` | 否 | SABA 体育极速投注组件 ID。不带则获取默认卡片组件；测试环境经典组件 ID：`1lx2FADe`。 |
| `is_widget_login` | `bool` | 否 | SABA 体育极速投注组件是否登录。`true` 为登录模式，`false` 或不带为未登录模式。 |
| `event_id` | `string` | 否 | SABA 体育直播场次 ID。 |
| `is_streaming_login` | `bool` | 否 | SABA 体育直播是否登录。`true` 为登录模式，`false` 或不带为未登录模式。 |
| `sign` | `string` | 是 | 请求签名。 |
| `request_time` | `int` | 是 | 请求时间戳，单位秒。 |
| `operator_lobby_url` | `string` | 是 | 客户端站点 URL。 |

## 签名规则

```text
md5(request_time + secret_key + "launchgame" + operator_code)
```

示例：

```text
md5(1694617425XXXXlaunchgameCMUT_V2)
```

验签工具：

```text
https://testcase.gscplusmd.com
```

## 请求示例

```json
{
  "operator_code": "CMUT_V2",
  "member_account": "s0350",
  "password": "e10adc3949ba59abbe56e057f20f883e",
  "nickname": "test123",
  "currency": "IDR",
  "game_code": null,
  "product_code": 1001,
  "game_type": "Slot",
  "language_code": 0,
  "ip": "127.0.0.1",
  "platform": "WEB",
  "sign": "977e0ad6dd5c9f953a5b7681d2fa9fb8",
  "request_time": 1694617425,
  "operator_lobby_url": "https://URL"
}
```

## 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | `int` | 运营商代码。 |
| `message` | `string` | 错误消息。 |
| `url` | `string` | 用于启动游戏或提供商游戏大厅的网址。 |
| `content` | `string` | 用于显示特定提供商游戏的 HTML 内容。 |

## 响应示例

```json
{
  "code": 200,
  "message": "",
  "url": "https://dev-test.spribe.io/games/launch/aviator?currency=USD&lang=EN&user=test9&operator=efinity&token=NCVKnX2cTPDDfLUfQ7UtXB"
}
```

# 3.2 游戏类型 (Game Type)

`game_type` 用于启动游戏和游戏列表筛选。同步游戏时也会按该字段映射到本地分类。

| 代码 | 描述                  |
| --- |---------------------|
| `SLOT` | Slot                |
| `LIVE_CASINO` | Live Casino         |
| `SPORT_BOOK` | Sport Book          |
| `VIRTUAL_SPORT` | Virtual Sport       |
| `LOTTERY` | Lottery             |
| `QIPAI` | Qipai               |
| `P2P` | P2P                 |
| `FISHING` | Fishing             |
| `COCK_FIGHTING` | Cock Fighting       |
| `BONUS` | Bonus               |
| `SPORT_BOOK` | SPORTBOOK           |
| `POKER` | Poker               |
| `OTHERS` / `OTHER` | Others              |
| `LIVE_CASINO_PREMIUM` | Live Casino Premium |

# 3.4 游戏列表 (Game List)

用于获取运营商与 GSC+ 签约的所有游戏。仅会返回已签定产品下的游戏。

## Endpoint

```http
GET {{operator_url}}/api/operators/provider-games
```

## 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `product_code` | `int` | 是 | 产品唯一标识符。 |
| `operator_code` | `string` | 是 | 运营商唯一标识，作为后台登录用户名。 |
| `game_type` | `string` | 否 | 游戏类型。 |
| `sign` | `string` | 是 | 请求签名。 |
| `request_time` | `int64` | 是 | 请求时间戳。 |
| `offset` | `int` | 否 | 本次检索的起始记录号。 |
| `size` | `int` | 否 | 本次检索的记录条数；不带则返回全部。 |

## 签名规则

```text
md5(request_time + secret_key + "gamelist" + operator_code)
```

示例：

```text
md5(1694617425XXXXgamelistCMUT_V2)
```

验签工具：

```text
https://testcase.gscplusmd.com
```

## 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | `int` | 操作员代码，`0` 表示成功。 |
| `message` | `string` | 错误信息。 |
| `provider_games` | `array` | 游戏列表。 |
| `pagination` | `object` | 分页信息。 |

## 响应示例

```json
{
  "code": 0,
  "message": "",
  "provider_games": [
    {
      "game_code": "aviator",
      "game_name": "Aviator",
      "game_type": "POKER",
      "image_url": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/aviator.png",
      "product_id": 1,
      "product_code": 1138,
      "support_currency": "MXN",
      "status": "ACTIVATED",
      "allow_free_round": true,
      "lang_name": {
        "0": "Aviator",
        "1": "Aviator",
        "12": "Aviator"
      },
      "lang_icon": {
        "0": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/aviator.png",
        "1": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/aviator.png",
        "12": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/aviator.png"
      },
      "created_at": 1738570027673
    },
    {
      "game_code": "dice",
      "game_name": "Dice",
      "game_type": "POKER",
      "image_url": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/dice.png",
      "product_id": 1,
      "product_code": 1138,
      "support_currency": "MXN",
      "status": "ACTIVAT",
      "allow_free_round": true,
      "lang_name": {
        "0": "Dice",
        "1": "Dice",
        "12": "Dice"
      },
      "lang_icon": {
        "0": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/dice.png",
        "1": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/dice.png",
        "12": "https://images.gscplusmd.com/statics/staging/images/games/1/POKER/dice.png"
      },
      "created_at": 1738570027673
    }
  ],
  "pagination": {
    "size": 10,
    "offset": 0,
    "total": "2000"
  }
}
```
