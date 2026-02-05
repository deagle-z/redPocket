import { reactive } from "vue";
import type { FormRules } from "element-plus";

/** 自定义表单规则校验 */
export const formRules = reactive(<FormRules>{
  tenantId: [{ required: true, message: "租户ID为必填项", trigger: "blur" }],
  username: [
    { required: true, message: "账号为必填项", trigger: "blur" },
    { max: 64, message: "账号长度不能超过64个字符", trigger: "blur" }
  ],
  passwordHash: [
    { required: true, message: "密码哈希为必填项", trigger: "blur" },
    { max: 255, message: "密码哈希长度不能超过255个字符", trigger: "blur" }
  ],
  passwordAlgo: [
    { required: true, message: "密码算法为必填项", trigger: "blur" },
    { max: 32, message: "密码算法长度不能超过32个字符", trigger: "blur" }
  ],
  email: [{ max: 128, message: "邮箱长度不能超过128个字符", trigger: "blur" }],
  mobile: [
    { max: 32, message: "手机号长度不能超过32个字符", trigger: "blur" }
  ],
  roleCode: [
    { required: true, message: "角色编码为必填项", trigger: "change" },
    { max: 64, message: "角色编码长度不能超过64个字符", trigger: "blur" }
  ],
  remark: [
    { max: 255, message: "备注长度不能超过255个字符", trigger: "blur" }
  ],
  twofaSecret: [
    { max: 128, message: "2FA密钥长度不能超过128个字符", trigger: "blur" }
  ]
});
