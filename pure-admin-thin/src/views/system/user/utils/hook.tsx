import "./reset.css";
import dayjs from "dayjs";
import editForm from "../form/index.vue";
import { zxcvbn } from "@zxcvbn-ts/core";
import { message } from "@/utils/message";
import { usePublicHooks } from "../../hooks";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import type { FormItemProps } from "../utils/types";
import { getKeyList, isAllEmpty, deviceDetection } from "@pureadmin/utils";
import {
  getUserList,
  getRoleList,
  setUser,
  delUsers,
  setBalance,
  checkCashHistory
} from "@/api/system";
import {
  ElForm,
  ElInput,
  ElFormItem,
  ElProgress,
  ElMessageBox,
  ElTable,
  ElTableColumn
} from "element-plus";
import {
  type Ref,
  h,
  ref,
  toRaw,
  watch,
  computed,
  reactive,
  onMounted
} from "vue";

export function useUser(tableRef: Ref, treeRef: Ref) {
  const form = reactive({
    username: "",
    enabled: null
  });
  const formRef = ref();
  const ruleFormRef = ref();
  const dataList = ref([]);
  const loading = ref(true);
  const switchLoadMap = ref({});
  const { switchStyle } = usePublicHooks();
  const higherDeptOptions = ref();
  const treeData = ref([]);
  const treeLoading = ref(true);
  const selectedNum = ref(0);
  const setValue = ref(0);
  const recordData = ref([]);
  const pagination = reactive<PaginationProps>({
    total: 1,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const pagination1 = reactive<PaginationProps>({
    total: 1,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    {
      label: "勾选列", // 如果需要表格多选，此处label必须设置
      type: "selection",
      fixed: "left",
      reserveSelection: true // 数据刷新后保留选项
    },
    {
      label: "用户编号",
      prop: "id",
      width: 90
    },
    {
      label: "用户名称",
      prop: "username",
      minWidth: 130
    },
    {
      label: "用户昵称",
      prop: "nickName",
      minWidth: 130
    },
    {
      label: "性别",
      prop: "gender",
      minWidth: 90,
      cellRenderer: ({ row, props }) => (
        <el-tag
          size={props.size}
          type={row.gender === 1 ? "danger" : null}
          effect="plain"
        >
          {row.gender === 1 ? "女" : "男"}
        </el-tag>
      )
    },
    {
      label: "权限",
      prop: "roles",
      minWidth: 90
    },
    {
      label: "状态",
      prop: "enabled",
      minWidth: 90,
      cellRenderer: scope => (
        <el-switch
          size={scope.props.size === "small" ? "small" : "default"}
          loading={switchLoadMap.value[scope.index]?.loading}
          v-model={scope.row.enabled}
          active-value={true}
          inactive-value={false}
          active-text="已启用"
          inactive-text="已停用"
          inline-prompt
          style={switchStyle.value}
          onChange={() => onChange(scope as any)}
        />
      )
    },
    {
      label: "创建时间",
      minWidth: 90,
      prop: "createdAt",
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
  const buttonClass = computed(() => {
    return [
      "!h-[20px]",
      "reset-margin",
      "!text-gray-500",
      "dark:!text-white",
      "dark:hover:!text-primary"
    ];
  });
  // 重置的新密码
  const pwdForm = reactive({
    newPwd: "",
    amount: "",
    cashMark: ""
  });
  const pwdProgress = [
    { color: "#e74242", text: "非常弱" },
    { color: "#EFBD47", text: "弱" },
    { color: "#ffa500", text: "一般" },
    { color: "#1bbf1b", text: "强" },
    { color: "#008000", text: "非常强" }
  ];
  // 当前密码强度（0-4）
  const curScore = ref();
  const roleOptions = ref([]);

  function onChange({ row, index }) {
    ElMessageBox.confirm(
      `确认要<strong>${
        row.enabled ? "启用" : "停用"
      }</strong><strong style='color:var(--el-color-primary)'>${
        row.username
      }</strong>用户吗?`,
      "系统提示",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
        dangerouslyUseHTMLString: true,
        draggable: true
      }
    )
      .then(() => {
        switchLoadMap.value[index] = Object.assign(
          {},
          switchLoadMap.value[index],
          {
            loading: true
          }
        );
        setUser(row)
          .then(() => {
            message("已成功修改用户状态", {
              type: "success"
            });
          })
          .catch(() => {
            message("修改用户状态失败", {
              type: "error"
            });
          })
          .finally(() => {
            onSearch();
            switchLoadMap.value[index] = false;
          });
      })
      .catch(() => {
        row.enabled = !row.enabled;
      });
  }

  function handleUpdate(row) {
    console.log(row);
  }

  async function handleDelete(row) {
    if (row.username === "admin") {
      message(`admin用户无法删除`, { type: "error" });
      return;
    }
    await delUsers({ ids: [row.id] });
    message(`已删除用户编号为 ${row.id} 的数据`, { type: "success" });
    onSearch();
  }

  function handleSizeChange(val: number, val1: string) {
    pagination.pageSize = val;

    onSearch();
    console.log(`${val} items per page`);
  }
  function handleSizeChange1(val: number) {
    pagination1.pageSize = val;
    rechargeRecord1(setValue.value);
    // searchRechargeRecord(setValue.value);
  }

  function handleCurrentChange(val: number, val1: string) {
    pagination.currentPage = val - 1;

    onSearch();
    console.log(`current page: ${val}`);
  }
  function handleCurrentChange1(val: number, val1: string) {
    pagination1.currentPage = val - 1;
    rechargeRecord1(setValue.value);
  }

  /** 当CheckBox选择项发生变化时会触发该事件 */
  function handleSelectionChange(val) {
    selectedNum.value = val.length;
    // 重置表格高度
    tableRef.value.setAdaptive();
  }

  /** 取消选择 */
  function onSelectionCancel() {
    selectedNum.value = 0;
    // 用于多选表格，清空用户的选择
    tableRef.value.getTableRef().clearSelection();
  }

  /** 批量删除 */
  async function onbatchDel() {
    // 返回当前选中的行
    const curSelected = tableRef.value.getTableRef().getSelectionRows();
    // 接下来根据实际业务，通过选中行的某项数据，比如下面的id，调用接口进行批量删除
    await delUsers({ ids: getKeyList(curSelected, "id") });
    message(`已删除用户编号为 ${getKeyList(curSelected, "id")} 的数据`, {
      type: "success"
    });
    tableRef.value.getTableRef().clearSelection();
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const { data } = await getUserList({
      ...toRaw(form),
      ...toRaw(pagination)
    });
    dataList.value = data.list || [];
    pagination.total = data.total;
    pagination.pageSize = data.pageSize;
    pagination.currentPage = data.currentPage;

    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  const searchRechargeRecord = async row => {
    let params = {
      currentPage: pagination1.currentPage ? pagination1.currentPage : 0,
      pageSize: pagination1.pageSize ? pagination1.pageSize : 10,
      userId: row
    };

    return await checkCashHistory(params).then(res => {
      if (res.code === 200) {
        let tableData = res.data;
        return (
          tableData || { list: [], total: 0, pageSize: 10, currentPage: 0 }
        );
      } else {
        message(`暂无数据`, {
          type: "error"
        });
        return { list: [], total: 0, pageSize: 10, currentPage: 0 };
      }
    });
  };

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    // treeRef.value.onTreeReset();
    onSearch();
  };

  function formatHigherDeptOptions(treeList) {
    // 根据返回数据的 enabled 字段值判断追加是否禁用disabled字段，返回处理后的树结构，用于上级部门级联选择器的展示（实际开发中也是如此，不可能前端需要的每个字段后端都会返回，这时需要前端自行根据后端返回的某些字段做逻辑处理）
    if (!treeList || !treeList.length) return;
    const newTreeList = [];
    for (let i = 0; i < treeList.length; i++) {
      treeList[i].disabled = !treeList[i].enabled;
      formatHigherDeptOptions(treeList[i].children);
      newTreeList.push(treeList[i]);
    }
    return newTreeList;
  }

  function openDialog(title = "新增", row?: FormItemProps) {
    addDialog({
      title: `${title}用户`,
      props: {
        formInline: {
          title,
          id: row?.id ?? 0,
          higherDeptOptions: formatHigherDeptOptions(higherDeptOptions.value),
          nickName: row?.nickName ?? "",
          username: row?.username ?? "",
          enabled: row?.enabled ?? true,
          userType: row?.userType ?? 0,
          gender: row?.gender ?? 0,
          roles: row?.roles ?? [],
          mark: row?.mark ?? "",
          roleOptions: roleOptions
        }
      },
      width: "46%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;

        function chores() {
          setUser(curData)
            .then(() => {
              message(`您${title}了用户名称为${curData.username}的这条数据`, {
                type: "success"
              });
            })
            .catch(error => {
              const msg = error?.response?.data?.message;
              message(`${title} ${curData.username} 失败--${msg}`, {
                type: "error"
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

  watch(
    pwdForm,
    ({ newPwd }) =>
      (curScore.value = isAllEmpty(newPwd) ? -1 : zxcvbn(newPwd).score)
  );

  /** 重置密码 */
  function handleReset(row) {
    addDialog({
      title: `重置 ${row.username} 用户的密码`,
      width: "30%",
      draggable: true,
      closeOnClickModal: false,
      fullscreen: deviceDetection(),
      contentRenderer: () => (
        <>
          <ElForm ref={ruleFormRef} model={pwdForm}>
            <ElFormItem
              prop="newPwd"
              rules={[
                {
                  required: true,
                  message: "请输入新密码",
                  trigger: "blur"
                }
              ]}
            >
              <ElInput
                clearable
                show-password
                type="password"
                v-model={pwdForm.newPwd}
                placeholder="请输入新密码"
              />
            </ElFormItem>
          </ElForm>
          <div class="mt-4 flex">
            {pwdProgress.map(({ color, text }, idx) => (
              <div
                class="w-[19vw]"
                style={{ marginLeft: idx !== 0 ? "4px" : 0 }}
              >
                <ElProgress
                  striped
                  striped-flow
                  duration={curScore.value === idx ? 6 : 0}
                  percentage={curScore.value >= idx ? 100 : 0}
                  color={color}
                  stroke-width={10}
                  show-text={false}
                />
                <p
                  class="text-center"
                  style={{ color: curScore.value === idx ? color : "" }}
                >
                  {text}
                </p>
              </div>
            ))}
          </div>
        </>
      ),
      closeCallBack: () => (pwdForm.newPwd = ""),
      beforeSure: done => {
        ruleFormRef.value.validate(valid => {
          if (valid) {
            // 表单规则校验通过
            message(`已成功重置 ${row.username} 用户的密码`, {
              type: "success"
            });
            console.log(pwdForm.newPwd);
            done();
            onSearch();
          }
        });
      }
    });
  }
  /** 余额管理 */
  function balanceManagement(row) {
    addDialog({
      title: `管理${row.username} 用户的余额`,
      width: "30%",
      draggable: true,
      closeOnClickModal: false,
      fullscreen: deviceDetection(),
      contentRenderer: () => (
        <>
          <ElForm ref={ruleFormRef} model={pwdForm}>
            <ElFormItem
              prop="amount"
              rules={[
                {
                  required: true,
                  message: "请输入金额",
                  trigger: "blur"
                }
              ]}
            >
              <ElInput
                clearable
                style="width: 150px"
                v-model={pwdForm.amount}
                onChange={value => {
                  pwdForm.amount = value.replace(/[^\d]/g, "");
                }}
                placeholder="请输入金额"
              />
            </ElFormItem>
            <ElFormItem style="{margin-top: 15px; }" prop="cashMark">
              <ElInput
                type="textarea"
                style="{width: 240px; resize: 'none'}"
                rows={6}
                v-model={pwdForm.cashMark}
                placeholder="请输入备注信息"
              />
            </ElFormItem>
          </ElForm>
        </>
      ),
      closeCallBack: () => ((pwdForm.amount = ""), (pwdForm.cashMark = "")),
      beforeSure: done => {
        if (pwdForm.amount === "") return message("请输入金额");
        let curData = {
          amount: Number(pwdForm.amount),
          cashMark: pwdForm.cashMark,
          userId: row.id
        };

        setBalance(curData)
          .then(() => {
            message(`您修改了角色名称为${row.username}用户的余额`, {
              type: "success"
            });
            done();
            onSearch();
          })
          .finally(() => {
            done();
            onSearch();
          });
      }
    });
  }
  /** 充值记录 */
  async function rechargeRecord(row) {
    setValue.value = row.id; // 保存当前行数据，以便在切换分页使用
    const result = await searchRechargeRecord(row.id);
    recordData.value = result;
    pagination1.total = result.total;
    pagination1.pageSize = result.pageSize;
    pagination1.currentPage = result.currentPage;

    let dynamicColumns = [
      { prop: "startAmount", label: "起始金额" },
      { prop: "amount", label: "修改后金额" },
      { prop: "cashMark", label: "备注" }
    ];
    addDialog({
      title: `用户：${row.username} 的充值记录`,
      width: "30%",
      draggable: true,
      closeOnClickModal: false,
      hideFooter: true,
      fullscreen: deviceDetection(),
      contentRenderer: () => (
        <>
          <div style="height:504px;display:flex;flex-direction:'column'">
            <pure-table
              ref="tableRef"
              table-layout="auto"
              align-whole="center"
              show-overflow-tooltip
              row-key="id"
              style="flex:1;display:flex;flex-direction:'column';height:100% !important;"
              adaptive
              pagination={{
                ...pagination1,
                currentPage: pagination1.currentPage + 1
              }}
              data={recordData.value.list}
              columns={dynamicColumns}
              onPageSizeChange={handleSizeChange1}
              onPageCurrentChange={handleCurrentChange1}
            >
              <pure-table-column
                prop="startAmount"
                label="起始金额"
                width="180"
              />
              <pure-table-column prop="amount" label="修改后金额" width="180" />
              <pure-table-column prop="cashMark" label="备注" width="180" />
            </pure-table>
          </div>
        </>
      ),
      closeCallBack: () => (data = [])
    });
  }

  /** 充值记录 */
  async function rechargeRecord1(row) {
    const result = await searchRechargeRecord(row);
    recordData.value = result;
    pagination1.total = result.total;
    pagination1.pageSize = result.pageSize;
    pagination1.currentPage = result.currentPage;
  }
  onMounted(async () => {
    treeLoading.value = true;
    onSearch();
    roleOptions.value = (await getRoleList({})).data.list;
    // loading.value = false;
  });

  return {
    form,
    loading,
    columns,
    dataList,
    treeData,
    treeLoading,
    selectedNum,
    pagination,
    buttonClass,
    setValue,
    recordData,
    deviceDetection,
    onSearch,
    resetForm,
    onbatchDel,
    openDialog,
    handleUpdate,
    rechargeRecord,
    handleDelete,
    handleReset,
    balanceManagement,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
