import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  fieldKey: [{ required: true, message: "字段Key为必填项", trigger: "blur" }],
  fieldLabel: [{ required: true, message: "字段名称为必填项", trigger: "blur" }],
  fieldType: [{ required: true, message: "字段类型为必填项", trigger: "change" }],
  dataType: [{ required: true, message: "数据类型为必填项", trigger: "change" }]
});
