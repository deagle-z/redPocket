import dayjs from "dayjs";
import editForm from "../form.vue";
import { handleTree } from "@/utils/tree";
import { message } from "@/utils/message";
import { transformI18n } from "@/plugins/i18n";
import { addDialog } from "@/components/ReDialog";
import type { FormItemProps } from "../utils/types";
import type { PaginationProps } from "@pureadmin/table";
import { getKeyList, deviceDetection } from "@pureadmin/utils";
import _ from "lodash";
import {
  delRole,
  getRoleList,
  getRoleMenu,
  getRoleMenuIds,
  setRole
} from "@/api/system";
import { type Ref, reactive, ref, onMounted, h, toRaw, watch } from "vue";

export function useRole(treeRef: Ref) {
  const form = reactive({
    name: "",
    code: ""
  });
  const curRow = ref();
  const formRef = ref();
  const dataList = ref([]);
  const treeIds = ref([]);
  const treeData = ref([]);
  const isShow = ref(false);
  const loading = ref(true);
  const isLinkage = ref(false);
  const treeSearchValue = ref();
  const isExpandAll = ref(false);
  const isSelectAll = ref(false);
  const treeProps = {
    value: "id",
    label: "nameCode",
    children: "children"
  };
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    {
      label: "角色编号",
      prop: "id"
    },
    {
      label: "角色名称",
      prop: "name"
    },
    {
      label: "角色标识",
      prop: "code"
    },
    {
      label: "备注",
      prop: "description",
      minWidth: 160
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
      width: 210,
      slot: "operation"
    }
  ];

  function handleDelete(row) {
    delRole(row.id)
      .then(() => {
        message(`您删除了角色名称为${row.name}的这条数据`, { type: "success" });
      })
      .finally(() => {
        onSearch();
      });
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    onSearch();
    console.log(`${val} items per page`);
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val - 1;
    onSearch();
    console.log(`current page: ${val}`);
  }

  function handleSelectionChange(val) {
    console.log("handleSelectionChange", val);
  }

  async function onSearch() {
    loading.value = true;
    const { data } = await getRoleList({
      ...toRaw(form),
      ...toRaw(pagination)
    });
    dataList.value = data.list;
    pagination.total = data.total;
    pagination.pageSize = data.pageSize;
    pagination.currentPage = data.currentPage;

    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  };

  function openDialog(title = "新增", row?: FormItemProps) {
    addDialog({
      title: `${title}角色`,
      props: {
        formInline: {
          id: row?.id ?? 0,
          name: row?.name ?? "",
          code: row?.code ?? "",
          description: row?.description ?? "",
          menuIds: row?.menuIds
        }
      },
      width: "40%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        function chores() {
          setRole(curData)
            .then(() => {
              message(`您${title}了角色名称为${curData.name}的这条数据`, {
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
            // 表单规则校验通过
            if (title === "新增") {
              // 实际开发先调用新增接口，再进行下面操作
              chores();
            } else {
              // 实际开发先调用修改接口，再进行下面操作
              chores();
            }
          }
        });
      }
    });
  }

  /** 菜单角色 */
  async function handleMenu(row?: any) {
    const { id } = row;
    if (id) {
      curRow.value = row;
      isShow.value = true;
      const { data } = await getRoleMenuIds(id);
      treeRef.value.setCheckedKeys(data);
    } else {
      curRow.value = null;
      isShow.value = false;
    }
  }

  /** 高亮当前角色选中行 */
  function rowStyle({ row: { id } }) {
    return {
      cursor: "pointer",
      background: id === curRow.value?.id ? "var(--el-fill-color-light)" : ""
    };
  }

  /** 菜单角色-保存 */
  function handleSave() {
    const { name } = curRow.value;
    console.log(treeRef.value.getCheckedKeys());
    curRow.value.menuIds = treeRef.value.getCheckedKeys();
    setRole(curRow.value)
      .then(() => {
        message(`角色名称为${name}的菜单角色修改成功`, {
          type: "success"
        });
      })
      .finally(() => {});
  }

  // 保存原始树形数据的备份
  const deepCloneRef = refObj => {
    const rawData = toRaw(refObj.value);
    return _.cloneDeep(rawData);
  };
  let originalTreeData = [];
  /** 数据角色 可自行开发 */
  const onQueryChanged = (query: string) => {
    if (!query.trim()) {
      treeData.value = JSON.parse(JSON.stringify(originalTreeData));
      treeRef.value!.filter("");
      return;
    }
    const tempTree = JSON.parse(JSON.stringify(treeData.value));
    const filteredTree = tempTree.filter(rootNode => {
      return filterAndModifyTree(rootNode, query);
    });
    treeData.value = filteredTree;
    treeRef.value.setExpandedKeys(treeIds.value);
  };

  function filterAndModifyTree(node, query) {
    const originalName = node.nameCode || "";
    const i18nName = transformI18n(node.nameCode) || "";
    const isMatched = originalName.includes(query) || i18nName.includes(query);
    if (node.children && node.children.length > 0) {
      node.children = node.children.filter(child => {
        return filterAndModifyTree(child, query);
      });
    }
    const shouldKeep = isMatched || (node.children && node.children.length > 0);
    return shouldKeep;
  }

  const filterMethod = (query: string, node) => {
    return transformI18n(node.title)!.includes(query);
  };

  onMounted(async () => {
    onSearch();
    const { data } = await getRoleMenu({ code: "", name: "" });
    treeIds.value = getKeyList(data, "id");
    treeData.value = handleTree(data);
    originalTreeData = deepCloneRef(treeData);
  });

  watch(isExpandAll, val => {
    val
      ? treeRef.value.setExpandedKeys(treeIds.value)
      : treeRef.value.setExpandedKeys([]);
  });

  watch(isSelectAll, val => {
    val
      ? treeRef.value.setCheckedKeys(treeIds.value)
      : treeRef.value.setCheckedKeys([]);
  });

  return {
    form,
    isShow,
    curRow,
    loading,
    columns,
    rowStyle,
    dataList,
    treeData,
    treeProps,
    isLinkage,
    pagination,
    isExpandAll,
    isSelectAll,
    treeSearchValue,
    // buttonClass,
    onSearch,
    resetForm,
    openDialog,
    handleMenu,
    handleSave,
    handleDelete,
    filterMethod,
    transformI18n,
    onQueryChanged,
    // handleDatabase,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
