import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  configKey: [{ required: true, message: "配置Key不能为空", trigger: "blur" }],
  configValue: [{ required: true, message: "配置值不能为空", trigger: "blur" }]
});
