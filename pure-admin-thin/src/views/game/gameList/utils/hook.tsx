import dayjs from "dayjs";
import editForm from "@/components/AppGameEditForm/index.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import type { PaginationProps } from "@pureadmin/table";
import { type Ref, h, onMounted, reactive, ref, toRaw } from "vue";
import {
  getAppGameList,
  setAppGame,
  syncAppGames,
  type AppGame
} from "@/api/appGame";
import type { FormItemProps } from "./types";

const gameTypeOptions = [
  { label: "拉霸/电子", value: 0 },
  { label: "Mini游戏", value: 1 },
  { label: "视讯游戏", value: 2 },
  { label: "捕鱼游戏", value: 3 },
  { label: "彩票游戏", value: 4 },
  { label: "体育游戏", value: 5 }
];

const syncPlatformOptions = [
  { label: "HG", value: "hg" },
  { label: "GSC", value: "gsc" },
  { label: "GGR", value: "ggr" }
];

function getGameTypeLabel(value?: number | null) {
  const match = gameTypeOptions.find(item => item.value === value);
  return match ? match.label : (value ?? "-");
}

function formatTime(value?: string | null) {
  if (!value) return "-";
  return dayjs(value).format("YYYY-MM-DD HH:mm:ss");
}

function renderImage(url?: string | null) {
  if (!url) return <span>-</span>;
  return (
    <el-image
      src={url}
      preview-src-list={[url]}
      preview-teleported
      fit="cover"
      style="width: 52px; height: 52px; border-radius: 6px"
    />
  );
}

function toOptionalString(value?: string | null) {
  const text = value?.trim() || "";
  return text || null;
}

export function useAppGame(tableRef: Ref) {
  const form = reactive({
    gameName: "",
    platformCode: "",
    thirdGameId: "",
    thirdGameCategory: "",
    type: undefined as number | undefined,
    homeShow: undefined as number | undefined,
    hot: undefined as number | undefined,
    categoryCode: "" as string | undefined,
    disabledFlag: undefined as number | undefined
  });
  const syncForm = reactive({
    language: "en",
    platformCode: "hg"
  });
  const formRef = ref();
  const dataList = ref<AppGame[]>([]);
  const loading = ref(true);
  const syncLoading = ref(false);
  const savingRows = ref<Record<number, boolean>>({});
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "gameId", minWidth: 90 },
    {
      label: "图标",
      prop: "gameIcon",
      minWidth: 90,
      cellRenderer: ({ row }) => renderImage(row.gameIcon)
    },
    {
      label: "游戏名称",
      prop: "gameName",
      minWidth: 220,
      formatter: ({ gameName }) => gameName || "-"
    },
    {
      label: "第三方游戏ID",
      prop: "thirdGameId",
      minWidth: 160,
      formatter: ({ thirdGameId }) => thirdGameId || "-"
    },
    {
      label: "平台",
      prop: "platformCode",
      minWidth: 100,
      formatter: ({ platformCode }) => platformCode || "-"
    },
    {
      label: "类型",
      prop: "type",
      minWidth: 110,
      formatter: ({ type }) => getGameTypeLabel(type)
    },
    {
      label: "厂商",
      prop: "thirdGameCategory",
      minWidth: 120,
      formatter: ({ thirdGameCategory }) => thirdGameCategory || "-"
    },
    {
      label: "分类",
      prop: "categoryCode",
      minWidth: 120,
      formatter: ({ categoryCode }) => categoryCode || "-"
    },
    {
      label: "热门",
      prop: "hot",
      minWidth: 95,
      cellRenderer: ({ row }) => (
        <el-switch
          modelValue={row.hot === 1}
          loading={savingRows.value[row.gameId]}
          inline-prompt
          active-text="热门"
          inactive-text="普通"
          onChange={(value: boolean) =>
            handleRowPatch(row, { hot: value ? 1 : 0 })
          }
        />
      )
    },
    {
      label: "首页展示",
      prop: "homeShow",
      minWidth: 110,
      cellRenderer: ({ row }) => (
        <el-switch
          modelValue={row.homeShow === 1}
          loading={savingRows.value[row.gameId]}
          inline-prompt
          active-text="展示"
          inactive-text="隐藏"
          onChange={(value: boolean) =>
            handleRowPatch(row, { homeShow: value ? 1 : 0 })
          }
        />
      )
    },
    {
      label: "状态",
      prop: "disabledFlag",
      minWidth: 95,
      cellRenderer: ({ row }) => (
        <el-switch
          modelValue={row.disabledFlag !== 1}
          loading={savingRows.value[row.gameId]}
          inline-prompt
          active-text="启用"
          inactive-text="禁用"
          onChange={(value: boolean) =>
            handleRowPatch(row, { disabledFlag: value ? 0 : 1 })
          }
        />
      )
    },
    {
      label: "排序",
      prop: "sort",
      minWidth: 120,
      cellRenderer: ({ row }) => (
        <el-input-number
          modelValue={row.sort ?? 0}
          min={0}
          controls-position="right"
          size="small"
          style="width: 96px"
          disabled={savingRows.value[row.gameId]}
          onChange={(value: number) =>
            handleRowPatch(row, { sort: Number(value || 0) })
          }
        />
      )
    },
    {
      label: "更新时间",
      prop: "updateTime",
      minWidth: 170,
      formatter: ({ updateTime }) => formatTime(updateTime)
    },
    {
      label: "创建时间",
      prop: "createTime",
      minWidth: 170,
      formatter: ({ createTime }) => formatTime(createTime)
    },
    { label: "操作", fixed: "right", width: 110, slot: "operation" }
  ];

  async function fetchList() {
    loading.value = true;
    try {
      const { data } = await getAppGameList({
        ...toRaw(form),
        currentPage: pagination.currentPage,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = data?.currentPage || 0;
    } catch (error) {
      console.error("获取游戏列表失败", error);
      message("获取游戏列表失败", { type: "error" });
    } finally {
      loading.value = false;
      tableRef.value?.setAdaptive?.();
    }
  }

  function onSearch() {
    pagination.currentPage = 0;
    fetchList();
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 0;
    fetchList();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val - 1;
    fetchList();
  }

  async function handleSync() {
    syncLoading.value = true;
    try {
      const { data } = await syncAppGames(toRaw(syncForm));
      message(
        `同步完成：总数 ${data.total}，新增 ${data.created}，更新 ${data.updated}，跳过 ${data.skipped}`,
        { type: "success" }
      );
      onSearch();
    } catch (error) {
      console.error("同步游戏列表失败", error);
      message("同步游戏列表失败", { type: "error" });
    } finally {
      syncLoading.value = false;
    }
  }

  async function handleRowPatch(row: AppGame, patch: Partial<AppGame>) {
    if (!row.gameId || savingRows.value[row.gameId]) return;
    savingRows.value[row.gameId] = true;
    try {
      const next = { ...row, ...patch };
      await setAppGame({
        gameId: next.gameId,
        gameName: toOptionalString(next.gameName),
        categoryCode: toOptionalString(next.categoryCode),
        showIndex: Number(next.showIndex || 0),
        hot: Number(next.hot || 0),
        homeShow: Number(next.homeShow || 0),
        type: typeof next.type === "number" ? next.type : null,
        parentId: next.parentId || null,
        platformCode: toOptionalString(next.platformCode),
        thirdGameId: toOptionalString(next.thirdGameId),
        thirdGameName:
          toOptionalString(next.thirdGameName) ||
          toOptionalString(next.gameName),
        thirdGameCategory: toOptionalString(next.thirdGameCategory),
        horizontalImage: toOptionalString(next.horizontalImage),
        gameIcon: toOptionalString(next.gameIcon),
        typeIcon: toOptionalString(next.typeIcon),
        typeActiveIcon: toOptionalString(next.typeActiveIcon),
        tenantId: next.tenantId || null,
        sort: Number(next.sort || 0),
        disabledFlag: Number(next.disabledFlag || 0),
        deletedFlag: Number(next.deletedFlag || 0),
        remark: toOptionalString(next.remark)
      });
      Object.assign(row, patch);
      message("修改成功", { type: "success" });
    } catch (error) {
      console.error("保存游戏失败", error);
      message("保存游戏失败", { type: "error" });
      fetchList();
    } finally {
      savingRows.value[row.gameId] = false;
    }
  }

  function openDialog(title = "修改", row?: AppGame) {
    addDialog({
      title: `${title}游戏`,
      props: {
        formInline: buildEditFormData(title, row)
      },
      width: "62%",
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
            await saveAppGameForm(curData);
            message(`已${title}游戏 ${curData.gameName}`, { type: "success" });
            done();
            onSearch();
          } catch (error) {
            console.error("保存游戏失败", error);
            message("保存游戏失败", { type: "error" });
          }
        });
      }
    });
  }

  function buildEditFormData(title: string, row?: AppGame): FormItemProps {
    return {
      title,
      gameId: row?.gameId ?? 0,
      gameName: row?.gameName ?? "",
      categoryCode: row?.categoryCode ?? "",
      showIndex: row?.showIndex ?? 0,
      hot: row?.hot ?? 0,
      homeShow: row?.homeShow ?? 0,
      type: typeof row?.type === "number" ? row.type : null,
      parentId: row?.parentId ?? null,
      platformCode: row?.platformCode ?? "",
      thirdGameId: row?.thirdGameId ?? "",
      thirdGameName: row?.thirdGameName ?? row?.gameName ?? "",
      thirdGameCategory: row?.thirdGameCategory ?? "",
      horizontalImage: row?.horizontalImage ?? "",
      gameIcon: row?.gameIcon ?? "",
      typeIcon: row?.typeIcon ?? "",
      typeActiveIcon: row?.typeActiveIcon ?? "",
      tenantId: row?.tenantId ?? null,
      sort: row?.sort ?? 0,
      disabledFlag: row?.disabledFlag ?? 0,
      deletedFlag: row?.deletedFlag ?? 0,
      remark: row?.remark ?? ""
    };
  }

  async function saveAppGameForm(curData: FormItemProps) {
    await setAppGame({
      gameId: curData.gameId || undefined,
      gameName: curData.gameName.trim(),
      categoryCode: toOptionalString(curData.categoryCode),
      showIndex: Number(curData.showIndex || 0),
      hot: Number(curData.hot || 0),
      homeShow: Number(curData.homeShow || 0),
      type: typeof curData.type === "number" ? curData.type : null,
      parentId: curData.parentId || null,
      platformCode: curData.platformCode.trim(),
      thirdGameId: curData.thirdGameId.trim(),
      thirdGameName:
        toOptionalString(curData.thirdGameName) || curData.gameName.trim(),
      thirdGameCategory: toOptionalString(curData.thirdGameCategory),
      horizontalImage: toOptionalString(curData.horizontalImage),
      gameIcon: toOptionalString(curData.gameIcon),
      typeIcon: toOptionalString(curData.typeIcon),
      typeActiveIcon: toOptionalString(curData.typeActiveIcon),
      tenantId: curData.tenantId || null,
      sort: Number(curData.sort || 0),
      disabledFlag: Number(curData.disabledFlag || 0),
      deletedFlag: Number(curData.deletedFlag || 0),
      remark: toOptionalString(curData.remark)
    });
  }

  onMounted(() => {
    fetchList();
  });

  return {
    form,
    syncForm,
    loading,
    syncLoading,
    columns,
    dataList,
    pagination,
    gameTypeOptions,
    syncPlatformOptions,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    openDialog,
    handleSync
  };
}
