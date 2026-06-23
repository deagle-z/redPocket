墨西哥代收
#创建代收订单
#HTTP请求地址
TIP

POST : $baseUrl + /pay/save

#HTTP请求头
请求头	必选	值
Content-Type	Y	application/json
#HTTP请求体
参数名	必选	签名	类型	说明
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
推荐 UUID 或 时间戳.
trade_type	Y	Y	string[6]	请参考 代收TradeType
order_amount	Y	Y	int	订单金额(分), 金额为正整数.
示例: 10.00 * 100 = 1000
out_trade_no	Y	Y	string[8,64]	商户单号 (商户系统必须保证唯一).
notify_url	Y	Y	string[12,128]	回调地址(可使用路径，禁止携带参数)
back_url	Y	Y	string[12,128]	交易完成重定向地址
identity_name	N	Y	string[4,64]	付款人名字
identity_type	N	Y	string[2,12]	固定值: PHONE
identity	N	Y	string[8,64]	付款人手机号码
sign	Y	N	string[32]	签名
#HTTP请求示例
{
  "app_id": "1111111111111111",
  "back_url": "https://return.test.com",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "out_trade_no": "22222222222222222",
  "trade_type": "MX0001",
  "identity_name": "TEST",
  "identity_type": "PHONE",
  "identity": "9876543210",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
#HTTP响应体
参数名	必选	签名	类型	说明
code	Y	N	string[3]	响应码. 200 为 成功
msg	Y	N	string[0,128]	响应消息. 如果不为空，则为错误原因.
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
notify_url	Y	Y	string[12,128]	下单时提交的 $notify_url
order_amount	Y	Y	int	订单金额(分)
order_fee	Y	Y	int	交易手续费(分)
out_trade_no	Y	Y	string[8,64]	商户单号
pay_url	Y	Y	string[12,128]	收银台支付地址
trade_no	Y	Y	string[16,32]	平台单号
trade_state	Y	Y	int	请参考 订单TradeState
trade_type	Y	Y	string[6]	请参考 代收TradeType
sign	Y	N	string[32]	签名
#HTTP响应示例
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
  "trade_type": "MX0001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}

异步回调通知
#HTTP请求地址
TIP

POST : $notify_url (回调地址由商户下单时透传)

#HTTP请求头
请求头	必选	值
Content-Type	Y	application/json
#HTTP请求体
参数名	必选	签名	类型	说明
code	Y	N	string[3]	响应码. 200 为 成功
msg	Y	N	string[0,128]	响应消息. 如果不为空，则为错误原因.
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
notify_url	Y	Y	string[12,128]	下单时提交的 $notify_url
order_amount	Y	Y	int	订单金额(分)
order_fee	Y	Y	int	交易手续费(分)
out_trade_no	Y	Y	string[8,64]	商户单号
trade_no	Y	Y	string[16,32]	平台单号
trade_state	Y	Y	int	请参考 订单TradeState
trade_type	Y	Y	string[6]	请参考 代收TradeType
sign	Y	N	string[32]	签名
#HTTP请求示例
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
  "trade_type": "MX0001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
#回调通知响应
WARNING

商户处理回调后返回 SUCCESS 或 OK(忽略大小写). 对于其他响应，则认为回调失败，将以10、30、60、120、360秒递增时间间隔的方式重试，总共重试5次回调。

墨西哥代付
#创建代付订单
#HTTP请求地址
TIP

POST : $baseUrl + /wd/save

#HTTP请求头
请求头	必选	值
Content-Type	Y	application/json
#HTTP请求体
参数名	必选	签名	类型	说明
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
推荐 UUID 或 时间戳.
trade_type	Y	Y	string[6]	请参考 代付TradeType.
CLABE转账: MX0001,银行转账: MX0000
order_amount	Y	Y	int	订单金额(分), 金额为正整数.
示例: 10.00 * 100 = 1000
out_trade_no	Y	Y	string[8,64]	商户单号 (商户系统必须保证唯一).
notify_url	Y	Y	string[12,128]	回调地址(可使用路径，禁止携带参数)
bank_code	Y	Y	string[6]	参考 代付BankCode.
bank_owner	Y	Y	string[2,64]	受益人名字
bank_account	Y	Y	string[8,64]	受益人银行账号. CLABE账号18位，银行账号16位
bank_attach	N	Y	string[0,128]	银行附加信息，如无说明可不填
identity_type	Y	Y	string[2,12]	固定值: PHONE
identity	Y	Y	string[8,64]	受益人开户手机号码
sign	Y	N	string[32]	签名
#HTTP请求示例
SPEI-CLABE转账(trade_type:MX0001)

{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "out_trade_no": "22222222222222222",
  "trade_type": "MX0001",
  "bank_code": "MXA001",
  "bank_owner": "Test",
  "bank_account": "012180001234567891",
  "identity_type": "PHONE",
  "identity": "5512345678",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
银行转账(trade_type:MX0000)

{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "notify_url": "https://notify.test.com",
  "order_amount": 10000,
  "out_trade_no": "22222222222222222",
  "trade_type": "MX0000",
  "bank_code": "MXA001",
  "bank_owner": "Test",
  "bank_account": "1234567812345678",
  "identity_type": "PHONE",
  "identity": "5512345678",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
#HTTP响应体
参数名	必选	签名	类型	说明
code	Y	N	string[3]	响应码. 200 为 成功
msg	Y	N	string[0,128]	响应消息. 如果不为空，则为错误原因.
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
notify_url	Y	Y	string[12,128]	下单时提交的 $notify_url
order_amount	Y	Y	int	订单金额(分)
order_fee	Y	Y	int	交易手续费(分)
out_trade_no	Y	Y	string[8,64]	商户单号
trade_no	Y	Y	string[16,32]	平台单号
trade_state	Y	Y	int	请参考 订单TradeState
trade_type	Y	Y	string[6]	请参考 代付TradeType
sign	Y	N	string[32]	签名
#HTTP响应示例
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
  "trade_type": "MX0001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
WARNING

请求代付超时或未响应上述通用状态，不代表请求失败，可以通过查询接口判断订单是否创建

#查询代付订单
#HTTP请求地址
TIP

POST : $baseUrl + /wd/query

#HTTP请求头
请求头	必选	值
Content-Type	Y	application/json
#HTTP请求体
Parameter	Required	Signed	Type	Description
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
out_trade_no	Y	Y	string[8,64]	商户单号
sign	Y	N	string[32]	签名
#HTTP请求示例
{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "out_trade_no": "22222222222222222",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
#HTTP响应体
参数名	必选	签名	类型	说明
code	Y	N	string[3]	响应码. 200 为 成功
msg	Y	N	string[0,128]	响应消息. 如果不为空，则为错误原因.
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
notify_url	Y	Y	string[12,128]	下单时提交的 $notify_url
order_amount	Y	Y	int	订单金额(分)
order_fee	Y	Y	int	交易手续费(分)
out_trade_no	Y	Y	string[8,64]	商户单号
trade_no	Y	Y	string[16,32]	平台单号
trade_state	Y	Y	int	请参考 订单TradeState
trade_type	Y	Y	string[6]	请参考 代付TradeType
sign	Y	N	string[32]	签名
#HTTP响应示例
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
  "trade_type": "MX0001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
#异步回调通知
#HTTP请求地址
TIP

POST : $notify_url (回调地址由商户下单时透传)

#HTTP请求头
请求头	必选	值
Content-Type	Y	application/json
#HTTP请求体
参数名	必选	签名	类型	说明
code	Y	N	string[3]	响应码. 200 为 成功
msg	Y	N	string[0,128]	响应消息. 如果不为空，则为错误原因.
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
notify_url	Y	Y	string[12,128]	下单时提交的 $notify_url
order_amount	Y	Y	int	订单金额(分)
order_fee	Y	Y	int	交易手续费(分)
out_trade_no	Y	Y	string[8,64]	商户单号
trade_no	Y	Y	string[16,32]	平台单号
trade_state	Y	Y	int	请参考 订单TradeState
trade_type	Y	Y	string[6]	请参考 代付TradeType
sign	Y	N	string[32]	签名
#HTTP请求示例
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
  "trade_type": "MX0001",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
#回调通知响应
WARNING

商户处理回调后返回 SUCCESS 或 OK(忽略大小写). 对于其他响应，则认为回调失败，将以10、30、60、120、360秒递增时间间隔的方式重试，总共重试5次回调。

#查询商户余额
#HTTP请求地址
TIP

POST : $baseUrl + /pay/balance

请求频率限制10秒一次

#HTTP请求头
请求头	必选	值
Content-Type	Y	application/json
#HTTP请求体
参数名	必选	签名	类型	说明
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
sign	Y	N	string[32]	签名
#HTTP请求示例
{
  "app_id": "1111111111111111",
  "nonce_str": "6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c6c",
  "sign": "A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2A2"
}
#HTTP响应体
参数名	必选	签名	类型	说明
code	Y	N	string[3]	响应码. 200 为 成功
msg	Y	N	string[0,128]	响应消息. 如果不为空，则为错误原因.
app_id	Y	Y	string[16,32]	APP ID
nonce_str	Y	Y	string[8,32]	随机串
balance	Y	Y	int	可用余额(分)
total_balance	Y	Y	int	总余额(分)
freeze_balance	Y	Y	int	冻结余额(分)
sign	Y	N	string[32]	签名
#HTTP响应示例
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

代收TradeType
值	说明
MX0001	SPE

代付TradeType
值	说明
MX0001	SPEI(Clabe账号)
MX0000	Bank (银行账号)


代付BankCode
Bank Name	Bank Code
ABC CAPITAL	MXA001
ACTINVER	MXA002
AFIRME	MXA003
ALTERNATIVOS	MXA004
ARCUS FI	MXA005
ASP INTEGRA OPC	MXA006
AUTOFIN	MXA007
AZTECA	MXA008
Albo	MXA009
BaBien	MXB001
BAJIO	MXB002
BANAMEX	MXB003
BANCO COVALTO	MXB004
BANCO S3	MXB005
BANCOMEXT	MXB006
BANCOPPEL	MXB007
BANCREA	MXB008
BANJERCITO	MXB009
BANK OF AMERICA	MXB010
BANK OF CHINA	MXB011
BANKAOOL	MXB012
BANOBRAS	MXB013
BANORTE	MXB014
BANREGIO	MXB015
BANSI	MXB016
BARCLAYS	MXB017
BBASE	MXB018
BBVA MEXICO	MXB019
BMONEX	MXB020
CAJA POP MEXICA	MXC001
CAJA TELEFONIST	MXC002
CB INTERCAM	MXC003
CI BOLSA	MXC004
CIBANCO	MXC005
CITI MEXICO	MXC006
CLS	MXC007
COMPARTAMOS	MXC008
CONSUBANCO	MXC009
CREDICAPITAL	MXC010
CREDICLUB	MXC011
CRISTOBAL COLON	MXC012
Cuenca	MXC013
Credit Suisse	MXC014
DONDE	MXD001
FINAMEX	MXF001
FINCOMUN	MXF002
FOMPED	MXF003
FONDEADORA	MXF004
FONDO (FIRA)	MXF005
GBM	MXG001
HIPOTECARIA FED	MXH001
HSBC	MXH002
ICBC	MXI001
INBURSA	MXI002
INDEVAL	MXI003
INMOBILIARIO	MXI004
INTERCAM BANCO	MXI005
INVERCAP	MXI006
INVEX	MXI007
JP MORGAN	MXJ001
KUSPIT	MXK001
LIBERTAD	MXL001
MASARI	MXM001
Mercado Pago W	MXM002
MIFEL	MXM003
MIZUHO BANK	MXM004
MONEXCB	MXM005
MUFG	MXM006
MULTIVA BANCO	MXM007
MexPago	MXM008
NAFIN	MXN001
NU MEXICO	MXN002
NVIO	MXN003
PAGATODO	MXP001
Peibo	MXP002
PROFUTURO	MXP003
SABADELL	MXS001
SANTANDER	MXS002
SCOTIABANK	MXS003
SHINHAN	MXS004
SPIN BY OXXO	MXS005
STP	MXS006
TESORED	MXT001
TRANSFER	MXT002
UNAGRA	MXU001
VALMEX	MXV001
VALUE	MXV002
VE POR MAS	MXV003
VECTOR	MXV004
VOLKSWAGEN	MXV005