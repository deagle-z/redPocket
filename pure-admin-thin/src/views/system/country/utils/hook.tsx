import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  delSysCountry,
  getSysCountryList,
  setSysCountry,
  type SysCountry
} from "@/api/country";
import type { SysCustomField } from "@/api/customField";
import { type Ref, h, reactive, ref, onMounted } from "vue";
import type { FormItemProps } from "./types";

function toOptionalString(value?: string | null) {
  const text = value?.trim() || "";
  return text || "";
}

function parseCustomFields(value?: string | null) {
  if (!value) return [] as SysCustomField[];
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? (parsed as SysCustomField[]) : [];
  } catch {
    return [];
  }
}

export function useCountry(tableRef: Ref) {
  const form = reactive({
    countryCode: "",
    countryName: "",
    currencyCode: "",
    status: null
  });

  const formRef = ref();
  const dataList = ref<SysCountry[]>([]);
  const loading = ref(true);

  const pagination = reactive({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    { label: "国家编码", prop: "countryCode", minWidth: 110 },
    { label: "国家中文名", prop: "countryNameCn", minWidth: 140 },
    { label: "国家英文名", prop: "countryNameEn", minWidth: 160 },
    { label: "币种编码", prop: "currencyCode", minWidth: 110 },
    {
      label: "货币符号",
      prop: "currencySymbol",
      minWidth: 100,
      formatter: ({ currencySymbol }) => currencySymbol || "-"
    },
    {
      label: "语言编码",
      prop: "languageCode",
      minWidth: 120,
      formatter: ({ languageCode }) => languageCode || "-"
    },
    {
      label: "时区",
      prop: "timezone",
      minWidth: 160,
      formatter: ({ timezone }) => timezone || "-"
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
      const { data } = await getSysCountryList({
        ...form,
        currentPage: pagination.currentPage - 1,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = (data?.currentPage ?? 0) + 1;
    } catch {
      message("获取国家列表失败", { type: "error" });
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

  async function handleDelete(row: SysCountry) {
    await delSysCountry(row.id);
    message(`已删除国家 ${row.countryNameCn}`, { type: "success" });
    onSearch();
  }

  function openDialog(title = "新增", row?: SysCountry) {
    addDialog({
      title: `${title}国家`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          countryCode: row?.countryCode ?? "",
          countryNameCn: row?.countryNameCn ?? "",
          countryNameEn: row?.countryNameEn ?? "",
          currencyCode: row?.currencyCode ?? "",
          currencySymbol: row?.currencySymbol ?? "",
          timezone: row?.timezone ?? "",
          languageCode: row?.languageCode ?? "",
          withdrawFields: parseCustomFields(row?.withdrawFields),
          rechargeFields: parseCustomFields(row?.rechargeFields),
          sort: row?.sort ?? 0,
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
              await setSysCountry({
                id: curData.id || undefined,
                countryCode: curData.countryCode.trim().toUpperCase(),
                countryNameCn: curData.countryNameCn.trim(),
                countryNameEn: curData.countryNameEn.trim(),
                currencyCode: curData.currencyCode.trim().toUpperCase(),
                currencySymbol: toOptionalString(curData.currencySymbol),
                timezone: toOptionalString(curData.timezone),
                languageCode: toOptionalString(curData.languageCode),
                withdrawFields:
                  curData.withdrawFields.length > 0
                    ? JSON.stringify(curData.withdrawFields)
                    : "",
                rechargeFields:
                  curData.rechargeFields.length > 0
                    ? JSON.stringify(curData.rechargeFields)
                    : "",
                sort: Number(curData.sort || 0),
                status: curData.status,
                remark: toOptionalString(curData.remark)
              });
              message(`您${title}了国家 ${curData.countryNameCn}`, {
                type: "success"
              });
              done();
              onSearch();
            } catch {
              message("保存国家失败", { type: "error" });
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
