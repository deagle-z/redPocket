<script setup lang="ts">
import { LineChart } from "echarts/charts";
import {
  GridComponent,
  LegendComponent,
  TooltipComponent
} from "echarts/components";
import { init, use, type EChartsType } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import {
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch
} from "vue";
import {
  getTenantDomainVisits,
  type DomainVisitHourly,
  type DomainVisitRow
} from "@/api/dashboard";

const visible = defineModel<boolean>({ required: true });

use([
  LineChart,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  CanvasRenderer
]);

const loading = ref(false);
const loadError = ref(false);
const date = ref("");
const pageViews = ref(0);
const uniqueVisitors = ref(0);
const list = ref<DomainVisitRow[]>([]);
const hourly = ref<DomainVisitHourly[]>([]);
const chartRef = ref<HTMLDivElement>();
const pagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});
let chart: EChartsType | undefined;

function clearResult() {
  date.value = "";
  pageViews.value = 0;
  uniqueVisitors.value = 0;
  list.value = [];
  hourly.value = [];
  pagination.total = 0;
  chart?.clear();
}

function renderChart() {
  if (!chartRef.value || hourly.value.length === 0) {
    chart?.clear();
    return;
  }
  chart ??= init(chartRef.value);
  chart.setOption({
    animationDuration: 280,
    tooltip: { trigger: "axis" },
    legend: { data: ["PV", "UV"] },
    grid: { left: 48, right: 24, top: 48, bottom: 42 },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: hourly.value.map(item => `${String(item.hour).padStart(2, "0")}:00`)
    },
    yAxis: { type: "value", minInterval: 1 },
    series: [
      {
        name: "PV",
        type: "line",
        smooth: true,
        symbolSize: 6,
        data: hourly.value.map(item => item.pageViews)
      },
      {
        name: "UV",
        type: "line",
        smooth: true,
        symbolSize: 6,
        data: hourly.value.map(item => item.uniqueVisitors)
      }
    ]
  });
}

async function loadVisits() {
  loading.value = true;
  loadError.value = false;
  clearResult();
  try {
    const res = await getTenantDomainVisits({
      currentPage: pagination.currentPage - 1,
      pageSize: pagination.pageSize
    });
    date.value = res.data.date;
    pageViews.value = res.data.pageViews;
    uniqueVisitors.value = res.data.uniqueVisitors;
    list.value = res.data.list;
    hourly.value = res.data.hourly;
    pagination.total = res.data.total;
    pagination.pageSize = res.data.pageSize;
    pagination.currentPage = res.data.currentPage + 1;
    await nextTick();
    renderChart();
  } catch {
    loadError.value = true;
  } finally {
    loading.value = false;
  }
}

function handleSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.currentPage = 1;
  void loadVisits();
}

function handleCurrentChange(page: number) {
  pagination.currentPage = page;
  void loadVisits();
}

function resizeChart() {
  chart?.resize();
}

watch(visible, isVisible => {
  if (!isVisible) return;
  pagination.currentPage = 1;
  void loadVisits();
});

onMounted(() => window.addEventListener("resize", resizeChart));
onBeforeUnmount(() => {
  window.removeEventListener("resize", resizeChart);
  chart?.dispose();
  chart = undefined;
});
</script>

<template>
  <el-dialog v-model="visible" title="今日域名访问明细" width="82%">
    <div class="domain-toolbar">
      <span>统计日期：{{ date || "-" }}</span>
      <span class="domain-scope">仅统计当前商户绑定域名</span>
    </div>

    <div class="domain-summary">
      <el-statistic title="今日访问量 PV" :value="pageViews" />
      <el-statistic title="今日访客数 UV" :value="uniqueVisitors" />
    </div>

    <div v-if="loadError" class="domain-error">
      <el-alert
        title="域名访问统计加载失败，请重试"
        type="error"
        show-icon
        :closable="false"
      />
      <el-button type="primary" plain @click="loadVisits">重新加载</el-button>
    </div>

    <div v-loading="loading" class="domain-chart-panel">
      <h3>24 小时访问趋势</h3>
      <div ref="chartRef" class="domain-chart" />
    </div>

    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="domain" label="访问域名" min-width="260" />
      <el-table-column
        prop="pageViews"
        label="今日 PV"
        min-width="120"
        sortable
      />
      <el-table-column
        prop="uniqueVisitors"
        label="今日 UV"
        min-width="120"
        sortable
      />
    </el-table>

    <div class="domain-pagination">
      <el-pagination
        v-model:current-page="pagination.currentPage"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="pagination.total"
        :disabled="loading"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </el-dialog>
</template>

<style scoped>
.domain-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  color: #667085;
  font-size: 14px;
}

.domain-scope {
  color: #2e90fa;
}

.domain-summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 220px));
  gap: 16px;
  margin-bottom: 16px;
}

.domain-chart-panel {
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.domain-error {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
}

.domain-error :deep(.el-alert) {
  flex: 1;
}

.domain-chart-panel h3 {
  margin: 0 0 8px;
  color: #182230;
  font-size: 16px;
}

.domain-chart {
  width: 100%;
  height: 300px;
}

.domain-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

@media (max-width: 768px) {
  .domain-toolbar {
    flex-direction: column;
  }

  .domain-summary {
    width: 100%;
  }
}
</style>
