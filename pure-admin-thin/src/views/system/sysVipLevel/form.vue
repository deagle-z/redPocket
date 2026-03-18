<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";
import Segmented from "@/components/ReSegmented";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    tenantId: 0,
    level: 1,
    levelName: "",
    agentTag: "",
    totalRechargeCount: null,
    totalRechargeAmount: null,
    totalValidBet: null,
    monthRechargeAmount: null,
    monthValidBet: null,
    upgradeBonusAmount: 0,
    upgradeType: 1,
    keepLevelCondition: 0,
    sort: 0,
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

const upgradeTypeOptions = [
  { label: "累计", value: 1 },
  { label: "当月", value: 2 }
];

const keepLevelOptions = [
  { label: "否", value: false },
  { label: "是", value: true }
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
    label-width="120px"
  >
    <el-row :gutter="18">
      <el-col :span="12">
        <el-form-item label="等级排序" prop="level">
          <el-input-number
            v-model="newFormInline.level"
            :min="1"
            :precision="0"
            class="!w-full"
            controls-position="right"
            placeholder="1=VIP0, 2=VIP1..."
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="等级名称" prop="levelName">
          <el-input
            v-model="newFormInline.levelName"
            clearable
            placeholder="如 VIP0 / VIP1"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="代理标签">
          <el-input
            v-model="newFormInline.agentTag"
            clearable
            placeholder="如 平台默认"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="升级方式">
          <el-select v-model="newFormInline.upgradeType" class="!w-full">
            <el-option
              v-for="item in upgradeTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>

      <el-col :span="24">
        <el-divider content-position="left">升级条件</el-divider>
      </el-col>

      <el-col :span="12">
        <el-form-item label="总充值次数">
          <el-input-number
            v-model="newFormInline.totalRechargeCount"
            :min="0"
            :precision="0"
            class="!w-full"
            controls-position="right"
            placeholder="不限则留空"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="总充值金额">
          <el-input-number
            v-model="newFormInline.totalRechargeAmount"
            :min="0"
            :precision="2"
            class="!w-full"
            controls-position="right"
            placeholder="不限则留空"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="总有效投注">
          <el-input-number
            v-model="newFormInline.totalValidBet"
            :min="0"
            :precision="2"
            class="!w-full"
            controls-position="right"
            placeholder="不限则留空"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="当月充值金额">
          <el-input-number
            v-model="newFormInline.monthRechargeAmount"
            :min="0"
            :precision="2"
            class="!w-full"
            controls-position="right"
            placeholder="不限则留空"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="当月有效投注">
          <el-input-number
            v-model="newFormInline.monthValidBet"
            :min="0"
            :precision="2"
            class="!w-full"
            controls-position="right"
            placeholder="不限则留空"
          />
        </el-form-item>
      </el-col>

      <el-col :span="24">
        <el-divider content-position="left">升级奖励</el-divider>
      </el-col>

      <el-col :span="12">
        <el-form-item label="升级赠送金额">
          <el-input-number
            v-model="newFormInline.upgradeBonusAmount"
            :min="0"
            :precision="2"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>

      <el-col :span="24">
        <el-divider content-position="left">其他设置</el-divider>
      </el-col>

      <el-col :span="12">
        <el-form-item label="需要保级">
          <Segmented
            :modelValue="newFormInline.keepLevelCondition === 1"
            :options="keepLevelOptions"
            @change="
              ({ option: { value } }) => {
                newFormInline.keepLevelCondition = value ? 1 : 0;
              }
            "
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="排序">
          <el-input-number
            v-model="newFormInline.sort"
            :min="0"
            :precision="0"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col v-if="newFormInline.title === '修改'" :span="12">
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
      <el-col :span="24">
        <el-form-item label="备注">
          <el-input
            v-model="newFormInline.remark"
            placeholder="请输入备注"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
</template>
