<script setup lang="ts">
import { onMounted, ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";
import Segmented from "@/components/ReSegmented";
import { getSysPayMethodList, type SysPayMethod } from "@/api/sysPayMethod";
import { getSysCountryList, type SysCountry } from "@/api/country";
import { getToken, formatToken } from "@/utils/auth";
import { message } from "@/utils/message";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    channelCode: "",
    channelName: "",
    channelType: "",
    providerType: "third_party",
    countryCode: "",
    icon: "",
    remark: "",
    sort: 0,
    status: 1,
    methodIds: []
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const methodLoading = ref(false);
const methodOptions = ref<SysPayMethod[]>([]);
const countryLoading = ref(false);
const countryOptions = ref<SysCountry[]>([]);

const statusOptions = [
  { label: "启用", value: false },
  { label: "停用", value: true }
];

const channelTypeOptions = [
  { label: "充值", value: "deposit" },
  { label: "提现", value: "withdraw" },
  { label: "充值+提现", value: "both" }
];

const providerTypeOptions = [
  { label: "三方", value: "third_party" },
  { label: "自有", value: "native" }
];

const uploadUrl = `${import.meta.env.VITE_BASE_URL}/api/v1/admin/upload`;
const uploading = ref(false);

function getUploadHeaders() {
  const token = getToken()?.accessToken;
  return token ? { Authorization: formatToken(token) } : {};
}

function handleBeforeUpload() {
  uploading.value = true;
  return true;
}

function handleUploadSuccess(response: any) {
  uploading.value = false;
  const url = response?.data?.url || response?.url;
  if (!url) {
    message("上传失败，未返回URL", { type: "error" });
    return;
  }
  newFormInline.value.icon = url;
  message("上传成功", { type: "success" });
}

function handleUploadError() {
  uploading.value = false;
  message("上传失败", { type: "error" });
}

async function loadMethods() {
  methodLoading.value = true;
  try {
    const { data } = await getSysPayMethodList({
      currentPage: 0,
      pageSize: 500
    });
    methodOptions.value = data?.list || [];
  } finally {
    methodLoading.value = false;
  }
}

async function loadCountries() {
  countryLoading.value = true;
  try {
    const { data } = await getSysCountryList({
      currentPage: 0,
      pageSize: 500
    });
    countryOptions.value = data?.list || [];
  } finally {
    countryLoading.value = false;
  }
}

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, uploading });

onMounted(() => {
  loadMethods();
  loadCountries();
});
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="110px"
  >
    <el-row :gutter="18">
      <el-col :span="12">
        <el-form-item label="通道编码" prop="channelCode">
          <el-input
            v-model="newFormInline.channelCode"
            clearable
            placeholder="如 cashfree / paytm"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="通道名称" prop="channelName">
          <el-input
            v-model="newFormInline.channelName"
            clearable
            placeholder="请输入通道名称"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="通道类型" prop="channelType">
          <el-select v-model="newFormInline.channelType" class="!w-full" placeholder="请选择通道类型">
            <el-option
              v-for="item in channelTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="提供方类型" prop="providerType">
          <el-select v-model="newFormInline.providerType" class="!w-full" placeholder="请选择提供方类型">
            <el-option
              v-for="item in providerTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="国家码">
          <el-select
            v-model="newFormInline.countryCode"
            filterable
            clearable
            class="!w-full"
            placeholder="请选择国家"
            :loading="countryLoading"
          >
            <el-option
              v-for="item in countryOptions"
              :key="item.countryCode"
              :label="`${item.countryCode} - ${item.countryNameEn}`"
              :value="item.countryCode"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="排序">
          <el-input-number
            v-model="newFormInline.sort"
            :min="0"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="图标">
          <div class="flex items-center gap-2 w-full">
            <el-input
              v-model="newFormInline.icon"
              clearable
              placeholder="图标URL"
              class="flex-1"
            />
            <el-upload
              :action="uploadUrl"
              :headers="getUploadHeaders()"
              :show-file-list="false"
              accept="image/*"
              :before-upload="handleBeforeUpload"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
            >
              <el-button type="primary" :loading="uploading">上传</el-button>
            </el-upload>
          </div>
          <el-image
            v-if="newFormInline.icon"
            class="mt-2"
            style="width: 64px; height: 64px"
            :src="newFormInline.icon"
            fit="contain"
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="绑定支付方式">
          <el-select
            v-model="newFormInline.methodIds"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            clearable
            class="!w-full"
            placeholder="请选择支付方式"
            :loading="methodLoading"
          >
            <el-option
              v-for="item in methodOptions"
              :key="item.id"
              :label="`${item.methodName} (${item.methodCode})`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="备注">
          <el-input
            v-model="newFormInline.remark"
            placeholder="请输入备注"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </el-col>
      <el-col v-if="newFormInline.title === '修改'" :span="24">
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
    </el-row>
  </el-form>
</template>
