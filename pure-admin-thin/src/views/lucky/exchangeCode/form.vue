<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    code: "",
    amount: 0,
    maxRedeemCount: 1,
    generateCount: 1,
    status: 1,
    remark: ""
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
    label-width="112px"
  >
    <el-form-item v-if="newFormInline.title === '新增'" label="兑换码" prop="code">
      <el-input
        v-model="newFormInline.code"
        maxlength="6"
        clearable
        placeholder="请输入6位数字，留空自动生成"
      />
    </el-form-item>

    <el-form-item
      v-if="newFormInline.title === '新增' && !newFormInline.code.trim()"
      label="生成数量"
      prop="generateCount"
    >
      <el-input-number
        v-model="newFormInline.generateCount"
        :min="1"
        :max="500"
        :step="1"
        :precision="0"
        class="!w-full"
        controls-position="right"
      />
      <div class="text-xs text-[var(--el-text-color-secondary)] mt-1">
        留空兑换码时按此数量批量随机生成（1-500）
      </div>
    </el-form-item>
    <el-form-item v-else label="兑换码">
      <el-input v-model="newFormInline.code" disabled />
    </el-form-item>

    <template v-if="newFormInline.title === '新增'">
      <el-form-item label="兑换金额" prop="amount">
        <el-input-number
          v-model="newFormInline.amount"
          :min="0.01"
          :precision="2"
          :step="1"
          class="!w-full"
          controls-position="right"
        />
      </el-form-item>
      <el-form-item label="可兑换次数" prop="maxRedeemCount">
        <el-input-number
          v-model="newFormInline.maxRedeemCount"
          :min="1"
          :step="1"
          :precision="0"
          class="!w-full"
          controls-position="right"
        />
      </el-form-item>
    </template>

    <el-form-item label="状态" prop="status">
      <el-radio-group v-model="newFormInline.status">
        <el-radio-button :label="1">启用</el-radio-button>
        <el-radio-button :label="0">停用</el-radio-button>
      </el-radio-group>
    </el-form-item>

    <el-form-item label="备注">
      <el-input
        v-model="newFormInline.remark"
        type="textarea"
        :rows="3"
        maxlength="255"
        show-word-limit
        placeholder="请输入备注"
      />
    </el-form-item>
  </el-form>
</template>
