<script setup lang="ts">
import { ref } from "vue";
import { useSysBanner } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";

defineOptions({
  name: "SystemBanner"
});

const formRef = ref();
const tableRef = ref();

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange,
  openDialog,
  handleDelete
} = useSysBanner(tableRef);
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="轮播图名称：" prop="bannerName">
        <el-input
          v-model="form.bannerName"
          placeholder="请输入轮播图名称"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="位置：" prop="position">
        <el-select
          v-model="form.position"
          placeholder="请选择位置"
          clearable
          class="!w-[130px]"
        >
          <el-option label="首页" value="home" />
          <el-option label="首页弹窗" value="popup" />
          <el-option label="活动页" value="activity" />
        </el-select>
      </el-form-item>
      <el-form-item label="平台：" prop="platform">
        <el-select
          v-model="form.platform"
          placeholder="请选择平台"
          clearable
          class="!w-[120px]"
        >
          <el-option label="全部" value="all" />
          <el-option label="Web" value="web" />
          <el-option label="App" value="app" />
          <el-option label="H5" value="h5" />
        </el-select>
      </el-form-item>
      <el-form-item label="跳转类型：" prop="jumpType">
        <el-select
          v-model="form.jumpType"
          placeholder="请选择类型"
          clearable
          class="!w-[130px]"
        >
          <el-option label="不跳转" value="none" />
          <el-option label="外部链接" value="url" />
          <el-option label="站内页面" value="internal" />
          <el-option label="商品" value="product" />
          <el-option label="活动" value="activity" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select
          v-model="form.status"
          placeholder="请选择状态"
          clearable
          class="!w-[110px]"
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

    <PureTableBar title="轮播图管理" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增轮播图
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
          :pagination="pagination"
          :header-cell-style="{
            background: 'var(--el-fill-color-light)',
            color: 'var(--el-text-color-primary)'
          }"
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
              :title="`是否确认删除轮播图 ${row.bannerName}?`"
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
      </template>
    </PureTableBar>
  </div>
</template>

<style scoped lang="scss">
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
