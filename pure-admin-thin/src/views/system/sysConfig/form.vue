<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    configKey: "",
    configValue: "",
    configDesc: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

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
    label-width="90px"
  >
    <el-form-item label="配置Key" prop="configKey">
      <el-input
        v-model="newFormInline.configKey"
        :disabled="newFormInline.id > 0"
        clearable
        placeholder="请输入配置Key（唯一）"
      />
    </el-form-item>
    <el-form-item label="配置值" prop="configValue">
      <el-input
        v-model="newFormInline.configValue"
        type="textarea"
        :rows="4"
        placeholder="请输入配置值"
      />
    </el-form-item>
    <el-form-item label="描述">
      <el-input
        v-model="newFormInline.configDesc"
        clearable
        placeholder="请输入描述（可选）"
      />
    </el-form-item>
  </el-form>
</template>
