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
    hostName: "",
    tablePrefix: "",
    hostMark: "",
    hostDesc: "",
    enabled: true
  })
});

const hostInfoFormRef = ref();
const newFormInline = ref(props.formInline);
const { switchStyle } = usePublicHooks();

function getRef() {
  return hostInfoFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="hostInfoFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="82px"
  >
    <el-form-item label="域名" prop="hostName">
      <el-input
        v-model="newFormInline.hostName"
        placeholder="请输入域名"
        :disabled="newFormInline.title !== '新增'"
      />
    </el-form-item>

    <el-form-item
      v-if="newFormInline.title === '新增'"
      label="表名前缀"
      prop="tablePrefix"
    >
      <el-input
        v-model="newFormInline.tablePrefix"
        clearable
        placeholder="请输入表名前缀"
      />
    </el-form-item>
    <el-form-item label="域名状态">
      <el-switch
        v-model="newFormInline.enabled"
        inline-prompt
        :active-value="true"
        :inactive-value="false"
        active-text="启用"
        inactive-text="停用"
        :style="switchStyle"
      />
    </el-form-item>
    <el-form-item label="域名标识" prop="hostMark">
      <el-input
        v-model="newFormInline.hostMark"
        clearable
        placeholder="请输入域名标识"
      />
    </el-form-item>

    <el-form-item label="域名备注">
      <el-input
        v-model="newFormInline.hostDesc"
        placeholder="请输入域名备注信息"
        type="textarea"
      />
    </el-form-item>
  </el-form>
</template>
