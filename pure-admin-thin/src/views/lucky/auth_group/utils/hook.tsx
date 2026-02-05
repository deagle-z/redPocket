import dayjs from "dayjs";
import editForm from "../form.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { FormItemProps } from "../utils/types";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import { getAuthGroups, setAuthGroup, delAuthGroup, type AuthGroup } from "@/api/authGroup";
import { type Ref, reactive, ref, onMounted, h, toRaw } from "vue";
import { ElMessageBox, ElTag } from "element-plus";
import { usePublicHooks } from "@/views/system/hooks";

export function useAuthGroup(treeRef: Ref) {
  const form = reactive({
    groupId: undefined as number | undefined,
    status: undefined as number | undefined
  });
  const curRow = ref();
  const formRef = ref();
  const dataList = ref<AuthGroup[]>([]);
  const loading = ref(true);
  const switchLoadMap = ref({});
  const { switchStyle } = usePublicHooks();
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
      width: 80
    },
    {
      label: "群组ID",
      prop: "groupId",
      minWidth: 120
    },
    {
      label: "群组名称",
      prop: "groupName",
      minWidth: 150
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
      cellRenderer: scope => (
        <el-switch
          size={scope.props.size === "small" ? "small" : "default"}
          loading={switchLoadMap.value[scope.index]?.loading}
          v-model={scope.row.status}
          active-value={1}
          inactive-value={0}
          active-text="启用"
          inactive-text="禁用"
          inline-prompt
          style={switchStyle.value}
          onChange={() => onChange(scope as any)}
        />
      )
    },
    {
      label: "客服URL",
      prop: "serviceUrl",
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      label: "充值URL",
      prop: "rechargeUrl",
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      label: "玩法URL",
      prop: "channelUrl",
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      label: "发包图片",
      prop: "sendPacketImage",
      minWidth: 160,
      formatter: ({ sendPacketImage }) => (sendPacketImage ? "已配置" : "-")
    },
    {
      label: "中雷倍数",
      prop: "loseRate",
      minWidth: 100
    },
    {
      label: "数量配置",
      prop: "numConfig",
      minWidth: 120
    },
    {
      label: "发包抽成(%)",
      prop: "sendCommission",
      minWidth: 120
    },
    {
      label: "抢包抽成(%)",
      prop: "grabbingCommission",
      minWidth: 120
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) => dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "操作",
      fixed: "right",
      width: 150,
      slot: "operation"
    }
  ];

  function onChange({ row, index }) {
    ElMessageBox.confirm(
      `确认要${row.status === 1 ? "启用" : "禁用"}群组 ${row.groupName} 吗？`,
      "系统提示",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning"
      }
    )
      .then(() => {
        switchLoadMap.value[index] = Object.assign(
          switchLoadMap.value[index] || {},
          {
            loading: true
          }
        );
        setAuthGroup(row)
          .then(() => {
            switchLoadMap.value[index] = Object.assign(
              switchLoadMap.value[index] || {},
              {
                loading: false
              }
            );
            message(`已${row.status === 1 ? "启用" : "禁用"}群组 ${row.groupName}`, {
              type: "success"
            });
          })
          .catch(() => {
            switchLoadMap.value[index] = Object.assign(
              switchLoadMap.value[index] || {},
              {
                loading: false
              }
            );
          });
      })
      .catch(() => {
        row.status === 1 ? (row.status = 0) : (row.status = 1);
      });
  }

  function handleDelete(row: AuthGroup) {
    ElMessageBox.confirm(`确认要删除群组 ${row.groupName} 吗？`, "系统提示", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning"
    })
      .then(() => {
        delAuthGroup(row.id)
          .then(() => {
            message(`您删除了群组 ${row.groupName}`, { type: "success" });
          })
          .finally(() => {
            onSearch();
          });
      })
      .catch(() => {});
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
      const { data } = await getAuthGroups({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list;
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取授权群组列表失败", error);
      message("获取授权群组列表失败", { type: "error" });
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
      title: `${title}授权群组`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          groupId: row?.groupId ?? 0,
          groupName: row?.groupName ?? "",
          status: row?.status ?? 1,
          serviceUrl: row?.serviceUrl ?? "",
          rechargeUrl: row?.rechargeUrl ?? "",
          channelUrl: row?.channelUrl ?? "",
          sendPacketImage: row?.sendPacketImage ?? "",
          loseRate: row?.loseRate ?? 1.8,
          numConfig: row?.numConfig ?? "3",
          sendCommission: row?.sendCommission ?? 2,
          grabbingCommission: row?.grabbingCommission ?? 3
        }
      },
      width: "50%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        function chores() {
          setAuthGroup(curData)
            .then(() => {
              message(`您${title}了群组 ${curData.groupName}`, {
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
            console.log("curData", curData);
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
    onSearch,
    resetForm,
    openDialog,
    handleDelete,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
