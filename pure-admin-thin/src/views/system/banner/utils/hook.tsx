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
import type {
  BannerCountryFormItem,
  BannerI18nFormItem,
  FormItemProps
} from "./types";

const positionMap: Record<string, string> = {
  home: "首页",
  popup: "首页弹窗",
  activity: "活动页",
  member_center: "会员中心"
};

const platformMap: Record<string, string> = {
  all: "全部",
  web: "Web",
  app: "App",
  h5: "H5",
  pc: "PC"
};

const jumpTypeMap: Record<string, string> = {
  none: "不跳转",
  url: "外部链接",
  internal: "站内页面",
  product: "商品",
  activity: "活动",
  category: "分类"
};

const bannerTypeMap: Record<string, string> = {
  image: "图片",
  video: "视频",
  popup: "弹窗"
};

const displayTypeMap: Record<string, string> = {
  banner: "轮播",
  popup: "弹窗",
  float: "悬浮"
};

function createEmptyI18nItem(): BannerI18nFormItem {
  return {
    id: 0,
    tenantId: 0,
    bannerId: 0,
    languageCode: "zh-CN",
    countryCode: "",
    title: "",
    subTitle: "",
    description: "",
    buttonText: "",
    imageUrl: "",
    thumbUrl: "",
    bgImageUrl: "",
    iconUrl: "",
    videoUrl: "",
    jumpValue: "",
    textColor: "",
    buttonColor: "",
    bgColor: "",
    status: 1,
    remark: ""
  };
}

function createEmptyCountryItem(): BannerCountryFormItem {
  return {
    id: 0,
    tenantId: 0,
    bannerId: 0,
    countryCode: "",
    status: 1,
    remark: ""
  };
}

function getDefaultI18n(row: SysBanner) {
  return row.i18nList?.find(item => item.status === 1) || row.i18nList?.[0];
}

function normalizeOptional(value?: string | null) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : null;
}

function mapI18nItem(item: BannerI18nFormItem, tenantId: number) {
  return {
    id: item.id || undefined,
    tenantId,
    bannerId: item.bannerId || undefined,
    languageCode: item.languageCode.trim(),
    countryCode: normalizeOptional(item.countryCode),
    title: normalizeOptional(item.title),
    subTitle: normalizeOptional(item.subTitle),
    description: normalizeOptional(item.description),
    buttonText: normalizeOptional(item.buttonText),
    imageUrl: item.imageUrl.trim(),
    thumbUrl: normalizeOptional(item.thumbUrl),
    bgImageUrl: normalizeOptional(item.bgImageUrl),
    iconUrl: normalizeOptional(item.iconUrl),
    videoUrl: normalizeOptional(item.videoUrl),
    jumpValue: normalizeOptional(item.jumpValue),
    textColor: normalizeOptional(item.textColor),
    buttonColor: normalizeOptional(item.buttonColor),
    bgColor: normalizeOptional(item.bgColor),
    status: item.status,
    remark: normalizeOptional(item.remark)
  };
}

function mapCountryItem(item: BannerCountryFormItem, tenantId: number) {
  return {
    id: item.id || undefined,
    tenantId,
    bannerId: item.bannerId || undefined,
    countryCode: item.countryCode.trim(),
    status: item.status,
    remark: normalizeOptional(item.remark)
  };
}

function mapRowToForm(row?: SysBanner): FormItemProps {
  return {
    title: row ? "修改" : "新增",
    id: row?.id ?? 0,
    tenantId: row?.tenantId ?? 0,
    bannerName: row?.bannerName ?? "",
    bannerCode: row?.bannerCode ?? "",
    position: row?.position ?? "home",
    platform: row?.platform ?? "all",
    bannerType: row?.bannerType ?? "image",
    jumpType: row?.jumpType ?? "none",
    displayType: row?.displayType ?? "banner",
    openMode: row?.openMode ?? "current",
    sort: row?.sort ?? 0,
    status: row?.status ?? 1,
    startTime: row?.startTime
      ? dayjs(row.startTime).format("YYYY-MM-DD HH:mm:ss")
      : "",
    endTime: row?.endTime
      ? dayjs(row.endTime).format("YYYY-MM-DD HH:mm:ss")
      : "",
    version: row?.version ?? 1,
    remark: row?.remark ?? "",
    i18nList: row?.i18nList?.map(item => ({
      id: item.id ?? 0,
      tenantId: item.tenantId ?? row.tenantId ?? 0,
      bannerId: item.bannerId ?? row.id ?? 0,
      languageCode: item.languageCode ?? "",
      countryCode: item.countryCode ?? "",
      title: item.title ?? "",
      subTitle: item.subTitle ?? "",
      description: item.description ?? "",
      buttonText: item.buttonText ?? "",
      imageUrl: item.imageUrl ?? "",
      thumbUrl: item.thumbUrl ?? "",
      bgImageUrl: item.bgImageUrl ?? "",
      iconUrl: item.iconUrl ?? "",
      videoUrl: item.videoUrl ?? "",
      jumpValue: item.jumpValue ?? "",
      textColor: item.textColor ?? "",
      buttonColor: item.buttonColor ?? "",
      bgColor: item.bgColor ?? "",
      status: item.status ?? 1,
      remark: item.remark ?? ""
    })) ?? [createEmptyI18nItem()],
    countryList: row?.countryList?.map(item => ({
      id: item.id ?? 0,
      tenantId: item.tenantId ?? row.tenantId ?? 0,
      bannerId: item.bannerId ?? row.id ?? 0,
      countryCode: item.countryCode ?? "",
      status: item.status ?? 1,
      remark: item.remark ?? ""
    })) ?? [createEmptyCountryItem()]
  };
}

export function useSysBanner(tableRef: Ref) {
  const form = reactive({
    tenantId: undefined as number | undefined,
    bannerName: "",
    bannerCode: "",
    position: "",
    platform: "",
    bannerType: "",
    jumpType: "",
    displayType: "",
    languageCode: "",
    countryCode: "",
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
      label: "编码",
      prop: "bannerCode",
      minWidth: 140,
      formatter: ({ bannerCode }) => bannerCode || "-"
    },
    {
      label: "默认语言",
      prop: "languageCode",
      minWidth: 100,
      formatter: row => getDefaultI18n(row)?.languageCode || "-"
    },
    {
      label: "主图",
      prop: "imageUrl",
      minWidth: 110,
      cellRenderer: scope => {
        const i18n = getDefaultI18n(scope.row);
        if (!i18n?.imageUrl) return "-";
        return h(ElImage, {
          src: i18n.thumbUrl || i18n.imageUrl,
          previewSrcList: [i18n.imageUrl],
          previewTeleported: true,
          fit: "cover",
          style:
            "width: 64px; height: 36px; border-radius: 6px; border: 1px solid var(--el-border-color-light);"
        });
      }
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
      label: "类型",
      prop: "bannerType",
      minWidth: 90,
      formatter: ({ bannerType }) => bannerTypeMap[bannerType] ?? bannerType
    },
    {
      label: "展示",
      prop: "displayType",
      minWidth: 90,
      formatter: ({ displayType }) => displayTypeMap[displayType] ?? displayType
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
      formatter: row => getDefaultI18n(row)?.jumpValue || "-"
    },
    {
      label: "国家",
      prop: "countryList",
      minWidth: 140,
      formatter: ({ countryList }) =>
        countryList?.map(item => item.countryCode).join(", ") || "-"
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
      message("获取Banner列表失败", { type: "error" });
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
    message(`已删除Banner ${row.bannerName}`, { type: "success" });
    onSearch();
  }

  async function openDialog(title = "新增", row?: SysBanner) {
    addDialog({
      title: `${title}Banner`,
      props: {
        formInline: mapRowToForm(row)
      },
      width: "75%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef }),
      beforeSure: async (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        FormRef.validate(async valid => {
          if (!valid) return;
          try {
            const tenantId = Number(curData.tenantId || 0);
            await setSysBanner({
              id: curData.id || undefined,
              tenantId,
              bannerName: curData.bannerName.trim(),
              bannerCode: normalizeOptional(curData.bannerCode),
              position: curData.position,
              platform: curData.platform,
              bannerType: curData.bannerType,
              jumpType: curData.jumpType,
              displayType: curData.displayType,
              openMode: curData.openMode,
              sort: Number(curData.sort || 0),
              status: curData.status,
              startTime: curData.startTime
                ? dayjs(curData.startTime).unix()
                : null,
              endTime: curData.endTime ? dayjs(curData.endTime).unix() : null,
              version: Number(curData.version || 1),
              remark: normalizeOptional(curData.remark),
              i18nList: curData.i18nList
                .filter(
                  item => item.languageCode.trim() && item.imageUrl.trim()
                )
                .map(item => mapI18nItem(item, tenantId)),
              countryList: curData.countryList
                .filter(item => item.countryCode.trim())
                .map(item => mapCountryItem(item, tenantId))
            });
            message(`您${title}了Banner ${curData.bannerName}`, {
              type: "success"
            });
            done();
            onSearch();
          } catch {
            message("保存Banner失败", { type: "error" });
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
