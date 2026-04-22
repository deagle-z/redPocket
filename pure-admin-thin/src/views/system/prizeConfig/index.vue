<script setup lang="ts">
import { ref, onMounted } from "vue";
import {
  getPrizePoolBalance,
  getPrizePoolConfig,
  setPrizePoolBalance,
  setPrizePoolConfig
} from "@/api/prizePoolConfig";
import { message } from "@/utils/message";
import Segmented from "@/components/ReSegmented";

defineOptions({ name: "SystemPrizeConfig" });

const loading = ref(false);
const saving = ref(false);
const balanceSaving = ref(false);
const poolId = ref<number>(1);
const formRef = ref();
const luckyBalance = ref<number>(0);

const form = ref({
  id: 0,
  poolId: 1,
  probabilities: "",
  amounts: "",
  totalProbability: 100,
  count: 1,
  peerAmount: 0,
  status: 1,
  remark: ""
});

const formRules = {
  probabilities: [
    { required: true, message: "请输入概率列表", trigger: "blur" }
  ],
  amounts: [{ required: true, message: "请输入金额列表", trigger: "blur" }],
  totalProbability: [
    { required: true, message: "请输入概率总和", trigger: "blur" }
  ]
};

const statusOptions = [
  { label: "启用", value: false },
  { label: "停用", value: true }
];

async function loadConfig() {
  if (!poolId.value) return;
  loading.value = true;
  try {
    const [{ data }, balanceResp] = await Promise.all([
      getPrizePoolConfig(poolId.value),
      getPrizePoolBalance("lucky")
    ]);
    luckyBalance.value = Number(balanceResp?.data?.balance ?? 0);
    if (data?.id) {
      form.value = {
        id: data.id,
        poolId: data.poolId,
        probabilities: data.probabilities,
        amounts: data.amounts,
        totalProbability: data.totalProbability,
        count: data.count ?? 1,
        peerAmount: data.peerAmount ?? 0,
        status: data.status,
        remark: data.remark ?? ""
      };
    } else {
      form.value = {
        id: 0,
        poolId: poolId.value,
        probabilities: "",
        amounts: "",
        totalProbability: 100,
        count: 1,
        peerAmount: 0,
        status: 1,
        remark: ""
      };
    }
  } catch {
    message("获取配置失败", { type: "error" });
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  formRef.value?.validate(async (valid: boolean) => {
    if (!valid) return;
    const probList = form.value.probabilities
      .split("|")
      .filter(s => s.trim() !== "");
    const amountList = form.value.amounts
      .split("|")
      .filter(s => s.trim() !== "");
    if (probList.length !== amountList.length) {
      message(
        `概率列表(${probList.length}个)与金额列表(${amountList.length}个)数量不一致`,
        { type: "warning" }
      );
      return;
    }
    saving.value = true;
    try {
      await setPrizePoolConfig({
        id: form.value.id || undefined,
        poolId: form.value.poolId,
        probabilities: form.value.probabilities.trim(),
        amounts: form.value.amounts.trim(),
        totalProbability: Number(form.value.totalProbability),
        count: Number(form.value.count),
        peerAmount: Number(form.value.peerAmount),
        status: form.value.status,
        remark: form.value.remark?.trim() || null
      });
      message("保存成功", { type: "success" });
      loadConfig();
    } catch {
      message("保存失败", { type: "error" });
    } finally {
      saving.value = false;
    }
  });
}

async function handleSaveBalance() {
  balanceSaving.value = true;
  try {
    await setPrizePoolBalance({
      poolCode: "lucky",
      balance: Number(luckyBalance.value),
      remark: "后台手动修改 lucky 奖池余额"
    });
    message("lucky奖池余额保存成功", { type: "success" });
    loadConfig();
  } catch {
    message("lucky奖池余额保存失败", { type: "error" });
  } finally {
    balanceSaving.value = false;
  }
}

onMounted(() => loadConfig());
</script>

<template>
  <div class="main p-4">
    <el-card shadow="never">
      <!-- <template #header>
        <div class="flex items-center gap-3">
          <span class="font-semibold text-base">奖池概率配置</span>
          <el-input-number
            v-model="poolId"
            :min="1"
            controls-position="right"
            style="width: 140px"
            placeholder="奖池 ID"
            @change="loadConfig"
          />
          <el-button @click="loadConfig">加载</el-button>
        </div>
      </template> -->

      <el-form
        ref="formRef"
        v-loading="loading"
        :model="form"
        :rules="formRules"
        label-width="110px"
        style="max-width: 700px"
      >
        <el-form-item label="lucky奖池余额">
          <div class="flex items-center gap-2 w-full">
            <el-input-number
              v-model="luckyBalance"
              :min="0"
              :precision="6"
              controls-position="right"
              style="width: 220px"
              placeholder="请输入lucky奖池余额"
            />
            <el-button
              type="primary"
              :loading="balanceSaving"
              @click="handleSaveBalance"
            >
              保存余额
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="概率列表" prop="probabilities">
          <el-input
            v-model="form.probabilities"
            clearable
            placeholder="用 | 分隔，如 2|10|30|58"
          />
        </el-form-item>

        <el-form-item label="金额列表" prop="amounts">
          <el-input
            v-model="form.amounts"
            clearable
            placeholder="用 | 分隔，如 0.5|2|10|50，与概率列表一一对应"
          />
        </el-form-item>

        <!-- <el-form-item label="概率总和" prop="totalProbability">
          <el-input-number
            v-model="form.totalProbability"
            :min="1"
            :max="1000"
            controls-position="right"
            style="width: 160px"
            placeholder="必须等于100"
          />
        </el-form-item> -->

        <el-form-item label="每批人数" prop="count">
          <el-input-number
            v-model="form.count"
            :min="1"
            controls-position="right"
            style="width: 160px"
            placeholder="多少人算一批"
          />
        </el-form-item>

        <el-form-item label="单次抽奖金额" prop="peerAmount">
          <el-input-number
            v-model="form.peerAmount"
            :min="0"
            :precision="2"
            controls-position="right"
            style="width: 160px"
            placeholder="单次抽奖需要金额"
          />
        </el-form-item>

        <el-form-item label="状态">
          <Segmented
            :modelValue="form.status !== 1"
            :options="statusOptions"
            @change="
              ({ option: { value } }) => {
                form.status = value ? 0 : 1;
              }
            "
          />
        </el-form-item>

        <el-form-item label="备注">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="2"
            placeholder="请输入备注"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">
            保存
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>
