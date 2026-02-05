import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import { roleOptions, statusOptions } from "./enums";
import { delSysTenantUser, getSysTenantUserList, setSysTenantUser } from "@/api/tenantUser";
import { type Ref, reactive, ref, onMounted, h, toRaw } from "vue";
import type { FormItemProps } from "./types";
import { ElTag } from "element-plus";
import type { SysTenantUser } from "@/api/tenantUser";

function getStatusLabel(status: number) {
  const match = statusOptions.find(item => item.value === status);
  return match ? match.label : "-";
}

function getStatusType(status: number) {
  if (status === 1) return "success";
  if (status === 0) return "danger";
  if (status === 2) return "warning";
  return "info";
}

export function useTenantUser(tableRef: Ref) {
  const form = reactive({
    tenantId: undefined as number | undefined,
    username: "",
    email: "",
    mobile: "",
    roleCode: "",
    isOwner: undefined as boolean | undefined,
    status: undefined as number | undefined,
    require2fa: undefined as boolean | undefined
  });
  const formRef = ref();
  const dataList = ref<SysTenantUser[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    { label: "租户ID", prop: "tenantId", minWidth: 90 },
    { label: "账号", prop: "username", minWidth: 140 },
    { label: "邮箱", prop: "email", minWidth: 160, formatter: ({ email }) => email || "-" },
    { label: "手机号", prop: "mobile", minWidth: 140, formatter: ({ mobile }) => mobile || "-" },
    { label: "角色", prop: "roleCode", minWidth: 120 },
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
      label: "Owner",
      prop: "isOwner",
      minWidth: 90,
      formatter: ({ isOwner }) => (isOwner ? "是" : "否")
    },
    {
      label: "2FA",
      prop: "require2fa",
      minWidth: 90,
      formatter: ({ require2fa }) => (require2fa ? "是" : "否")
    },
    {
      label: "最后登录",
      prop: "lastLoginAt",
      minWidth: 160,
      formatter: ({ lastLoginAt }) =>
        lastLoginAt ? dayjs(lastLoginAt).format("YYYY-MM-DD HH:mm:ss") : "-"
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
      width: 160,
      slot: "operation"
    }
  ];

  function handleDelete(row) {
    delSysTenantUser(row.id)
      .then(() => {
        message(`您删除了账号 ${row.username} 的这条数据`, { type: "success" });
      })
      .finally(() => {
        onSearch();
      });
  }

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

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getSysTenantUserList({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取租户用户失败", error);
      message("获取租户用户失败", { type: "error" });
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

  function openDialog(title = "新增", row?: FormItemProps) {
    addDialog({
      title: `${title}租户用户`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          tenantId: row?.tenantId ?? 0,
          username: row?.username ?? "",
          passwordHash: row?.passwordHash ?? "",
          passwordAlgo: row?.passwordAlgo ?? "bcrypt",
          email: row?.email ?? "",
          mobile: row?.mobile ?? "",
          roleCode: row?.roleCode ?? "member",
          isOwner: row?.isOwner ?? false,
          status: row?.status ?? 1,
          require2fa: row?.require2fa ?? false,
          twofaSecret: row?.twofaSecret ?? "",
          remark: row?.remark ?? ""
        }
      },
      width: "45%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        function chores() {
          setSysTenantUser(curData)
            .then(() => {
              message(`您${title}了账号 ${curData.username}`, { type: "success" });
            })
            .finally(() => {
              done();
              onSearch();
            });
        }

        FormRef.validate(valid => {
          if (valid) {
            chores();
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
    statusOptions,
    roleOptions,
    onSearch,
    resetForm,
    openDialog,
    handleDelete,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
