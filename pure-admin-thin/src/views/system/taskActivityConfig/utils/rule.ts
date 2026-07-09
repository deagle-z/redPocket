import { reactive } from "vue";
import type { FormRules } from "element-plus";

const positiveNumberValidator = (_rule, value, callback) => {
  if (value === null || value === undefined || Number(value) <= 0) {
    callback(new Error("必须大于0"));
    return;
  }
  callback();
};

export const formRules = reactive<FormRules>({
  configTitle: [
    { required: true, message: "任务标题不能为空", trigger: "blur" }
  ],
  levelCode: [{ required: true, message: "请选择任务等级", trigger: "change" }],
  requiredInviteCount: [
    { required: true, validator: positiveNumberValidator, trigger: "blur" }
  ],
  requiredRechargeAmount: [
    { required: true, validator: positiveNumberValidator, trigger: "blur" }
  ],
  durationMinutes: [
    { required: true, validator: positiveNumberValidator, trigger: "blur" }
  ],
  rewardAmount: [
    { required: true, validator: positiveNumberValidator, trigger: "blur" }
  ]
});
