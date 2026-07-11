# Hopopay 接口文档

## 1. 接口类型

- 请求方式：统一使用 `POST` 提交
- 请求头：`Content-Type: application/json; charset=utf-8`
- 数据格式：请求参数、异步回调返回数据均为 JSON 结构
- 字符编码：全局统一 UTF-8

## 2. 签名规则

商户后台分配两组凭证：

- `merchant_id`：商户号
- `secret`：签名密钥，需妥善保管，禁止泄露

### 通用规则

- 接口请求签名、服务端回调验签使用同一套加密算法。
- 唯一区别：回调验签时 `signature` 参数不参与签名拼接。

### 签名完整步骤

1. 参数排序：将所有待签参数按照参数名正向升序排序。
2. 循环拼接原始待签串 `SIGNSTRING`。
3. 遍历每组参数键值 `k-v`：
   - 若参数值 `v` 为数组类型：先进行 JSON 编码，再执行 Base64 编码。
   - 普通字符串 / 数字类型：直接 Base64 编码。
4. 拼接格式：`base64_encode(k)base64_encode(v)&`，多组参数连续拼接。
5. 拼接密钥生成 `SIGNSTRING2`：在 `SIGNSTRING` 末尾直接拼接商户密钥 `secret`。
6. MD5 生成签名：对 `SIGNSTRING2` 进行 MD5 哈希计算，即为最终 `signature`。

## 3. 代收接口

请求网关：`https://{网关地址}/api/payment/payin`

### 请求参数

| 参数名 | 必传 | 类型 | 说明 |
| --- | --- | --- | --- |
| `merchant_id` | 是 | 字符串 | 商户号 |
| `app_user_id` | 是 | 字符串 | 商户用户 ID，最大不超过 32 位 |
| `app_order_id` | 是 | 字符串 | 商户订单号，最大不超过 32 位 |
| `amount` | 是 | number | 订单金额，精确到小数点后两位 |
| `timestamp` | 是 | number | 时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `nonce` | 是 | 字符串 | 随机字符串，最大 16 位 |
| `ip` | 是 | 字符串 | 用户客户端真实 IP |
| `extra_params` | 是 | 字符串 | JSON 字符串 |
| `extra_params.pay_type` | 是 | 字符串 | 代收方式：`cashapp` / `paypal` / `applepay` / `googlepay` |
| `extra_params.platform` | 是 | 字符串 | 设备：`android` / `ios` |
| `callback_url` | 是 | 字符串 | 支付结果接收回调的地址 |
| `return_url` | 是 | 字符串 | 界面支付成功后返回的地址 |
| `signature` | 是 | 字符串 | 签名字符串，不参与签名 |

### 请求 Body 示例

```json
{
  "merchant_id": "t7YDHo3sdsaashskcP",
  "app_user_id": "9888998",
  "app_order_id": "20260301212312231",
  "amount": 20.0,
  "timestamp": 1764744945,
  "nonce": "376299859b5fb4ed",
  "ip": "172.65.2.4",
  "extra_params": "{\"pay_type\":\"paypal\",\"platform\":\"android\"}",
  "callback_url": "https://paynotify.aa.notify.com/paynotify",
  "return_url": "https://pay.url.com",
  "signature": "897sdfasdaf8ds0asd9fas0da"
}
```

### 同步成功响应示例

```jsonc
{
  "info": "request ok!",
  "status": 1, // 1 表示响应成功，其他表示失败
  "data": {
    "step": "webview", // 支付方式
    "order_no": "20251203-073443-440654-228614140",
    "app_order_id": "202603010001231",
    "status": 3, // 3 表示创建订单成功，2 或其他表示失败
    "url": "https://paytest.paypgae.com", // 支付跳转地址
    "monitor": {
      "sub_string": ["callback", "redirect"]
    },
    "qrcode": "" // 支付二维码，可能没有，用上面 url 地址
  },
  "request_id": "1764747283-643b7d40-49c9-4611-969c-075341240552"
}
```

### 同步失败响应示例

```json
{
  "info": "request failed!",
  "status": 0,
  "data": null,
  "error": {
    "error_code": 9001,
    "error_msg": "Inner system error!"
  },
  "request_id": "1764747044-89a25822-beae-465d-bc96-c55f2bb9a252"
}
```

## 4. 代收异步回调通知

使用 `Content-Type: application/json; charset=utf-8` 接收报文，商户需要返回 `ok` 或 `OK`。如果没有接收到 `ok` 或 `OK`，将继续通知，一共通知 5 次。

### 请求参数

| 参数名 | 必传 | 类型 | 说明 |
| --- | --- | --- | --- |
| `status` | 是 | number | `0` 订单创建，`1` 支付中，`2` 失败，`3` 成功，`4` 退款中，`5` 退款，`9` 取消 |
| `order_no` | 是 | 字符串 | 平台订单号 |
| `app_order_id` | 是 | 字符串 | 商户订单号 |
| `amount` | 是 | number | 订单金额 |
| `fee` | 是 | number | 订单手续费 |
| `net_amount` | 是 | number | 订单实际商户到账金额 |
| `pay_amount` | 是 | number | 用户实际支付金额 |
| `pay_at` | 是 | number | 支付成功时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `time` | 是 | number | 时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `type` | 是 | 字符串 | 默认 `order_callback`，类型 |
| `signature` | 是 | 字符串 | 签名字符串，不参与签名 |

### 代收异步响应成功示例

```json
{
  "status": 3,
  "order_no": "20251210-062556-770818-850175937",
  "amount": 10,
  "net_amount": 9.62,
  "pay_amount": 10,
  "fee": 0.28,
  "app_order_id": "313",
  "pay_at": 1765358762,
  "message": "success",
  "time": 1765358854,
  "type": "payout_order_callback",
  "signature": "543e8fcb62xxxdfe2d28149d834c2"
}
```

### 代收异步响应失败示例

```json
{
  "status": 2,
  "order_no": "20251210-055526-798913-392745806",
  "amount": 10,
  "net_amount": 0,
  "fee": 0,
  "app_order_id": "202603010001231",
  "pay_at": null,
  "message": "fail",
  "time": 1765356930,
  "type": "payout_order_callback",
  "signature": "f977faedb5d9sdf4c6470eeb0ae5"
}
```

## 5. 代付接口

请求网关：`https://{网关地址}/api/payment/payout`

### 请求参数

| 参数名 | 必传 | 类型 | 说明 |
| --- | --- | --- | --- |
| `merchant_id` | 是 | 字符串 | 商户号 |
| `app_user_id` | 是 | 字符串 | 商户用户 ID，最大不超过 32 位 |
| `app_order_id` | 是 | 字符串 | 商户订单号，最大不超过 32 位 |
| `amount` | 是 | number | 订单金额，精确到小数点后两位 |
| `timestamp` | 是 | number | 时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `nonce` | 是 | 字符串 | 随机字符串，最大 16 位 |
| `ip` | 是 | 字符串 | 用户客户端真实 IP |
| `bank_code` | 是 | 字符串 | 代付编码：传 `cashapp` 或者 `paypal` |
| `bank_number` | 是 | 字符串 | 收款人账号。Cash App 标签账号，例如 `$abcd1234`；PayPal 邮箱账号，例如 `1111@gmail.com` |
| `bank_account_name` | 是 | 字符串 | 收款人真实姓名 |
| `extra_params` | 是 | 字符串 | JSON 字符串 |
| `extra_params.platform` | 是 | 字符串 | 设备：`android` / `ios` |
| `callback_url` | 是 | 字符串 | 支付结果接收回调的地址 |
| `signature` | 是 | 字符串 | 签名字符串，不参与签名 |

### 请求 Body 示例

```json
{
  "merchant_id": "t7YDHo3safdashskcP",
  "app_user_id": "123",
  "app_order_id": "123",
  "amount": 15.0,
  "timestamp": 1764744945,
  "nonce": "376299859b5fb4ed",
  "bank_account_name": "Rogerio fonte de azeredo",
  "bank_code": "cashapp",
  "bank_number": "03445699784",
  "extra_params": "{\"platform\":\"android\"}",
  "callback_url": "https://api.servernotify.com/api/paynotify",
  "signature": "94ac5ae092a9a7ff8d82db063d611f0d"
}
```

### 同步成功响应示例

```jsonc
{
  "info": "request ok!",
  "status": 1, // 1 表示响应成功，其他表示失败
  "data": {
    "step": "complete",
    "order_no": "20251203-073443-440654-228614140",
    "app_order_id": "123",
    "status": 4 // 4 表示创建订单成功，3 表示打款成功，2 或其他表示失败
  },
  "request_id": "1764747283-643b7d40-49c9-4611-969c-075341240552"
}
```

### 同步失败响应示例

```json
{
  "info": "request failed!",
  "status": 0,
  "data": null,
  "error": {
    "error_code": 9001,
    "error_msg": "Inner system error!"
  },
  "request_id": "1764747044-89a25822-beae-465d-bc96-c55f2bb9a252"
}
```

## 6. 代付异步回调通知

使用 `Content-Type: application/json; charset=utf-8` 接收报文，商户需要返回 `ok` 或 `OK`。如果没有接收到 `ok` 或 `OK`，将继续通知，一共通知 5 次。

### 请求参数

| 参数名 | 必传 | 类型 | 说明 |
| --- | --- | --- | --- |
| `status` | 是 | 字符串 | `0` 订单创建，`1` 支付中，`2` 失败，`3` 成功，`4` 打款中，`5` 拒绝，`6` 退款，`9` 取消 |
| `order_no` | 是 | 字符串 | 平台订单号 |
| `app_order_id` | 是 | 字符串 | 商户订单号 |
| `amount` | 是 | number | 订单金额 |
| `fee` | 是 | number | 订单手续费 |
| `net_amount` | 是 | number | 订单实际扣除金额 |
| `pay_at` | 是 | number | 支付成功时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `time` | 是 | number | 时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `type` | 是 | 字符串 | 默认 `payout_order_callback`，类型 |
| `signature` | 是 | 字符串 | 签名字符串，不参与签名 |

### 代付异步响应成功示例

```json
{
  "status": 3,
  "order_no": "20251210-062556-770818-850175937",
  "amount": 9.99,
  "net_amount": 9.98,
  "fee": 0.01,
  "app_order_id": "131312311342",
  "pay_at": 1765358762,
  "message": "success",
  "time": 1765358854,
  "type": "payout_order_callback",
  "signature": "543e8fcb62xxxdfe2d28149d834c2"
}
```

### 代付异步响应失败示例

```json
{
  "status": 2,
  "order_no": "20251210-055526-798913-392745806",
  "amount": 9.99,
  "net_amount": 0,
  "fee": 0,
  "app_order_id": "309",
  "pay_at": null,
  "message": "fail",
  "time": 1765356930,
  "type": "payout_order_callback",
  "signature": "f977faedb5d9sdf4c6470eeb0ae5"
}
```

## 7. 代收订单查询接口

请求网关：`https://{网关地址}/api/payin/order/query`

### 请求参数

| 参数名 | 必传 | 类型 | 说明 |
| --- | --- | --- | --- |
| `merchant_id` | 是 | 字符串 | 商户号 |
| `app_order_id` | 是 | 字符串 | 商户订单号，最大不超过 32 位 |
| `timestamp` | 是 | number | 时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `nonce` | 是 | 字符串 | 随机字符串，最大 16 位 |
| `order_no` | 是 | 字符串 | 代收接口返回的 `order_no` |
| `signature` | 是 | 字符串 | 签名字符串，不参与签名 |

### 请求 Body 示例

```json
{
  "merchant_id": "t7YDHo3YlJphskcP",
  "app_order_id": "Bx15sn0002141",
  "order_no": "20251203-073045-071690-039613366",
  "timestamp": 1765448478,
  "nonce": "bb4d1752-e3d7-4f03-ba44-ef70e8dadc97",
  "signature": "fa916cc2a96c3193109fa6cea41dfbd7"
}
```

### 同步响应示例

```jsonc
{
  "info": "request ok!",
  "status": 1, // 1 成功，其他失败
  "data": {
    "status": 0, // 0 订单创建，1 支付中，2 失败，3 成功，4 退款中，5 退款，9 取消
    "order_no": "20251203-073045-071690-039613366",
    "amount": 15, // 订单金额
    "pay_amount": 15, // 用户实际支付金额
    "net_amount": 0, // 订单商户到账金额
    "fee": 0, // 订单手续费
    "app_order_id": "Bx15sn0002141",
    "create_at": 1764757845,
    "pay_at": null,
    "time": 1765448479
  },
  "request_id": "1765448479-2413c01b-497b-4725-b9ee-edb122f77c5b"
}
```

## 8. 代付订单查询接口

请求网关：`https://{网关地址}/api/payout/order/query`

### 请求参数

| 参数名 | 必传 | 类型 | 说明 |
| --- | --- | --- | --- |
| `merchant_id` | 是 | 字符串 | 商户号 |
| `app_order_id` | 是 | 字符串 | 商户订单号，最大不超过 32 位 |
| `timestamp` | 是 | number | 时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `nonce` | 是 | 字符串 | 随机字符串，最大 16 位 |
| `order_no` | 是 | 字符串 | 代付接口返回的 `order_no` |
| `signature` | 是 | 字符串 | 签名字符串，不参与签名 |

### 请求 Body 示例

```json
{
  "merchant_id": "t7YDHo3YlJphskcP",
  "app_order_id": "Bx15sn0002141",
  "order_no": "20251203-073045-071690-039613366",
  "timestamp": 1765448478,
  "nonce": "bb4d1752-e3d7-4f03-ba44-ef70e8dadc97",
  "signature": "fa916cc2a96c3193109fa6cea41dfbd7"
}
```

### 同步响应示例

```jsonc
{
  "info": "request ok!",
  "status": 1, // 1 成功，其他失败
  "data": {
    "status": 0, // 0 订单创建，1 支付中，2 失败，3 成功，4 打款中，5 拒绝，6 退款，9 取消
    "order_no": "20251203-073045-071690-039613366",
    "amount": 15, // 订单金额
    "net_amount": 0, // 订单商户扣款金额
    "fee": 0, // 订单手续费
    "app_order_id": "Bx15sn0002141",
    "create_at": 1764757845,
    "pay_at": null,
    "time": 1765448479
  },
  "request_id": "1765448479-2413c01b-497b-4725-b9ee-edb122f77c5b"
}
```

## 9. 商户钱包查询接口

请求网关：`https://{网关地址}/api/merchant/wallet/query`

### 请求参数

| 参数名 | 必传 | 类型 | 说明 |
| --- | --- | --- | --- |
| `merchant_id` | 是 | 字符串 | 商户号 |
| `timestamp` | 是 | number | 时间戳，unix timestamp，从 1970 年 1 月 1 日至今的秒数 |
| `nonce` | 是 | 字符串 | 随机字符串，最大 16 位 |
| `signature` | 是 | 字符串 | 签名字符串，不参与签名 |

### 请求 Body 示例

```json
{
  "merchant_id": "t7YDHo3YlJphskcP",
  "timestamp": 1765448480,
  "nonce": "64c68a5c-a8b7-4f4c-a1a8-fdaef4f7e27a",
  "signature": "1343ca44b5d78c6355b78cb4fee0f93a"
}
```

### 同步响应示例

```jsonc
{
  "info": "request ok!",
  "status": 1, // 1 成功，其他失败
  "data": {
    "money": "0.00000000", // 商户余额
    "total_money": "0.00000000", // 累计代收额
    "withdraw_money": "0.00000000", // 正在回款额
    "total_withdraw_money": "0.00000000", // 累计回款额
    "total_payout_money": "200.00000000" // 累计代付额
  },
  "request_id": "1765448480-af6304b5-90fe-42ce-a5d3-2256449541e5"
}
```
