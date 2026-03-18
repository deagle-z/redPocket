import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  levelName: [{ required: true, message: "等级名称为必填项", trigger: "blur" }],
  level: [{ required: true, message: "等级排序为必填项", trigger: "blur" }]
});
