<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import { FormProps } from "./utils/types";
import ReCol from "@/components/ReCol";
import { usePublicHooks } from "@/views/system/hooks";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "",
    id: 0,
    apkPackage: "",
    name: "",
    url: "",
    isActive: 1
  })
});

const pkManagerFormRef = ref();
const newFormInline = ref(props.formInline);
const { switchStyle } = usePublicHooks();

function getRef() {
  return pkManagerFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="pkManagerFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="100px"
  >
    <el-form-item label="APK包名" prop="apkPackage">
      <el-input
        v-model="newFormInline.apkPackage"
        placeholder="请输入APK包名"
        clearable
      />
    </el-form-item>

    <el-form-item label="名称" prop="name">
      <el-input
        v-model="newFormInline.name"
        placeholder="请输入名称"
        clearable
      />
    </el-form-item>

    <el-form-item label="URL" prop="url">
      <el-input
        v-model="newFormInline.url"
        placeholder="请输入WebView展示的URL"
        clearable
        type="textarea"
        :rows="3"
      />
    </el-form-item>

    <el-form-item label="状态">
      <el-switch
        v-model="newFormInline.isActive"
        inline-prompt
        :active-value="1"
        :inactive-value="0"
        active-text="启用"
        inactive-text="停用"
        :style="switchStyle"
      />
    </el-form-item>
  </el-form>
</template>


