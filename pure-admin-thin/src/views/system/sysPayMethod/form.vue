<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";
import Segmented from "@/components/ReSegmented";
import { getToken, formatToken } from "@/utils/auth";
import { message } from "@/utils/message";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    methodCode: "",
    methodName: "",
    icon: "",
    sort: 0,
    status: 1,
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

const statusOptions = [
  { label: "启用", value: false },
  { label: "停用", value: true }
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

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, uploading });
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
        <el-form-item label="方式编码" prop="methodCode">
          <el-input
            v-model="newFormInline.methodCode"
            clearable
            placeholder="如 upi / pix / usdt_trc20"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="方式名称" prop="methodName">
          <el-input
            v-model="newFormInline.methodName"
            clearable
            placeholder="请输入方式名称"
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
      <el-col :span="12" />
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
