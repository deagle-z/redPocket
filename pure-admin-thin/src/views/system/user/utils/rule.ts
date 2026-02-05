import { reactive } from "vue";
import type { FormRules } from "element-plus";


const validatePasswordStrength = (rule: any, value: string, callback: any) => {
  if (value.length < 6 || value.length > 18) {
    callback(new Error('密码长度必须为6-18位'));
    return;
  }
  const hasDigit = /\d/.test(value);
  const hasLetter = /[a-zA-Z]/.test(value);
  const hasSymbol = /[^a-zA-Z0-9]/.test(value);
  const typeCount = [hasDigit, hasLetter, hasSymbol].filter(Boolean).length;
  if (typeCount < 2) {
    callback(new Error('密码必须包含数字、字母、符号中的至少两种'));
    return;
  }
  callback();
};

/** 自定义表单规则校验 */
export const formRules = reactive(<FormRules>{
  nickName: [{ required: true, message: "用户昵称为必填项", trigger: "blur" }],
  username: [{ required: true, message: "用户名称为必填项", trigger: "blur" }],
  password: [{ required: true, message: "用户密码为必填项", trigger: "blur" }, { validator: validatePasswordStrength, trigger: "blur" }]
});


