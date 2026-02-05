<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import { FormProps } from "./utils/types";
import Segmented from "@/components/ReSegmented";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    dictName: "",
    dictType: "",
    description: "",
    status: 0,
    menuIds: []
  })
});

import { dictTypeOptions } from "./utils/enums";

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
    label-width="82px"
  >
    <el-form-item label="字典名称" prop="dictName">
      <el-input
        v-model="newFormInline.dictName"
        clearable
        placeholder="请输入字典名称"
      />
    </el-form-item>

    <el-form-item label="字典类型" prop="dictType">
      <el-input
        v-model="newFormInline.dictType"
        clearable
        placeholder="请输入字典类型"
      />
    </el-form-item>

    <el-form-item label="描述">
      <el-input
        v-model="newFormInline.description"
        placeholder="请输入备注信息"
        type="textarea"
      />
    </el-form-item>
    <el-form-item v-if="newFormInline.title === '修改'" label="状态">
      <Segmented
        :modelValue="newFormInline.status != 0 ? 0 : 1"
        :options="dictTypeOptions"
        @change="
          ({ option: { value } }) => {
            newFormInline.status = value ? 1 : 0;
          }
        "
      />
    </el-form-item>
  </el-form>
</template>
