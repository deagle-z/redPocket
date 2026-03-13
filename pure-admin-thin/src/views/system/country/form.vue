<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";
import Segmented from "@/components/ReSegmented";
import { getSysCustomFieldList, type SysCustomField } from "@/api/customField";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    countryCode: "",
    countryNameCn: "",
    countryNameEn: "",
    currencyCode: "",
    currencySymbol: "",
    timezone: "",
    languageCode: "",
    withdrawFields: [],
    rechargeFields: [],
    sort: 0,
    status: 1,
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);
const customFieldLoading = ref(false);
const customFieldOptions = ref<SysCustomField[]>([]);

const statusOptions = [
  { label: "启用", value: false },
  { label: "停用", value: true }
];

const mergedCustomFieldOptions = computed(() => {
  const map = new Map<string, SysCustomField>();
  customFieldOptions.value.forEach(item => {
    map.set(item.fieldKey, item);
  });
  newFormInline.value.withdrawFields.forEach(item => {
    map.set(item.fieldKey, item);
  });
  newFormInline.value.rechargeFields.forEach(item => {
    map.set(item.fieldKey, item);
  });
  return Array.from(map.values());
});

async function loadCustomFields() {
  customFieldLoading.value = true;
  try {
    const { data } = await getSysCustomFieldList({
      currentPage: 0,
      pageSize: 500,
      status: 1
    });
    customFieldOptions.value = data?.list || [];
  } finally {
    customFieldLoading.value = false;
  }
}

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });

onMounted(() => {
  loadCustomFields();
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
        <el-form-item label="国家编码" prop="countryCode">
          <el-input v-model="newFormInline.countryCode" clearable placeholder="请输入国家编码" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="币种编码" prop="currencyCode">
          <el-input v-model="newFormInline.currencyCode" clearable placeholder="请输入币种编码" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="中文名" prop="countryNameCn">
          <el-input v-model="newFormInline.countryNameCn" clearable placeholder="请输入国家中文名" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="英文名" prop="countryNameEn">
          <el-input v-model="newFormInline.countryNameEn" clearable placeholder="请输入国家英文名" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="货币符号">
          <el-input v-model="newFormInline.currencySymbol" clearable placeholder="请输入货币符号" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="语言编码">
          <el-input v-model="newFormInline.languageCode" clearable placeholder="请输入语言编码" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="时区">
          <el-input v-model="newFormInline.timezone" clearable placeholder="请输入时区" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="排序">
          <el-input-number v-model="newFormInline.sort" :min="0" class="!w-full" controls-position="right" />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="提现字段">
          <el-select
            v-model="newFormInline.withdrawFields"
            multiple
            value-key="fieldKey"
            filterable
            collapse-tags
            collapse-tags-tooltip
            clearable
            class="!w-full"
            placeholder="请选择提现字段"
            :loading="customFieldLoading"
          >
            <el-option
              v-for="item in mergedCustomFieldOptions"
              :key="item.fieldKey"
              :label="`${item.fieldLabel} (${item.fieldKey})`"
              :value="item"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="充值字段">
          <el-select
            v-model="newFormInline.rechargeFields"
            multiple
            value-key="fieldKey"
            filterable
            collapse-tags
            collapse-tags-tooltip
            clearable
            class="!w-full"
            placeholder="请选择充值字段"
            :loading="customFieldLoading"
          >
            <el-option
              v-for="item in mergedCustomFieldOptions"
              :key="`recharge-${item.fieldKey}`"
              :label="`${item.fieldLabel} (${item.fieldKey})`"
              :value="item"
            />
          </el-select>
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
