import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  bannerName: [
    { required: true, message: "轮播图名称为必填项", trigger: "blur" }
  ],
  position: [{ required: true, message: "位置为必填项", trigger: "change" }],
  platform: [{ required: true, message: "平台为必填项", trigger: "change" }],
  imageUrl: [{ required: true, message: "图片地址为必填项", trigger: "blur" }],
  jumpType: [
    { required: true, message: "跳转类型为必填项", trigger: "change" }
  ]
});
