<script setup lang="ts">
import dayjs from "dayjs";
import { computed, h, onMounted, reactive, ref } from "vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import {
  getAppUserBetRecordList,
  type AppUserBetRecord
} from "@/api/appUserBetRecord";

import Refresh from "@iconify-icons/ep/refresh";
import Search from "@iconify-icons/ep/search";

defineOptions({
  name: "GameRebatRecord"
});

const tableRef = ref();
const loading = ref(false);
const dataList = ref<AppUserBetRecord[]>([]);
const defaultDateRange = (): [string, string] => [
  dayjs().startOf("day").format("YYYY-MM-DD HH:mm:ss"),
  dayjs().endOf("day").format("YYYY-MM-DD HH:mm:ss")
];
const dateRange = ref<[string, string] | null>(defaultDateRange());

const form = reactive({
  uid: undefined as number | undefined,
  userId: undefined as number | undefined,
  gameName: "",
  gameId: "",
  platformCode: "",
  roundId: "",
  traceId: "",
  roundEnd: undefined as number | undefined
});

const pagination = reactive({
  total: 0,
  pageSize: 20,
  currentPage: 0,
  pageSizes: [20, 50, 100],
  background: true
});

const summary = computed(() => {
  return dataList.value.reduce(
    (acc, item) => {
      acc.bet += Number(item.betAmount || 0);
      acc.win += Number(item.winAmount || 0);
      return acc;
    },
    { bet: 0, win: 0 }
  );
});

function formatMoney(value?: number | null) {
  return Number(value || 0).toFixed(2);
}

function formatTime(value?: string | null) {
  return value ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "-";
}

function timeToUnixSecond(value?: string) {
  if (!value) return undefined;
  return dayjs(value).unix();
}

const columns: TableColumnList = [
  { label: "ID", prop: "id", minWidth: 90 },
  {
    label: "UID",
    prop: "uid",
    minWidth: 120,
    formatter: ({ uid }) => uid ?? "-"
  },
  {
    label: "用户ID",
    prop: "userId",
    minWidth: 110,
    formatter: ({ userId }) => userId ?? "-"
  },
  {
    label: "平台",
    prop: "platformCode",
    minWidth: 100,
    formatter: ({ platformCode }) => platformCode || "-"
  },
  {
    label: "游戏ID",
    prop: "gameId",
    minWidth: 150,
    formatter: ({ gameId }) => gameId || "-"
  },
  {
    label: "游戏名称",
    prop: "gameName",
    minWidth: 190,
    formatter: ({ gameName }) => gameName || "-"
  },
  {
    label: "下注金额",
    prop: "betAmount",
    minWidth: 120,
    cellRenderer: ({ row }) =>
      h("span", { class: "amount amount--bet" }, formatMoney(row.betAmount))
  },
  {
    label: "派彩金额",
    prop: "winAmount",
    minWidth: 120,
    cellRenderer: ({ row }) =>
      h("span", { class: "amount amount--win" }, formatMoney(row.winAmount))
  },
  {
    label: "局号",
    prop: "roundId",
    minWidth: 180,
    formatter: ({ roundId }) => roundId || "-"
  },
  {
    label: "订单号",
    prop: "traceId",
    minWidth: 190,
    formatter: ({ traceId }) => traceId || "-"
  },
  {
    label: "对局状态",
    prop: "roundEnd",
    minWidth: 105,
    cellRenderer: ({ row }) =>
      h("el-tag", { type: row.roundEnd === 1 ? "success" : "warning" }, () =>
        row.roundEnd === 1 ? "已结束" : "进行中"
      )
  },
  {
    label: "下注时间",
    prop: "date",
    minWidth: 170,
    formatter: ({ date }) => formatTime(date)
  },
  {
    label: "备注",
    prop: "remark",
    minWidth: 220,
    formatter: ({ remark }) => remark || "-"
  }
];

function buildSearchPayload() {
  return {
    currentPage: pagination.currentPage,
    pageSize: pagination.pageSize,
    uid: form.uid || undefined,
    userId: form.userId || undefined,
    gameName: form.gameName.trim() || undefined,
    gameId: form.gameId.trim() || undefined,
    platformCode: form.platformCode.trim() || undefined,
    roundId: form.roundId.trim() || undefined,
    traceId: form.traceId.trim() || undefined,
    roundEnd: typeof form.roundEnd === "number" ? form.roundEnd : undefined,
    startTime: timeToUnixSecond(dateRange.value?.[0]),
    endTime: timeToUnixSecond(dateRange.value?.[1])
  };
}

async function fetchList() {
  loading.value = true;
  try {
    const { data } = await getAppUserBetRecordList(buildSearchPayload());
    dataList.value = data?.list || [];
    pagination.total = data?.total || 0;
    pagination.pageSize = data?.pageSize || pagination.pageSize;
    pagination.currentPage = data?.currentPage || 0;
  } catch (error) {
    console.error("获取下注记录失败", error);
    message("获取下注记录失败", { type: "error" });
  } finally {
    loading.value = false;
    tableRef.value?.setAdaptive?.();
  }
}

function onSearch() {
  pagination.currentPage = 0;
  fetchList();
}

function resetSearch() {
  Object.assign(form, {
    uid: undefined,
    userId: undefined,
    gameName: "",
    gameId: "",
    platformCode: "",
    roundId: "",
    traceId: "",
    roundEnd: undefined
  });
  dateRange.value = defaultDateRange();
  onSearch();
}

function handleSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.currentPage = 0;
  fetchList();
}

function handleCurrentChange(page: number) {
  pagination.currentPage = page - 1;
  fetchList();
}

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="rebat-record-page">
    <el-form
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="UID">
        <el-input
          v-model.number="form.uid"
          placeholder="用户UID"
          clearable
          class="!w-[140px]"
          @keyup.enter="onSearch"
        />
      </el-form-item>
      <el-form-item label="用户ID">
        <el-input
          v-model.number="form.userId"
          placeholder="用户ID"
          clearable
          class="!w-[130px]"
          @keyup.enter="onSearch"
        />
      </el-form-item>
      <el-form-item label="游戏名称">
        <el-input
          v-model="form.gameName"
          placeholder="游戏名称"
          clearable
          class="!w-[170px]"
          @keyup.enter="onSearch"
        />
      </el-form-item>
      <el-form-item label="游戏ID">
        <el-input
          v-model="form.gameId"
          placeholder="第三方游戏ID"
          clearable
          class="!w-[150px]"
          @keyup.enter="onSearch"
        />
      </el-form-item>
      <el-form-item label="平台">
        <el-input
          v-model="form.platformCode"
          placeholder="如 hg"
          clearable
          class="!w-[110px]"
          @keyup.enter="onSearch"
        />
      </el-form-item>
      <el-form-item label="局号">
        <el-input
          v-model="form.roundId"
          placeholder="roundId"
          clearable
          class="!w-[170px]"
          @keyup.enter="onSearch"
        />
      </el-form-item>
      <el-form-item label="订单号">
        <el-input
          v-model="form.traceId"
          placeholder="traceId / tid"
          clearable
          class="!w-[180px]"
          @keyup.enter="onSearch"
        />
      </el-form-item>
      <el-form-item label="对局状态">
        <el-select
          v-model="form.roundEnd"
          placeholder="全部"
          clearable
          class="!w-[120px]"
        >
          <el-option label="进行中" :value="0" />
          <el-option label="已结束" :value="1" />
        </el-select>
      </el-form-item>
      <el-form-item label="下注时间">
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          value-format="YYYY-MM-DD HH:mm:ss"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          class="!w-[340px]"
        />
      </el-form-item>
      <el-form-item>
        <el-button
          type="primary"
          :icon="useRenderIcon(Search)"
          :loading="loading"
          @click="onSearch"
        >
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetSearch">
          重置
        </el-button>
      </el-form-item>
    </el-form>

    <div class="summary-row">
      <span>当前页下注：{{ summary.bet.toFixed(2) }}</span>
      <span>当前页派奖：{{ summary.win.toFixed(2) }}</span>
    </div>

    <PureTableBar title="下注记录" :columns="columns" @refresh="fetchList">
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          ref="tableRef"
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
          border
          align-whole="center"
          showOverflowTooltip
          row-key="id"
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
        />
      </template>
    </PureTableBar>
  </div>
</template>

<style scoped lang="scss">
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}

.summary-row {
  display: flex;
  gap: 18px;
  align-items: center;
  padding: 10px 18px;
  margin-bottom: 8px;
  color: var(--el-text-color-regular);
  background: var(--el-bg-color);
  border-radius: 6px;
}

:deep(.amount) {
  font-weight: 600;
}

:deep(.amount--bet) {
  color: var(--el-color-danger);
}

:deep(.amount--win) {
  color: var(--el-color-success);
}
</style>
