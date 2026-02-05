import { reactive } from "vue";
import type { FormRules } from "element-plus";

/** 自定义表单规则校验 */
export const formRules = reactive(<FormRules>{
  hostName: [{ required: true, message: "域名为必填项", trigger: "blur" }],
  tablePrefix: [
    { required: true, message: "表名前缀为必填项", trigger: "blur" }
  ]
});
