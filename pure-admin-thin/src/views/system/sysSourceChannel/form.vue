<script setup lang="ts">
import { ref } from "vue";
import Segmented from "@/components/ReSegmented";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    tenantId: 0,
    channelCode: "",
    channelName: "",
    parentId: null,
    level: 1,
    status: 1,
    sort: 0,
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

const statusOptions = [
  { label: "启用", value: false },
  { label: "停用", value: true }
];

const levelOptions = [
  { label: "一级渠道", value: 1 },
  { label: "二级渠道", value: 2 }
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
    label-width="110px"
  >
    <el-row :gutter="18">
      <!-- <el-col :span="12">
        <el-form-item label="渠道层级" prop="level">
          <el-select
            v-model="newFormInline.level"
            class="!w-full"
            placeholder="请选择渠道层级"
          >
            <el-option
              v-for="item in levelOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col> -->
      <el-col :span="12">
        <el-form-item label="渠道编码" prop="channelCode">
          <el-input
            v-model="newFormInline.channelCode"
            :disabled="true"
            :placeholder="
              newFormInline.title === '新增'
                ? '保存后自动生成 8 位唯一编码'
                : '渠道编码由系统生成'
            "
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="渠道名称" prop="channelName">
          <el-input
            v-model="newFormInline.channelName"
            clearable
            placeholder="请输入渠道名称"
          />
        </el-form-item>
      </el-col>
      <!-- <el-col :span="12">
        <el-form-item label="父渠道ID">
          <el-input-number
            v-model="newFormInline.parentId"
            :min="1"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col> -->
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
