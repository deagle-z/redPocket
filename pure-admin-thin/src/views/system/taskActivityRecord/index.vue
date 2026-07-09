<script setup lang="ts">
import { ref } from "vue";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { PureTableBar } from "@/components/RePureTableBar";
import { useTaskActivityRecord } from "./utils/hook";

import Refresh from "@iconify-icons/ep/refresh";

defineOptions({ name: "SystemTaskActivityRecord" });

const formRef = ref();
const tableRef = ref();

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  levelOptions,
  recordStatusOptions,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange
} = useTaskActivityRecord(tableRef);
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="用户UID" prop="uid">
        <el-input
          v-model="form.uid"
          placeholder="请输入用户UID"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="配置ID" prop="configId">
        <el-input-number
          v-model="form.configId"
          :min="1"
          :precision="0"
          controls-position="right"
          class="!w-[140px]"
        />
      </el-form-item>
      <el-form-item label="任务标题" prop="title">
        <el-input
          v-model="form.title"
          placeholder="请输入任务标题"
          clearable
          class="!w-[200px]"
        />
      </el-form-item>
      <el-form-item label="任务等级" prop="levelCode">
        <el-select
          v-model="form.levelCode"
          placeholder="全部"
          clearable
          class="!w-[160px]"
        >
          <el-option
            v-for="item in levelOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select
          v-model="form.status"
          placeholder="全部"
          clearable
          class="!w-[140px]"
        >
          <el-option
            v-for="item in recordStatusOptions"
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

    <PureTableBar title="任务活动记录" :columns="columns" @refresh="onSearch">
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
        />
      </template>
    </PureTableBar>
  </div>
</template>

<style lang="scss" scoped>
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}

:deep(.task-progress-cell) {
  display: grid;
  grid-template-columns: 52px 1fr;
  gap: 8px;
  align-items: center;
}
</style>
