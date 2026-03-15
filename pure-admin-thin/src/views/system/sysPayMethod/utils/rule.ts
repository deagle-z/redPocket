import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  methodCode: [{ required: true, message: "方式编码为必填项", trigger: "blur" }],
  methodName: [{ required: true, message: "方式名称为必填项", trigger: "blur" }]
});
