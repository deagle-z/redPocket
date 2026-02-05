import { reactive } from "vue";
import type { FormRules } from "element-plus";

/** 自定义表单规则校验 */
export const formRules = reactive(<FormRules>{
  apkPackage: [
    { required: true, message: "APK包名为必填项", trigger: "blur" },
    { max: 255, message: "APK包名长度不能超过255个字符", trigger: "blur" }
  ],
  name: [
    { required: true, message: "名称为必填项", trigger: "blur" },
    { max: 255, message: "名称长度不能超过255个字符", trigger: "blur" }
  ],
  url: [
    { required: true, message: "URL为必填项", trigger: "blur" },
    { max: 2048, message: "URL长度不能超过2048个字符", trigger: "blur" }
  ]
});


