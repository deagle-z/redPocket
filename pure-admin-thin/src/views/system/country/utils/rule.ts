import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  countryCode: [{ required: true, message: "国家编码为必填项", trigger: "blur" }],
  countryNameCn: [{ required: true, message: "国家中文名为必填项", trigger: "blur" }],
  countryNameEn: [{ required: true, message: "国家英文名为必填项", trigger: "blur" }],
  currencyCode: [{ required: true, message: "币种编码为必填项", trigger: "blur" }]
});
