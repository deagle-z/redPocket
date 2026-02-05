<script setup lang="ts">
import { ref } from "vue";
import { useWithdrawOrderBr } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";

import Refresh from "@iconify-icons/ep/refresh";
import Close from "@iconify-icons/ep/close";
import Check from "@iconify-icons/ep/check";

defineOptions({
  name: "SystemWithdrawOrderBr"
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
  approveOrder,
  rejectOrder
} = useWithdrawOrderBr(tableRef);
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
        <el-form-item label="用户ID：" prop="userId">
          <el-input
            v-model.number="form.userId"
            placeholder="请输入用户ID（可选）"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="订单号：" prop="orderNo">
          <el-input
            v-model="form.orderNo"
            placeholder="平台订单号"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="商户单号：" prop="merchantOrderNo">
          <el-input
            v-model="form.merchantOrderNo"
            placeholder="商户订单号"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="状态：" prop="status">
          <el-select
            v-model="form.status"
            placeholder="请选择状态"
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
        <el-form-item label="渠道：" prop="channel">
          <el-input
            v-model="form.channel"
            placeholder="提现渠道"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="子渠道：" prop="payMethod">
          <el-input
            v-model="form.payMethod"
            placeholder="支付方式"
            clearable
            class="!w-[180px]"
          />
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

      <PureTableBar title="提现记录（巴西）" :columns="columns" @refresh="onSearch">
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
                v-if="row.status === 0"
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(Check)"
                @click="approveOrder(row)"
              >
                审核通过
              </el-button>
              <el-button
                v-if="row.status === 0"
                class="reset-margin"
                link
                type="danger"
                :size="size"
                :icon="useRenderIcon(Close)"
                @click="rejectOrder(row)"
              >
                驳回
              </el-button>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>
  </div>
</template>
