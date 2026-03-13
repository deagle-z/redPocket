import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  delSysCustomField,
  getSysCustomFieldList,
  setSysCustomField,
  type SysCustomField
} from "@/api/customField";
import { type Ref, h, reactive, ref, onMounted } from "vue";
import type { FormItemProps } from "./types";

function toOptionalString(value?: string | null) {
  const text = value?.trim() || "";
  return text || "";
}

function toOptionalNumber(value?: number | null) {
  return typeof value === "number" && !Number.isNaN(value) ? value : null;
}

export function useCustomField(tableRef: Ref) {
  const form = reactive({
    fieldKey: "",
    fieldLabel: "",
    fieldType: "",
    dataType: "",
    status: null
  });

  const formRef = ref();
  const dataList = ref<SysCustomField[]>([]);
  const loading = ref(true);

  const pagination = reactive({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    { label: "字段Key", prop: "fieldKey", minWidth: 180 },
    { label: "字段名称", prop: "fieldLabel", minWidth: 150 },
    { label: "字段类型", prop: "fieldType", minWidth: 110 },
    { label: "数据类型", prop: "dataType", minWidth: 110 },
    {
      label: "必填",
      prop: "isRequired",
      minWidth: 80,
      cellRenderer: scope => (
        <el-tag
          size="small"
          type={scope.row.isRequired === 1 ? "danger" : "info"}
          effect="plain"
        >
          {scope.row.isRequired === 1 ? "是" : "否"}
        </el-tag>
      )
    },
    {
      label: "敏感",
      prop: "isSensitive",
      minWidth: 80,
      cellRenderer: scope => (
        <el-tag
          size="small"
          type={scope.row.isSensitive === 1 ? "warning" : "info"}
          effect="plain"
        >
          {scope.row.isSensitive === 1 ? "是" : "否"}
        </el-tag>
      )
    },
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
      const { data } = await getSysCustomFieldList({
        ...form,
        currentPage: pagination.currentPage - 1,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = (data?.currentPage ?? 0) + 1;
    } catch {
      message("获取自定义字段列表失败", { type: "error" });
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

  async function handleDelete(row: SysCustomField) {
    await delSysCustomField(row.id);
    message(`已删除字段 ${row.fieldLabel}`, { type: "success" });
    onSearch();
  }

  function openDialog(title = "新增", row?: SysCustomField) {
    addDialog({
      title: `${title}自定义字段`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          fieldKey: row?.fieldKey ?? "",
          fieldLabel: row?.fieldLabel ?? "",
          fieldPlaceholder: row?.fieldPlaceholder ?? "",
          fieldType: row?.fieldType ?? "input",
          dataType: row?.dataType ?? "string",
          defaultValue: row?.defaultValue ?? "",
          isRequired: row?.isRequired ?? 0,
          isSensitive: row?.isSensitive ?? 0,
          maxLength: row?.maxLength ?? null,
          minLength: row?.minLength ?? null,
          regexRule: row?.regexRule ?? "",
          errorTips: row?.errorTips ?? "",
          optionsJson: row?.optionsJson ?? "",
          status: row?.status ?? 1,
          remark: row?.remark ?? ""
        }
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
              await setSysCustomField({
                id: curData.id || undefined,
                fieldKey: curData.fieldKey.trim(),
                fieldLabel: curData.fieldLabel.trim(),
                fieldPlaceholder: toOptionalString(curData.fieldPlaceholder),
                fieldType: curData.fieldType,
                dataType: curData.dataType,
                defaultValue: toOptionalString(curData.defaultValue),
                isRequired: curData.isRequired,
                isSensitive: curData.isSensitive,
                maxLength: toOptionalNumber(curData.maxLength),
                minLength: toOptionalNumber(curData.minLength),
                regexRule: toOptionalString(curData.regexRule),
                errorTips: toOptionalString(curData.errorTips),
                optionsJson: toOptionalString(curData.optionsJson),
                status: curData.status,
                remark: toOptionalString(curData.remark)
              });
              message(`您${title}了字段 ${curData.fieldLabel}`, {
                type: "success"
              });
              done();
              onSearch();
            } catch {
              message("保存自定义字段失败", { type: "error" });
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
