import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import { ElTag, ElMessageBox } from "element-plus";
import { getTgUserList, setTgUserStatus, type TgUser } from "@/api/tgUser";

const statusOptions = [
  { label: "正常", value: 1 },
  { label: "禁用", value: 0 },
  { label: "删除", value: -1 }
];

function getStatusLabel(status: number) {
  const match = statusOptions.find(item => item.value === status);
  return match ? match.label : "-";
}

function getStatusType(status: number) {
  if (status === 1) return "success";
  if (status === 0) return "warning";
  if (status === -1) return "info";
  return "info";
}

function formatNullable(val?: string | null) {
  return val && val !== "" ? val : "-";
}

function formatMoney(val?: number | null) {
  if (val === null || val === undefined || Number.isNaN(Number(val))) return "0";
  return String(val);
}

function formatPercent(val?: number | null) {
  if (typeof val !== "number" || Number.isNaN(val)) return "0.00%";
  return `${val.toFixed(2)}%`;
}

export function useTgUser(_tableRef: Ref) {
  const form = reactive({
    uid: "",
    username: "",
    firstName: "",
    parentUid: "",
    inviteCode: "",
    status: undefined as number | undefined
  });
  const dataList = ref<TgUser[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    {
      label: "ID",
      prop: "id",
      minWidth: 80
    },
    {
      label: "用户UID",
      prop: "uid",
      minWidth: 120,
      formatter: ({ uid }) => formatNullable(uid)
    },
    {
      label: "代理名称",
      prop: "tenantName",
      minWidth: 140,
      formatter: ({ tenantName }) => formatNullable(tenantName)
    },
    {
      label: "手机号",
      prop: "phone",
      minWidth: 140,
      formatter: ({ phone }) => formatNullable(phone)
    },
    {
      label: "用户名",
      prop: "username",
      minWidth: 140,
      formatter: ({ username }) => formatNullable(username)
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 160,
      formatter: ({ remark }) => formatNullable(remark)
    },
    {
      label: "余额",
      prop: "balance",
      minWidth: 120,
      formatter: ({ balance }) => formatMoney(balance)
    },
    {
      label: "赠送余额",
      prop: "giftAmount",
      minWidth: 120,
      formatter: ({ giftAmount }) => formatMoney(giftAmount)
    },
    {
      label: "累计赠送",
      prop: "giftTotal",
      minWidth: 120,
      formatter: ({ giftTotal }) => formatMoney(giftTotal)
    },
    {
      label: "可用返水",
      prop: "rebateAmount",
      minWidth: 120,
      formatter: ({ rebateAmount }) => formatMoney(rebateAmount)
    },
    {
      label: "累计返水",
      prop: "rebateTotalAmount",
      minWidth: 120,
      formatter: ({ rebateTotalAmount }) => formatMoney(rebateTotalAmount)
    },
    {
      label: "返佣比例",
      prop: "rebateRate",
      minWidth: 110,
      formatter: ({ rebateRate }) => formatPercent(rebateRate)
    },
    {
      label: "上级UID",
      prop: "parentUid",
      minWidth: 100,
      formatter: ({ parentUid }) => formatNullable(parentUid)
    },
    {
      label: "邀请码",
      prop: "inviteCode",
      minWidth: 120,
      formatter: ({ inviteCode }) => formatNullable(inviteCode)
    },
    {
      label: "昵称",
      prop: "firstName",
      minWidth: 140,
      formatter: ({ firstName }) => formatNullable(firstName)
    },
    {
      label: "密码",
      prop: "passwordPlain",
      minWidth: 120,
      formatter: ({ passwordPlain }) => formatNullable(passwordPlain)
    },
    {
      label: "注册IP",
      prop: "ip",
      minWidth: 140,
      formatter: ({ ip }) => formatNullable(ip)
    },
    {
      label: "地区",
      prop: "region",
      minWidth: 90,
      formatter: ({ region }) => formatNullable(region)
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
      cellRenderer: scope => (
        <ElTag type={getStatusType(scope.row.status)} effect="plain">
          {getStatusLabel(scope.row.status)}
        </ElTag>
      )
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "操作",
      fixed: "right",
      width: 340,
      slot: "operation"
    }
  ];

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val - 1;
    onSearch();
  }

  function handleSelectionChange(val) {
    console.log("handleSelectionChange", val);
  }

  async function updateStatus(row: TgUser, status: number) {
    const actionText = status === 1 ? "启用" : "禁用";
    try {
      await ElMessageBox.confirm(
        `确认要${actionText}用户 <strong>${formatNullable(
          row.username
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
      await setTgUserStatus({ id: row.id, status });
      message(`已${actionText}用户`, { type: "success" });
      onSearch();
    } catch (error) {
      if (error !== "cancel") {
        message(`${actionText}用户失败`, { type: "error" });
      }
    }
  }

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getTgUserList({
        isBot: false,
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取TG用户失败", error);
      message("获取TG用户失败", { type: "error" });
    } finally {
      setTimeout(() => {
        loading.value = false;
      }, 500);
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
    pagination,
    statusOptions,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange,
    updateStatus
  };
}
