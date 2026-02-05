<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import { FormProps } from "./utils/types";
import ReCol from "@/components/ReCol";
import { usePublicHooks } from "@/views/system/hooks";
import { getToken, formatToken } from "@/utils/auth";
import { message } from "@/utils/message";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "",
    id: 0,
    groupId: 0,
    groupName: "",
    status: 1,
    serviceUrl: "",
    rechargeUrl: "",
    channelUrl: "",
    sendPacketImage: "",
    loseRate: 1.8,
    numConfig: "3",
    sendCommission: 2,
    grabbingCommission: 3
  })
});

const authGroupFormRef = ref();
const newFormInline = ref(props.formInline);
const { switchStyle } = usePublicHooks();
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

function handleUploadSuccess(response: any) {
  const url = response?.data?.url || response?.url;
  if (!url) {
    message("上传失败，未返回URL", { type: "error" });
    return;
  }
  newFormInline.value.sendPacketImage = url;
  message("上传成功", { type: "success" });
}

function handleUploadError() {
  message("上传失败", { type: "error" });
}

function getRef() {
  return authGroupFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="authGroupFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="100px"
  >
    <el-form-item label="群组ID" prop="groupId">
      <el-input-number
        v-model="newFormInline.groupId"
        placeholder="请输入群组ID"
        :min="1"
        style="width: 100%"
      />
    </el-form-item>

    <el-form-item label="群组名称" prop="groupName">
      <el-input
        v-model="newFormInline.groupName"
        placeholder="请输入群组名称"
        clearable
      />
    </el-form-item>

    <el-form-item label="状态">
      <el-switch
        v-model="newFormInline.status"
        inline-prompt
        :active-value="1"
        :inactive-value="0"
        active-text="启用"
        inactive-text="禁用"
        :style="switchStyle"
      />
    </el-form-item>

    <el-form-item label="客服URL">
      <el-input
        v-model="newFormInline.serviceUrl"
        placeholder="请输入客服URL（可选）"
        clearable
        type="textarea"
        :rows="2"
      />
    </el-form-item>

    <el-form-item label="充值URL">
      <el-input
        v-model="newFormInline.rechargeUrl"
        placeholder="请输入充值URL（可选）"
        clearable
        type="textarea"
        :rows="2"
      />
    </el-form-item>

    <el-form-item label="玩法URL">
      <el-input
        v-model="newFormInline.channelUrl"
        placeholder="请输入玩法URL（可选）"
        clearable
        type="textarea"
        :rows="2"
      />
    </el-form-item>

    <el-form-item label="发包图片">
      <el-input
        v-model="newFormInline.sendPacketImage"
        placeholder="发红包图片URL"
        clearable
      />
      <el-upload
        class="mt-2"
        :action="uploadUrl"
        :headers="getUploadHeaders()"
        :before-upload="handleBeforeUpload"
        :show-file-list="false"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
      >
        <el-button type="primary">上传图片</el-button>
      </el-upload>
      <el-image
        v-if="newFormInline.sendPacketImage"
        class="mt-2"
        style="width: 120px; height: 120px"
        :src="newFormInline.sendPacketImage"
        fit="cover"
      />
    </el-form-item>

    <el-form-item label="中雷倍数" prop="loseRate">
      <el-input-number
        v-model="newFormInline.loseRate"
        placeholder="请输入中雷倍数"
        :precision="2"
        :step="0.1"
        :min="0.1"
        :max="10"
        style="width: 100%"
      />
    </el-form-item>

    <el-form-item label="数量配置" prop="numConfig">
      <el-input
        v-model="newFormInline.numConfig"
        placeholder="请输入数量配置，如：3 或 3|9"
        clearable
      />
    </el-form-item>

    <el-form-item label="发包抽成(%)" prop="sendCommission">
      <el-input-number
        v-model="newFormInline.sendCommission"
        placeholder="请输入发包中雷抽成百分比"
        :min="0"
        :max="100"
        style="width: 100%"
      />
    </el-form-item>

    <el-form-item label="抢包抽成(%)" prop="grabbingCommission">
      <el-input-number
        v-model="newFormInline.grabbingCommission"
        placeholder="请输入抢红包抽成百分比"
        :min="0"
        :max="100"
        style="width: 100%"
      />
    </el-form-item>
  </el-form>
</template>
