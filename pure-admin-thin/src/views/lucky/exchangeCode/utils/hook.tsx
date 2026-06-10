import dayjs from "dayjs";
import editForm from "../form.vue";
import { h, onMounted, reactive, ref, toRaw, type Ref } from "vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import type { PaginationProps } from "@pureadmin/table";
import type { FormItemProps } from "./types";
import {
  delExchangeCodeAdmin,
  getExchangeCodeListAdmin,
  setExchangeCodeAdmin,
  type ExchangeCode
} from "@/api/exchangeCode";

export function useExchangeCode(tableRef: Ref) {
  const form = reactive<{
    code: string;
    status: 1 | 0 | null;
  }>({
    code: "",
    status: null
  });
  const formRef = ref();
  const dataList = ref<ExchangeCode[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "id", width: 80 },
    { label: "兑换码", prop: "code", minWidth: 120 },
    {
      label: "金额",
      prop: "amount",
      minWidth: 100,
      formatter: ({ amount }) => Number(amount || 0).toFixed(2)
    },
    { label: "总次数", prop: "maxRedeemCount", minWidth: 100 },
    { label: "已兑换", prop: "redeemCount", minWidth: 100 },
    {
      label: "剩余",
      minWidth: 100,
      formatter: row =>
        String(
          Math.max(
            0,
            Number(row.maxRedeemCount || 0) - Number(row.redeemCount || 0)
          )
        )
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
      cellRenderer: scope => (
        <el-tag
          size="small"
          type={scope.row.status === 1 ? "success" : "info"}
          effect="plain"
        >
          {scope.row.status === 1 ? "启用" : "停用"}
        </el-tag>
      )
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 180,
      showOverflowTooltip: true
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 180,
      formatter: ({ createdAt }) =>
        createdAt ? dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss") : "-"
    },
    { label: "操作", fixed: "right", width: 220, slot: "operation" }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getExchangeCodeListAdmin({
        ...toRaw(form),
        currentPage: pagination.currentPage - 1,
        pageSize: pagination.pageSize
      });
      dataList.value = (data?.list || []) as ExchangeCode[];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = (data?.currentPage ?? 0) + 1;
    } finally {
      loading.value = false;
      tableRef.value?.setAdaptive?.();
    }
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    pagination.currentPage = 1;
    onSearch();
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 1;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  async function handleDelete(row: ExchangeCode) {
    await delExchangeCodeAdmin(row.id);
    message(`已删除兑换码 [${row.code}]`, { type: "success" });
    onSearch();
  }

  async function handleToggleStatus(row: ExchangeCode) {
    const nextStatus = row.status === 1 ? 0 : 1;
    await setExchangeCodeAdmin({
      id: row.id,
      status: nextStatus,
      remark: row.remark || ""
    });
    message(`${nextStatus === 1 ? "启用" : "停用"}成功`, { type: "success" });
    onSearch();
  }

  function openDialog(title = "新增", row?: ExchangeCode) {
    addDialog({
      title: `${title}兑换码`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          code: row?.code ?? "",
          amount: row?.amount ?? 0,
          maxRedeemCount: row?.maxRedeemCount ?? 1,
          generateCount: 1,
          status: row?.status === 0 ? 0 : 1,
          remark: row?.remark ?? ""
        } as FormItemProps
      },
      width: "520px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () =>
        h(editForm, { ref: formRef, formInline: null as any }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;
        FormRef.validate(async valid => {
          if (!valid) return;
          try {
            if (curData.id > 0) {
              await setExchangeCodeAdmin({
                id: curData.id,
                status: curData.status,
                remark: curData.remark?.trim() || ""
              });
            } else {
              const code = curData.code?.trim() || "";
              await setExchangeCodeAdmin({
                code,
                amount: Number(curData.amount || 0),
                maxRedeemCount: Number(curData.maxRedeemCount || 0),
                generateCount: code ? 1 : Number(curData.generateCount || 1),
                status: curData.status,
                remark: curData.remark?.trim() || ""
              });
            }
            message(`${title}成功`, { type: "success" });
            done();
            onSearch();
          } catch (err) {
            message((err as Error)?.message || "保存兑换码失败", {
              type: "error"
            });
          }
        });
      }
    });
  }

  onMounted(onSearch);

  return {
    form,
    loading,
    columns,
    dataList,
    pagination,
    onSearch,
    resetForm,
    openDialog,
    handleDelete,
    handleToggleStatus,
    handleSizeChange,
    handleCurrentChange
  };
}
