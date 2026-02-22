<script setup lang="ts">
import { reactive, ref } from "vue";
import { useTgUser } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import { message } from "@/utils/message";
import {
  getTgUserListWithSubStats,
  getTgUserSubStatsSummary,
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
  subWithdrawAmount: 0
});
const subStatsPagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

function formatMoney(val?: number | null) {
  if (typeof val !== "number" || Number.isNaN(val)) return "0.000";
  return val.toFixed(3);
}

async function loadSubStatsSummary() {
  if (!currentUser.value) return;
  subStatsSummaryLoading.value = true;
  try {
    const { data } = await getTgUserSubStatsSummary({
      parentId: currentUser.value.id
    });
    subStatsSummary.subRechargeAmount = Number(data?.subRechargeAmount ?? 0);
    subStatsSummary.subFlowAmount = Number(data?.subFlowAmount ?? 0);
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
    const { data } = await getTgUserListWithSubStats({
      parentId: currentUser.value.id,
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
        <el-form-item label="TG用户ID：" prop="tgId">
          <el-input
            v-model.number="form.tgId"
            placeholder="请输入TG用户ID"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="用户名：" prop="username">
          <el-input
            v-model="form.username"
            placeholder="Telegram用户名"
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
        <el-form-item label="上级ID：" prop="parentId">
          <el-input
            v-model.number="form.parentId"
            placeholder="邀请人ID"
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

      <PureTableBar title="Telegram用户" :columns="columns" @refresh="onSearch">
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
      :title="`下级统计汇总（TG用户ID: ${currentUser?.tgId ?? '-'}）`"
      width="78%"
      destroy-on-close
    >
      <el-skeleton :loading="subStatsSummaryLoading" animated :rows="2">
        <el-row :gutter="12" class="mb-3">
          <el-col :span="8">
            <el-card shadow="hover">
              <div class="stat-title">充值金额之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subRechargeAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover">
              <div class="stat-title">流水之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subFlowAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover">
              <div class="stat-title">提现金额之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subWithdrawAmount) }}
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-skeleton>

      <el-table :data="subStatsList" border stripe v-loading="subStatsLoading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tgId" label="TG用户ID" min-width="140" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="firstName" label="昵称" min-width="120" />
        <el-table-column prop="subRechargeAmount" label="下级充值" min-width="120">
          <template #default="{ row }">
            {{ formatMoney(row.subRechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="subFlowAmount" label="下级流水" min-width="120">
          <template #default="{ row }">
            {{ formatMoney(row.subFlowAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="subWithdrawAmount" label="下级提现" min-width="120">
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
</style>
