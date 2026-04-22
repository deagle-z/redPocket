签名规范
所有接口的签名规范：
将所有参数按照字段名的 ASCII 码(字典序)从小到大排序后使用 QueryString 的格式(即key1=value1&key2=value2…) 拼接成签名串,空参数和sign不参与签名

HmacSHA256加密格式案例(案例只做参考,加密后的值可能不准确)
代付下单接口需要对HmacSHA256加密后的sign值再进行RSA-1024私钥加密,其他接口均只需使用HmacSHA256加密和验签
(代付下单HmacSHA256+rsa,代付通知,代收下单,代收通知,代收查询,代付查询等均是HmacSHA256)
加密前:
bankCode=abc&busiCode=103001&currency=INR&email=tom@gmail.com&merNo=xxxxxxxxx&merOrderNo=1703664046297&name=tom&notifyUrl=https://xxx.xxx.xxx/xxx/xxxxxxxx&orderAmount=1000&pageUrl=https://xxx.xxxx.com&phone=9001941197&timestamp=1703664046000
加密后:
3dc928559f8a3657e759c6fda27178071beed1b444df9f76b4538d6985cd6cb1

代付下单RSA加密格式案例(案例只做参考,加密后的值可能不准确)
1.正式号使用工具类或者在线网址生成RSA密钥对,生成RSA-1024位,密钥格式:PKCS#8,私钥自己保存,公钥上传至商户后台。(测试号使用文档提供的密钥即可)
2.代付加密流程,使用HmacSHA256对签名串加密后得到signA,再对signA进行RSA私钥加密并base64编码后得到sign。
加密前:
accName=tomjrui&accNo=123123156412&bankCode=SCB&busiCode=201001&currency=THB&email=tom@gmail.com&extend=tom@gmail.com&merNo=xxxxxxxxx&merOrderNo=1703663635553&notifyUrl=https://xxx.xxx.com/temp/acqNotify&orderAmount=100&phone=9001941197&province=ICIC1234567&timestamp=1703663635553
signA:
aa5db6a30e8a27607904cf8b75ddd368e26960ee34d149ea6d53c81c0230c70d
对signA进行RSA加密并base64编码后的sign:
gse2VBCsjN73ICHV0GgaeL3vnvp+4QBhuiH5iVcAh4HN0cQ8G0DrYx/udW724f66ShnB4pHYSfhBYsemfsDnzt583VqJ7OLrCqW0Q9WFengFtQO3IGSn//shzl5pzGzpEKF3PKfoBV6HXY5QUBXNB8zJs/wTuQDVDcx69LfIXPM=


支付编码(busiCode)
印尼网银支付	印尼OVO电子钱包	印尼LINKAJA电子钱包	印尼扫码	印尼DANA电子钱包	印尼BNI网银	印尼卡卡
104001	104002	104003	104004	104005	104007	104008