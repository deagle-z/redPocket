import JSEncrypt from "jsencrypt/bin/jsencrypt.min";

// 密钥对生成 http://web.chacuo.net/netrsakeypair

const publicKey =
  "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArf1XPQ5gOa5VtRgdEnICG7gzZsCbFu3g13Uct4xrXV5a4tNBxVz8taNpOQ8QmjtaIsAR4DSuNOq6UJWE7ZuPFbfFzbMeI+4S8UwrBciJGcXlZgxNSSAOPbloqXoWFomfWlLAfoDj9MoZSlNf68CUcGyzQErLmYrPYlX5vLupSAEY157NlLyPEFngEIKLCRz/YpSnxwksAT/WmDw+fQKhP7xvTFuknS9aKFeV6Iy+IrY1pdQysvHO4IgGKGpuMK+1fElsyt6F2Dw41vk+qIm+I9CaZUFJbxJl1mNYGESosCqfvjj2OyRKoxxQtElwN7A6IOxpoKzQuomiYpNTsnoGBQIDAQAB";

export function encrypt(txt) {
  const encryptor = new JSEncrypt();
  encryptor.setPublicKey(publicKey);
  return encryptor.encrypt(txt);
}
