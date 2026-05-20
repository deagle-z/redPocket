<script setup lang="ts">
import { reactive, ref } from "vue";
import { useTgUser } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import { message } from "@/utils/message";
import {
  getTgUserList,
  getTgUserSubStatsSummary,
  setTgUserRebateRate,
  setTgUserRemark,
  type TgUser
} from "@/api/tgUser";

import Refresh from "@iconify-icons/ep/refresh";

defineOptions({
  name: "SystemTgUser"
});

const formRef = ref();
const tableRef = ref();

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  statusOptions,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange,
  handleSelectionChange,
  updateStatus
} = useTgUser(tableRef);

const subStatsDialogVisible = ref(false);
const subStatsLoading = ref(false);
const subStatsSummaryLoading = ref(false);
const currentUser = ref<TgUser | null>(null);
const subStatsList = ref<TgUser[]>([]);
const subStatsSummary = reactive({
  subRechargeAmount: 0,
  subFlowAmount: 0,
  subProfitAmount: 0,
  subWithdrawAmount: 0
});
const rebateRateDialogVisible = ref(false);
const rebateRateSaving = ref(false);
const rebateRateForm = reactive({
  id: 0,
  tgId: 0,
  username: "",
  firstName: "",
  rebateRate: 0
});
const remarkDialogVisible = ref(false);
const remarkSaving = ref(false);
const remarkForm = reactive({
  id: 0,
  tgId: 0,
  username: "",
  firstName: "",
  remark: ""
});
const subStatsPagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

function formatMoney(val?: number | null) {
  if (val === null || val === undefined || Number.isNaN(Number(val)))
    return "0";
  return String(val);
}

function formatName(row: {
  id?: number;
  username?: string | null;
  firstName?: string | null;
}) {
  return row.firstName || row.username || `ID:${row.id || "-"}`;
}

function openRebateRateDialog(row: TgUser) {
  rebateRateForm.id = row.id;
  rebateRateForm.tgId = row.tgId;
  rebateRateForm.username = row.username || "";
  rebateRateForm.firstName = row.firstName || "";
  rebateRateForm.rebateRate = Number(row.rebateRate ?? 0);
  rebateRateDialogVisible.value = true;
}

function openRemarkDialog(row: TgUser) {
  remarkForm.id = row.id;
  remarkForm.tgId = row.tgId;
  remarkForm.username = row.username || "";
  remarkForm.firstName = row.firstName || "";
  remarkForm.remark = row.remark || "";
  remarkDialogVisible.value = true;
}

async function submitRemark() {
  if (!remarkForm.id) return;
  const remark = String(remarkForm.remark || "").trim();
  if ([...remark].length > 255) {
    message("备注不能超过 255 个字符", { type: "warning" });
    return;
  }
  remarkSaving.value = true;
  try {
    await setTgUserRemark({
      id: remarkForm.id,
      remark
    });
    message("备注修改成功", { type: "success" });
    remarkDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("修改备注失败", error);
    message("修改备注失败", { type: "error" });
  } finally {
    remarkSaving.value = false;
  }
}

async function submitRebateRate() {
  if (!rebateRateForm.id) return;
  const rebateRate = Number(rebateRateForm.rebateRate);
  if (Number.isNaN(rebateRate) || rebateRate < 0 || rebateRate > 100) {
    message("返佣比例必须在 0 到 100 之间", { type: "warning" });
    return;
  }
  rebateRateSaving.value = true;
  try {
    await setTgUserRebateRate({
      id: rebateRateForm.id,
      rebateRate
    });
    message("返佣比例修改成功", { type: "success" });
    rebateRateDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("修改返佣比例失败", error);
    message("修改返佣比例失败", { type: "error" });
  } finally {
    rebateRateSaving.value = false;
  }
}

function buildQueryPayload() {
  const payload: Record<string, any> = {};
  payload.isBot = false;
  if (currentUser.value) payload.parentId = currentUser.value.id;
  return payload;
}

async function loadSubStatsSummary() {
  subStatsSummaryLoading.value = true;
  try {
    const { data } = await getTgUserSubStatsSummary(buildQueryPayload());
    subStatsSummary.subRechargeAmount = Number(data?.subRechargeAmount ?? 0);
    subStatsSummary.subFlowAmount = Number(data?.subFlowAmount ?? 0);
    subStatsSummary.subProfitAmount = Number(data?.subProfitAmount ?? 0);
    subStatsSummary.subWithdrawAmount = Number(data?.subWithdrawAmount ?? 0);
  } catch (error) {
    console.error("获取下级汇总失败", error);
    message("获取下级汇总失败", { type: "error" });
  } finally {
    subStatsSummaryLoading.value = false;
  }
}

async function loadSubStatsList() {
  if (!currentUser.value) return;
  subStatsLoading.value = true;
  try {
    const { data } = await getTgUserList({
      ...buildQueryPayload(),
      currentPage: subStatsPagination.currentPage - 1,
      pageSize: subStatsPagination.pageSize
    });
    subStatsList.value = data.list || [];
    subStatsPagination.total = data.total || 0;
    subStatsPagination.pageSize = data.pageSize || subStatsPagination.pageSize;
    subStatsPagination.currentPage = (data.currentPage || 0) + 1;
  } catch (error) {
    console.error("获取下级统计列表失败", error);
    message("获取下级统计列表失败", { type: "error" });
  } finally {
    subStatsLoading.value = false;
  }
}

async function openSubStatsDialog(row: TgUser) {
  currentUser.value = row;
  subStatsDialogVisible.value = true;
  subStatsPagination.currentPage = 1;
  await Promise.all([loadSubStatsSummary(), loadSubStatsList()]);
}

function handleSubStatsPageSizeChange(size: number) {
  subStatsPagination.pageSize = size;
  subStatsPagination.currentPage = 1;
  loadSubStatsList();
}

function handleSubStatsCurrentChange(page: number) {
  subStatsPagination.currentPage = page;
  loadSubStatsList();
}
</script>

<template>
  <div :class="['flex', 'justify-between', deviceDetection() && 'flex-wrap']">
    <div :class="[deviceDetection() ? ['w-full', 'mt-2'] : 'w-[calc(100%)]']">
      <el-form
        ref="formRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
      >
        <el-form-item label="用户UID：" prop="uid">
          <el-input
            v-model="form.uid"
            placeholder="请输入用户UID"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="用户名：" prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="昵称：" prop="firstName">
          <el-input
            v-model="form.firstName"
            placeholder="展示名"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="手机号：" prop="phone">
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="上级UID：" prop="parentUid">
          <el-input
            v-model="form.parentUid"
            placeholder="邀请人UID"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="邀请码：" prop="inviteCode">
          <el-input
            v-model="form.inviteCode"
            placeholder="邀请码"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="状态：" prop="status">
          <el-select
            v-model="form.status"
            placeholder="请选择"
            clearable
            class="!w-[180px]"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri:search-line')"
            :loading="loading"
            @click="onSearch"
          >
            搜索
          </el-button>
          <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar title="用户" :columns="columns" @refresh="onSearch">
        <template v-slot="{ size, dynamicColumns }">
          <pure-table
            ref="tableRef"
            border
            align-whole="center"
            showOverflowTooltip
            table-layout="auto"
            :loading="loading"
            :size="size"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="{
              ...pagination,
              size,
              currentPage: pagination.currentPage + 1
            }"
            :paginationSmall="size === 'small'"
            :header-cell-style="{
              background: 'var(--el-table-row-hover-bg-color)',
              color: 'var(--el-text-color-primary)'
            }"
            @page-size-change="handleSizeChange"
            @page-current-change="handleCurrentChange"
            @selection-change="handleSelectionChange"
          >
            <template #operation="{ row }">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                @click="openRebateRateDialog(row)"
              >
                修改返佣
              </el-button>
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                @click="openRemarkDialog(row)"
              >
                修改备注
              </el-button>
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                @click="openSubStatsDialog(row)"
              >
                下级统计
              </el-button>
              <el-button
                v-if="row.status !== 1"
                class="reset-margin"
                link
                type="primary"
                :size="size"
                @click="updateStatus(row, 1)"
              >
                启用
              </el-button>
              <el-button
                v-if="row.status === 1"
                class="reset-margin"
                link
                type="danger"
                :size="size"
                @click="updateStatus(row, 0)"
              >
                禁用
              </el-button>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>

    <el-dialog
      v-model="subStatsDialogVisible"
      :title="`下级统计汇总（用户ID: ${currentUser?.tgId ?? '-'}）`"
      width="78%"
      destroy-on-close
    >
      <el-skeleton :loading="subStatsSummaryLoading" animated :rows="2">
        <el-row :gutter="12" class="mb-3">
          <el-col :span="6">
            <el-card shadow="hover">
              <div class="stat-title">充值金额之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subRechargeAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <div class="stat-title">流水之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subFlowAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <div class="stat-title">盈利之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subProfitAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <div class="stat-title">提现金额之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subWithdrawAmount) }}
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-skeleton>

      <el-table v-loading="subStatsLoading" :data="subStatsList" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tgId" label="用户ID" min-width="140" />
        <el-table-column prop="uid" label="用户UID" min-width="120" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="firstName" label="昵称" min-width="120" />
        <el-table-column
          prop="subRechargeAmount"
          label="下级充值"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatMoney(row.subRechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="subFlowAmount" label="下级流水" min-width="120">
          <template #default="{ row }">
            {{ formatMoney(row.subFlowAmount) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="subProfitAmount"
          label="下级盈利"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatMoney(row.subProfitAmount) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="subWithdrawAmount"
          label="下级提现"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatMoney(row.subWithdrawAmount) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="dialog-pagination">
        <el-pagination
          v-model:current-page="subStatsPagination.currentPage"
          v-model:page-size="subStatsPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :small="deviceDetection()"
          :background="true"
          layout="total, sizes, prev, pager, next, jumper"
          :total="subStatsPagination.total"
          @size-change="handleSubStatsPageSizeChange"
          @current-change="handleSubStatsCurrentChange"
        />
      </div>
    </el-dialog>

    <el-dialog
      v-model="rebateRateDialogVisible"
      title="修改返佣比例"
      width="420px"
      destroy-on-close
    >
      <el-form :model="rebateRateForm" label-width="96px">
        <el-form-item label="用户">
          <span>{{ formatName(rebateRateForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ rebateRateForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="返佣比例">
          <el-input-number
            v-model="rebateRateForm.rebateRate"
            :min="0"
            :max="100"
            :precision="2"
            :step="1"
            controls-position="right"
            class="!w-full"
          />
          <div class="form-tip">单位：%，范围 0 - 100</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rebateRateDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="rebateRateSaving"
          @click="submitRebateRate"
        >
          保存
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="remarkDialogVisible"
      title="修改备注"
      width="460px"
      destroy-on-close
    >
      <el-form :model="remarkForm" label-width="96px">
        <el-form-item label="用户">
          <span>{{ formatName(remarkForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ remarkForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="remarkForm.remark"
            type="textarea"
            :rows="4"
            maxlength="255"
            show-word-limit
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="remarkDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="remarkSaving" @click="submitRemark">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}

.stat-title {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.stat-value {
  margin-top: 6px;
  font-size: 22px;
  font-weight: 600;
}

.dialog-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}

.form-tip {
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.3;
}
</style>
