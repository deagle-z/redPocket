import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  delSysPayMethod,
  getSysPayMethodList,
  setSysPayMethod,
  type SysPayMethod
} from "@/api/sysPayMethod";
import { type Ref, h, reactive, ref, onMounted } from "vue";
import type { FormItemProps } from "./types";

function toOptionalString(value?: string | null) {
  const text = value?.trim() || "";
  return text || null;
}

export function useSysPayMethod(tableRef: Ref) {
  const form = reactive({
    methodCode: "",
    methodName: "",
    status: null
  });

  const formRef = ref();
  const dataList = ref<SysPayMethod[]>([]);
  const loading = ref(true);

  const pagination = reactive({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    { label: "方式编码", prop: "methodCode", minWidth: 160 },
    { label: "方式名称", prop: "methodName", minWidth: 150 },
    {
      label: "图标",
      prop: "icon",
      minWidth: 80,
      formatter: ({ icon }) => icon || "-"
    },
    { label: "排序", prop: "sort", minWidth: 80 },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
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
    {
      label: "更新时间",
      prop: "updatedAt",
      minWidth: 180,
      formatter: ({ updatedAt }) =>
        dayjs(updatedAt).format("YYYY-MM-DD HH:mm:ss")
    },
    { label: "操作", fixed: "right", width: 180, slot: "operation" }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getSysPayMethodList({
        ...form,
        currentPage: pagination.currentPage - 1,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = (data?.currentPage ?? 0) + 1;
    } catch {
      message("获取支付方式列表失败", { type: "error" });
    } finally {
      loading.value = false;
      tableRef.value?.setAdaptive?.();
    }
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    pagination.currentPage = 1;
    onSearch();
  };

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 1;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  async function handleDelete(row: SysPayMethod) {
    await delSysPayMethod(row.id);
    message(`已删除支付方式 ${row.methodName}`, { type: "success" });
    onSearch();
  }

  function openDialog(title = "新增", row?: SysPayMethod) {
    addDialog({
      title: `${title}支付方式`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          methodCode: row?.methodCode ?? "",
          methodName: row?.methodName ?? "",
          icon: row?.icon ?? "",
          sort: row?.sort ?? 0,
          status: row?.status ?? 1,
          remark: row?.remark ?? ""
        }
      },
      width: "50%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef }),
      beforeSure: async (done, { options }) => {
        if (formRef.value.uploading) {
          message("图片上传中，请稍候", { type: "warning" });
          return;
        }
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        FormRef.validate(async valid => {
          if (valid) {
            try {
              await setSysPayMethod({
                id: curData.id || undefined,
                methodCode: curData.methodCode.trim(),
                methodName: curData.methodName.trim(),
                icon: toOptionalString(curData.icon),
                sort: Number(curData.sort || 0),
                status: curData.status,
                remark: toOptionalString(curData.remark)
              });
              message(`您${title}了支付方式 ${curData.methodName}`, {
                type: "success"
              });
              done();
              onSearch();
            } catch {
              message("保存支付方式失败", { type: "error" });
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
