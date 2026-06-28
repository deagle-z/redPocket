<script setup lang="ts">
import dayjs from "dayjs";
import { computed, onMounted, reactive, ref } from "vue";
import Refresh from "@iconify-icons/ep/refresh";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import {
  getTenantDashboardAgentRanks,
  getTenantDashboardOnlineUsers,
  getTenantDashboardRegisterUsers,
  getTenantDashboardRechargeUsers,
  getTenantDashboardRechargeOrders,
  getTenantDashboardWithdrawOrders,
  getTenantDashboardMonthlyBalances,
  getTenantDashboardStats,
  type TenantDashboardAgentRank,
  type TenantDashboardStats,
  type TenantDashboardUserDetail,
  type TenantDashboardOrderDetail,
  type TenantDashboardMonthlyBalance
} from "@/api/dashboard";

defineOptions({
  name: "Welcome"
});

const loading = ref(false);
const emptyPeriodStats = () => ({
  rechargeAmount: 0,
  betAmount: 0,
  withdrawAmount: 0,
  rebateAmount: 0,
  platformPumpAmount: 0,
  rechargeUsers: 0,
  repeatRechargeUsers: 0,
  registerUsers: 0
});

const stats = ref<TenantDashboardStats>({
  today: emptyPeriodStats(),
  yesterday: emptyPeriodStats(),
  month: emptyPeriodStats(),
  totalPlatformPumpAmount: 0,
  totalRegisterUsers: 0,
  onlineUsers: 0
});

type DetailType =
  | "online"
  | "todayRechargeUsers"
  | "monthRechargeUsers"
  | "todayRegisterUsers"
  | "yesterdayRegisterUsers"
  | "monthRegisterUsers"
  | "totalRegisterUsers";

const detailDialogVisible = ref(false);
const detailLoading = ref(false);
const detailType = ref<DetailType>("online");
const detailTitle = ref("");
const detailList = ref<TenantDashboardUserDetail[]>([]);
const detailPagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

type OrderType = "recharge" | "withdraw";
type OrderPeriod = "today" | "yesterday";

const orderDialogVisible = ref(false);
const orderLoading = ref(false);
const orderType = ref<OrderType>("recharge");
const orderPeriod = ref<OrderPeriod>("today");
const orderTotalAmount = ref(0);
const orderList = ref<TenantDashboardOrderDetail[]>([]);
const orderPagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

const monthlyBalanceDialogVisible = ref(false);
const monthlyBalanceLoading = ref(false);
const monthlyBalanceYear = ref(String(dayjs().year()));
const monthlyBalanceList = ref<TenantDashboardMonthlyBalance[]>([]);
const monthlyBalanceTotal = reactive({
  rechargeAmount: 0,
  withdrawAmount: 0,
  balanceAmount: 0
});

const agentRankDialogVisible = ref(false);
const agentRankLoading = ref(false);
const agentRankList = ref<TenantDashboardAgentRank[]>([]);
const agentRankPagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

const orderDialogTitle = computed(() =>
  orderType.value === "recharge" ? "充值总额明细" : "提现总额明细"
);
const isWithdrawOrder = computed(() => orderType.value === "withdraw");
const monthBalanceAmount = computed(
  () =>
    Number(stats.value.month.rechargeAmount || 0) -
    Number(stats.value.month.withdrawAmount || 0)
);

const formatAmount = (value?: number) => {
  return Number(value || 0).toLocaleString("zh-CN", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });
};

const formatDateTime = (value?: string | null) => {
  return value ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "-";
};

const formatNullable = (value?: string | null) => {
  return value && value !== "" ? value : "-";
};

const metricCards = computed(() => [
  {
    title: "当日充值总额",
    value: formatAmount(stats.value.today.rechargeAmount),
    unit: "",
    tone: "green",
    orderType: "recharge" as OrderType
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
    tone: "orange",
    orderType: "withdraw" as OrderType
  },
  {
    title: "当月提现总额",
    value: formatAmount(stats.value.month.withdrawAmount),
    unit: "",
    tone: "orange"
  },
  {
    title: "余额统计",
    value: formatAmount(monthBalanceAmount.value),
    unit: "",
    tone: "indigo",
    monthlyBalance: true
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
    title: "当日注册人数",
    value: String(stats.value.today.registerUsers || 0),
    unit: "人",
    tone: "purple",
    detailType: "todayRegisterUsers" as DetailType
  },
  {
    title: "昨日注册人数",
    value: String(stats.value.yesterday.registerUsers || 0),
    unit: "人",
    tone: "purple",
    detailType: "yesterdayRegisterUsers" as DetailType
  },
  {
    title: "当月注册人数",
    value: String(stats.value.month.registerUsers || 0),
    unit: "人",
    tone: "purple",
    detailType: "monthRegisterUsers" as DetailType
  },
  {
    title: "注册总人数",
    value: String(stats.value.totalRegisterUsers || 0),
    unit: "人",
    tone: "purple",
    detailType: "totalRegisterUsers" as DetailType
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
  },
  {
    title: "当日复充客户数",
    value: String(stats.value.today.repeatRechargeUsers || 0),
    unit: "人",
    tone: "slate"
  },
  {
    title: "当月复充客户数",
    value: String(stats.value.month.repeatRechargeUsers || 0),
    unit: "人",
    tone: "slate"
  }
]);

const detailPeriod = computed(() => {
  if (detailType.value.startsWith("yesterday")) return "yesterday";
  if (detailType.value.startsWith("month")) return "month";
  if (detailType.value.startsWith("total")) return "total";
  return "today";
});

const isRechargeDetail = computed(() =>
  detailType.value.includes("RechargeUsers")
);

const isRegisterDetail = computed(() =>
  detailType.value.includes("RegisterUsers")
);

const isOnlineDetail = computed(() => detailType.value === "online");
const detailPageCount = computed(() => {
  return Math.max(
    1,
    Math.ceil(detailPagination.total / detailPagination.pageSize)
  );
});
const agentRankPageCount = computed(() => {
  return Math.max(
    1,
    Math.ceil(agentRankPagination.total / agentRankPagination.pageSize)
  );
});

function getDetailExpectedTotal(type: DetailType) {
  if (type === "online") return stats.value.onlineUsers || 0;
  if (type === "todayRegisterUsers")
    return stats.value.today.registerUsers || 0;
  if (type === "yesterdayRegisterUsers") {
    return stats.value.yesterday.registerUsers || 0;
  }
  if (type === "monthRegisterUsers")
    return stats.value.month.registerUsers || 0;
  if (type === "totalRegisterUsers") return stats.value.totalRegisterUsers || 0;
  if (type === "todayRechargeUsers")
    return stats.value.today.rechargeUsers || 0;
  if (type === "monthRechargeUsers")
    return stats.value.month.rechargeUsers || 0;
  return 0;
}

async function loadStats() {
  loading.value = true;
  try {
    const res = await getTenantDashboardStats();
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
    let res;
    if (isOnlineDetail.value) {
      res = await getTenantDashboardOnlineUsers(payload);
    } else if (isRegisterDetail.value) {
      res = await getTenantDashboardRegisterUsers({
        ...payload,
        period: detailPeriod.value
      });
    } else {
      res = await getTenantDashboardRechargeUsers({
        ...payload,
        period: detailPeriod.value
      });
    }
    detailList.value = res.data?.list || [];
    detailPagination.total = Math.max(
      Number(res.data?.total ?? 0),
      getDetailExpectedTotal(detailType.value),
      detailList.value.length
    );
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
  detailPagination.total = getDetailExpectedTotal(item.detailType);
  detailDialogVisible.value = true;
  await loadDetail();
}

async function loadOrderDetail() {
  orderLoading.value = true;
  try {
    const payload = {
      currentPage: orderPagination.currentPage - 1,
      pageSize: orderPagination.pageSize,
      period: orderPeriod.value
    };
    const res =
      orderType.value === "recharge"
        ? await getTenantDashboardRechargeOrders(payload)
        : await getTenantDashboardWithdrawOrders(payload);
    orderList.value = res.data?.list || [];
    orderTotalAmount.value = Number(res.data?.totalAmount ?? 0);
    orderPagination.total = Number(res.data?.total ?? 0);
    orderPagination.pageSize = res.data?.pageSize || orderPagination.pageSize;
    orderPagination.currentPage = (res.data?.currentPage || 0) + 1;
  } finally {
    orderLoading.value = false;
  }
}

async function openOrderDetail(type: OrderType) {
  orderType.value = type;
  orderPeriod.value = "today";
  orderPagination.currentPage = 1;
  orderPagination.total = 0;
  orderTotalAmount.value = 0;
  orderDialogVisible.value = true;
  await loadOrderDetail();
}

async function handleOrderPeriodChange() {
  orderPagination.currentPage = 1;
  await loadOrderDetail();
}

function handleOrderSizeChange(size: number) {
  orderPagination.pageSize = size;
  orderPagination.currentPage = 1;
  loadOrderDetail();
}

function handleOrderCurrentChange(page: number) {
  orderPagination.currentPage = page;
  loadOrderDetail();
}

function handleCardClick(item: {
  title: string;
  detailType?: DetailType;
  orderType?: OrderType;
  monthlyBalance?: boolean;
}) {
  if (item.monthlyBalance) {
    openMonthlyBalance();
    return;
  }
  if (item.orderType) {
    openOrderDetail(item.orderType);
    return;
  }
  openDetail(item);
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

function handleDetailPrevPage() {
  if (detailLoading.value || detailPagination.currentPage <= 1) return;
  handleDetailCurrentChange(detailPagination.currentPage - 1);
}

function handleDetailNextPage() {
  if (
    detailLoading.value ||
    detailPagination.currentPage >= detailPageCount.value
  ) {
    return;
  }
  handleDetailCurrentChange(detailPagination.currentPage + 1);
}

async function loadMonthlyBalance() {
  monthlyBalanceLoading.value = true;
  try {
    const res = await getTenantDashboardMonthlyBalances(
      Number(monthlyBalanceYear.value)
    );
    monthlyBalanceList.value = res.data?.list || [];
    monthlyBalanceTotal.rechargeAmount = Number(
      res.data?.totalRechargeAmount ?? 0
    );
    monthlyBalanceTotal.withdrawAmount = Number(
      res.data?.totalWithdrawAmount ?? 0
    );
    monthlyBalanceTotal.balanceAmount = Number(
      res.data?.totalBalanceAmount ?? 0
    );
  } finally {
    monthlyBalanceLoading.value = false;
  }
}

async function openMonthlyBalance() {
  monthlyBalanceDialogVisible.value = true;
  await loadMonthlyBalance();
}

async function handleMonthlyBalanceYearChange() {
  await loadMonthlyBalance();
}

async function loadAgentRanks() {
  agentRankLoading.value = true;
  try {
    const res = await getTenantDashboardAgentRanks({
      currentPage: agentRankPagination.currentPage - 1,
      pageSize: agentRankPagination.pageSize
    });
    agentRankList.value = res.data?.list || [];
    agentRankPagination.total = Number(res.data?.total ?? 0);
    agentRankPagination.pageSize =
      res.data?.pageSize || agentRankPagination.pageSize;
    agentRankPagination.currentPage = (res.data?.currentPage || 0) + 1;
  } finally {
    agentRankLoading.value = false;
  }
}

async function openAgentRankDialog() {
  agentRankPagination.currentPage = 1;
  agentRankPagination.total = 0;
  agentRankDialogVisible.value = true;
  await loadAgentRanks();
}

function handleAgentRankSizeChange(size: number) {
  agentRankPagination.pageSize = size;
  agentRankPagination.currentPage = 1;
  loadAgentRanks();
}

function handleAgentRankCurrentChange(page: number) {
  agentRankPagination.currentPage = page;
  loadAgentRanks();
}

function handleAgentRankPrevPage() {
  if (agentRankLoading.value || agentRankPagination.currentPage <= 1) return;
  handleAgentRankCurrentChange(agentRankPagination.currentPage - 1);
}

function handleAgentRankNextPage() {
  if (
    agentRankLoading.value ||
    agentRankPagination.currentPage >= agentRankPageCount.value
  ) {
    return;
  }
  handleAgentRankCurrentChange(agentRankPagination.currentPage + 1);
}

onMounted(() => {
  loadStats();
});
</script>

<template>
  <div class="tenant-home">
    <div class="tenant-home__header">
      <div>
        <h1>首页</h1>
        <p>商户经营数据</p>
      </div>
      <div class="tenant-home__actions">
        <el-button type="primary" plain @click="openAgentRankDialog">
          代理排行
        </el-button>
        <el-button
          :icon="useRenderIcon(Refresh)"
          :loading="loading"
          @click="loadStats"
        >
          刷新
        </el-button>
      </div>
    </div>

    <el-skeleton :loading="loading" animated :rows="6">
      <div class="metric-grid">
        <div
          v-for="item in metricCards"
          :key="item.title"
          class="metric-card"
          :class="[
            `metric-card--${item.tone}`,
            (item.detailType || item.orderType || item.monthlyBalance) &&
              'metric-card--clickable'
          ]"
          @click="handleCardClick(item)"
        >
          <div class="metric-card__label">{{ item.title }}</div>
          <div class="metric-card__value">
            <span>{{ item.value }}</span>
            <em>{{ item.unit }}</em>
          </div>
        </div>
      </div>
    </el-skeleton>

    <el-dialog v-model="agentRankDialogVisible" title="代理排行" width="80%">
      <div class="detail-pager">
        <div class="detail-pager__meta">
          共 {{ agentRankPagination.total }} 条，第
          {{ agentRankPagination.currentPage }} / {{ agentRankPageCount }} 页
        </div>
        <div class="detail-pager__actions">
          <span>每页</span>
          <el-select
            v-model="agentRankPagination.pageSize"
            size="small"
            class="detail-pager__size"
            :disabled="agentRankLoading"
            @change="handleAgentRankSizeChange"
          >
            <el-option :value="10" label="10" />
            <el-option :value="20" label="20" />
            <el-option :value="50" label="50" />
            <el-option :value="100" label="100" />
          </el-select>
          <el-button
            size="small"
            :disabled="agentRankLoading || agentRankPagination.currentPage <= 1"
            @click="handleAgentRankPrevPage"
          >
            上一页
          </el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="
              agentRankLoading ||
              agentRankPagination.currentPage >= agentRankPageCount
            "
            @click="handleAgentRankNextPage"
          >
            下一页
          </el-button>
        </div>
      </div>

      <el-table :data="agentRankList" border stripe v-loading="agentRankLoading">
        <el-table-column label="排名" width="80">
          <template #default="{ $index }">
            {{
              (agentRankPagination.currentPage - 1) *
                agentRankPagination.pageSize +
              $index +
              1
            }}
          </template>
        </el-table-column>
        <el-table-column prop="uid" label="代理UID" min-width="110">
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
          <template #default="{ row }">
            {{ formatNullable(row.phone) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="subRechargeAmount"
          label="下级充值金额"
          min-width="140"
          sortable
        >
          <template #default="{ row }">
            {{ formatAmount(row.subRechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="subRechargeUsers"
          label="充值下级数"
          min-width="110"
        />
        <el-table-column
          prop="subRechargeCount"
          label="充值笔数"
          min-width="100"
        />
        <el-table-column
          prop="lastRechargeAt"
          label="最后充值时间"
          min-width="170"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.lastRechargeAt) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="detail-pagination">
        <el-pagination
          v-model:current-page="agentRankPagination.currentPage"
          v-model:page-size="agentRankPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :background="true"
          :disabled="agentRankLoading"
          :hide-on-single-page="false"
          layout="total, sizes, prev, pager, next, jumper"
          :total="agentRankPagination.total"
          @size-change="handleAgentRankSizeChange"
          @current-change="handleAgentRankCurrentChange"
        />
      </div>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" :title="detailTitle" width="76%">
      <div class="detail-pager">
        <div class="detail-pager__meta">
          共 {{ detailPagination.total }} 条，第
          {{ detailPagination.currentPage }} / {{ detailPageCount }} 页
        </div>
        <div class="detail-pager__actions">
          <span>每页</span>
          <el-select
            v-model="detailPagination.pageSize"
            size="small"
            class="detail-pager__size"
            :disabled="detailLoading"
            @change="handleDetailSizeChange"
          >
            <el-option :value="10" label="10" />
            <el-option :value="20" label="20" />
            <el-option :value="50" label="50" />
            <el-option :value="100" label="100" />
          </el-select>
          <el-button
            size="small"
            :disabled="detailLoading || detailPagination.currentPage <= 1"
            @click="handleDetailPrevPage"
          >
            上一页
          </el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="
              detailLoading || detailPagination.currentPage >= detailPageCount
            "
            @click="handleDetailNextPage"
          >
            下一页
          </el-button>
        </div>
      </div>

      <el-table :data="detailList" border stripe v-loading="detailLoading">
        <el-table-column prop="id" label="ID" width="90" />
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
          <template #default="{ row }">{{
            formatNullable(row.phone)
          }}</template>
        </el-table-column>
        <el-table-column
          v-if="isRechargeDetail"
          prop="parentUid"
          label="上级UID"
          min-width="110"
        >
          <template #default="{ row }">
            {{ formatNullable(row.parentUid) }}
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="余额" min-width="120">
          <template #default="{ row }">
            {{ formatAmount(row.balance) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="isRechargeDetail"
          prop="rechargeAmount"
          label="充值金额"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatAmount(row.rechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="isRechargeDetail"
          prop="rechargeCount"
          label="充值笔数"
          min-width="100"
        />
        <el-table-column
          v-if="isRechargeDetail"
          prop="lastRechargeAt"
          label="最后充值时间"
          min-width="170"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.lastRechargeAt) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="isOnlineDetail"
          prop="lastActiveAt"
          label="最后在线时间"
          min-width="170"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.lastActiveAt) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="isRegisterDetail"
          prop="registeredAt"
          label="注册时间"
          min-width="170"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.registeredAt) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="detail-pagination">
        <el-pagination
          v-model:current-page="detailPagination.currentPage"
          v-model:page-size="detailPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :background="true"
          :disabled="detailLoading"
          :hide-on-single-page="false"
          layout="total, sizes, prev, pager, next, jumper"
          :total="detailPagination.total"
          @size-change="handleDetailSizeChange"
          @current-change="handleDetailCurrentChange"
        />
      </div>
    </el-dialog>

    <el-dialog
      v-model="monthlyBalanceDialogVisible"
      title="余额统计"
      width="720px"
    >
      <div class="order-toolbar">
        <el-date-picker
          v-model="monthlyBalanceYear"
          type="year"
          value-format="YYYY"
          placeholder="选择年份"
          :clearable="false"
          :disabled="monthlyBalanceLoading"
          @change="handleMonthlyBalanceYearChange"
        />
        <div class="order-toolbar__total">
          年度结余：
          <strong>{{ formatAmount(monthlyBalanceTotal.balanceAmount) }}</strong>
        </div>
      </div>

      <el-table
        :data="monthlyBalanceList"
        border
        stripe
        v-loading="monthlyBalanceLoading"
      >
        <el-table-column prop="month" label="月份" min-width="120" />
        <el-table-column prop="rechargeAmount" label="充值金额" min-width="150">
          <template #default="{ row }">
            {{ formatAmount(row.rechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="withdrawAmount" label="提现金额" min-width="150">
          <template #default="{ row }">
            {{ formatAmount(row.withdrawAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="balanceAmount" label="结余" min-width="150">
          <template #default="{ row }">
            {{ formatAmount(row.balanceAmount) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="order-toolbar__total monthly-balance-total">
        <span
          >充值合计：{{
            formatAmount(monthlyBalanceTotal.rechargeAmount)
          }}</span
        >
        <span
          >提现合计：{{
            formatAmount(monthlyBalanceTotal.withdrawAmount)
          }}</span
        >
        <span
          >结余合计：{{ formatAmount(monthlyBalanceTotal.balanceAmount) }}</span
        >
      </div>
    </el-dialog>

    <el-dialog
      v-model="orderDialogVisible"
      :title="orderDialogTitle"
      width="80%"
    >
      <div class="order-toolbar">
        <el-radio-group
          v-model="orderPeriod"
          :disabled="orderLoading"
          @change="handleOrderPeriodChange"
        >
          <el-radio-button value="today">今天</el-radio-button>
          <el-radio-button value="yesterday">昨天</el-radio-button>
        </el-radio-group>
        <div class="order-toolbar__total">
          {{ orderPeriod === "today" ? "今日" : "昨日" }}合计：
          <strong>{{ formatAmount(orderTotalAmount) }}</strong>
          <span>（{{ orderPagination.total }} 笔）</span>
        </div>
      </div>

      <el-table :data="orderList" border stripe v-loading="orderLoading">
        <el-table-column prop="orderNo" label="订单号" min-width="200" />
        <el-table-column prop="uid" label="UID" min-width="110">
          <template #default="{ row }">{{ formatNullable(row.uid) }}</template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="120">
          <template #default="{ row }">
            {{ formatNullable(row.username) }}
          </template>
        </el-table-column>
        <el-table-column prop="firstName" label="昵称" min-width="120">
          <template #default="{ row }">
            {{ formatNullable(row.firstName) }}
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="金额" min-width="120">
          <template #default="{ row }">
            {{ formatAmount(row.amount) }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="isWithdrawOrder"
          prop="fee"
          label="手续费"
          min-width="110"
        >
          <template #default="{ row }">{{ formatAmount(row.fee) }}</template>
        </el-table-column>
        <el-table-column
          v-if="isWithdrawOrder"
          prop="netAmount"
          label="到账金额"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatAmount(row.netAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="channel" label="渠道" min-width="100">
          <template #default="{ row }">
            {{ formatNullable(row.channel) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="time"
          :label="isWithdrawOrder ? '打款时间' : '支付时间'"
          min-width="170"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.time) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="detail-pagination">
        <el-pagination
          v-model:current-page="orderPagination.currentPage"
          v-model:page-size="orderPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :background="true"
          :disabled="orderLoading"
          :hide-on-single-page="false"
          layout="total, sizes, prev, pager, next, jumper"
          :total="orderPagination.total"
          @size-change="handleOrderSizeChange"
          @current-change="handleOrderCurrentChange"
        />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.tenant-home {
  min-height: calc(100vh - 140px);
  padding: 24px;
  background: #f6f8fb;
}

.tenant-home__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.tenant-home__header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 650;
  line-height: 32px;
  color: #182230;
}

.tenant-home__header p {
  margin: 4px 0 0;
  font-size: 14px;
  color: #667085;
}

.tenant-home__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  justify-content: flex-end;
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

.metric-card--indigo {
  border-top: 3px solid #6172f3;
}

.detail-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.detail-pager {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.detail-pager__meta {
  color: #606266;
  font-size: 14px;
  line-height: 24px;
}

.detail-pager__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #606266;
  font-size: 14px;
}

.detail-pager__size {
  width: 86px;
}

.order-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.order-toolbar__total {
  color: #606266;
  font-size: 14px;
}

.order-toolbar__total strong {
  color: #101828;
  font-size: 18px;
}

.order-toolbar__total span {
  color: #98a2b3;
}

.monthly-balance-total {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
