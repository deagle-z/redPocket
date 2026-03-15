import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  channelCode: [{ required: true, message: "通道编码为必填项", trigger: "blur" }],
  channelName: [{ required: true, message: "通道名称为必填项", trigger: "blur" }],
  channelType: [{ required: true, message: "通道类型为必填项", trigger: "change" }],
  providerType: [{ required: true, message: "提供方类型为必填项", trigger: "change" }]
});
