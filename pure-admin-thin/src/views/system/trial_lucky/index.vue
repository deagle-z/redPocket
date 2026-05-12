<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { getTrialLuckyListAdmin, type TrialLuckyMoney } from "@/api/trialLucky";

defineOptions({
  name: "TrialLuckyRecord"
});

const activeName = ref("lucky");
const luckyLoading = ref(false);
const luckyList = ref<TrialLuckyMoney[]>([]);

const luckySearch = reactive({
  senderId: undefined as number | undefined,
  status: undefined as number | undefined,
  currentPage: 0,
  pageSize: 10,
  total: 0
});

function formatMoney(value?: number) {
  return Number(value || 0).toFixed(2);
}

async function loadLucky() {
  luckyLoading.value = true;
  try {
    const { data } = await getTrialLuckyListAdmin(luckySearch);
    luckyList.value = data?.list || [];
    luckySearch.total = data?.total || 0;
  } finally {
    luckyLoading.value = false;
  }
}

function resetLucky() {
  luckySearch.senderId = undefined;
  luckySearch.status = undefined;
  luckySearch.currentPage = 0;
  loadLucky();
}

onMounted(() => {
  loadLucky();
});
</script>

<template>
  <div class="main">
    <el-card shadow="never">
      <el-tabs v-model="activeName">
        <el-tab-pane label="试玩红包" name="lucky">
          <el-form :inline="true">
            <el-form-item label="发送者ID">
              <el-input-number
                v-model="luckySearch.senderId"
                :min="1"
                controls-position="right"
              />
            </el-form-item>
            <el-form-item label="状态">
              <el-select
                v-model="luckySearch.status"
                clearable
                placeholder="全部"
                style="width: 130px"
              >
                <el-option label="进行中" :value="1" />
                <el-option label="已完成" :value="2" />
                <el-option label="已过期" :value="3" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="loadLucky">查询</el-button>
              <el-button @click="resetLucky">重置</el-button>
            </el-form-item>
          </el-form>
          <el-table v-loading="luckyLoading" :data="luckyList" border>
            <el-table-column prop="id" label="ID" width="90" />
            <el-table-column prop="senderName" label="发送者" min-width="150" />
            <el-table-column prop="senderType" label="类型" width="90" />
            <el-table-column label="金额" width="120">
              <template #default="{ row }">{{
                formatMoney(row.amount)
              }}</template>
            </el-table-column>
            <el-table-column label="已抢" width="120">
              <template #default="{ row }">{{
                formatMoney(row.received)
              }}</template>
            </el-table-column>
            <el-table-column prop="number" label="个数" width="80" />
            <el-table-column prop="thunder" label="雷号" width="80" />
            <el-table-column label="玩法" width="90">
              <template #default="{ row }">{{
                row.gameMode === 1 ? "单双" : "雷号"
              }}</template>
            </el-table-column>
            <el-table-column prop="loseRate" label="倍率" width="90" />
            <el-table-column prop="status" label="状态" width="90" />
            <el-table-column
              prop="createdAt"
              label="创建时间"
              min-width="180"
            />
          </el-table>
          <el-pagination
            class="mt-4"
            :current-page="luckySearch.currentPage + 1"
            :page-size="luckySearch.pageSize"
            :total="luckySearch.total"
            layout="total, sizes, prev, pager, next"
            @current-change="
              page => {
                luckySearch.currentPage = page - 1;
                loadLucky();
              }
            "
            @size-change="
              size => {
                luckySearch.pageSize = size;
                luckySearch.currentPage = 0;
                loadLucky();
              }
            "
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>
