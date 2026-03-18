<script setup lang="ts">
import { ref } from "vue";
import { useSysVipLevel } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";

defineOptions({
  name: "SystemSysVipLevel"
});

const formRef = ref();
const tableRef = ref();

const {
  loading,
  columns,
  dataList,
  onSearch,
  resetForm,
  openDialog,
  handleDelete
} = useSysVipLevel(tableRef);
</script>

<template>
  <div class="main">
    <div class="flex items-center justify-between mb-2 px-2">
      <span class="text-sm text-[var(--el-text-color-secondary)]">
        等级由低到高排列，等级数字越小越低（如 level=1 对应 VIP0）
      </span>
      <div class="flex gap-2">
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">
          刷新
        </el-button>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增等级
        </el-button>
      </div>
    </div>

    <PureTableBar title="VIP等级管理" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增等级
        </el-button>
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
          :pagination="{ hide: true }"
          :header-cell-style="{
            background: 'var(--el-fill-color-light)',
            color: 'var(--el-text-color-primary)'
          }"
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
            <el-popconfirm
              :title="`是否确认删除等级 ${row.levelName}?`"
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

<style scoped lang="scss">
</style>
