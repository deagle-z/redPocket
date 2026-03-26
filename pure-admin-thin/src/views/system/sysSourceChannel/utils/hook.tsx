import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  delSysSourceChannel,
  getSysSourceChannelList,
  setSysSourceChannel,
  type SysSourceChannel
} from "@/api/sysSourceChannel";
import { type Ref, h, reactive, ref, onMounted } from "vue";
import type { FormItemProps } from "./types";

export function useSysSourceChannel(tableRef: Ref) {
  const form = reactive({
    tenantId: undefined as number | undefined,
    channelCode: "",
    channelName: "",
    level: null as number | null,
    status: null as number | null
  });

  const formRef = ref();
  const dataList = ref<SysSourceChannel[]>([]);
  const loading = ref(true);

  const pagination = reactive({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const levelMap: Record<number, string> = {
    1: "一级渠道",
    2: "二级渠道"
  };

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    { label: "渠道编码", prop: "channelCode", minWidth: 180 },
    { label: "渠道名称", prop: "channelName", minWidth: 160 },
    // {
    //   label: "父渠道ID",
    //   prop: "parentId",
    //   minWidth: 100,
    //   formatter: ({ parentId }) => parentId ?? "-"
    // },
    // {
    //   label: "渠道层级",
    //   prop: "level",
    //   minWidth: 100,
    //   formatter: ({ level }) => levelMap[level] ?? level
    // },
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
      const { data } = await getSysSourceChannelList({
        ...form,
        currentPage: pagination.currentPage - 1,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = (data?.currentPage ?? 0) + 1;
    } catch {
      message("获取来源渠道列表失败", { type: "error" });
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

  async function handleDelete(row: SysSourceChannel) {
    await delSysSourceChannel(row.id);
    message(`已删除来源渠道 ${row.channelName}`, { type: "success" });
    onSearch();
  }

  async function openDialog(title = "新增", row?: SysSourceChannel) {
    addDialog({
      title: `${title}来源渠道`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          tenantId: row?.tenantId ?? 0,
          channelCode: row?.channelCode ?? "",
          channelName: row?.channelName ?? "",
          parentId: row?.parentId ?? null,
          level: row?.level ?? 1,
          status: row?.status ?? 1,
          sort: row?.sort ?? 0,
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
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        FormRef.validate(async valid => {
          if (valid) {
            try {
              await setSysSourceChannel({
                id: curData.id || undefined,
                tenantId: Number(curData.tenantId || 0),
                channelCode: curData.channelCode.trim() || undefined,
                channelName: curData.channelName.trim(),
                parentId: curData.parentId || null,
                level: Number(curData.level || 1),
                status: curData.status,
                sort: Number(curData.sort || 0),
                remark: curData.remark.trim() || null
              });
              message(`您${title}了来源渠道 ${curData.channelName}`, {
                type: "success"
              });
              done();
              onSearch();
            } catch {
              message("保存来源渠道失败", { type: "error" });
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
