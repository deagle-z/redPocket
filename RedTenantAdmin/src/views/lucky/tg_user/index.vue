<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useTgUser } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import { ElMessageBox } from "element-plus";
import { message } from "@/utils/message";
import { BACKSTAGE_DISPLAY_CURRENCY } from "@/utils/currency";
import {
  getSysCountryList,
  getTenantCountryRechargeInfo,
  type AppRechargeChannelItem,
  type SysCountry
} from "@/api/country";
import {
  addTgUserRebateAmount,
  createTgUserRechargeOrderV2,
  getTgUserList,
  getTgUserSubStatsSummary,
  setTgUserRebateRate,
  setTgUserRebateType,
  setTgUserRechargeRebateRates,
  setTgUserRemark,
  type TgUser,
  type TgUserRechargeOrderAppBack
} from "@/api/tgUser";

import Refresh from "@iconify-icons/ep/refresh";

defineOptions({
  name: "SystemTgUser"
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
  updateStatus,
  updateRebateWithdrawDisabled
} = useTgUser(tableRef);

const subStatsDialogVisible = ref(false);
const subStatsLoading = ref(false);
const subStatsSummaryLoading = ref(false);
const currentUser = ref<TgUser | null>(null);
const subStatsList = ref<TgUser[]>([]);
const subStatsSummary = reactive({
  subRechargeAmount: 0,
  subFlowAmount: 0,
  subProfitAmount: 0,
  subWithdrawAmount: 0,
  rechargeUsers: 0,
  validUsers: 0
});
const rebateRateDialogVisible = ref(false);
const rebateRateSaving = ref(false);
const rebateRateForm = reactive({
  id: 0,
  tgId: 0,
  username: "",
  firstName: "",
  rebateRate: 0
});
const rebateTypeDialogVisible = ref(false);
const rebateTypeSaving = ref(false);
const rebateTypeForm = reactive({
  id: 0,
  tgId: 0,
  username: "",
  firstName: "",
  rebateType: 1
});
const rechargeRatesDialogVisible = ref(false);
const rechargeRatesSaving = ref(false);
const rechargeRatesForm = reactive({
  id: 0,
  tgId: 0,
  username: "",
  firstName: "",
  useDefault: true,
  r1: 40,
  r2: 45,
  r3: 50
});
const rebateAmountDialogVisible = ref(false);
const rebateAmountSaving = ref(false);
const rebateAmountForm = reactive({
  id: 0,
  tgId: 0,
  username: "",
  firstName: "",
  currentAmount: 0,
  amount: 0
});
const remarkDialogVisible = ref(false);
const remarkSaving = ref(false);
const remarkForm = reactive({
  id: 0,
  tgId: 0,
  username: "",
  firstName: "",
  remark: ""
});
type RechargeFieldOption = {
  label: string;
  value: string;
};

type RechargeField = {
  fieldKey: string;
  fieldLabel: string;
  fieldPlaceholder?: string | null;
  fieldType?: string | null;
  dataType?: string | null;
  isRequired?: number | boolean | null;
  defaultValue?: string | null;
  minLength?: number | null;
  maxLength?: number | null;
  regexRule?: string | null;
  errorTips?: string | null;
  optionsJson?: string | null;
};

const rechargeDialogVisible = ref(false);
const rechargeConfigLoading = ref(false);
const rechargeSaving = ref(false);
const rechargeCountries = ref<SysCountry[]>([]);
const rechargeChannels = ref<AppRechargeChannelItem[]>([]);
const rechargeFields = ref<RechargeField[]>([]);
const rechargeFieldValues = reactive<Record<string, string>>({});
const rechargeResult = ref<TgUserRechargeOrderAppBack | null>(null);
const rechargeMinAmount = ref(1);
const rechargeForm = reactive({
  userId: 0,
  tgId: 0,
  username: "",
  firstName: "",
  amount: 100,
  countryCode: "",
  currency: "",
  channel: "",
  payMethod: "",
  merchantOrderNo: ""
});
const rechargeDisplayCurrency = computed(() =>
  rechargeForm.currency ? BACKSTAGE_DISPLAY_CURRENCY : ""
);
const subStatsPagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});
const selectedRechargeChannel = computed(
  () =>
    rechargeChannels.value.find(
      channel => channel.channelCode === rechargeForm.channel
    ) || null
);
const availablePayMethods = computed(
  () => selectedRechargeChannel.value?.methods || []
);

function formatMoney(val?: number | null) {
  if (val === null || val === undefined || Number.isNaN(Number(val)))
    return "0";
  return String(val);
}

function formatCount(val?: number | null) {
  if (val === null || val === undefined || Number.isNaN(Number(val)))
    return "0";
  return String(Math.trunc(Number(val)));
}

function formatName(row: {
  id?: number;
  username?: string | null;
  firstName?: string | null;
}) {
  return row.firstName || row.username || `ID:${row.id || "-"}`;
}

function openRebateRateDialog(row: TgUser) {
  rebateRateForm.id = row.id;
  rebateRateForm.tgId = row.tgId;
  rebateRateForm.username = row.username || "";
  rebateRateForm.firstName = row.firstName || "";
  rebateRateForm.rebateRate = Number(row.rebateRate ?? 0);
  rebateRateDialogVisible.value = true;
}

function openRebateTypeDialog(row: TgUser) {
  rebateTypeForm.id = row.id;
  rebateTypeForm.tgId = row.tgId;
  rebateTypeForm.username = row.username || "";
  rebateTypeForm.firstName = row.firstName || "";
  rebateTypeForm.rebateType = Number(row.rebateType ?? 1) === 2 ? 2 : 1;
  rebateTypeDialogVisible.value = true;
}

function openRechargeRatesDialog(row: TgUser) {
  rechargeRatesForm.id = row.id;
  rechargeRatesForm.tgId = row.tgId;
  rechargeRatesForm.username = row.username || "";
  rechargeRatesForm.firstName = row.firstName || "";
  const raw = String(row.rechargeRebateRates || "").trim();
  const parts = raw ? raw.split(",").map(s => Number(s.trim())) : [];
  if (parts.length >= 3 && parts.every(n => Number.isFinite(n) && n >= 0)) {
    rechargeRatesForm.useDefault = false;
    rechargeRatesForm.r1 = parts[0];
    rechargeRatesForm.r2 = parts[1];
    rechargeRatesForm.r3 = parts[2];
  } else {
    rechargeRatesForm.useDefault = true;
    rechargeRatesForm.r1 = 40;
    rechargeRatesForm.r2 = 45;
    rechargeRatesForm.r3 = 50;
  }
  rechargeRatesDialogVisible.value = true;
}

async function submitRechargeRates() {
  if (!rechargeRatesForm.id) return;
  let rechargeRebateRates = "";
  if (!rechargeRatesForm.useDefault) {
    const vals = [
      rechargeRatesForm.r1,
      rechargeRatesForm.r2,
      rechargeRatesForm.r3
    ].map(Number);
    if (vals.some(n => Number.isNaN(n) || n < 0 || n > 100)) {
      message("各档位比例必须在 0 到 100 之间", { type: "warning" });
      return;
    }
    rechargeRebateRates = vals.join(",");
  }
  rechargeRatesSaving.value = true;
  try {
    await setTgUserRechargeRebateRates({
      id: rechargeRatesForm.id,
      rechargeRebateRates
    });
    message("充值返佣档位修改成功", { type: "success" });
    rechargeRatesDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("修改充值返佣档位失败", error);
    message("修改充值返佣档位失败", { type: "error" });
  } finally {
    rechargeRatesSaving.value = false;
  }
}

function openRebateAmountDialog(row: TgUser) {
  rebateAmountForm.id = row.id;
  rebateAmountForm.tgId = row.tgId;
  rebateAmountForm.username = row.username || "";
  rebateAmountForm.firstName = row.firstName || "";
  rebateAmountForm.currentAmount = Number(row.rebateAmount ?? 0);
  rebateAmountForm.amount = 0;
  rebateAmountDialogVisible.value = true;
}

async function submitRebateAmount() {
  if (!rebateAmountForm.id) return;
  const amount = Number(rebateAmountForm.amount);
  if (Number.isNaN(amount) || amount <= 0) {
    message("加佣金金额必须大于 0", { type: "warning" });
    return;
  }
  rebateAmountSaving.value = true;
  try {
    await addTgUserRebateAmount({
      id: rebateAmountForm.id,
      amount
    });
    message("佣金增加成功", { type: "success" });
    rebateAmountDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("增加佣金失败", error);
    message("增加佣金失败", { type: "error" });
  } finally {
    rebateAmountSaving.value = false;
  }
}

function openRemarkDialog(row: TgUser) {
  remarkForm.id = row.id;
  remarkForm.tgId = row.tgId;
  remarkForm.username = row.username || "";
  remarkForm.firstName = row.firstName || "";
  remarkForm.remark = row.remark || "";
  remarkDialogVisible.value = true;
}

function normalizeOption(option: unknown): RechargeFieldOption | null {
  if (!option || typeof option !== "object") return null;
  const record = option as Record<string, unknown>;
  const label = String(
    record.label ?? record.name ?? record.value ?? ""
  ).trim();
  const value = String(
    record.value ?? record.code ?? record.label ?? ""
  ).trim();
  return label && value ? { label, value } : null;
}

function normalizeRechargeFields(parsed: unknown): RechargeField[] {
  if (!Array.isArray(parsed)) return [];
  return parsed
    .filter(item => item && typeof item === "object")
    .map(item => item as RechargeField)
    .filter(field => String(field.fieldKey || "").trim() !== "");
}

function parseRechargeFields(raw?: string | null | unknown[]): RechargeField[] {
  if (Array.isArray(raw)) return normalizeRechargeFields(raw);
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw) as unknown;
    return normalizeRechargeFields(parsed);
  } catch {
    return [];
  }
}

function fieldOptions(field: RechargeField) {
  if (!field.optionsJson) return [];
  try {
    const parsed = JSON.parse(field.optionsJson) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.map(normalizeOption).filter(Boolean) as RechargeFieldOption[];
  } catch {
    return [];
  }
}

function getFieldInputType(field: RechargeField) {
  return field.fieldType === "number" || field.dataType === "number"
    ? "number"
    : "text";
}

function clearRechargeFieldValues() {
  Object.keys(rechargeFieldValues).forEach(key => {
    delete rechargeFieldValues[key];
  });
}

async function syncRechargeCountryConfig() {
  const country = rechargeCountries.value.find(
    item => item.countryCode === rechargeForm.countryCode
  );
  rechargeForm.currency = country?.currencyCode || "";
  rechargeChannels.value = [];
  rechargeFields.value = [];
  rechargeMinAmount.value = 1;
  if (rechargeForm.countryCode) {
    rechargeConfigLoading.value = true;
    try {
      const { data } = await getTenantCountryRechargeInfo(
        rechargeForm.countryCode
      );
      rechargeChannels.value = [...(data?.channels || [])].sort(
        (a, b) => a.sort - b.sort
      );
      rechargeFields.value = parseRechargeFields(data?.rechargeFields || []);
      rechargeMinAmount.value = Math.max(1, Number(data?.minAmount || 1));
    } catch (error) {
      console.error("加载国家充值配置失败", error);
      message("加载国家充值配置失败", { type: "error" });
    } finally {
      rechargeConfigLoading.value = false;
    }
  } else {
    rechargeFields.value = parseRechargeFields(country?.rechargeFields);
  }
  clearRechargeFieldValues();
  rechargeFields.value.forEach(field => {
    rechargeFieldValues[field.fieldKey] = String(field.defaultValue ?? "");
  });
  if (
    rechargeForm.channel &&
    !rechargeChannels.value.some(
      channel => channel.channelCode === rechargeForm.channel
    )
  ) {
    rechargeForm.channel = "";
  }
  if (!rechargeForm.channel) {
    rechargeForm.channel = rechargeChannels.value[0]?.channelCode || "";
  }
  rechargeForm.payMethod =
    selectedRechargeChannel.value?.methods?.[0]?.methodCode || "";
}

async function loadRechargeConfig() {
  if (rechargeCountries.value.length) return;
  rechargeConfigLoading.value = true;
  try {
    const countryResp = await getSysCountryList({
      currentPage: 0,
      pageSize: 200,
      status: 1
    });
    rechargeCountries.value = countryResp.data?.list || [];
  } catch (error) {
    console.error("加载充值配置失败", error);
    message("加载充值配置失败", { type: "error" });
  } finally {
    rechargeConfigLoading.value = false;
  }
}

function selectDefaultRechargeCountry(row: TgUser) {
  const preferred = String(row.region || "").toUpperCase();
  const preferredCountry = preferred
    ? rechargeCountries.value.find(item => item.countryCode === preferred)
    : undefined;
  const usCountry = rechargeCountries.value.find(
    item => item.countryCode === "US"
  );
  return (
    preferredCountry?.countryCode ||
    usCountry?.countryCode ||
    rechargeCountries.value[0]?.countryCode ||
    ""
  );
}

async function openRechargeDialog(row: TgUser) {
  rechargeDialogVisible.value = true;
  rechargeResult.value = null;
  rechargeForm.userId = row.id;
  rechargeForm.tgId = row.tgId;
  rechargeForm.username = row.username || "";
  rechargeForm.firstName = row.firstName || "";
  rechargeForm.amount = 100;
  rechargeForm.payMethod = "";
  rechargeForm.merchantOrderNo = `tenant_${Date.now()}_${row.id}`;
  await loadRechargeConfig();
  rechargeForm.countryCode = selectDefaultRechargeCountry(row);
  await syncRechargeCountryConfig();
}

async function handleRechargeCountryChange() {
  rechargeResult.value = null;
  await syncRechargeCountryConfig();
}

function handleRechargeChannelChange() {
  rechargeResult.value = null;
  rechargeForm.payMethod =
    selectedRechargeChannel.value?.methods?.[0]?.methodCode || "";
}

function buildRechargeExtraFields() {
  const entries = rechargeFields.value
    .map(field => [
      field.fieldKey,
      String(rechargeFieldValues[field.fieldKey] ?? "").trim()
    ])
    .filter(([, value]) => value);
  return entries.length
    ? (Object.fromEntries(entries) as Record<string, string>)
    : undefined;
}

function validateRechargeForm() {
  if (!rechargeForm.userId) return "请选择用户";
  if (
    Number.isNaN(Number(rechargeForm.amount)) ||
    rechargeForm.amount < rechargeMinAmount.value
  ) {
    return `充值金额最小为 ${rechargeMinAmount.value}`;
  }
  if (!rechargeForm.countryCode) return "请选择国家";
  if (!rechargeForm.channel) return "请选择充值通道";
  for (const field of rechargeFields.value) {
    const value = String(rechargeFieldValues[field.fieldKey] ?? "").trim();
    if (field.isRequired && !value) {
      return `${field.fieldLabel}不能为空`;
    }
    if (field.minLength && value && value.length < field.minLength) {
      return field.errorTips || `${field.fieldLabel}格式不正确`;
    }
    if (field.maxLength && value && value.length > field.maxLength) {
      return field.errorTips || `${field.fieldLabel}格式不正确`;
    }
    if (field.regexRule && value) {
      try {
        const regex = new RegExp(field.regexRule);
        if (!regex.test(value)) {
          return field.errorTips || `${field.fieldLabel}格式不正确`;
        }
      } catch {
        return field.errorTips || `${field.fieldLabel}格式不正确`;
      }
    }
  }
  return "";
}

async function submitRechargeOrder(confirmUnfinishedActivityCycle = false) {
  const errorMessage = validateRechargeForm();
  if (errorMessage) {
    message(errorMessage, { type: "warning" });
    return;
  }
  rechargeSaving.value = true;
  try {
    const { data } = await createTgUserRechargeOrderV2({
      userId: rechargeForm.userId,
      amount: Number(rechargeForm.amount),
      channel: rechargeForm.channel,
      payMethod: rechargeForm.payMethod.trim() || undefined,
      currency: rechargeForm.currency,
      countryCode: rechargeForm.countryCode,
      merchantOrderNo: rechargeForm.merchantOrderNo.trim() || undefined,
      extraFields: buildRechargeExtraFields(),
      confirmUnfinishedActivityCycle
    });
    if (data?.needConfirmUnfinishedActivityCycle) {
      await ElMessageBox.confirm(
        "该用户存在未完成活动流水，确认继续创建充值订单？",
        "系统提示",
        {
          confirmButtonText: "继续创建",
          cancelButtonText: "取消",
          type: "warning"
        }
      );
      await submitRechargeOrder(true);
      return;
    }
    rechargeResult.value = data;
    message("支付订单创建成功", { type: "success" });
  } catch (error) {
    if (error !== "cancel") {
      console.error("创建支付订单失败", error);
      message("创建支付订单失败", { type: "error" });
    }
  } finally {
    rechargeSaving.value = false;
  }
}

async function copyRechargePayUrl() {
  const url = rechargeResult.value?.payUrl || "";
  if (!url) return;
  await navigator.clipboard.writeText(url);
  message("支付链接已复制", { type: "success" });
}

function openRechargePayUrl() {
  const url = rechargeResult.value?.payUrl || "";
  if (!url) return;
  window.open(url, "_blank");
}

async function submitRemark() {
  if (!remarkForm.id) return;
  const remark = String(remarkForm.remark || "").trim();
  if ([...remark].length > 255) {
    message("备注不能超过 255 个字符", { type: "warning" });
    return;
  }
  remarkSaving.value = true;
  try {
    await setTgUserRemark({
      id: remarkForm.id,
      remark
    });
    message("备注修改成功", { type: "success" });
    remarkDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("修改备注失败", error);
    message("修改备注失败", { type: "error" });
  } finally {
    remarkSaving.value = false;
  }
}

async function submitRebateRate() {
  if (!rebateRateForm.id) return;
  const rebateRate = Number(rebateRateForm.rebateRate);
  if (Number.isNaN(rebateRate) || rebateRate < 0 || rebateRate > 100) {
    message("返佣比例必须在 0 到 100 之间", { type: "warning" });
    return;
  }
  rebateRateSaving.value = true;
  try {
    await setTgUserRebateRate({
      id: rebateRateForm.id,
      rebateRate
    });
    message("返佣比例修改成功", { type: "success" });
    rebateRateDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("修改返佣比例失败", error);
    message("修改返佣比例失败", { type: "error" });
  } finally {
    rebateRateSaving.value = false;
  }
}

async function submitRebateType() {
  if (!rebateTypeForm.id) return;
  const rebateType = Number(rebateTypeForm.rebateType) === 2 ? 2 : 1;
  rebateTypeSaving.value = true;
  try {
    await setTgUserRebateType({
      id: rebateTypeForm.id,
      rebateType
    });
    message("返水方式修改成功", { type: "success" });
    rebateTypeDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("修改返水方式失败", error);
    message("修改返水方式失败", { type: "error" });
  } finally {
    rebateTypeSaving.value = false;
  }
}

function buildQueryPayload() {
  const payload: Record<string, any> = {};
  payload.isBot = false;
  if (currentUser.value) payload.parentId = currentUser.value.id;
  return payload;
}

async function loadSubStatsSummary() {
  subStatsSummaryLoading.value = true;
  try {
    const { data } = await getTgUserSubStatsSummary(buildQueryPayload());
    subStatsSummary.subRechargeAmount = Number(data?.subRechargeAmount ?? 0);
    subStatsSummary.subFlowAmount = Number(data?.subFlowAmount ?? 0);
    subStatsSummary.subProfitAmount = Number(data?.subProfitAmount ?? 0);
    subStatsSummary.subWithdrawAmount = Number(data?.subWithdrawAmount ?? 0);
    subStatsSummary.rechargeUsers = Number(data?.rechargeUsers ?? 0);
    subStatsSummary.validUsers = Number(data?.validUsers ?? 0);
  } catch (error) {
    console.error("获取下级汇总失败", error);
    message("获取下级汇总失败", { type: "error" });
  } finally {
    subStatsSummaryLoading.value = false;
  }
}

async function loadSubStatsList() {
  if (!currentUser.value) return;
  subStatsLoading.value = true;
  try {
    const { data } = await getTgUserList({
      ...buildQueryPayload(),
      currentPage: subStatsPagination.currentPage - 1,
      pageSize: subStatsPagination.pageSize
    });
    subStatsList.value = data.list || [];
    subStatsPagination.total = data.total || 0;
    subStatsPagination.pageSize = data.pageSize || subStatsPagination.pageSize;
    subStatsPagination.currentPage = (data.currentPage || 0) + 1;
  } catch (error) {
    console.error("获取下级统计列表失败", error);
    message("获取下级统计列表失败", { type: "error" });
  } finally {
    subStatsLoading.value = false;
  }
}

async function openSubStatsDialog(row: TgUser) {
  currentUser.value = row;
  subStatsDialogVisible.value = true;
  subStatsPagination.currentPage = 1;
  await Promise.all([loadSubStatsSummary(), loadSubStatsList()]);
}

function handleSubStatsPageSizeChange(size: number) {
  subStatsPagination.pageSize = size;
  subStatsPagination.currentPage = 1;
  loadSubStatsList();
}

function handleSubStatsCurrentChange(page: number) {
  subStatsPagination.currentPage = page;
  loadSubStatsList();
}
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
        <el-form-item label="用户UID：" prop="uid">
          <el-input
            v-model="form.uid"
            placeholder="请输入用户UID"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="用户名：" prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="昵称：" prop="firstName">
          <el-input
            v-model="form.firstName"
            placeholder="展示名"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="手机号：" prop="phone">
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="上级UID：" prop="parentUid">
          <el-input
            v-model="form.parentUid"
            placeholder="邀请人UID"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="邀请码：" prop="inviteCode">
          <el-input
            v-model="form.inviteCode"
            placeholder="邀请码"
            clearable
            class="!w-[180px]"
          />
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

      <PureTableBar title="用户" :columns="columns" @refresh="onSearch">
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
              <div
                style="
                  display: flex;
                  flex-wrap: wrap;
                  gap: 4px 8px;
                  align-items: center;
                "
              >
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  @click="openRebateRateDialog(row)"
                >
                  修改返佣
                </el-button>
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  @click="openRebateTypeDialog(row)"
                >
                  返水方式
                </el-button>
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  @click="openRechargeRatesDialog(row)"
                >
                  充值档位
                </el-button>
                <el-button
                  class="reset-margin"
                  link
                  :type="
                    Number(row.rebateWithdrawDisabled) === 1
                      ? 'success'
                      : 'danger'
                  "
                  :size="size"
                  @click="
                    updateRebateWithdrawDisabled(
                      row,
                      Number(row.rebateWithdrawDisabled) === 1 ? 0 : 1
                    )
                  "
                >
                  {{
                    Number(row.rebateWithdrawDisabled) === 1
                      ? "允许提现"
                      : "禁止提现"
                  }}
                </el-button>
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  @click="openRemarkDialog(row)"
                >
                  修改备注
                </el-button>
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  @click="openSubStatsDialog(row)"
                >
                  下级统计
                </el-button>
                <el-button
                  v-if="row.status !== 1"
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  @click="updateStatus(row, 1)"
                >
                  启用
                </el-button>
                <el-button
                  v-if="row.status === 1"
                  class="reset-margin"
                  link
                  type="danger"
                  :size="size"
                  @click="updateStatus(row, 0)"
                >
                  禁用
                </el-button>
              </div>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>

    <el-dialog
      v-model="subStatsDialogVisible"
      :title="`下级统计汇总（用户ID: ${currentUser?.tgId ?? '-'}）`"
      width="78%"
      destroy-on-close
    >
      <el-skeleton :loading="subStatsSummaryLoading" animated :rows="2">
        <el-row :gutter="12" class="mb-3">
          <el-col :xs="24" :sm="12" :md="8" :lg="4">
            <el-card shadow="hover">
              <div class="stat-title">充值金额之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subRechargeAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="4">
            <el-card shadow="hover">
              <div class="stat-title">流水之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subFlowAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="4">
            <el-card shadow="hover">
              <div class="stat-title">盈利之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subProfitAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="4">
            <el-card shadow="hover">
              <div class="stat-title">提现金额之和</div>
              <div class="stat-value">
                {{ formatMoney(subStatsSummary.subWithdrawAmount) }}
              </div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="4">
            <el-card shadow="hover">
              <div class="stat-title">充值人数</div>
              <div class="stat-value">
                {{ formatCount(subStatsSummary.rechargeUsers) }}
              </div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="4">
            <el-card shadow="hover">
              <div class="stat-title">有效用户数</div>
              <div class="stat-value">
                {{ formatCount(subStatsSummary.validUsers) }}
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-skeleton>

      <el-table v-loading="subStatsLoading" :data="subStatsList" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tgId" label="用户ID" min-width="140" />
        <el-table-column prop="uid" label="用户UID" min-width="120" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="firstName" label="昵称" min-width="120" />
        <el-table-column
          prop="subRechargeAmount"
          label="下级充值"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatMoney(row.subRechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="subFlowAmount" label="下级流水" min-width="120">
          <template #default="{ row }">
            {{ formatMoney(row.subFlowAmount) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="subProfitAmount"
          label="下级盈利"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatMoney(row.subProfitAmount) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="subWithdrawAmount"
          label="下级提现"
          min-width="120"
        >
          <template #default="{ row }">
            {{ formatMoney(row.subWithdrawAmount) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="dialog-pagination">
        <el-pagination
          v-model:current-page="subStatsPagination.currentPage"
          v-model:page-size="subStatsPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :small="deviceDetection()"
          :background="true"
          layout="total, sizes, prev, pager, next, jumper"
          :total="subStatsPagination.total"
          @size-change="handleSubStatsPageSizeChange"
          @current-change="handleSubStatsCurrentChange"
        />
      </div>
    </el-dialog>

    <el-dialog
      v-model="rechargeDialogVisible"
      title="手动拉起支付订单"
      width="560px"
      destroy-on-close
    >
      <el-form
        v-loading="rechargeConfigLoading"
        :model="rechargeForm"
        label-width="104px"
      >
        <el-form-item label="用户">
          <span>{{ formatName(rechargeForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ rechargeForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="充值金额" required>
          <el-input-number
            v-model="rechargeForm.amount"
            :min="rechargeMinAmount"
            :precision="0"
            :step="1"
            controls-position="right"
            class="!w-full"
            @change="rechargeResult = null"
          />
          <div class="form-tip">
            后台拉起支付最小金额 {{ rechargeMinAmount }}，满 100 才赠送
          </div>
        </el-form-item>
        <el-form-item label="国家" required>
          <el-select
            v-model="rechargeForm.countryCode"
            filterable
            class="!w-full"
            placeholder="请选择国家"
            @change="handleRechargeCountryChange"
          >
            <el-option
              v-for="country in rechargeCountries"
              :key="country.countryCode"
              :label="`${country.countryNameCn} (${country.countryCode})`"
              :value="country.countryCode"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="币种">
          <el-input :model-value="rechargeDisplayCurrency" readonly />
        </el-form-item>
        <el-form-item label="充值通道" required>
          <el-select
            v-model="rechargeForm.channel"
            filterable
            class="!w-full"
            placeholder="请选择充值通道"
            @change="handleRechargeChannelChange"
          >
            <el-option
              v-for="channel in rechargeChannels"
              :key="channel.channelCode"
              :label="`${channel.channelName} (${channel.channelCode})`"
              :value="channel.channelCode"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="支付方式">
          <el-select
            v-if="availablePayMethods.length"
            v-model="rechargeForm.payMethod"
            filterable
            clearable
            class="!w-full"
            placeholder="请选择支付方式"
            @change="rechargeResult = null"
          >
            <el-option
              v-for="method in availablePayMethods"
              :key="method.methodCode"
              :label="`${method.methodName} (${method.methodCode})`"
              :value="method.methodCode"
            />
          </el-select>
          <el-input
            v-else
            v-model="rechargeForm.payMethod"
            placeholder="可选，按通道要求填写"
            clearable
            @input="rechargeResult = null"
          />
        </el-form-item>
        <el-form-item label="商户单号">
          <el-input
            v-model="rechargeForm.merchantOrderNo"
            placeholder="可选"
            clearable
            @input="rechargeResult = null"
          />
        </el-form-item>
        <template v-for="field in rechargeFields" :key="field.fieldKey">
          <el-form-item
            :label="field.fieldLabel"
            :required="!!field.isRequired"
          >
            <el-select
              v-if="field.fieldType === 'select'"
              v-model="rechargeFieldValues[field.fieldKey]"
              filterable
              clearable
              class="!w-full"
              :placeholder="
                field.fieldPlaceholder || `请选择${field.fieldLabel}`
              "
              @change="rechargeResult = null"
            >
              <el-option
                v-for="option in fieldOptions(field)"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-input
              v-else-if="field.fieldType === 'textarea'"
              v-model="rechargeFieldValues[field.fieldKey]"
              type="textarea"
              :rows="3"
              :maxlength="field.maxLength || undefined"
              :placeholder="
                field.fieldPlaceholder || `请输入${field.fieldLabel}`
              "
              @input="rechargeResult = null"
            />
            <el-input
              v-else
              v-model="rechargeFieldValues[field.fieldKey]"
              :type="getFieldInputType(field)"
              :maxlength="field.maxLength || undefined"
              :placeholder="
                field.fieldPlaceholder || `请输入${field.fieldLabel}`
              "
              clearable
              @input="rechargeResult = null"
            />
          </el-form-item>
        </template>
        <template v-if="rechargeResult">
          <el-divider>创建结果</el-divider>
          <el-form-item label="订单号">
            <el-input :model-value="rechargeResult.orderNo" readonly />
          </el-form-item>
          <el-form-item label="支付链接">
            <el-input :model-value="rechargeResult.payUrl || ''" readonly>
              <template #append>
                <el-button
                  :disabled="!rechargeResult.payUrl"
                  @click="copyRechargePayUrl"
                >
                  复制
                </el-button>
              </template>
            </el-input>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="rechargeDialogVisible = false">关闭</el-button>
        <el-button
          v-if="rechargeResult?.payUrl"
          type="success"
          @click="openRechargePayUrl"
        >
          打开支付链接
        </el-button>
        <el-button
          type="primary"
          :loading="rechargeSaving"
          :disabled="rechargeConfigLoading"
          @click="submitRechargeOrder()"
        >
          创建支付订单
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="rebateRateDialogVisible"
      title="修改返佣比例"
      width="420px"
      destroy-on-close
    >
      <el-form :model="rebateRateForm" label-width="96px">
        <el-form-item label="用户">
          <span>{{ formatName(rebateRateForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ rebateRateForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="返佣比例">
          <el-input-number
            v-model="rebateRateForm.rebateRate"
            :min="0"
            :max="100"
            :precision="2"
            :step="1"
            controls-position="right"
            class="!w-full"
          />
          <div class="form-tip">单位：%，范围 0 - 100</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rebateRateDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="rebateRateSaving"
          @click="submitRebateRate"
        >
          保存
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="rebateTypeDialogVisible"
      title="修改充值返水方式"
      width="460px"
      destroy-on-close
    >
      <el-form :model="rebateTypeForm" label-width="96px">
        <el-form-item label="用户">
          <span>{{ formatName(rebateTypeForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ rebateTypeForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="返水方式">
          <el-radio-group v-model="rebateTypeForm.rebateType">
            <el-radio :value="1">返水余额</el-radio>
            <el-radio :value="2">可用余额</el-radio>
          </el-radio-group>
          <div class="form-tip">
            返水余额：充值返水进可用返水余额，走佣金转余额提现；可用余额：充值返水直接进可用余额，并附加
            v2 提现流水批次限制。仅影响“充值返水”，不影响投注返水/红包返水。
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rebateTypeDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="rebateTypeSaving"
          @click="submitRebateType"
        >
          保存
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="rechargeRatesDialogVisible"
      title="修改充值返佣档位"
      width="460px"
      destroy-on-close
    >
      <el-form :model="rechargeRatesForm" label-width="120px">
        <el-form-item label="用户">
          <span>{{ formatName(rechargeRatesForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ rechargeRatesForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="使用默认配置">
          <el-switch v-model="rechargeRatesForm.useDefault" />
          <div class="form-tip">
            开启=使用系统默认档位（sys_config），关闭=为该用户单独配置；只影响充值返佣比例。
          </div>
        </el-form-item>
        <template v-if="!rechargeRatesForm.useDefault">
          <el-form-item label="第1次充值(%)">
            <el-input-number
              v-model="rechargeRatesForm.r1"
              :min="0"
              :max="100"
              :precision="2"
              :step="1"
              controls-position="right"
              class="!w-full"
            />
          </el-form-item>
          <el-form-item label="第2次充值(%)">
            <el-input-number
              v-model="rechargeRatesForm.r2"
              :min="0"
              :max="100"
              :precision="2"
              :step="1"
              controls-position="right"
              class="!w-full"
            />
          </el-form-item>
          <el-form-item label="第3次及以上(%)">
            <el-input-number
              v-model="rechargeRatesForm.r3"
              :min="0"
              :max="100"
              :precision="2"
              :step="1"
              controls-position="right"
              class="!w-full"
            />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="rechargeRatesDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="rechargeRatesSaving"
          @click="submitRechargeRates"
        >
          保存
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="rebateAmountDialogVisible"
      title="给用户加佣金"
      width="420px"
      destroy-on-close
    >
      <el-form :model="rebateAmountForm" label-width="96px">
        <el-form-item label="用户">
          <span>{{ formatName(rebateAmountForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ rebateAmountForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="当前佣金">
          <span>{{ formatMoney(rebateAmountForm.currentAmount) }}</span>
        </el-form-item>
        <el-form-item label="加佣金额">
          <el-input-number
            v-model="rebateAmountForm.amount"
            :min="0"
            :precision="2"
            :step="1"
            controls-position="right"
            class="!w-full"
          />
          <div class="form-tip">
            只增加可用佣金和累计佣金，不修改余额或其他字段
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rebateAmountDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="rebateAmountSaving"
          @click="submitRebateAmount"
        >
          确认加佣金
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="remarkDialogVisible"
      title="修改备注"
      width="460px"
      destroy-on-close
    >
      <el-form :model="remarkForm" label-width="96px">
        <el-form-item label="用户">
          <span>{{ formatName(remarkForm) }}</span>
        </el-form-item>
        <el-form-item label="用户ID">
          <span>{{ remarkForm.tgId || "-" }}</span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="remarkForm.remark"
            type="textarea"
            :rows="4"
            maxlength="255"
            show-word-limit
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="remarkDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="remarkSaving" @click="submitRemark">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}

.stat-title {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.stat-value {
  margin-top: 6px;
  font-size: 22px;
  font-weight: 600;
}

.dialog-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}

.form-tip {
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.3;
}
</style>
