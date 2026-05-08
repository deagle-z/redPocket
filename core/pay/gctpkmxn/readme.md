代付编码(busiCode)
墨西哥代付
207001
请求参数
bankCode银行编码Stringrequired
请输入 银行编码
出款类型为debitcard, phonenum。需要填写银行编码
identityType出款类型Stringrequired
请输入 出款类型
需要填写,分为clabe, debitcard, phonenum，需要让收款人自己选择这三种其中之一，选择对应类型，acc_no则对应该类型出款账号，其中手机号方式不用加区号
accName姓名Stringrequired
请输入 姓名
姓名
accNo卡号Stringrequired
请输入 卡号
卡号
busiCode支付业务编码Stringrequired
请输入 支付业务编码
按照表格编码填写
currency币种Stringrequired
MXN
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
签名
请求数据
copy
{
bankCode:""
identityType:""
accName:""
accNo:""
busiCode:""
currency:"MXN"
email:""
merNo:""
merOrderNo:""
notifyUrl:""
orderAmount:""
phone:""
timestamp:""
sign:""
}
响应体
copy
{}
返回参数
参数名
参数名称
类型
说明
data
承载数据
Object
code=200有数据
code
接口状态码
integer(int32)
code=200或code=500代表请求成功,订单状态以data.status参数为准,code!=200、500则表示请求失败订单不入库（响应异常不能作为失败处理，比如响应超时或者httpCode响应502）
msg
接口状态信息
String
接口状态信息
