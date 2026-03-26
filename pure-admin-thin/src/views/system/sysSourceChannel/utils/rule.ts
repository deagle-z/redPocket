import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  channelName: [
    { required: true, message: "渠道名称为必填项", trigger: "blur" }
  ],
  level: [{ required: true, message: "渠道层级为必填项", trigger: "change" }]
});
