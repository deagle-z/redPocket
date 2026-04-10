import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const formRules = reactive<FormRules>({
  bannerName: [
    { required: true, message: "轮播图名称为必填项", trigger: "blur" }
  ],
  position: [{ required: true, message: "位置为必填项", trigger: "change" }],
  platform: [{ required: true, message: "平台为必填项", trigger: "change" }],
  bannerType: [
    { required: true, message: "Banner类型为必填项", trigger: "change" }
  ],
  jumpType: [
    { required: true, message: "跳转类型为必填项", trigger: "change" }
  ],
  displayType: [
    { required: true, message: "展示类型为必填项", trigger: "change" }
  ],
  openMode: [
    { required: true, message: "打开方式为必填项", trigger: "change" }
  ],
  i18nList: [
    {
      validator: (_, value, callback) => {
        if (!Array.isArray(value) || value.length === 0) {
          callback(new Error("至少需要配置一条多语言内容"));
          return;
        }
        const invalid = value.find(
          item => !item?.languageCode?.trim() || !item?.imageUrl?.trim()
        );
        if (invalid) {
          callback(new Error("多语言项必须填写语言编码和主图"));
          return;
        }
        callback();
      },
      trigger: "change"
    }
  ],
  countryList: [
    {
      validator: (_, value, callback) => {
        if (!Array.isArray(value) || value.length === 0) {
          callback(new Error("至少需要配置一个投放国家"));
          return;
        }
        const invalid = value.find(item => !item?.countryCode?.trim());
        if (invalid) {
          callback(new Error("投放国家不能为空"));
          return;
        }
        callback();
      },
      trigger: "change"
    }
  ]
});
