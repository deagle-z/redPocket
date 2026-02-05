<script setup lang="ts">
import { useTenant } from "./utils/hook";
import { ref } from "vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";

import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";

defineOptions({
  name: "SystemTenant"
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
  tenantTypeOptions,
  onSearch,
  resetForm,
  openDialog,
  handleDelete,
  handleSizeChange,
  handleCurrentChange,
  handleSelectionChange
} = useTenant(tableRef);
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
        <el-form-item label="租户编码：" prop="tenantCode">
          <el-input
            v-model="form.tenantCode"
            placeholder="租户唯一编码"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="租户名称：" prop="tenantName">
          <el-input
            v-model="form.tenantName"
            placeholder="租户名称"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="租户类型：" prop="tenantType">
          <el-select
            v-model="form.tenantType"
            placeholder="请选择"
            clearable
            class="!w-[180px]"
          >
            <el-option
              v-for="item in tenantTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态：" prop="status">
          <el-select
            v-model="form.status"
            placeholder="请选择"
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
        <el-form-item label="套餐标识：" prop="planCode">
          <el-input
            v-model="form.planCode"
            placeholder="free/pro/enterprise"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item>
          <!-- <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增租户
        </el-button> -->
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

      <div :class="['flex', 'w-full', deviceDetection() ? 'flex-wrap' : '']">
        <PureTableBar
          class="w-full"
          title="租户管理"
          :columns="columns"
          @refresh="onSearch"
        >
          <template #buttons>
            <el-button
              type="primary"
              :icon="useRenderIcon(AddFill)"
              @click="openDialog()"
            >
              新增租户
            </el-button>
          </template>
          <template v-slot="{ size, dynamicColumns }">
            <div class="w-full overflow-x-auto">
              <pure-table
                ref="tableRef"
                align-whole="center"
                show-overflow-tooltip="true"
                row-key="id"
                adaptive
                :adaptiveConfig="{ offsetBottom: 108 }"
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
                  <el-popconfirm
                    :title="`是否确认删除租户 ${row.tenantName} 这条数据`"
                    @confirm="handleDelete(row)"
                  >
                    <template #reference>
                      <el-button
                        class="reset-margin"
                        link
                        type="primary"
                        :size="size"
                        :icon="useRenderIcon(Delete)"
                      >
                        删除
                      </el-button>
                    </template>
                  </el-popconfirm>
                </template>
              </pure-table>
            </div>
          </template>
        </PureTableBar>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
:deep(.el-dropdown-menu__item i) {
  margin: 0;
}

:deep(.el-button:focus-visible) {
  outline: none;
}

.main-content {
  margin: 24px 24px 0 !important;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
