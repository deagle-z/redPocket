import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import { tenantTypeOptions, statusOptions } from "./enums";
import {
  delSysTenant,
  getSysTenantList,
  setSysTenant,
  type SysTenant
} from "@/api/tenant";
import { type Ref, reactive, ref, onMounted, h, toRaw } from "vue";
import type { FormItemProps } from "./types";
import { ElTag } from "element-plus";

function getTenantTypeLabel(val: number) {
  const match = tenantTypeOptions.find(item => item.value === val);
  return match ? match.label : "-";
}

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

export function useTenant(tableRef: Ref) {
  const form = reactive({
    tenantCode: "",
    tenantName: "",
    tenantType: undefined as number | undefined,
    status: undefined as number | undefined,
    ownerUserId: undefined as number | undefined,
    planCode: ""
  });
  const formRef = ref();
  const dataList = ref<SysTenant[]>([]);
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
      label: "租户编码",
      prop: "tenantCode",
      minWidth: 160,
      showOverflowTooltip: true
    },
    {
      label: "租户名称",
      prop: "tenantName",
      minWidth: 180,
      showOverflowTooltip: true
    },
    {
      label: "类型",
      prop: "tenantType",
      minWidth: 100,
      formatter: ({ tenantType }) => getTenantTypeLabel(tenantType)
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
      label: "套餐",
      prop: "planCode",
      minWidth: 120,
      formatter: ({ planCode }) => planCode || "-"
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 180,
      showOverflowTooltip: true,
      formatter: ({ remark }) => remark || "-"
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "更新时间",
      prop: "updatedAt",
      minWidth: 160,
      formatter: ({ updatedAt }) =>
        dayjs(updatedAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "操作",
      fixed: "right",
      width: 160,
      slot: "operation"
    }
  ];

  function handleDelete(row) {
    delSysTenant(row.id)
      .then(() => {
        message(`您删除了租户 ${row.tenantName} 的这条数据`, {
          type: "success"
        });
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
      const { data } = await getSysTenantList({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取租户列表失败", error);
      message("获取租户列表失败", { type: "error" });
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
      title: `${title}租户`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          tenantCode: row?.tenantCode ?? "",
          tenantName: row?.tenantName ?? "",
          tenantType: row?.tenantType ?? 1,
          status: row?.status ?? 1,
          loginAccount: row?.loginAccount ?? "",
          loginPassword: "",
          ownerUserId: row?.ownerUserId ?? undefined,
          planCode: row?.planCode ?? "",
          timezone: row?.timezone ?? "UTC",
          locale: row?.locale ?? "en-US",
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
          setSysTenant(curData)
            .then(() => {
              message(`您${title}了租户 ${curData.tenantName}`, {
                type: "success"
              });
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
    tenantTypeOptions,
    onSearch,
    resetForm,
    openDialog,
    handleDelete,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
