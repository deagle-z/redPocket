代收参数
name姓名Stringrequired
请输入 姓名
案例:zhang san(必须是字母),如没有可以固定上传
merNo商户号Stringrequired
请输入 商户号
商户编号
merOrderNo商户订单号Stringrequired
请输入 商户订单号
商户必须保证商户单号唯一,我方不保证商户单号唯一。
email邮箱Stringrequired
请输入 邮箱
案例:test@gmail.com(需要符合邮箱格式),如没有可以固定上传
phone手机号Stringrequired
请输入 手机号
纯数字,不需要添加区号
orderAmount订单金额Stringrequired
请输入 订单金额
金额,两位小数
currency币种编码Stringrequired
IDR
最上方有编码信息
busiCode支付类型编码Stringrequired
请输入 支付类型编码
按照表格编码填写
pageUrl支付成功跳转地址Stringrequired
请输入 支付成功跳转地址
支付成功,页面跳转地址。
notifyUrl通知地址Stringrequired
请输入 通知地址
支付成功后，平台主动通知商家系统，商家系统必须指定接收通知的地址。
timestamp时间戳Stringrequired
当前UTC 13位时间戳,5分钟内有效
sign数字签名Stringrequired


代付参数
accName姓名Stringrequired
请输入 姓名
姓名
accNo卡号Stringrequired
请输入 卡号
卡号
bankCode银行编码Stringrequired
请输入 银行编码
银行编码,左侧目录有对应国家银行编码
busiCode支付业务编码Stringrequired
请输入 支付业务编码
按照表格编码填写
currency币种Stringrequired
IDR
币种
email邮箱Stringrequired
请输入 邮箱
邮箱
merNo商户号Stringrequired
请输入 商户号
商户号
merOrderNo商户订单号Stringrequired
请输入 商户订单号
商户必须保证商户单号唯一,我方不保证商户单号唯一。
notifyUrl回调地址Stringrequired
请输入 回调地址
回调地址
orderAmount订单金额Stringrequired
请输入 订单金额
订单金额
phone手机号Stringrequired
请输入 手机号
纯数字，不需要添加区号
timestamp时间戳Stringrequired
当前UTC 13位时间戳,5分钟内有效
sign签名Stringrequired
