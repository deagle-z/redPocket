import { reactive } from "vue";
import type { FormRules } from "element-plus";

const domainPattern =
  /^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/;

/** 自定义表单规则校验 */
export const formRules = reactive(<FormRules>{
  tenantCode: [
    { required: true, message: "租户编码为必填项", trigger: "blur" },
    { max: 64, message: "租户编码长度不能超过64个字符", trigger: "blur" }
  ],
  tenantName: [
    { required: true, message: "租户名称为必填项", trigger: "blur" },
    { max: 128, message: "租户名称长度不能超过128个字符", trigger: "blur" }
  ],
  tenantType: [
    { required: true, message: "请选择租户类型", trigger: "change" }
  ],
  status: [{ required: true, message: "请选择状态", trigger: "change" }],
  timezone: [
    { required: true, message: "时区为必填项", trigger: "blur" },
    { max: 64, message: "时区长度不能超过64个字符", trigger: "blur" }
  ],
  locale: [
    { required: true, message: "默认语言为必填项", trigger: "blur" },
    { max: 32, message: "默认语言长度不能超过32个字符", trigger: "blur" }
  ],
  planCode: [
    { max: 64, message: "套餐标识长度不能超过64个字符", trigger: "blur" }
  ],
  bindDomain: [
    { max: 128, message: "绑定域名长度不能超过128个字符", trigger: "blur" },
    {
      validator: (_rule, value, callback) => {
        if (!value || domainPattern.test(String(value))) {
          callback();
          return;
        }
        callback(new Error("请输入合法域名，如 example.com"));
      },
      trigger: "blur"
    }
  ],
  remark: [
    { max: 255, message: "备注长度不能超过255个字符", trigger: "blur" }
  ]
});
