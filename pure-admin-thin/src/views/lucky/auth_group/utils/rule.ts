import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  groupId: [
    {
      required: true,
      message: "群组ID不能为空",
      trigger: "blur"
    },
    {
      type: "number",
      message: "群组ID必须为数字",
      trigger: "blur"
    }
  ],
  groupName: [
    {
      required: true,
      message: "群组名称不能为空",
      trigger: "blur"
    }
  ],
  status: [
    {
      required: true,
      message: "状态不能为空",
      trigger: "change"
    }
  ],
  loseRate: [
    {
      required: true,
      message: "中雷倍数不能为空",
      trigger: "blur"
    },
    {
      type: "number",
      message: "中雷倍数必须为数字",
      trigger: "blur"
    }
  ],
  numConfig: [
    {
      required: true,
      message: "数量配置不能为空",
      trigger: "blur"
    }
  ],
  sendCommission: [
    {
      required: true,
      message: "发包抽成不能为空",
      trigger: "blur"
    },
    {
      type: "number",
      message: "发包抽成必须为数字",
      trigger: "blur"
    }
  ],
  grabbingCommission: [
    {
      required: true,
      message: "抢包抽成不能为空",
      trigger: "blur"
    },
    {
      type: "number",
      message: "抢包抽成必须为数字",
      trigger: "blur"
    }
  ]
});
