import type { FormRules } from "element-plus";

export const formRules = {
  gameName: [{ required: true, message: "请输入游戏名称", trigger: "blur" }],
  platformCode: [
    { required: true, message: "请输入平台编码", trigger: "blur" }
  ],
  thirdGameId: [
    { required: true, message: "请输入第三方游戏ID", trigger: "blur" }
  ]
} satisfies FormRules;
