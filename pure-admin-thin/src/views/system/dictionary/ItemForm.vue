<script setup lang="ts">
import { ref } from "vue";
import ReCol from "@/components/ReCol";
import { formRules, itemFormRules } from "./utils/rule";
import { ItemFormProps } from "./utils/types";
import { transformI18n } from "@/plugins/i18n";
import { IconSelect } from "@/components/ReIcon";
import Segmented from "@/components/ReSegmented";
import ReAnimateSelector from "@/components/ReAnimateSelector";

const props = withDefaults(defineProps<ItemFormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    code: "",
    color: "",
    dictLabel: "",
    dictType: "",
    dictValue: "",
    is_default: 0,
    remark: "",
    sort: 0,
    status: 0
  })
});
import { dictItemOptions } from "./utils/enums";

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
    :rules="itemFormRules"
    label-width="82px"
  >
    <el-row dis :gutter="30">
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="字典类型" prop="dictType">
          <el-input v-model="newFormInline.dictType" clearable disabled />
        </el-form-item>
      </re-col>
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="字典名称" prop="dictLabel">
          <el-input
            v-model="newFormInline.dictLabel"
            clearable
            placeholder="请输入字典项名称"
          />
        </el-form-item>
      </re-col>
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="字典项值" prop="dictValue">
          <el-input
            v-model="newFormInline.dictValue"
            clearable
            placeholder="请输入字典项值"
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="字典代码" prop="code">
          <el-input
            v-model="newFormInline.code"
            clearable
            placeholder="请输入字典代码"
          />
        </el-form-item>
      </re-col>
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            clearable
            placeholder="请输入备注"
          />
        </el-form-item>
      </re-col>
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="排序">
          <el-input-number
            v-model="newFormInline.sort"
            class="!w-full"
            :min="1"
            :max="9999"
            controls-position="right"
          />
        </el-form-item>
      </re-col>
      <re-col
        v-if="newFormInline.title === '修改'"
        :value="12"
        :xs="24"
        :sm="24"
      >
        ,
        <el-form-item label="状态">
          <Segmented
            :modelValue="newFormInline.status != 0 ? 0 : 1"
            :options="dictItemOptions"
            @change="
              ({ option: { value } }) => {
                newFormInline.status = value ? 1 : 0;
              }
            "
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
