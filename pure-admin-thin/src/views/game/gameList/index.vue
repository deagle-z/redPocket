<script setup lang="ts">
import { ref } from "vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import { useAppGame } from "./utils/hook";

import Refresh from "@iconify-icons/ep/refresh";
import EditPen from "@iconify-icons/ep/edit-pen";
import Download from "@iconify-icons/ep/download";

defineOptions({
  name: "GameList"
});

const formRef = ref();
const tableRef = ref();

const {
  form,
  syncForm,
  loading,
  syncLoading,
  columns,
  dataList,
  pagination,
  gameTypeOptions,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange,
  openDialog,
  handleSync
} = useAppGame(tableRef);
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
        <el-form-item label="游戏名称：" prop="gameName">
          <el-input
            v-model="form.gameName"
            placeholder="游戏名称"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="平台：" prop="platformCode">
          <el-input
            v-model="form.platformCode"
            placeholder="如 hg"
            clearable
            class="!w-[130px]"
          />
        </el-form-item>
        <el-form-item label="第三方ID：" prop="thirdGameId">
          <el-input
            v-model="form.thirdGameId"
            placeholder="第三方游戏ID"
            clearable
            class="!w-[170px]"
          />
        </el-form-item>
        <el-form-item label="厂商：" prop="thirdGameCategory">
          <el-input
            v-model="form.thirdGameCategory"
            placeholder="厂商"
            clearable
            class="!w-[130px]"
          />
        </el-form-item>
        <el-form-item label="类型：" prop="type">
          <el-select
            v-model="form.type"
            placeholder="请选择"
            clearable
            class="!w-[140px]"
          >
            <el-option
              v-for="item in gameTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="是否热门：" prop="hot">
          <el-select
            v-model="form.hot"
            placeholder="请选择"
            clearable
            class="!w-[120px]"
          >
            <el-option label="热门" :value="1" />
            <el-option label="普通" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类：" prop="categoryCode">
          <el-input
            v-model="form.categoryCode"
            placeholder="分类"
            clearable
            class="!w-[160px]"
          />
        </el-form-item>
        <el-form-item label="首页展示：" prop="homeShow">
          <el-select
            v-model="form.homeShow"
            placeholder="请选择"
            clearable
            class="!w-[120px]"
          >
            <el-option label="展示" :value="1" />
            <el-option label="不展示" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态：" prop="disabledFlag">
          <el-select
            v-model="form.disabledFlag"
            placeholder="请选择"
            clearable
            class="!w-[110px]"
          >
            <el-option label="启用" :value="0" />
            <el-option label="禁用" :value="1" />
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

      <PureTableBar title="游戏列表" :columns="columns" @refresh="onSearch">
        <template #buttons>
          <div class="sync-tools">
            <el-input
              v-model="syncForm.language"
              placeholder="语言"
              clearable
              class="!w-[90px]"
            />
            <el-input
              v-model="syncForm.platformCode"
              placeholder="平台"
              clearable
              class="!w-[100px]"
            />
            <el-button
              type="primary"
              :icon="useRenderIcon(Download)"
              :loading="syncLoading"
              @click="handleSync"
            >
              同步游戏
            </el-button>
          </div>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <pure-table
            ref="tableRef"
            adaptive
            :adaptiveConfig="{ offsetBottom: 108 }"
            border
            align-whole="center"
            showOverflowTooltip
            row-key="gameId"
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
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>
  </div>
</template>

<style scoped lang="scss">
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}

.sync-tools {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
