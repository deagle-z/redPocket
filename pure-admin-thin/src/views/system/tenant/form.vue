<script setup lang="ts">
import { ref } from "vue";
import ReCol from "@/components/ReCol";
import { formRules } from "./utils/rule";
import { FormProps } from "./utils/types";
import { tenantTypeOptions, statusOptions } from "./utils/enums";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "",
    id: 0,
    tenantCode: "",
    tenantName: "",
    tenantType: 1,
    status: 1,
    ownerUserId: undefined,
    planCode: "",
    timezone: "UTC",
    locale: "en-US",
    remark: ""
  })
});

const tenantFormRef = ref();
const newFormInline = ref(props.formInline);

function getRef() {
  return tenantFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="tenantFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="100px"
  >
    <el-row :gutter="20">
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="租户编码" prop="tenantCode">
          <el-input
            v-model="newFormInline.tenantCode"
            placeholder="请输入租户编码"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="租户名称" prop="tenantName">
          <el-input
            v-model="newFormInline.tenantName"
            placeholder="请输入租户名称"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="租户类型" prop="tenantType">
          <el-select
            v-model="newFormInline.tenantType"
            placeholder="请选择"
            class="w-full"
          >
            <el-option
              v-for="item in tenantTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="状态" prop="status">
          <el-select
            v-model="newFormInline.status"
            placeholder="请选择"
            class="w-full"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="套餐标识" prop="planCode">
          <el-input
            v-model="newFormInline.planCode"
            placeholder="free/pro/enterprise"
            clearable
          />
        </el-form-item>
      </re-col>

      <!-- <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="时区" prop="timezone">
          <el-input
            v-model="newFormInline.timezone"
            placeholder="UTC"
            clearable
          />
        </el-form-item>
      </re-col> -->

      <!-- <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="默认语言" prop="locale">
          <el-input
            v-model="newFormInline.locale"
            placeholder="en-US"
            clearable
          />
        </el-form-item>
      </re-col> -->

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            placeholder="备注"
            clearable
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
