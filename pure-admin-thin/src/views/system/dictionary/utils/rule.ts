import { reactive } from "vue";
import type { FormRules } from "element-plus";

/** 自定义表单规则校验 */
export const itemFormRules  = reactive(<FormRules>{
  dictLabel: [{ required: true, message: "字典项类型为必填项", trigger: "blur" }],
  dictType: [{ required: true, message: "字典项名称为必填项", trigger: "blur" }],
  dictValue: [{ required: true, message: "字典项值为必填项", trigger: "blur" }],
  code: [{ required: true, message: "字典代码为必填项", trigger: "blur" }],
});

export const formRules = reactive(<FormRules>{
  dictName: [{ required: true, message: "字典项类型为必填项", trigger: "blur" }],
  dictType: [{ required: true, message: "字典项名称为必填项", trigger: "blur" }],
});
