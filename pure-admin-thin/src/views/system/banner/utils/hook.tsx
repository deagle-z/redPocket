import dayjs from "dayjs";
import editForm from "../form.vue";
import { ElImage } from "element-plus";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  delSysBanner,
  getSysBannerList,
  setSysBanner,
  type SysBanner
} from "@/api/sysBanner";
import { type Ref, h, onMounted, reactive, ref } from "vue";
import type { FormItemProps } from "./types";

const positionMap: Record<string, string> = {
  home: "首页",
  popup: "首页弹窗",
  activity: "活动页"
};

const platformMap: Record<string, string> = {
  all: "全部",
  web: "Web",
  app: "App",
  h5: "H5"
};

const jumpTypeMap: Record<string, string> = {
  none: "不跳转",
  url: "外部链接",
  internal: "站内页面",
  product: "商品",
  activity: "活动"
};

export function useSysBanner(tableRef: Ref) {
  const form = reactive({
    tenantId: undefined as number | undefined,
    bannerName: "",
    position: "",
    platform: "",
    jumpType: "",
    status: null as number | null
  });

  const formRef = ref();
  const dataList = ref<SysBanner[]>([]);
  const loading = ref(true);

  const pagination = reactive({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    { label: "名称", prop: "bannerName", minWidth: 160 },
    {
      label: "图片",
      prop: "imageUrl",
      minWidth: 110,
      cellRenderer: scope =>
        h(ElImage, {
          src: scope.row.thumbUrl || scope.row.imageUrl,
          previewSrcList: [scope.row.imageUrl],
          previewTeleported: true,
          fit: "cover",
          style:
            "width: 64px; height: 36px; border-radius: 6px; border: 1px solid var(--el-border-color-light);"
        })
    },
    {
      label: "位置",
      prop: "position",
      minWidth: 100,
      formatter: ({ position }) => positionMap[position] ?? position
    },
    {
      label: "平台",
      prop: "platform",
      minWidth: 90,
      formatter: ({ platform }) => platformMap[platform] ?? platform
    },
    {
      label: "跳转类型",
      prop: "jumpType",
      minWidth: 100,
      formatter: ({ jumpType }) => jumpTypeMap[jumpType] ?? jumpType
    },
    {
      label: "跳转值",
      prop: "jumpValue",
      minWidth: 180,
      formatter: ({ jumpValue }) => jumpValue || "-"
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
      label: "投放时间",
      prop: "startTime",
      minWidth: 220,
      formatter: ({ startTime, endTime }) =>
        `${startTime ? dayjs(startTime).format("YYYY-MM-DD HH:mm:ss") : "-"} ~ ${
          endTime ? dayjs(endTime).format("YYYY-MM-DD HH:mm:ss") : "-"
        }`
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
      const { data } = await getSysBannerList({
        ...form,
        currentPage: pagination.currentPage - 1,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = (data?.currentPage ?? 0) + 1;
    } catch {
      message("获取轮播图列表失败", { type: "error" });
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

  async function handleDelete(row: SysBanner) {
    await delSysBanner(row.id);
    message(`已删除轮播图 ${row.bannerName}`, { type: "success" });
    onSearch();
  }

  async function openDialog(title = "新增", row?: SysBanner) {
    addDialog({
      title: `${title}轮播图`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          tenantId: row?.tenantId ?? 0,
          bannerName: row?.bannerName ?? "",
          position: row?.position ?? "home",
          platform: row?.platform ?? "all",
          imageUrl: row?.imageUrl ?? "",
          thumbUrl: row?.thumbUrl ?? "",
          jumpType: row?.jumpType ?? "none",
          jumpValue: row?.jumpValue ?? "",
          sort: row?.sort ?? 0,
          status: row?.status ?? 1,
          startTime: row?.startTime ? dayjs(row.startTime).format("YYYY-MM-DD HH:mm:ss") : "",
          endTime: row?.endTime ? dayjs(row.endTime).format("YYYY-MM-DD HH:mm:ss") : "",
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
              await setSysBanner({
                id: curData.id || undefined,
                tenantId: Number(curData.tenantId || 0),
                bannerName: curData.bannerName.trim(),
                position: curData.position,
                platform: curData.platform,
                imageUrl: curData.imageUrl.trim(),
                thumbUrl: curData.thumbUrl.trim() || null,
                jumpType: curData.jumpType,
                jumpValue: curData.jumpValue.trim() || null,
                sort: Number(curData.sort || 0),
                status: curData.status,
                startTime: curData.startTime ? dayjs(curData.startTime).unix() : null,
                endTime: curData.endTime ? dayjs(curData.endTime).unix() : null,
                remark: curData.remark.trim() || null
              });
              message(`您${title}了轮播图 ${curData.bannerName}`, {
                type: "success"
              });
              done();
              onSearch();
            } catch {
              message("保存轮播图失败", { type: "error" });
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
