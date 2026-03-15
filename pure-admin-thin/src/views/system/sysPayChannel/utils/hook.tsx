import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  delSysPayChannel,
  getSysPayChannelList,
  getSysPayChannelMethods,
  setSysPayChannel,
  setSysPayChannelMethods,
  type SysPayChannel
} from "@/api/sysPayChannel";
import { type Ref, h, reactive, ref, onMounted } from "vue";
import type { FormItemProps } from "./types";

export function useSysPayChannel(tableRef: Ref) {
  const form = reactive({
    channelCode: "",
    channelName: "",
    channelType: "",
    providerType: "",
    countryCode: "",
    status: null
  });

  const formRef = ref();
  const dataList = ref<SysPayChannel[]>([]);
  const loading = ref(true);

  const pagination = reactive({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const channelTypeMap: Record<string, string> = {
    deposit: "充值",
    withdraw: "提现",
    both: "充值+提现"
  };

  const providerTypeMap: Record<string, string> = {
    third_party: "三方",
    native: "自有"
  };

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    { label: "通道编码", prop: "channelCode", minWidth: 160 },
    { label: "通道名称", prop: "channelName", minWidth: 150 },
    {
      label: "通道类型",
      prop: "channelType",
      minWidth: 120,
      formatter: ({ channelType }) => channelTypeMap[channelType] ?? channelType
    },
    {
      label: "提供方",
      prop: "providerType",
      minWidth: 100,
      formatter: ({ providerType }) =>
        providerTypeMap[providerType] ?? providerType
    },
    {
      label: "国家码",
      prop: "countryCode",
      minWidth: 90,
      formatter: ({ countryCode }) => countryCode || "-"
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
      const { data } = await getSysPayChannelList({
        ...form,
        currentPage: pagination.currentPage - 1,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = (data?.currentPage ?? 0) + 1;
    } catch {
      message("获取支付通道列表失败", { type: "error" });
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

  async function handleDelete(row: SysPayChannel) {
    await delSysPayChannel(row.id);
    message(`已删除通道 ${row.channelName}`, { type: "success" });
    onSearch();
  }

  async function openDialog(title = "新增", row?: SysPayChannel) {
    let methodIds: number[] = [];
    if (row?.id) {
      try {
        const { data } = await getSysPayChannelMethods(row.id);
        methodIds = (data || []).map(item => item.methodId);
      } catch {
        // ignore
      }
    }

    addDialog({
      title: `${title}支付通道`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          channelCode: row?.channelCode ?? "",
          channelName: row?.channelName ?? "",
          channelType: row?.channelType ?? "",
          providerType: row?.providerType ?? "third_party",
          countryCode: row?.countryCode ?? "",
          icon: row?.icon ?? "",
          remark: row?.remark ?? "",
          sort: row?.sort ?? 0,
          status: row?.status ?? 1,
          methodIds
        }
      },
      width: "60%",
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
              const result = await setSysPayChannel({
                id: curData.id || undefined,
                channelCode: curData.channelCode.trim().toUpperCase(),
                channelName: curData.channelName.trim(),
                channelType: curData.channelType,
                providerType: curData.providerType,
                countryCode: curData.countryCode.trim() || null,
                icon: curData.icon.trim() || null,
                remark: curData.remark.trim() || null,
                sort: Number(curData.sort || 0),
                status: curData.status
              });
              const channelId = result.data?.id ?? curData.id;
              if (channelId) {
                await setSysPayChannelMethods({
                  channelId,
                  methodIds: curData.methodIds
                });
              }
              message(`您${title}了通道 ${curData.channelName}`, {
                type: "success"
              });
              done();
              onSearch();
            } catch {
              message("保存支付通道失败", { type: "error" });
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
