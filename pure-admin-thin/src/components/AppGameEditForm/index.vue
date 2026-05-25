<script setup lang="ts">
import { ref } from "vue";
import Segmented from "@/components/ReSegmented";
import { formRules } from "@/views/game/gameList/utils/rule";
import type { FormProps } from "@/views/game/gameList/utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "修改",
    gameId: 0,
    gameName: "",
    categoryCode: "",
    showIndex: 0,
    hot: 0,
    homeShow: 0,
    type: null,
    parentId: null,
    platformCode: "",
    thirdGameId: "",
    thirdGameName: "",
    thirdGameCategory: "",
    horizontalImage: "",
    gameIcon: "",
    typeIcon: "",
    typeActiveIcon: "",
    tenantId: null,
    sort: 0,
    disabledFlag: 0,
    deletedFlag: 0,
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

const gameTypeOptions = [
  { label: "拉霸/电子", value: 0 },
  { label: "Mini游戏", value: 1 },
  { label: "视讯游戏", value: 2 },
  { label: "捕鱼游戏", value: 3 },
  { label: "彩票游戏", value: 4 }
];

const yesNoOptions = [
  { label: "是", value: false },
  { label: "否", value: true }
];

const statusOptions = [
  { label: "启用", value: false },
  { label: "禁用", value: true }
];

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="120px"
  >
    <el-row :gutter="18">
      <el-col :span="12">
        <el-form-item label="游戏名称" prop="gameName">
          <el-input
            v-model="newFormInline.gameName"
            clearable
            placeholder="请输入游戏名称"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="平台编码" prop="platformCode">
          <el-input
            v-model="newFormInline.platformCode"
            clearable
            placeholder="如 hg"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="第三方游戏ID" prop="thirdGameId">
          <el-input
            v-model="newFormInline.thirdGameId"
            clearable
            placeholder="第三方游戏ID"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="第三方名称">
          <el-input
            v-model="newFormInline.thirdGameName"
            clearable
            placeholder="第三方游戏名称"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="厂商">
          <el-input
            v-model="newFormInline.thirdGameCategory"
            clearable
            placeholder="厂商"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="分类编码">
          <el-input
            v-model="newFormInline.categoryCode"
            clearable
            placeholder="如 slots / casino"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="游戏类型">
          <el-select
            v-model="newFormInline.type"
            clearable
            placeholder="请选择"
            class="!w-full"
          >
            <el-option
              v-for="item in gameTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="排序">
          <el-input-number
            v-model="newFormInline.sort"
            :min="0"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="展示序号">
          <el-input-number
            v-model="newFormInline.showIndex"
            :min="0"
            class="!w-full"
            controls-position="right"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="热门">
          <Segmented
            :modelValue="newFormInline.hot !== 1"
            :options="yesNoOptions"
            @change="
              ({ option: { value } }) => {
                newFormInline.hot = value ? 0 : 1;
              }
            "
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="首页展示">
          <Segmented
            :modelValue="newFormInline.homeShow !== 1"
            :options="yesNoOptions"
            @change="
              ({ option: { value } }) => {
                newFormInline.homeShow = value ? 0 : 1;
              }
            "
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="状态">
          <Segmented
            :modelValue="newFormInline.disabledFlag === 1"
            :options="statusOptions"
            @change="
              ({ option: { value } }) => {
                newFormInline.disabledFlag = value ? 1 : 0;
              }
            "
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="横版图">
          <el-input
            v-model="newFormInline.horizontalImage"
            clearable
            placeholder="横版图URL"
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="游戏图标">
          <el-input
            v-model="newFormInline.gameIcon"
            clearable
            placeholder="游戏图标URL"
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="备注">
          <el-input
            v-model="newFormInline.remark"
            placeholder="请输入备注"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
</template>
