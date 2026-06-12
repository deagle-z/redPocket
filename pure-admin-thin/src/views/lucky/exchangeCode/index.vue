<script setup lang="ts">
import { ref } from "vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { useExchangeCode } from "./utils/hook";

import AddFill from "@iconify-icons/ri/add-circle-line";
import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import Download from "@iconify-icons/ep/download";

defineOptions({ name: "LuckyExchangeCode" });

const formRef = ref();
const tableRef = ref();

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  multipleSelection,
  onSearch,
  resetForm,
  openDialog,
  handleDelete,
  handleToggleStatus,
  handleSizeChange,
  handleCurrentChange,
  handleSelectionChange,
  handleExport,
  handleBatchDelete
} = useExchangeCode(tableRef);
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="兑换码" prop="code">
        <el-input
          v-model="form.code"
          maxlength="6"
          clearable
          placeholder="请输入兑换码"
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select
          v-model="form.status"
          clearable
          placeholder="全部"
          class="!w-[140px]"
        >
          <el-option label="启用" :value="1" />
          <el-option label="停用" :value="0" />
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

    <PureTableBar title="兑换码管理" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增兑换码
        </el-button>
        <el-button
          type="success"
          :icon="useRenderIcon(Download)"
          :disabled="!multipleSelection.length"
          @click="handleExport"
        >
          导出选中{{
            multipleSelection.length ? `(${multipleSelection.length})` : ""
          }}
        </el-button>
        <el-button
          type="danger"
          :icon="useRenderIcon(Delete)"
          :disabled="!multipleSelection.length"
          @click="handleBatchDelete"
        >
          批量删除{{
            multipleSelection.length ? `(${multipleSelection.length})` : ""
          }}
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          ref="tableRef"
          row-key="id"
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
            currentPage: pagination.currentPage
          }"
          :paginationSmall="size === 'small'"
          :header-cell-style="{
            background: 'var(--el-fill-color-light)',
            color: 'var(--el-text-color-primary)'
          }"
          @selection-change="handleSelectionChange"
          @page-size-change="handleSizeChange"
          @page-current-change="handleCurrentChange"
        >
          <template #operation="{ row }">
            <el-button
              class="reset-margin"
              link
              type="primary"
              :size="size"
              :icon="useRenderIcon(EditPen)"
              @click="openDialog('修改', row)"
            >
              修改
            </el-button>
            <el-button
              class="reset-margin"
              link
              :type="row.status === 1 ? 'warning' : 'success'"
              :size="size"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? "停用" : "启用" }}
            </el-button>
            <el-popconfirm
              :title="`确认删除兑换码 [${row.code}] 吗？`"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button
                  class="reset-margin"
                  link
                  type="danger"
                  :size="size"
                  :icon="useRenderIcon(Delete)"
                >
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </pure-table>
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
</style>
