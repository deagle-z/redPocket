<script setup lang="ts">
import { ref } from "vue";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";
import { levelOptions, statusOptions } from "./utils/hook";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    configTitle: "",
    subTitle: "",
    levelCode: "primary",
    levelName: "初级任务",
    sort: 0,
    status: 1,
    taskType: "invite_recharge",
    requiredInviteCount: 1,
    requiredRechargeAmount: 50,
    durationMinutes: 1440,
    rewardAmount: 0,
    rewardCurrency: "USD",
    rewardTarget: "rebate",
    startAt: null,
    endAt: null,
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

function handleLevelChange(value: "primary" | "middle" | "advanced") {
  const match = levelOptions.find(item => item.value === value);
  if (match && !newFormInline.value.levelName) {
    newFormInline.value.levelName = match.label;
  }
}

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
        <el-form-item label="任务标题" prop="configTitle">
          <el-input
            v-model="newFormInline.configTitle"
            clearable
            placeholder="如 初级任务"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="任务等级" prop="levelCode">
          <el-select
            v-model="newFormInline.levelCode"
            class="!w-full"
            @change="handleLevelChange"
          >
            <el-option
              v-for="item in levelOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="等级名称">
          <el-input
            v-model="newFormInline.levelName"
            clearable
            placeholder="默认跟随任务等级"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="状态">
          <el-radio-group v-model="newFormInline.status">
            <el-radio-button
              v-for="item in statusOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-col>

      <el-col :span="24">
        <el-form-item label="副标题">
          <el-input
            v-model="newFormInline.subTitle"
            clearable
            placeholder="如 邀请3位好友注册并充值≥50USD"
          />
        </el-form-item>
      </el-col>

      <el-col :span="24">
        <el-divider content-position="left">任务规则</el-divider>
      </el-col>
      <el-col :span="8">
        <el-form-item label="邀请人数" prop="requiredInviteCount">
          <el-input-number
            v-model="newFormInline.requiredInviteCount"
            :min="1"
            :precision="0"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="充值金额" prop="requiredRechargeAmount">
          <el-input-number
            v-model="newFormInline.requiredRechargeAmount"
            :min="0"
            :precision="2"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="限制分钟" prop="durationMinutes">
          <el-input-number
            v-model="newFormInline.durationMinutes"
            :min="1"
            :precision="0"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>

      <el-col :span="24">
        <el-divider content-position="left">奖励规则</el-divider>
      </el-col>
      <el-col :span="8">
        <el-form-item label="奖励金额" prop="rewardAmount">
          <el-input-number
            v-model="newFormInline.rewardAmount"
            :min="0"
            :precision="2"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="奖励币种">
          <el-input
            v-model="newFormInline.rewardCurrency"
            clearable
            placeholder="USD"
          />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="奖励账户">
          <el-input model-value="佣金账户" disabled />
        </el-form-item>
      </el-col>

      <el-col :span="24">
        <el-divider content-position="left">有效期</el-divider>
      </el-col>
      <el-col :span="12">
        <el-form-item label="开始时间">
          <el-date-picker
            v-model="newFormInline.startAt"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="不限制"
            class="!w-full"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="结束时间">
          <el-date-picker
            v-model="newFormInline.endAt"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="不限制"
            class="!w-full"
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
      <el-col :span="24">
        <el-form-item label="备注">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="2"
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
</template>
