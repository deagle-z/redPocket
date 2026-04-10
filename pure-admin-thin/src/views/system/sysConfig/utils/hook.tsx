import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { FormItemProps } from "./types";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import { getSysConfigs, setSysConfig, delSysConfig } from "@/api/sys_config";
import { reactive, ref, onMounted, h, toRaw } from "vue";
import dayjs from "dayjs";

export function useSysConfig() {
  const form = reactive({
    configKey: "",
    configDesc: ""
  });
  const formRef = ref();
  const dataList = ref([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "id", width: 80 },
    {
      label: "配置Key",
      prop: "configKey",
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      label: "配置值",
      prop: "configValue",
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      label: "描述",
      prop: "configDesc",
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      label: "更新时间",
      prop: "updatedAt",
      minWidth: 160,
      formatter: ({ updatedAt }) =>
        dayjs(updatedAt).format("YYYY-MM-DD HH:mm:ss")
    },
    { label: "操作", fixed: "right", width: 160, slot: "operation" }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getSysConfigs({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list ?? [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
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

  function handleDelete(row) {
    delSysConfig(row.id)
      .then(() => {
        message(`已删除配置 [${row.configKey}]`, { type: "success" });
      })
      .finally(() => onSearch());
  }

  function openDialog(title = "新增", row?: FormItemProps) {
    addDialog({
      title: `${title}系统配置`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          configKey: row?.configKey ?? "",
          configValue: row?.configValue ?? "",
          configDesc: row?.configDesc ?? ""
        }
      },
      width: "500px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;
        FormRef.validate(valid => {
          if (!valid) return;
          setSysConfig(curData)
            .then(() => {
              message(`${title}成功`, { type: "success" });
              done();
              onSearch();
            })
            .catch(err => {
              message(err?.message ?? "操作失败", { type: "error" });
            });
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
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
