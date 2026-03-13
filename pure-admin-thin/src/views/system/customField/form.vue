<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";
import Segmented from "@/components/ReSegmented";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    fieldKey: "",
    fieldLabel: "",
    fieldPlaceholder: "",
    fieldType: "input",
    dataType: "string",
    defaultValue: "",
    isRequired: 0,
    isSensitive: 0,
    maxLength: null,
    minLength: null,
    regexRule: "",
    errorTips: "",
    optionsJson: "",
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

const fieldTypeOptions = [
  "input",
  "select",
  "textarea",
  "number",
  "date",
  "file",
  "switch",
  "radio",
  "checkbox"
];

const dataTypeOptions = ["string", "int", "decimal", "bool", "json", "date"];

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
    label-width="110px"
  >
    <el-row :gutter="18">
      <el-col :span="12">
        <el-form-item label="字段Key" prop="fieldKey">
          <el-input v-model="newFormInline.fieldKey" clearable placeholder="请输入字段Key" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="字段名称" prop="fieldLabel">
          <el-input v-model="newFormInline.fieldLabel" clearable placeholder="请输入字段名称" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="字段类型" prop="fieldType">
          <el-select v-model="newFormInline.fieldType" class="!w-full">
            <el-option v-for="item in fieldTypeOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="数据类型" prop="dataType">
          <el-select v-model="newFormInline.dataType" class="!w-full">
            <el-option v-for="item in dataTypeOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="占位提示">
          <el-input v-model="newFormInline.fieldPlaceholder" clearable placeholder="请输入占位提示" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="默认值">
          <el-input v-model="newFormInline.defaultValue" clearable placeholder="请输入默认值" />
        </el-form-item>
      </el-col>
      <el-col :span="6">
        <el-form-item label="最小长度">
          <el-input-number v-model="newFormInline.minLength" :min="0" class="!w-full" controls-position="right" />
        </el-form-item>
      </el-col>
      <el-col :span="6">
        <el-form-item label="最大长度">
          <el-input-number v-model="newFormInline.maxLength" :min="0" class="!w-full" controls-position="right" />
        </el-form-item>
      </el-col>
      <el-col :span="6">
        <el-form-item label="是否必填">
          <el-switch v-model="newFormInline.isRequired" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-col>
      <el-col :span="6">
        <el-form-item label="是否敏感">
          <el-switch v-model="newFormInline.isSensitive" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="校验正则">
          <el-input v-model="newFormInline.regexRule" clearable placeholder="请输入校验正则" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="错误提示">
          <el-input v-model="newFormInline.errorTips" clearable placeholder="请输入错误提示" />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="选项JSON">
          <el-input v-model="newFormInline.optionsJson" placeholder="请输入 JSON 字符串" type="textarea" :rows="4" />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="备注">
          <el-input v-model="newFormInline.remark" placeholder="请输入备注" type="textarea" :rows="3" />
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
