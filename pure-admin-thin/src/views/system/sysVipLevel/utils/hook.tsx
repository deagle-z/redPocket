import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  delSysVipLevel,
  getSysVipLevelList,
  setSysVipLevel,
  type SysVipLevel
} from "@/api/sysVipLevel";
import { type Ref, h, reactive, ref, onMounted } from "vue";
import type { FormItemProps } from "./types";

export function useSysVipLevel(tableRef: Ref) {
  const form = reactive({
    status: null
  });

  const formRef = ref();
  const dataList = ref<SysVipLevel[]>([]);
  const loading = ref(true);

  const pagination = reactive({
    total: 0,
    pageSize: 200,
    currentPage: 1,
    background: true
  });

  const upgradeTypeMap: Record<number, string> = {
    1: "累计",
    2: "当月"
  };

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 70 },
    { label: "等级", prop: "level", minWidth: 70 },
    { label: "等级名称", prop: "levelName", minWidth: 100 },
    { label: "代理标签", prop: "agentTag", minWidth: 120, formatter: ({ agentTag }) => agentTag || "-" },
    {
      label: "升级方式",
      prop: "upgradeType",
      minWidth: 90,
      formatter: ({ upgradeType }) => upgradeTypeMap[upgradeType] ?? "-"
    },
    { label: "总充值次数", prop: "totalRechargeCount", minWidth: 100, formatter: ({ totalRechargeCount }) => totalRechargeCount ?? "-" },
    { label: "总充值金额", prop: "totalRechargeAmount", minWidth: 110, formatter: ({ totalRechargeAmount }) => totalRechargeAmount ?? "-" },
    { label: "总有效投注", prop: "totalValidBet", minWidth: 110, formatter: ({ totalValidBet }) => totalValidBet ?? "-" },
    { label: "当月充值", prop: "monthRechargeAmount", minWidth: 100, formatter: ({ monthRechargeAmount }) => monthRechargeAmount ?? "-" },
    { label: "当月投注", prop: "monthValidBet", minWidth: 100, formatter: ({ monthValidBet }) => monthValidBet ?? "-" },
    { label: "升级奖励", prop: "upgradeBonusAmount", minWidth: 100 },
    { label: "排序", prop: "sort", minWidth: 70 },
    {
      label: "状态",
      prop: "status",
      minWidth: 80,
      cellRenderer: scope => (
        <el-tag
          size="small"
          type={scope.row.status === 1 ? "success" : "danger"}
          effect="plain"
        >
          {scope.row.status === 1 ? "启用" : "停用"}
        </el-tag>
      )
    },
    { label: "操作", fixed: "right", width: 150, slot: "operation" }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getSysVipLevelList({
        ...form,
        currentPage: 0,
        pageSize: pagination.pageSize
      });
      // 等级低的在前（升序）
      dataList.value = (data?.list || []).sort((a, b) => a.level - b.level);
      pagination.total = data?.total || 0;
    } catch {
      message("获取VIP等级列表失败", { type: "error" });
    } finally {
      loading.value = false;
      tableRef.value?.setAdaptive?.();
    }
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  };

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  async function handleDelete(row: SysVipLevel) {
    await delSysVipLevel(row.id);
    message(`已删除 ${row.levelName}`, { type: "success" });
    onSearch();
  }

  function openDialog(title = "新增", row?: SysVipLevel) {
    addDialog({
      title: `${title}VIP等级`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          tenantId: row?.tenantId ?? 0,
          level: row?.level ?? 1,
          levelName: row?.levelName ?? "",
          agentTag: row?.agentTag ?? "",
          totalRechargeCount: row?.totalRechargeCount ?? null,
          totalRechargeAmount: row?.totalRechargeAmount ?? null,
          totalValidBet: row?.totalValidBet ?? null,
          monthRechargeAmount: row?.monthRechargeAmount ?? null,
          monthValidBet: row?.monthValidBet ?? null,
          upgradeBonusAmount: row?.upgradeBonusAmount ?? 0,
          upgradeType: row?.upgradeType ?? 1,
          keepLevelCondition: row?.keepLevelCondition ?? 0,
          sort: row?.sort ?? 0,
          status: row?.status ?? 1,
          remark: row?.remark ?? ""
        } as FormItemProps
      },
      width: "60%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef }),
      beforeSure: async (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        FormRef.validate(async valid => {
          if (valid) {
            try {
              await setSysVipLevel({
                id: curData.id || undefined,
                tenantId: curData.tenantId,
                level: curData.level,
                levelName: curData.levelName.trim(),
                agentTag: curData.agentTag.trim() || null,
                totalRechargeCount: curData.totalRechargeCount,
                totalRechargeAmount: curData.totalRechargeAmount,
                totalValidBet: curData.totalValidBet,
                monthRechargeAmount: curData.monthRechargeAmount,
                monthValidBet: curData.monthValidBet,
                upgradeBonusAmount: Number(curData.upgradeBonusAmount || 0),
                upgradeType: curData.upgradeType,
                keepLevelCondition: curData.keepLevelCondition,
                sort: Number(curData.sort || 0),
                status: curData.status,
                remark: curData.remark.trim() || null
              });
              message(`您${title}了 ${curData.levelName}`, { type: "success" });
              done();
              onSearch();
            } catch {
              message("保存VIP等级失败", { type: "error" });
            }
          }
        });
      }
    });
  }

  onMounted(() => {
    onSearch();
  });

  return {
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
  };
}
