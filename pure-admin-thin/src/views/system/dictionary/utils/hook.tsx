import dayjs from "dayjs";
import editForm from "../form.vue";
import editItemForm from "../Itemform.vue";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import {
  getDictTypes,
  getDictItems,
  delDictItems,
  delDictTypes,
  setDictType,
  setDictItem
} from "@/api/system";
import { type Ref, h, ref, toRaw, reactive, onMounted, computed } from "vue";
import type { FormItemProps, ItemFormItemProps } from "./types";

export function useDictType(tableRef: Ref) {
  const form = reactive({
    dictName: "",
    dictType: "",
    status: null
  });

  const formRef = ref();
  const itemFormRef = ref();
  const dataList = ref([]);
  const loading = ref(true);
  const selectedNum = ref(0);

  const itemsData = reactive({});

  const itemColumns: TableColumnList = [
    {
      label: "id",
      prop: "id",
      width: 70
    },
    {
      label: "字典类型",
      prop: "dictType",
      minWidth: 120
    },
    {
      label: "字典项名称",
      prop: "dictLabel",
      minWidth: 120
    },
    {
      label: "字典项值",
      prop: "dictValue",
      minWidth: 120
    },
    {
      label: "字典代码",
      prop: "code",
      minWidth: 120
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 80,
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
      label: "排序",
      prop: "sort",
      minWidth: 80
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 120,
      showOverflowTooltip: true
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "操作",
      fixed: "right",
      width: 250,
      slot: "operation"
    }
  ];

  const buttonClass = computed(() => {
    return [
      "!h-[20px]",
      "reset-margin",
      "!text-gray-500",
      "dark:!text-white",
      "dark:hover:!text-primary"
    ];
  });

  const columns: TableColumnList = [
    {
      label: "",
      type: "expand",
      width: 70,
      slot: "expand"
    },
    {
      label: "id",
      prop: "id",
      width: 70
    },
    {
      label: "字典名称",
      prop: "dictName",
      minWidth: 120
    },
    {
      label: "字典类型",
      prop: "dictType",
      minWidth: 120
    },
    {
      label: "描述",
      prop: "description",
      minWidth: 150,
      showOverflowTooltip: true
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 80,
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
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "操作",
      fixed: "right",
      width: 180,
      slot: "operation"
    }
  ];

  async function handleItemsDelete(row) {
    await delDictItems({ ids: [row.id] });
    message(`已删除字典项 ${row.dictLabel}`, { type: "success" });
    onSearch();
  }

  async function handleDelete(row) {
    await delDictTypes({ ids: [row.id] });
    message(`已删除字典类型 ${row.dictName}`, { type: "success" });
    onSearch();
  }

  function handleSizeChange() {
    onSearch();
  }

  function handleCurrentChange() {
    onSearch();
  }

  function handleSelectionChange(val) {
    selectedNum.value = val.length;
    // 保证tableRef.value可以调用到clearSelection方法
    tableRef.value.setAdaptive();
  }

  function onSelectionCancel() {
    selectedNum.value = 0;
    // 用于清空选中项
    tableRef.value.getTableRef().clearSelection();
  }

  async function onbatchDel() {
    // 返回当前选中的行
    const curSelected = tableRef.value.getTableRef().getSelectionRows();
    await delDictTypes({ ids: curSelected.map(item => item.id) });
    message(`已删除${curSelected.length}条数据`, { type: "success" });
    tableRef.value.getTableRef().clearSelection();
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getDictTypes({
        ...toRaw(form)
      });
      dataList.value = data || [];
    } catch (error) {
      message("当前请求频率过高，请稍后再试！", { type: "error" });
    } finally {
      loading.value = false;
    }
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  };

  function openDialog(title = "新增", row?: FormItemProps) {
    addDialog({
      title: `${title}字典`,
      props: {
        formInline: {
          title,
          dictName: row?.dictName ?? "",
          dictType: row?.dictType ?? "",
          description: row?.description ?? "",
          status: row?.status ?? 1,
          id: row?.id ?? 0
        }
      },
      width: "40%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef }),
      beforeSure: async (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        function chores() {
          message(`您${title}了字典类型名称为${curData.dictName}的这条数据`, {
            type: "success"
          });
          done(); // 关闭弹框
          onSearch(); // 刷新表格数据
        }

        FormRef.validate(async valid => {
          if (valid) {
            try {
              const response = await setDictType(curData);
              if (response.code === 200) {
                // 表单规则校验通过
                if (title === "新增") {
                  // 实际开发中这里的模拟接口换成真实接口即可
                  chores();
                } else {
                  // 实际开发中这里的模拟接口换成真实接口即可
                  chores();
                }
              } else {
                message(response.message || "操作失败", { type: "error" });
              }
            } catch (error) {
              console.error("添加异常:", error);
              let errorMsg = "请求出错";
              if (error.response) {
                const { status, data } = error.response;
                errorMsg = data.message + ",请修改新的字典类型!";
                if (status === 400) message(errorMsg, { type: "error" });
              }
            }
          }
        });
      }
    });
  }

  function openItemDialog(title = "新增", row?: ItemFormItemProps) {
    addDialog({
      title: `${title}字典项`,
      props: {
        formInline: {
          title,
          dictType: row?.dictType ?? "",
          dictLabel: row?.dictLabel ?? "",
          dictValue: row?.dictValue ?? "",
          code: row?.code ?? "",
          color: row?.color ?? "",
          status: row?.status ?? 1,
          isDefault: row?.isDefault ?? 0,
          sort: row?.sort ?? 0,
          remark: row?.remark ?? "",
          id: row?.id ?? 0
        }
      },
      width: "50%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editItemForm, { ref: itemFormRef }),
      beforeSure: async (done, { options }) => {
        const dictItemFormRef = itemFormRef.value.getRef();
        const curData = options.props.formInline as ItemFormItemProps;

        async function chores() {
          message(`您${title}了字典项${curData.dictLabel}`, {
            type: "success"
          });
          done();
          // onSearch();
          const { data } = await getDictItems({ dictType: row.dictType });
          itemsData[row.dictType] = data;
        }

        dictItemFormRef.validate(async valid => {
          if (valid) {
            try {
              const response = await setDictItem(curData);
              if (response.code === 200) {
                if (title === "新增") {
                  chores();
                } else {
                  chores();
                }
              } else {
                message(response.message || "操作失败", { type: "error" });
              }
            } catch (error) {
              console.error("添加异常:", error);
              let errorMsg = "请求出错";
              if (error.response) {
                const { status, data } = error.response;
                errorMsg = data.message + ",请修改新的字典项!";
                if (status === 400) message(errorMsg, { type: "error" });
              }
            }
          }
        });
      }
    });
  }

  async function handleExpandChange(row: any, expandedRows: any[] | boolean) {
    if (Array.isArray(expandedRows)) {
      console.log("当前所有展开的行:", expandedRows);
      const isExpanded = expandedRows.some(item => item.id === row.id);
      console.log(`行 ${row.id} 是否展开:`, isExpanded);
      if (isExpanded) {
        const { data } = await getDictItems({ dictType: row.dictType });
        itemsData[row.dictType] = data;
      }
    } else {
      if (expandedRows) {
        const { data } = await getDictItems({ dictType: row.dictType });
        itemsData[row.dictType] = data;
      }
    }
  }

  /** 数据权限 可自行开发 */
  // function handleDatabase() {}

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    itemColumns,
    columns,
    dataList,
    selectedNum,
    // buttonClass,
    deviceDetection,
    onSearch,
    resetForm,
    onbatchDel,
    openDialog,
    handleDelete,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange,
    handleExpandChange,
    itemsData,
    handleItemsDelete,
    buttonClass,
    openItemDialog
  };
}
