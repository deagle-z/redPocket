<script setup lang="ts">
import { ref, onMounted } from "vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import {
  getUserLotteryRecords,
  type UserLotteryRecord
} from "@/api/userLotteryRecord";

import Refresh from "@iconify-icons/ep/refresh";
import Search from "@iconify-icons/ep/search";

defineOptions({ name: "SystemUserLotteryRecord" });

const tableRef = ref();
const loading = ref(false);
const dataList = ref<UserLotteryRecord[]>([]);
const total = ref(0);

const searchForm = ref({
  userId: undefined as number | undefined,
  poolId: undefined as number | undefined,
  status: null as number | null,
  currentPage: 0,
  pageSize: 20
});

const statusMap: Record<number, { label: string; type: string }> = {
  0: { label: "待结算", type: "warning" },
  1: { label: "已发放", type: "success" },
  2: { label: "未中奖", type: "info" }
};

const columns: TableColumnList = [
  { label: "ID", prop: "id", minWidth: 80 },
  { label: "用户ID", prop: "userId", minWidth: 100 },
  { label: "奖池ID", prop: "poolId", minWidth: 90 },
  {
    label: "消耗金额",
    prop: "peerAmount",
    minWidth: 100,
    formatter: ({ peerAmount }) => Number(peerAmount ?? 0).toFixed(2)
  },
  {
    label: "中奖金额",
    prop: "awardAmount",
    minWidth: 100,
    formatter: ({ awardAmount }) => Number(awardAmount ?? 0).toFixed(2)
  },
  {
    label: "抽奖前余额",
    prop: "beforeBalance",
    minWidth: 110,
    formatter: ({ beforeBalance }) => Number(beforeBalance ?? 0).toFixed(2)
  },
  {
    label: "抽奖后余额",
    prop: "afterBalance",
    minWidth: 110,
    formatter: ({ afterBalance }) => Number(afterBalance ?? 0).toFixed(2)
  },
  {
    label: "状态",
    prop: "status",
    minWidth: 90,
    cellRenderer: scope => (
      <el-tag
        size="small"
        type={statusMap[scope.row.status]?.type ?? ""}
        effect="plain"
      >
        {statusMap[scope.row.status]?.label ?? scope.row.status}
      </el-tag>
    )
  },
  { label: "备注", prop: "remark", minWidth: 120 },
  { label: "创建时间", prop: "createdAt", minWidth: 160 }
];

async function onSearch() {
  loading.value = true;
  try {
    const { data } = await getUserLotteryRecords({
      ...searchForm.value,
      status: searchForm.value.status ?? undefined
    });
    dataList.value = data?.list || [];
    total.value = data?.total || 0;
  } catch {
    message("获取抽奖记录失败", { type: "error" });
  } finally {
    loading.value = false;
    tableRef.value?.setAdaptive?.();
  }
}

function onPageChange(page: number) {
  searchForm.value.currentPage = page - 1;
  onSearch();
}

function onSizeChange(size: number) {
  searchForm.value.pageSize = size;
  searchForm.value.currentPage = 0;
  onSearch();
}

onMounted(() => onSearch());
</script>

<template>
  <div class="main">
    <PureTableBar title="抽奖记录" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-form inline class="flex flex-wrap gap-2">
          <el-form-item label="用户ID">
            <el-input-number
              v-model="searchForm.userId"
              :min="1"
              controls-position="right"
              style="width: 130px"
              placeholder="用户ID"
              clearable
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select
              v-model="searchForm.status"
              clearable
              placeholder="全部"
              style="width: 110px"
            >
              <el-option label="待结算" :value="0" />
              <el-option label="已发放" :value="1" />
              <el-option label="未中奖" :value="2" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :icon="useRenderIcon(Search)"
              @click="onSearch"
            >
              查询
            </el-button>
            <el-button :icon="useRenderIcon(Refresh)" @click="onSearch">
              刷新
            </el-button>
          </el-form-item>
        </el-form>
      </template>

      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          ref="tableRef"
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
          align-whole="center"
          showOverflowTooltip
          row-key="id"
          table-layout="auto"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="{
            total,
            pageSize: searchForm.pageSize,
            currentPage: searchForm.currentPage + 1,
            pageSizes: [20, 50, 100],
            onChange: onPageChange,
            onSizeChange
          }"
          :header-cell-style="{
            background: 'var(--el-fill-color-light)',
            color: 'var(--el-text-color-primary)'
          }"
        />
      </template>
    </PureTableBar>
  </div>
</template>
