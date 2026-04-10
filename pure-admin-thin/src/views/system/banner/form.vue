<script setup lang="ts">
import { ref } from "vue";
import Segmented from "@/components/ReSegmented";
import { formRules } from "./utils/rule";
import type {
  BannerCountryFormItem,
  BannerI18nFormItem,
  FormProps
} from "./utils/types";
import { getToken, formatToken } from "@/utils/auth";
import { message } from "@/utils/message";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    tenantId: 0,
    bannerName: "",
    bannerCode: "",
    position: "home",
    platform: "all",
    bannerType: "image",
    jumpType: "none",
    displayType: "banner",
    openMode: "current",
    sort: 0,
    status: 1,
    startTime: "",
    endTime: "",
    version: 1,
    remark: "",
    i18nList: [
      {
        id: 0,
        tenantId: 0,
        bannerId: 0,
        languageCode: "zh-CN",
        countryCode: "",
        title: "",
        subTitle: "",
        description: "",
        buttonText: "",
        imageUrl: "",
        thumbUrl: "",
        bgImageUrl: "",
        iconUrl: "",
        videoUrl: "",
        jumpValue: "",
        textColor: "",
        buttonColor: "",
        bgColor: "",
        status: 1,
        remark: ""
      }
    ],
    countryList: [
      {
        id: 0,
        tenantId: 0,
        bannerId: 0,
        countryCode: "",
        status: 1,
        remark: ""
      }
    ]
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

const uploadUrl = `${import.meta.env.VITE_BASE_URL}/api/v1/admin/upload`;

function getUploadHeaders() {
  const token = getToken()?.accessToken;
  return token ? { Authorization: formatToken(token) } : {};
}

function handleBeforeUpload() {
  const token = getToken()?.accessToken;
  if (!token) {
    message("请先登录后再上传", { type: "error" });
    return false;
  }
  return true;
}

function createI18nItem(): BannerI18nFormItem {
  return {
    id: 0,
    tenantId: 0,
    bannerId: 0,
    languageCode: "",
    countryCode: "",
    title: "",
    subTitle: "",
    description: "",
    buttonText: "",
    imageUrl: "",
    thumbUrl: "",
    bgImageUrl: "",
    iconUrl: "",
    videoUrl: "",
    jumpValue: "",
    textColor: "",
    buttonColor: "",
    bgColor: "",
    status: 1,
    remark: ""
  };
}

function createCountryItem(): BannerCountryFormItem {
  return {
    id: 0,
    tenantId: 0,
    bannerId: 0,
    countryCode: "",
    status: 1,
    remark: ""
  };
}

function updateI18nImageField(
  index: number,
  field: keyof Pick<
    BannerI18nFormItem,
    "imageUrl" | "thumbUrl" | "bgImageUrl" | "iconUrl" | "videoUrl"
  >,
  response: any
) {
  const url = response?.data?.url || response?.url;
  if (!url) {
    message("上传失败，未返回URL", { type: "error" });
    return;
  }
  newFormInline.value.i18nList[index][field] = url;
  message("上传成功", { type: "success" });
}

function handleUploadError() {
  message("上传失败", { type: "error" });
}

function addI18nItem() {
  newFormInline.value.i18nList.push(createI18nItem());
}

function removeI18nItem(index: number) {
  if (newFormInline.value.i18nList.length <= 1) {
    message("至少保留一条多语言内容", { type: "warning" });
    return;
  }
  newFormInline.value.i18nList.splice(index, 1);
}

function addCountryItem() {
  newFormInline.value.countryList.push(createCountryItem());
}

function removeCountryItem(index: number) {
  if (newFormInline.value.countryList.length <= 1) {
    message("至少保留一个投放国家", { type: "warning" });
    return;
  }
  newFormInline.value.countryList.splice(index, 1);
}

const statusOptions = [
  { label: "启用", value: false },
  { label: "停用", value: true }
];

const positionOptions = [
  { label: "首页", value: "home" },
  { label: "首页弹窗", value: "popup" },
  { label: "活动页", value: "activity" },
  { label: "会员中心", value: "member_center" }
];

const platformOptions = [
  { label: "全部", value: "all" },
  { label: "Web", value: "web" },
  { label: "App", value: "app" },
  { label: "H5", value: "h5" },
  { label: "PC", value: "pc" }
];

const bannerTypeOptions = [
  { label: "图片", value: "image" },
  { label: "视频", value: "video" },
  { label: "弹窗", value: "popup" }
];

const jumpTypeOptions = [
  { label: "不跳转", value: "none" },
  { label: "外部链接", value: "url" },
  { label: "站内页面", value: "internal" },
  { label: "商品", value: "product" },
  { label: "活动", value: "activity" },
  { label: "分类", value: "category" }
];

const displayTypeOptions = [
  { label: "轮播", value: "banner" },
  { label: "弹窗", value: "popup" },
  { label: "悬浮", value: "float" }
];

const openModeOptions = [
  { label: "当前页", value: "current" },
  { label: "新窗口", value: "new_window" },
  { label: "小程序", value: "mini_app" }
];

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="120px"
  >
    <el-row :gutter="18">
      <el-col :span="12">
        <el-form-item label="Banner名称" prop="bannerName">
          <el-input
            v-model="newFormInline.bannerName"
            clearable
            placeholder="请输入Banner名称"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="Banner编码">
          <el-input
            v-model="newFormInline.bannerCode"
            clearable
            placeholder="请输入Banner编码"
          />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="位置" prop="position">
          <el-select
            v-model="newFormInline.position"
            class="!w-full"
            placeholder="请选择位置"
          >
            <el-option
              v-for="item in positionOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="平台" prop="platform">
          <el-select
            v-model="newFormInline.platform"
            class="!w-full"
            placeholder="请选择平台"
          >
            <el-option
              v-for="item in platformOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="Banner类型" prop="bannerType">
          <el-select
            v-model="newFormInline.bannerType"
            class="!w-full"
            placeholder="请选择Banner类型"
          >
            <el-option
              v-for="item in bannerTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="跳转类型" prop="jumpType">
          <el-select
            v-model="newFormInline.jumpType"
            class="!w-full"
            placeholder="请选择跳转类型"
          >
            <el-option
              v-for="item in jumpTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="展示类型" prop="displayType">
          <el-select
            v-model="newFormInline.displayType"
            class="!w-full"
            placeholder="请选择展示类型"
          >
            <el-option
              v-for="item in displayTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="打开方式" prop="openMode">
          <el-select
            v-model="newFormInline.openMode"
            class="!w-full"
            placeholder="请选择打开方式"
          >
            <el-option
              v-for="item in openModeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="开始时间">
          <el-date-picker
            v-model="newFormInline.startTime"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="请选择开始时间"
            class="!w-full"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="结束时间">
          <el-date-picker
            v-model="newFormInline.endTime"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="请选择结束时间"
            class="!w-full"
          />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="排序">
          <el-input-number
            v-model="newFormInline.sort"
            :min="0"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="版本号">
          <el-input-number
            v-model="newFormInline.version"
            :min="1"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col v-if="newFormInline.title === '修改'" :span="8">
        <el-form-item label="状态">
          <Segmented
            :modelValue="newFormInline.status !== 1"
            :options="statusOptions"
            @change="
              ({ option: { value } }) => {
                newFormInline.status = value ? 0 : 1;
              }
            "
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="备注">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-col>
    </el-row>

    <el-divider content-position="left">多语言内容</el-divider>
    <el-form-item prop="i18nList">
      <div class="w-full flex flex-col gap-4">
        <el-card
          v-for="(item, index) in newFormInline.i18nList"
          :key="item.id || `i18n-${index}`"
          shadow="never"
        >
          <template #header>
            <div class="flex items-center justify-between gap-3">
              <span>多语言项 {{ index + 1 }}</span>
              <div class="flex gap-2">
                <el-button type="primary" plain @click="addI18nItem">
                  新增语言
                </el-button>
                <el-button type="danger" plain @click="removeI18nItem(index)">
                  删除
                </el-button>
              </div>
            </div>
          </template>

          <el-row :gutter="18">
            <el-col :span="8">
              <el-form-item label="语言编码">
                <el-input
                  v-model="item.languageCode"
                  clearable
                  placeholder="如 zh-CN / en-US"
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="国家编码">
                <el-input
                  v-model="item.countryCode"
                  clearable
                  placeholder="为空表示通用"
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="状态">
                <el-select v-model="item.status" class="!w-full">
                  <el-option label="启用" :value="1" />
                  <el-option label="停用" :value="0" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="主标题">
                <el-input
                  v-model="item.title"
                  clearable
                  placeholder="请输入主标题"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="副标题">
                <el-input
                  v-model="item.subTitle"
                  clearable
                  placeholder="请输入副标题"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="按钮文案">
                <el-input
                  v-model="item.buttonText"
                  clearable
                  placeholder="请输入按钮文案"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="跳转值">
                <el-input
                  v-model="item.jumpValue"
                  clearable
                  placeholder="URL/路由/业务ID"
                />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item label="描述文案">
                <el-input
                  v-model="item.description"
                  type="textarea"
                  :rows="3"
                  placeholder="请输入描述文案"
                />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item label="主图URL">
                <el-input
                  v-model="item.imageUrl"
                  clearable
                  placeholder="请输入主图URL"
                />
                <el-upload
                  class="mt-2"
                  :action="uploadUrl"
                  :headers="getUploadHeaders()"
                  :before-upload="handleBeforeUpload"
                  :show-file-list="false"
                  :on-success="
                    response =>
                      updateI18nImageField(index, 'imageUrl', response)
                  "
                  :on-error="handleUploadError"
                >
                  <el-button type="primary">上传主图</el-button>
                </el-upload>
                <el-image
                  v-if="item.imageUrl"
                  class="mt-2"
                  style="width: 120px; height: 120px"
                  :src="item.imageUrl"
                  fit="cover"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="缩略图URL">
                <el-input
                  v-model="item.thumbUrl"
                  clearable
                  placeholder="请输入缩略图URL"
                />
                <el-upload
                  class="mt-2"
                  :action="uploadUrl"
                  :headers="getUploadHeaders()"
                  :before-upload="handleBeforeUpload"
                  :show-file-list="false"
                  :on-success="
                    response =>
                      updateI18nImageField(index, 'thumbUrl', response)
                  "
                  :on-error="handleUploadError"
                >
                  <el-button type="primary">上传缩略图</el-button>
                </el-upload>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="背景图URL">
                <el-input
                  v-model="item.bgImageUrl"
                  clearable
                  placeholder="请输入背景图URL"
                />
                <el-upload
                  class="mt-2"
                  :action="uploadUrl"
                  :headers="getUploadHeaders()"
                  :before-upload="handleBeforeUpload"
                  :show-file-list="false"
                  :on-success="
                    response =>
                      updateI18nImageField(index, 'bgImageUrl', response)
                  "
                  :on-error="handleUploadError"
                >
                  <el-button type="primary">上传背景图</el-button>
                </el-upload>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="图标URL">
                <el-input
                  v-model="item.iconUrl"
                  clearable
                  placeholder="请输入图标URL"
                />
                <el-upload
                  class="mt-2"
                  :action="uploadUrl"
                  :headers="getUploadHeaders()"
                  :before-upload="handleBeforeUpload"
                  :show-file-list="false"
                  :on-success="
                    response => updateI18nImageField(index, 'iconUrl', response)
                  "
                  :on-error="handleUploadError"
                >
                  <el-button type="primary">上传图标</el-button>
                </el-upload>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="视频URL">
                <el-input
                  v-model="item.videoUrl"
                  clearable
                  placeholder="请输入视频URL"
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="文字颜色">
                <el-input
                  v-model="item.textColor"
                  clearable
                  placeholder="#FFFFFF"
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="按钮颜色">
                <el-input
                  v-model="item.buttonColor"
                  clearable
                  placeholder="#409EFF"
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="背景色">
                <el-input
                  v-model="item.bgColor"
                  clearable
                  placeholder="#000000"
                />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item label="备注">
                <el-input
                  v-model="item.remark"
                  type="textarea"
                  :rows="2"
                  placeholder="请输入多语言备注"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-card>
      </div>
    </el-form-item>

    <el-divider content-position="left">投放国家</el-divider>
    <el-form-item prop="countryList">
      <div class="w-full flex flex-col gap-4">
        <el-card
          v-for="(item, index) in newFormInline.countryList"
          :key="item.id || `country-${index}`"
          shadow="never"
        >
          <template #header>
            <div class="flex items-center justify-between gap-3">
              <span>国家项 {{ index + 1 }}</span>
              <div class="flex gap-2">
                <el-button type="primary" plain @click="addCountryItem">
                  新增国家
                </el-button>
                <el-button
                  type="danger"
                  plain
                  @click="removeCountryItem(index)"
                >
                  删除
                </el-button>
              </div>
            </div>
          </template>

          <el-row :gutter="18">
            <el-col :span="8">
              <el-form-item label="国家编码">
                <el-input
                  v-model="item.countryCode"
                  clearable
                  placeholder="如 CN / TH / VN"
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="状态">
                <el-select v-model="item.status" class="!w-full">
                  <el-option label="启用" :value="1" />
                  <el-option label="停用" :value="0" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item label="备注">
                <el-input
                  v-model="item.remark"
                  type="textarea"
                  :rows="2"
                  placeholder="请输入国家投放备注"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-card>
      </div>
    </el-form-item>
  </el-form>
</template>
