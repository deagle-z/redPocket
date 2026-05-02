<script setup lang="ts">
import { ref } from "vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import { usePlatformProfitLedger } from "./utils/hook";

import Refresh from "@iconify-icons/ep/refresh";

defineOptions({
  name: "SystemPlatformProfitLedger"
});

const formRef = ref();
const tableRef = ref();

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  sourceTypeOptions,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange,
  handleSelectionChange
} = usePlatformProfitLedger(tableRef);
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
          <el-input-number
            v-model="form.userId"
            :min="1"
            :controls="false"
            placeholder="用户ID"
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="来源类型：" prop="sourceType">
          <el-select
            v-model="form.sourceType"
            placeholder="请选择"
            clearable
            class="!w-[180px]"
          >
            <el-option
              v-for="item in sourceTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="来源单号：" prop="sourceId">
          <el-input
            v-model="form.sourceId"
            placeholder="来源单号"
            clearable
            class="!w-[240px]"
          />
        </el-form-item>
        <el-form-item label="最小净额：" prop="minNet">
          <el-input-number
            v-model="form.minNet"
            :controls="false"
            placeholder="最小净额"
            class="!w-[160px]"
          />
        </el-form-item>
        <el-form-item label="最大净额：" prop="maxNet">
          <el-input-number
            v-model="form.maxNet"
            :controls="false"
            placeholder="最大净额"
            class="!w-[160px]"
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

      <PureTableBar title="平台抽水记录" :columns="columns" @refresh="onSearch">
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
          />
        </template>
      </PureTableBar>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
