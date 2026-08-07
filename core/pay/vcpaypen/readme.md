# VcPay 秘鲁（PE）接口文档

> 抓取自 <https://doc.vcpay.org/zh/pe/>，官方 Last Updated 2026/6/11。
> `$baseUrl` 需联系 VcPay 客服获取，配置在 `core.yaml` 的 `pay.vcpaypen.baseUrl`。
> 金额单位统一为 **分**（正整数）；`code == "200"` 才是成功。

---

## 一、代收

### 1.1 创建代收订单

> **TIP**
> POST : `$baseUrl` + `/pay/save`

#### HTTP 请求头

| 请求头 | 必选 | 值 |
| --- | --- | --- |
| Content-Type | Y | application/json |

#### HTTP 请求体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串，推荐 UUID 或时间戳 |
| trade_type | Y | Y | string[6] | 请参考[代收 TradeType](#代收-tradetype) |
| order_amount | Y | Y | int | 订单金额（**分**），正整数。示例：10.00 * 100 = 1000 |
| out_trade_no | Y | Y | string[8,64] | 商户单号（商户系统必须保证唯一） |
| notify_url | Y | Y | string[12,128] | 回调地址（可使用路径，禁止携带参数） |
| back_url | Y | Y | string[12,128] | 交易完成重定向地址 |
| identity_name | N | Y | string[4,64] | 付款人姓名 |
| identity_type | Y | Y | string[2,12] | 身份证：`DNI`，护照：`PAS`，外国人居住证：`CE`，税号：`RUC` |
| identity | Y | Y | string[8,64] | 付款人身份号码 |
| sign | Y | N | string[32] | 签名 |

身份类型（identity_type）取值：

| identity_type | 说明 | 对应身份号码格式 |
| --- | --- | --- |
| DNI | 身份证 | 通常为 8 位数字 |
| PAS | 护照 | 通常为 7-12 位数字 |
| CE | 外国人居留许可证 | 通常为 8-12 位字母数字 |
| RUC | 税号（纳税人注册号） | 通常为 11 位数字 |

#### HTTP 请求示例

```json
{
  "app_id": "1111111111111111",
  "back_url": "https://return.test.com",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "out_trade_no": "22222222222222222",
  "trade_type": "PEW001",
  "identity_name": "TEST",
  "identity_type": "DNI",
  "identity": "66666666",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

#### HTTP 响应体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| code | Y | N | string[3] | 响应码，`200` 为成功 |
| msg | Y | N | string[0,128] | 响应消息，如果不为空则为错误原因 |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| notify_url | Y | Y | string[12,128] | 下单时提交的 `$notify_url` |
| order_amount | Y | Y | int | 订单金额（分） |
| order_fee | Y | Y | int | 交易手续费（分） |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| pay_url | Y | Y | string[12,128] | 收银台支付地址 |
| trade_no | Y | Y | string[16,32] | 平台单号 |
| trade_state | Y | Y | int | 请参考[订单 TradeState](#订单-tradestate) |
| trade_type | Y | Y | string[6] | 请参考[代收 TradeType](#代收-tradetype) |
| sign | Y | N | string[32] | 签名 |

#### HTTP 响应示例

```json
{
  "code": "200",
  "msg": "SUCCESS",
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "order_fee": 0,
  "out_trade_no": "22222222222222222",
  "pay_url": "https://pay.test.com",
  "trade_no": "8888888888888888",
  "trade_state": 0,
  "trade_type": "PEW001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

> 官方文档此处响应示例误写为 `"trade_type": "PHW001"`，实际应为 `PEW001`。

### 1.2 查询代收订单

> **TIP**
> POST : `$baseUrl` + `/pay/query`
> **请求频率限制 10 秒一次**

#### HTTP 请求头

| 请求头 | 必选 | 值 |
| --- | --- | --- |
| Content-Type | Y | application/json |

#### HTTP 请求体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| sign | Y | N | string[32] | 签名 |

#### HTTP 请求示例

```json
{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "out_trade_no": "22222222222222222",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

#### HTTP 响应体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| code | Y | N | string[3] | 响应码，`200` 为成功 |
| msg | Y | N | string[0,128] | 响应消息，如果不为空则为错误原因 |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| notify_url | Y | Y | string[12,128] | 下单时提交的 `$notify_url` |
| order_amount | Y | Y | int | 订单金额（分） |
| order_fee | Y | Y | int | 交易手续费（分） |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| trade_no | Y | Y | string[16,32] | 平台单号 |
| trade_state | Y | Y | int | 请参考[订单 TradeState](#订单-tradestate) |
| trade_type | Y | Y | string[6] | 请参考[代收 TradeType](#代收-tradetype) |
| sign | Y | N | string[32] | 签名 |

#### HTTP 响应示例

```json
{
  "code": "200",
  "msg": "SUCCESS",
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "order_fee": 0,
  "out_trade_no": "22222222222222222",
  "trade_no": "8888888888888888",
  "trade_state": 1,
  "trade_type": "PEW001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

### 1.3 代收异步回调通知

> **TIP**
> POST : `$notify_url`（回调地址由商户下单时透传）

#### HTTP 请求头

| 请求头 | 必选 | 值 |
| --- | --- | --- |
| Content-Type | Y | application/json |

#### HTTP 请求体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| code | Y | N | string[3] | 响应码，`200` 为成功 |
| msg | Y | N | string[0,128] | 响应消息，如果不为空则为错误原因 |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| notify_url | Y | Y | string[12,128] | 下单时提交的 `$notify_url` |
| order_amount | Y | Y | int | 订单金额（分） |
| order_fee | Y | Y | int | 交易手续费（分） |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| trade_no | Y | Y | string[16,32] | 平台单号 |
| trade_state | Y | Y | int | 请参考[订单 TradeState](#订单-tradestate) |
| trade_type | Y | Y | string[6] | 请参考[代收 TradeType](#代收-tradetype) |
| sign | Y | N | string[32] | 签名 |

#### HTTP 请求示例

```json
{
  "code": "200",
  "msg": "SUCCESS",
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "order_fee": 0,
  "out_trade_no": "22222222222222222",
  "trade_no": "8888888888888888",
  "trade_state": 1,
  "trade_type": "PEW001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

#### 回调通知响应

> **WARNING**
> 商户处理回调后返回 `SUCCESS` 或 `OK`（忽略大小写）。对于其他响应，则认为回调失败，
> 将以 10、30、60、120、360 秒递增时间间隔的方式重试，总共重试 5 次回调。

---

## 二、代付

### 2.1 创建代付订单

> **TIP**
> POST : `$baseUrl` + `/wd/save`

#### HTTP 请求头

| 请求头 | 必选 | 值 |
| --- | --- | --- |
| Content-Type | Y | application/json |

#### HTTP 请求体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串，推荐 UUID 或时间戳 |
| trade_type | Y | Y | string[6] | 请参考[代付 TradeType](#代付-tradetype)。当为银行转账时，固定值：`PEN000` |
| order_amount | Y | Y | int | 订单金额（**分**），正整数。示例：10.00 * 100 = 1000 |
| out_trade_no | Y | Y | string[8,64] | 商户单号（商户系统必须保证唯一） |
| notify_url | Y | Y | string[12,128] | 回调地址（可使用路径，禁止携带参数） |
| bank_code | Y | Y | string[6] | 银行编码，参考[代付 BankCode](#代付-bankcode) |
| bank_owner | Y | Y | string[2,64] | 受益人名字 |
| bank_account | Y | Y | string[8,64] | 受益人银行账号（20 位银行 CCI 账号）。当为电子钱包转账时，请填写开户的手机号码（9 开头 9 位数字） |
| bank_account_type | Y | Y | string[2,12] | 受益人银行账号类型。储蓄账户：`SA`，活期账户：`CA` |
| bank_attach | N | Y | string[0,128] | 银行附加信息，如无说明可不填 |
| identity_type | Y | Y | string[2,12] | 身份证：`DNI`，护照：`PAS`，外国人居住证：`CE`，税号：`RUC` |
| identity | Y | Y | string[8,64] | 受益人身份号码 |
| sign | Y | N | string[32] | 签名 |

#### HTTP 请求示例

```json
{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "out_trade_no": "22222222222222222",
  "trade_type": "PEW001",
  "bank_code": "PEW001",
  "bank_owner": "Test",
  "bank_account": "906666666",
  "bank_account_type": "SA",
  "identity_type": "DNI",
  "identity": "66666666",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

#### HTTP 响应体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| code | Y | N | string[3] | 响应码，`200` 为成功 |
| msg | Y | N | string[0,128] | 响应消息，如果不为空则为错误原因 |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| notify_url | Y | Y | string[12,128] | 下单时提交的 `$notify_url` |
| order_amount | Y | Y | int | 订单金额（分） |
| order_fee | Y | Y | int | 交易手续费（分） |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| trade_no | Y | Y | string[16,32] | 平台单号 |
| trade_state | Y | Y | int | 请参考[订单 TradeState](#订单-tradestate) |
| trade_type | Y | Y | string[6] | 请参考[代付 TradeType](#代付-tradetype) |
| sign | Y | N | string[32] | 签名 |

#### HTTP 响应示例

```json
{
  "code": "200",
  "msg": "SUCCESS",
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "order_fee": 0,
  "out_trade_no": "22222222222222222",
  "trade_no": "8888888888888888",
  "trade_state": 0,
  "trade_type": "PEW001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

> **WARNING**
> 请求代付超时或未响应上述通用状态，不代表请求失败，可以通过查询接口判断订单是否创建。

### 2.2 查询代付订单

> **TIP**
> POST : `$baseUrl` + `/wd/query`

#### HTTP 请求头

| 请求头 | 必选 | 值 |
| --- | --- | --- |
| Content-Type | Y | application/json |

#### HTTP 请求体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| sign | Y | N | string[32] | 签名 |

#### HTTP 请求示例

```json
{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "out_trade_no": "22222222222222222",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

#### HTTP 响应体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| code | Y | N | string[3] | 响应码，`200` 为成功 |
| msg | Y | N | string[0,128] | 响应消息，如果不为空则为错误原因 |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| notify_url | Y | Y | string[12,128] | 下单时提交的 `$notify_url` |
| order_amount | Y | Y | int | 订单金额（分） |
| order_fee | Y | Y | int | 交易手续费（分） |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| trade_no | Y | Y | string[16,32] | 平台单号 |
| trade_state | Y | Y | int | 请参考[订单 TradeState](#订单-tradestate) |
| trade_type | Y | Y | string[6] | 请参考[代付 TradeType](#代付-tradetype) |
| sign | Y | N | string[32] | 签名 |

#### HTTP 响应示例

```json
{
  "code": "200",
  "msg": "SUCCESS",
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "order_fee": 0,
  "out_trade_no": "22222222222222222",
  "trade_no": "8888888888888888",
  "trade_state": 1,
  "trade_type": "PEW001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

### 2.3 代付异步回调通知

> **TIP**
> POST : `$notify_url`（回调地址由商户下单时透传）

#### HTTP 请求头

| 请求头 | 必选 | 值 |
| --- | --- | --- |
| Content-Type | Y | application/json |

#### HTTP 请求体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| code | Y | N | string[3] | 响应码，`200` 为成功 |
| msg | Y | N | string[0,128] | 响应消息，如果不为空则为错误原因 |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| notify_url | Y | Y | string[12,128] | 下单时提交的 `$notify_url` |
| order_amount | Y | Y | int | 订单金额（分） |
| order_fee | Y | Y | int | 交易手续费（分） |
| out_trade_no | Y | Y | string[8,64] | 商户单号 |
| trade_no | Y | Y | string[16,32] | 平台单号 |
| trade_state | Y | Y | int | 请参考[订单 TradeState](#订单-tradestate) |
| trade_type | Y | Y | string[6] | 请参考[代付 TradeType](#代付-tradetype) |
| sign | Y | N | string[32] | 签名 |

#### HTTP 请求示例

```json
{
  "code": "200",
  "msg": "SUCCESS",
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "order_fee": 0,
  "out_trade_no": "22222222222222222",
  "trade_no": "8888888888888888",
  "trade_state": 1,
  "trade_type": "PEW001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

#### 回调通知响应

> **WARNING**
> 商户处理回调后返回 `SUCCESS` 或 `OK`（忽略大小写）。对于其他响应，则认为回调失败，
> 将以 10、30、60、120、360 秒递增时间间隔的方式重试，总共重试 5 次回调。

---

## 三、查询商户余额

> **TIP**
> POST : `$baseUrl` + `/pay/balance`
> **请求频率限制 10 秒一次**

#### HTTP 请求头

| 请求头 | 必选 | 值 |
| --- | --- | --- |
| Content-Type | Y | application/json |

#### HTTP 请求体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| sign | Y | N | string[32] | 签名 |

#### HTTP 请求示例

```json
{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

#### HTTP 响应体

| 参数名 | 必选 | 签名 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| code | Y | N | string[3] | 响应码，`200` 为成功 |
| msg | Y | N | string[0,128] | 响应消息，如果不为空则为错误原因 |
| app_id | Y | Y | string[16,32] | APP ID |
| nonce_str | Y | Y | string[8,32] | 随机串 |
| balance | Y | Y | int | 可用余额（分） |
| total_balance | Y | Y | int | 总余额（分） |
| freeze_balance | Y | Y | int | 冻结余额（分） |
| sign | Y | N | string[32] | 签名 |

#### HTTP 响应示例

```json
{
  "code": "200",
  "msg": "SUCCESS",
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "balance": 1000000,
  "total_balance": 1000000,
  "freeze_balance": 0,
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
```

---

## 四、秘鲁附录

### 代收 TradeType

| 值 | 说明 |
| --- | --- |
| PEW001 | Yape |
| PEW002 | Plin |
| PET000 | Bank Transfer |
| PEP001 | PagoEfectivo |

### 代付 TradeType

| 值 | 说明 | 范围 |
| --- | --- | --- |
| PEW001 | Yape | E-Wallet |
| PEW002 | Plin | E-Wallet |
| PEN000 | Bank | E-bank |

### 代付 BankCode

| Bank Code | Bank Name |
| --- | --- |
| PEW001 | Yape |
| PEW002 | Plin |
| PEBA01 | Banco Azteca Perú |
| PEBB01 | BanBif |
| PEBB02 | Banco de la Nación |
| PEBB03 | BBVA Perú |
| PEBB04 | BCP (Banco de Crédito del Perú) |
| PEBB05 | Banco de Comercio |
| PEBC01 | Caja Arequipa |
| PEBC02 | Caja Cusco |
| PEBC03 | Caja Huancayo |
| PEBC04 | Caja Maynas |
| PEBC05 | Caja Metropolitana |
| PEBC06 | Caja Municipal Ica |
| PEBC07 | Caja Piura |
| PEBC08 | Caja Sullana |
| PEBC09 | Caja Tacna |
| PEBC10 | Caja Trujillo |
| PEBC11 | Banco Cencosud |
| PEBC12 | Citibank Perú |
| PEBF01 | Banco Falabella |
| PEBG01 | Banco GNB Perú |
| PEBI01 | Interbank |
| PEBI02 | ICBC Perú |
| PEBM01 | MiBanco |
| PEBP01 | Banco Pichincha |
| PEBR01 | Banco Ripley |
| PEBS01 | Santander Perú |
| PEBS02 | Scotiabank Perú |

> 官方文档该表表头写反了（写成 `Bank Name | Bank Code`），此处按实际数据修正。

### 订单 TradeState

| 值 | 说明 | 备注 |
| --- | --- | --- |
| 0 | 等待付款 | - |
| 1 | 支付成功 | 订单最终状态 |
| 2 | 支付失败 | 订单最终状态 |
| 3 | 支付中 | - |

### API 常见错误信息

| 错误码 | 说明 |
| --- | --- |
| 200 | 成功 |
| 500 | 系统错误 |
| 50100 | 交易错误 |
| 50101 | 余额不足 |
| 400 | 参数错误 |
| 40100 | 币种不支持 |
| 40101 | 参数错误：identity |
| 40102 | 参数错误：identity_type |
| 40103 | 参数错误：notify_url |
| 40104 | 参数错误：trade_type |
| 40105 | 参数错误：out_trade_no |
| 40106 | 参数错误：order_amount |
| 40107 | 参数错误：attach |
| 40108 | 参数错误：bank_owner |
| 40109 | 参数错误：bank_account |
| 40110 | 参数错误：nonce_str |
| 40111 | 签名错误 |
| 40112 | 参数错误：back_url |
| 40113 | 参数错误：bank_code |
| 40114 | 银行编码不支持或没开通 |
| 40200 | 对接应用不存在 |
| 40201 | 商户不存在或未启用 |
| 40202 | 通道不支持 |
| 40203 | 通道未配置 |
| 40204 | 通道配置错误 |
| 40300 | 订单不存在 |
| 40301 | 订单已完成 |
| 40302 | 订单已存在 |
| 40401 | 未加 IP 白名单 |
| 40402 | 请求频繁 |
| 40403 | 无操作权限 |
| 40404 | 请求地址不存在 |
| 40405 | 请求不合法 |
