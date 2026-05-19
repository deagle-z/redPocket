<script setup lang="ts">
import { h, onMounted, ref } from "vue";
import dayjs from "dayjs";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import {
  getCheckInRecordsAdmin,
  type CheckInRecord
} from "@/api/checkInRecord";

import Refresh from "@iconify-icons/ep/refresh";
import Search from "@iconify-icons/ep/search";

defineOptions({ name: "SystemCheckInRecord" });

const tableRef = ref();
const loading = ref(false);
const dataList = ref<CheckInRecord[]>([]);
const total = ref(0);
const dateRange = ref<[string, string] | null>(null);

const searchForm = ref({
  userUid: "",
  currentPage: 0,
  pageSize: 20
});

const columns: TableColumnList = [
  { label: "ID", prop: "id", minWidth: 80 },
  { label: "用户UID", prop: "userUid", minWidth: 120 },
  {
    label: "签到日期",
    prop: "checkInDate",
    minWidth: 120,
    formatter: ({ checkInDate }) => checkInDate || "-"
  },
  {
    label: "累计第几次",
    prop: "checkInSeq",
    minWidth: 110,
    formatter: ({ checkInSeq }) => `第 ${Number(checkInSeq || 0)} 次`
  },
  {
    label: "奖励金币",
    prop: "rewardAmount",
    minWidth: 110,
    cellRenderer: scope =>
      h(
        "span",
        { class: "font-semibold text-amber-500" },
        `+${Number(scope.row.rewardAmount ?? 0).toFixed(2)}`
      )
  },
  {
    label: "签到前余额",
    prop: "beforeBalance",
    minWidth: 120,
    formatter: ({ beforeBalance }) => Number(beforeBalance ?? 0).toFixed(2)
  },
  {
    label: "签到后余额",
    prop: "afterBalance",
    minWidth: 120,
    formatter: ({ afterBalance }) => Number(afterBalance ?? 0).toFixed(2)
  },
  {
    label: "创建时间",
    prop: "createdAt",
    minWidth: 170,
    formatter: ({ createdAt }) =>
      createdAt ? dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss") : "-"
  }
];

function buildSearchPayload() {
  return {
    ...searchForm.value,
    userUid: searchForm.value.userUid.trim() || undefined,
    startDate: dateRange.value?.[0],
    endDate: dateRange.value?.[1]
  };
}

async function onSearch() {
  loading.value = true;
  try {
    const { data } = await getCheckInRecordsAdmin(buildSearchPayload());
    dataList.value = data?.list || [];
    total.value = data?.total || 0;
    searchForm.value.pageSize = data?.pageSize || searchForm.value.pageSize;
    searchForm.value.currentPage = data?.currentPage || 0;
  } catch (error) {
    console.error("获取签到记录失败", error);
    message("获取签到记录失败", { type: "error" });
  } finally {
    loading.value = false;
    tableRef.value?.setAdaptive?.();
  }
}

function resetSearch() {
  searchForm.value.userUid = "";
  searchForm.value.currentPage = 0;
  dateRange.value = null;
  onSearch();
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
    <PureTableBar title="签到记录" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-form inline class="flex flex-wrap gap-2">
          <el-form-item label="用户UID">
            <el-input
              v-model="searchForm.userUid"
              style="width: 160px"
              placeholder="请输入用户UID"
              clearable
              @keyup.enter="onSearch"
            />
          </el-form-item>
          <el-form-item label="签到日期">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              style="width: 260px"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :icon="useRenderIcon(Search)"
              @click="onSearch"
            >
              查询
            </el-button>
            <el-button :icon="useRenderIcon(Refresh)" @click="resetSearch">
              重置
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
