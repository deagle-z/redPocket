import { reactive } from "vue";
import type { FormRules } from "element-plus";

function validateCode(_rule: unknown, value: string, callback: (error?: Error) => void) {
  const code = value?.trim() || "";
  if (code !== "" && !/^\d{6}$/.test(code)) {
    callback(new Error("兑换码必须为6位数字，留空则自动生成"));
    return;
  }
  callback();
}

export const formRules = reactive<FormRules>({
  code: [{ validator: validateCode, trigger: "blur" }],
  amount: [
    { required: true, message: "兑换金额不能为空", trigger: "blur" },
    { type: "number", min: 0.01, message: "兑换金额必须大于0", trigger: "blur" }
  ],
  maxRedeemCount: [
    { required: true, message: "可兑换次数不能为空", trigger: "blur" },
    { type: "number", min: 1, message: "可兑换次数必须大于0", trigger: "blur" }
  ],
  status: [{ required: true, message: "请选择状态", trigger: "change" }]
});
