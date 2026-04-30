<script setup lang="ts">
import dayjs from "dayjs";
import { computed, onMounted, reactive, ref } from "vue";
import Refresh from "@iconify-icons/ep/refresh";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import {
  getAdminDashboardOnlineUsers,
  getAdminDashboardRechargeUsers,
  getAdminDashboardStats,
  type AdminDashboardStats,
  type AdminDashboardUserDetail
} from "@/api/dashboard";

defineOptions({
  name: "Welcome"
});

const loading = ref(false);
const stats = ref<AdminDashboardStats>({
  today: {
    rechargeAmount: 0,
    betAmount: 0,
    withdrawAmount: 0,
    rebateAmount: 0,
    platformPumpAmount: 0,
    rechargeUsers: 0
  },
  month: {
    rechargeAmount: 0,
    betAmount: 0,
    withdrawAmount: 0,
    rebateAmount: 0,
    platformPumpAmount: 0,
    rechargeUsers: 0
  },
  totalPlatformPumpAmount: 0,
  onlineUsers: 0
});

type DetailType = "online" | "todayRechargeUsers" | "monthRechargeUsers";

const detailDialogVisible = ref(false);
const detailLoading = ref(false);
const detailType = ref<DetailType>("online");
const detailTitle = ref("");
const detailList = ref<AdminDashboardUserDetail[]>([]);
const detailPagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

const formatAmount = (value?: number) => {
  return Number(value || 0).toLocaleString("zh-CN", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });
};

const formatDateTime = (value?: string | null) => {
  return value ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "-";
};

const formatNullable = (value?: string | number | null) => {
  return value !== undefined && value !== null && value !== "" ? value : "-";
};

const metricCards = computed(() => [
  {
    title: "当日充值总额",
    value: formatAmount(stats.value.today.rechargeAmount),
    unit: "",
    tone: "green"
  },
  {
    title: "当月充值总额",
    value: formatAmount(stats.value.month.rechargeAmount),
    unit: "",
    tone: "green"
  },
  {
    title: "当日投注总额",
    value: formatAmount(stats.value.today.betAmount),
    unit: "",
    tone: "blue"
  },
  {
    title: "当月投注总额",
    value: formatAmount(stats.value.month.betAmount),
    unit: "",
    tone: "blue"
  },
  {
    title: "当日提现总额",
    value: formatAmount(stats.value.today.withdrawAmount),
    unit: "",
    tone: "orange"
  },
  {
    title: "当月提现总额",
    value: formatAmount(stats.value.month.withdrawAmount),
    unit: "",
    tone: "orange"
  },
  {
    title: "当月返佣总额",
    value: formatAmount(stats.value.month.rebateAmount),
    unit: "",
    tone: "cyan"
  },
  {
    title: "今日平台抽水",
    value: formatAmount(stats.value.today.platformPumpAmount),
    unit: "",
    tone: "red"
  },
  {
    title: "当月平台抽水",
    value: formatAmount(stats.value.month.platformPumpAmount),
    unit: "",
    tone: "red"
  },
  {
    title: "平台抽水总额",
    value: formatAmount(stats.value.totalPlatformPumpAmount),
    unit: "",
    tone: "red"
  },
  {
    title: "在线人数",
    value: String(stats.value.onlineUsers || 0),
    unit: "人",
    tone: "purple",
    detailType: "online" as DetailType
  },
  {
    title: "当日充值客户数",
    value: String(stats.value.today.rechargeUsers || 0),
    unit: "人",
    tone: "slate",
    detailType: "todayRechargeUsers" as DetailType
  },
  {
    title: "当月充值客户数",
    value: String(stats.value.month.rechargeUsers || 0),
    unit: "人",
    tone: "slate",
    detailType: "monthRechargeUsers" as DetailType
  }
]);

async function loadStats() {
  loading.value = true;
  try {
    const res = await getAdminDashboardStats();
    if (res?.data) {
      stats.value = res.data;
    }
  } finally {
    loading.value = false;
  }
}

async function loadDetail() {
  detailLoading.value = true;
  try {
    const payload = {
      currentPage: detailPagination.currentPage - 1,
      pageSize: detailPagination.pageSize
    };
    const res =
      detailType.value === "online"
        ? await getAdminDashboardOnlineUsers(payload)
        : await getAdminDashboardRechargeUsers({
            ...payload,
            period:
              detailType.value === "monthRechargeUsers" ? "month" : "today"
          });
    detailList.value = res.data?.list || [];
    detailPagination.total = res.data?.total || 0;
    detailPagination.pageSize = res.data?.pageSize || detailPagination.pageSize;
    detailPagination.currentPage = (res.data?.currentPage || 0) + 1;
  } finally {
    detailLoading.value = false;
  }
}

async function openDetail(item: { title: string; detailType?: DetailType }) {
  if (!item.detailType) {
    return;
  }
  detailType.value = item.detailType;
  detailTitle.value = item.title;
  detailPagination.currentPage = 1;
  detailDialogVisible.value = true;
  await loadDetail();
}

function handleDetailSizeChange(size: number) {
  detailPagination.pageSize = size;
  detailPagination.currentPage = 1;
  loadDetail();
}

function handleDetailCurrentChange(page: number) {
  detailPagination.currentPage = page;
  loadDetail();
}

onMounted(() => {
  loadStats();
});
</script>

<template>
  <div class="admin-home">
    <div class="admin-home__header">
      <div>
        <h1>首页</h1>
        <p>平台全部商户经营数据</p>
      </div>
      <el-button
        :icon="useRenderIcon(Refresh)"
        :loading="loading"
        @click="loadStats"
      >
        刷新
      </el-button>
    </div>

    <el-skeleton :loading="loading" animated :rows="6">
      <div class="metric-grid">
        <div
          v-for="item in metricCards"
          :key="item.title"
          class="metric-card"
          :class="[
            `metric-card--${item.tone}`,
            item.detailType && 'metric-card--clickable'
          ]"
          @click="openDetail(item)"
        >
          <div class="metric-card__label">{{ item.title }}</div>
          <div class="metric-card__value">
            <span>{{ item.value }}</span>
            <em>{{ item.unit }}</em>
          </div>
        </div>
      </div>
    </el-skeleton>

    <el-dialog v-model="detailDialogVisible" :title="detailTitle" width="76%">
      <el-table :data="detailList" border stripe v-loading="detailLoading">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="tenantId" label="商户ID" min-width="100">
          <template #default="{ row }">
            {{ formatNullable(row.tenantId) }}
          </template>
        </el-table-column>
        <el-table-column prop="uid" label="UID" min-width="110">
          <template #default="{ row }">{{ formatNullable(row.uid) }}</template>
        </el-table-column>
        <el-table-column prop="tgId" label="用户ID" min-width="130" />
        <el-table-column prop="username" label="用户名" min-width="130">
          <template #default="{ row }">
            {{ formatNullable(row.username) }}
          </template>
        </el-table-column>
        <el-table-column prop="firstName" label="昵称" min-width="130">
          <template #default="{ row }">
            {{ formatNullable(row.firstName) }}
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="130">
          <template #default="{ row }">{{ formatNullable(row.phone) }}</template>
        </el-table-column>
        <el-table-column prop="balance" label="余额" min-width="120">
          <template #default="{ row }">
            {{ formatAmount(row.balance) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="detailType !== 'online'"
          prop="rechargeAmount"
          label="充值金额"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatAmount(row.rechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="detailType !== 'online'"
          prop="rechargeCount"
          label="充值笔数"
          min-width="100"
        />
        <el-table-column
          v-if="detailType !== 'online'"
          prop="lastRechargeAt"
          label="最后充值时间"
          min-width="170"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.lastRechargeAt) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="detailType === 'online'"
          prop="lastActiveAt"
          label="最后在线时间"
          min-width="170"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.lastActiveAt) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="detail-pagination">
        <el-pagination
          v-model:current-page="detailPagination.currentPage"
          v-model:page-size="detailPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :background="true"
          layout="total, sizes, prev, pager, next, jumper"
          :total="detailPagination.total"
          @size-change="handleDetailSizeChange"
          @current-change="handleDetailCurrentChange"
        />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.admin-home {
  min-height: calc(100vh - 140px);
  padding: 24px;
  background: #f6f8fb;
}

.admin-home__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.admin-home__header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 650;
  line-height: 32px;
  color: #182230;
}

.admin-home__header p {
  margin: 4px 0 0;
  font-size: 14px;
  color: #667085;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}

.metric-card {
  min-height: 118px;
  padding: 18px;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 1px 2px rgb(16 24 40 / 4%);
}

.metric-card--clickable {
  cursor: pointer;
  transition:
    border-color 0.16s ease,
    box-shadow 0.16s ease,
    transform 0.16s ease;
}

.metric-card--clickable:hover {
  border-color: #b9c1cf;
  box-shadow: 0 8px 20px rgb(16 24 40 / 8%);
  transform: translateY(-1px);
}

.metric-card__label {
  font-size: 14px;
  line-height: 22px;
  color: #667085;
}

.metric-card__value {
  display: flex;
  gap: 8px;
  align-items: baseline;
  margin-top: 18px;
  color: #101828;
}

.metric-card__value span {
  min-width: 0;
  overflow-wrap: anywhere;
  font-size: 30px;
  font-weight: 700;
  line-height: 36px;
}

.metric-card__value em {
  font-size: 14px;
  font-style: normal;
  color: #98a2b3;
}

.metric-card--green {
  border-top: 3px solid #12b76a;
}

.metric-card--blue {
  border-top: 3px solid #2e90fa;
}

.metric-card--orange {
  border-top: 3px solid #f79009;
}

.metric-card--cyan {
  border-top: 3px solid #06aed4;
}

.metric-card--red {
  border-top: 3px solid #f04438;
}

.metric-card--purple {
  border-top: 3px solid #7a5af8;
}

.metric-card--slate {
  border-top: 3px solid #475467;
}

.detail-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
