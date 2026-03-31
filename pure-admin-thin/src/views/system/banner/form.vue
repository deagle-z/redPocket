<script setup lang="ts">
import { ref } from "vue";
import Segmented from "@/components/ReSegmented";
import { formRules } from "./utils/rule";
import type { FormProps } from "./utils/types";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "新增",
    id: 0,
    tenantId: 0,
    bannerName: "",
    position: "home",
    platform: "all",
    imageUrl: "",
    thumbUrl: "",
    jumpType: "none",
    jumpValue: "",
    sort: 0,
    status: 1,
    startTime: "",
    endTime: "",
    remark: ""
  })
});

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

const statusOptions = [
  { label: "启用", value: false },
  { label: "停用", value: true }
];

const positionOptions = [
  { label: "首页", value: "home" },
  { label: "首页弹窗", value: "popup" },
  { label: "活动页", value: "activity" }
];

const platformOptions = [
  { label: "全部", value: "all" },
  { label: "Web", value: "web" },
  { label: "App", value: "app" },
  { label: "H5", value: "h5" }
];

const jumpTypeOptions = [
  { label: "不跳转", value: "none" },
  { label: "外部链接", value: "url" },
  { label: "站内页面", value: "internal" },
  { label: "商品", value: "product" },
  { label: "活动", value: "activity" }
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
    label-width="110px"
  >
    <el-row :gutter="18">
      <el-col :span="12">
        <el-form-item label="轮播图名称" prop="bannerName">
          <el-input
            v-model="newFormInline.bannerName"
            clearable
            placeholder="请输入轮播图名称"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="位置" prop="position">
          <el-select
            v-model="newFormInline.position"
            class="!w-full"
            placeholder="请选择位置"
          >
            <el-option
              v-for="item in positionOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="平台" prop="platform">
          <el-select
            v-model="newFormInline.platform"
            class="!w-full"
            placeholder="请选择平台"
          >
            <el-option
              v-for="item in platformOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="图片地址" prop="imageUrl">
          <el-input
            v-model="newFormInline.imageUrl"
            clearable
            placeholder="请输入图片URL"
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="缩略图地址">
          <el-input
            v-model="newFormInline.thumbUrl"
            clearable
            placeholder="请输入缩略图URL"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="跳转类型" prop="jumpType">
          <el-select
            v-model="newFormInline.jumpType"
            class="!w-full"
            placeholder="请选择跳转类型"
          >
            <el-option
              v-for="item in jumpTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="跳转值">
          <el-input
            v-model="newFormInline.jumpValue"
            clearable
            placeholder="请输入跳转值"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="开始时间">
          <el-date-picker
            v-model="newFormInline.startTime"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="请选择开始时间"
            class="!w-full"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="结束时间">
          <el-date-picker
            v-model="newFormInline.endTime"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="请选择结束时间"
            class="!w-full"
          />
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
      <el-col v-if="newFormInline.title === '修改'" :span="12">
        <el-form-item label="状态">
          <Segmented
            :modelValue="newFormInline.status !== 1"
            :options="statusOptions"
            @change="
              ({ option: { value } }) => {
                newFormInline.status = value ? 0 : 1;
              }
            "
          />
        </el-form-item>
      </el-col>
      <el-col :span="24">
        <el-form-item label="备注">
          <el-input
            v-model="newFormInline.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
</template>
