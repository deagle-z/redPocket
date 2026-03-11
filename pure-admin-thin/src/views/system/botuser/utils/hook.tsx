import dayjs from "dayjs";
import { h, onMounted, reactive, ref, toRaw, type Ref } from "vue";
import { ElImage, ElMessageBox, ElTag } from "element-plus";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  delAdminTgUser,
  getAdminBotUserList,
  setAdminTgUserStatus,
  type TgUser
} from "@/api/tgUser";
import { getKeyList } from "@pureadmin/utils";

const statusOptions = [
  { label: "正常", value: 1 },
  { label: "禁用", value: 0 },
  { label: "删除", value: -1 }
];

function formatNullable(val?: string | null) {
  return val && val !== "" ? val : "-";
}

function formatMoney(val?: number | null) {
  if (typeof val !== "number" || Number.isNaN(val)) return "0.000";
  return val.toFixed(3);
}

function getStatusLabel(status: number) {
  return statusOptions.find(item => item.value === status)?.label || "-";
}

function getStatusType(status: number) {
  if (status === 1) return "success";
  if (status === 0) return "warning";
  return "info";
}

export function useBotUser(tableRef: Ref) {
  const form = reactive({
    tgId: undefined as number | undefined,
    username: "",
    firstName: "",
    status: undefined as number | undefined
  });
  const dataList = ref<TgUser[]>([]);
  const loading = ref(true);
  const selectedNum = ref(0);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });

  const columns: TableColumnList = [
    {
      label: "勾选列",
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: "ID",
      prop: "id",
      minWidth: 80
    },
    {
      label: "TG用户ID",
      prop: "tgId",
      minWidth: 160
    },
    {
      label: "头像",
      prop: "avatar",
      width: 90,
      cellRenderer: scope =>
        scope.row.avatar
          ? h(ElImage, {
              src: scope.row.avatar,
              fit: "cover",
              previewSrcList: [scope.row.avatar],
              previewTeleported: true,
              style: {
                width: "42px",
                height: "42px",
                borderRadius: "8px"
              }
            })
          : "-"
    },
    {
      label: "用户名",
      prop: "username",
      minWidth: 180,
      formatter: ({ username }) => formatNullable(username)
    },
    {
      label: "昵称",
      prop: "firstName",
      minWidth: 160,
      formatter: ({ firstName }) => formatNullable(firstName)
    },
    {
      label: "余额",
      prop: "balance",
      minWidth: 110,
      formatter: ({ balance }) => formatMoney(balance)
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: scope =>
        h(
          ElTag,
          { type: getStatusType(scope.row.status) },
          () => getStatusLabel(scope.row.status)
        )
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 180,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "操作",
      fixed: "right",
      width: 180,
      slot: "operation"
    }
  ];

  function handleSizeChange(size: number) {
    pagination.pageSize = size;
    onSearch();
  }

  function handleCurrentChange(page: number) {
    pagination.currentPage = page - 1;
    onSearch();
  }

  function handleSelectionChange(val) {
    selectedNum.value = val.length;
    tableRef.value?.setAdaptive?.();
  }

  function onSelectionCancel() {
    selectedNum.value = 0;
    tableRef.value?.getTableRef?.().clearSelection();
  }

  async function handleDelete(row: TgUser) {
    await delAdminTgUser(row.id);
    message(`已删除机器人编号为 ${row.id} 的数据`, { type: "success" });
    onSearch();
  }

  async function onBatchDel() {
    const curSelected = tableRef.value?.getTableRef?.().getSelectionRows?.() || [];
    const ids = getKeyList(curSelected, "id");
    await Promise.all(ids.map(id => delAdminTgUser(Number(id))));
    message(`已删除机器人编号为 ${ids} 的数据`, { type: "success" });
    tableRef.value?.getTableRef?.().clearSelection();
    selectedNum.value = 0;
    onSearch();
  }

  async function updateStatus(row: TgUser, status: number) {
    const actionText = status === 1 ? "启用" : "禁用";
    try {
      await ElMessageBox.confirm(
        `确认要${actionText}机器人 <strong>${formatNullable(
          row.firstName || row.username
        )}</strong> 吗?`,
        "系统提示",
        {
          confirmButtonText: "确定",
          cancelButtonText: "取消",
          type: "warning",
          dangerouslyUseHTMLString: true,
          draggable: true
        }
      );
      await setAdminTgUserStatus({ id: row.id, status });
      message(`已${actionText}机器人`, { type: "success" });
      onSearch();
    } catch (error) {
      if (error !== "cancel") {
        console.error(`${actionText}机器人失败`, error);
        message(`${actionText}机器人失败`, { type: "error" });
      }
    }
  }

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getAdminBotUserList({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = data?.currentPage || 0;
    } catch (error) {
      console.error("获取机器人列表失败", error);
      message("获取机器人列表失败", { type: "error" });
    } finally {
      setTimeout(() => {
        loading.value = false;
      }, 300);
    }
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  };

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    selectedNum,
    pagination,
    statusOptions,
    onSearch,
    resetForm,
    onBatchDel,
    onSelectionCancel,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange,
    updateStatus,
    handleDelete
  };
}
