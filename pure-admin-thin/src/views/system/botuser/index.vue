<script setup lang="ts">
import { reactive, ref } from "vue";
import { deviceDetection } from "@pureadmin/utils";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import {
  batchCreateBotUsers,
  batchUpdateBotUsers,
  type TgUser
} from "@/api/tgUser";
import { useBotUser } from "./utils/hook";
import { getToken, formatToken } from "@/utils/auth";

import Delete from "@iconify-icons/ep/delete";
import Edit from "@iconify-icons/ep/edit";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";

defineOptions({
  name: "SysBotUser"
});

type AvatarUploadState = {
  avatarUrlList: string[];
  avatarUploading: boolean;
  uploadingCount: number;
};

const formRef = ref();
const tableRef = ref();
const submitLoading = ref(false);
const fileInputLoading = ref(false);
const createDialogVisible = ref(false);
const editDialogVisible = ref(false);
const uploadUrl = `${import.meta.env.VITE_BASE_URL}/api/v1/admin/upload`;

const batchCreateForm = reactive({
  num: 10,
  randomName: false,
  nameFile: "",
  avatarText: ""
});

const batchEditForm = reactive({
  randomName: false,
  nameFile: "",
  avatarText: "",
  statusEnabled: false,
  status: 1
});

const createAvatarState = reactive<AvatarUploadState>({
  avatarUrlList: [],
  avatarUploading: false,
  uploadingCount: 0
});

const editAvatarState = reactive<AvatarUploadState>({
  avatarUrlList: [],
  avatarUploading: false,
  uploadingCount: 0
});

const {
  form,
  loading,
  columns,
  dataList,
  selectedNum,
  selectedRows,
  pagination,
  statusOptions,
  onSearch,
  resetForm,
  onBatchDel,
  onSelectionCancel,
  handleSizeChange,
  handleCurrentChange,
  handleSelectionChange,
  updateStatus,
  handleDelete
} = useBotUser(tableRef);

function parseLines(text: string) {
  return text
    .split(/\r?\n/)
    .map(item => item.trim())
    .filter(Boolean);
}

async function handleNameFileChange(
  uploadFile: any,
  targetForm: { nameFile: string }
) {
  const raw = uploadFile?.raw as File | undefined;
  if (!raw) return;
  fileInputLoading.value = true;
  try {
    targetForm.nameFile = await raw.text();
    message("名称文件已读取", { type: "success" });
  } catch (error) {
    console.error("读取名称文件失败", error);
    message("读取名称文件失败", { type: "error" });
  } finally {
    fileInputLoading.value = false;
  }
}

function syncAvatarText(
  targetForm: { avatarText: string },
  uploadState: AvatarUploadState
) {
  targetForm.avatarText = uploadState.avatarUrlList.join("\n");
}

function resetAvatarState(uploadState: AvatarUploadState) {
  uploadState.avatarUrlList = [];
  uploadState.uploadingCount = 0;
  uploadState.avatarUploading = false;
}

function resetBatchCreateForm() {
  batchCreateForm.num = 10;
  batchCreateForm.randomName = false;
  batchCreateForm.nameFile = "";
  batchCreateForm.avatarText = "";
  resetAvatarState(createAvatarState);
}

function resetBatchEditForm() {
  batchEditForm.randomName = false;
  batchEditForm.nameFile = "";
  batchEditForm.avatarText = "";
  batchEditForm.statusEnabled = false;
  batchEditForm.status = 1;
  resetAvatarState(editAvatarState);
}

function openBatchCreateDialog() {
  resetBatchCreateForm();
  createDialogVisible.value = true;
}

function openBatchEditDialog() {
  if (selectedRows.value.length === 0) {
    message("请先勾选要修改的机器人", { type: "warning" });
    return;
  }
  resetBatchEditForm();
  editDialogVisible.value = true;
}

function getUploadHeaders() {
  const token = getToken()?.accessToken;
  return token ? { Authorization: formatToken(token) } : {};
}

function handleBeforeAvatarUpload() {
  const token = getToken()?.accessToken;
  if (!token) {
    message("请先登录后再上传", { type: "error" });
    return false;
  }
  return true;
}

function ensureImageFile(file: File, uploadState: AvatarUploadState) {
  if (!file.type || !file.type.startsWith("image/")) {
    message(`仅支持图片格式：${file.name}`, { type: "error" });
    return false;
  }
  uploadState.uploadingCount += 1;
  uploadState.avatarUploading = uploadState.uploadingCount > 0;
  return true;
}

function finishAvatarUpload(uploadState: AvatarUploadState) {
  uploadState.uploadingCount = Math.max(0, uploadState.uploadingCount - 1);
  uploadState.avatarUploading = uploadState.uploadingCount > 0;
}

function handleAvatarUploadSuccess(
  response: any,
  targetForm: { avatarText: string },
  uploadState: AvatarUploadState
) {
  const url = response?.data?.url || response?.url;
  if (!url) {
    finishAvatarUpload(uploadState);
    message("上传失败，未返回URL", { type: "error" });
    return;
  }
  if (!uploadState.avatarUrlList.includes(url)) {
    uploadState.avatarUrlList.push(url);
    syncAvatarText(targetForm, uploadState);
  }
  finishAvatarUpload(uploadState);
  message("图片上传成功", { type: "success" });
}

function handleAvatarUploadError(uploadState: AvatarUploadState) {
  finishAvatarUpload(uploadState);
  message("图片上传失败", { type: "error" });
}

function handleAvatarUploadChange() {
  fileInputLoading.value = false;
}

function handleAvatarUploadExceed() {
  message("单次最多选择 100 张图片", { type: "warning" });
}

function clearAvatarList(
  targetForm: { avatarText: string },
  uploadState: AvatarUploadState
) {
  uploadState.avatarUrlList = [];
  syncAvatarText(targetForm, uploadState);
}

async function submitBatchCreate() {
  if (!Number.isInteger(batchCreateForm.num) || batchCreateForm.num <= 0) {
    message("生成数量必须大于 0", { type: "warning" });
    return;
  }
  if (
    !batchCreateForm.randomName &&
    parseLines(batchCreateForm.nameFile).length === 0
  ) {
    message("请上传或填写名称文件内容", { type: "warning" });
    return;
  }

  submitLoading.value = true;
  try {
    const { data } = await batchCreateBotUsers({
      num: batchCreateForm.num,
      randomName: batchCreateForm.randomName,
      nameFile: batchCreateForm.nameFile,
      avatarLinks: parseLines(batchCreateForm.avatarText)
    });
    message(`已生成 ${data?.count || 0} 个机器人`, { type: "success" });
    createDialogVisible.value = false;
    onSearch();
  } catch (error) {
    console.error("批量生成机器人失败", error);
    message("批量生成机器人失败", { type: "error" });
  } finally {
    submitLoading.value = false;
  }
}

async function submitBatchUpdate() {
  const ids = selectedRows.value.map((item: TgUser) => Number(item.id));
  if (ids.length === 0) {
    message("请先勾选要修改的机器人", { type: "warning" });
    return;
  }

  const nameList = parseLines(batchEditForm.nameFile);
  const avatarLinks = parseLines(batchEditForm.avatarText);
  const shouldUpdateName = batchEditForm.randomName || nameList.length > 0;
  const shouldUpdateAvatar = avatarLinks.length > 0;
  const shouldUpdateStatus = batchEditForm.statusEnabled;
  if (!shouldUpdateName && !shouldUpdateAvatar && !shouldUpdateStatus) {
    message("请至少提供一项修改内容", { type: "warning" });
    return;
  }

  submitLoading.value = true;
  try {
    const payload: {
      ids: number[];
      randomName: boolean;
      nameFile: string;
      avatarLinks: string[];
      status?: number;
    } = {
      ids,
      randomName: batchEditForm.randomName,
      nameFile: batchEditForm.nameFile,
      avatarLinks
    };
    if (shouldUpdateStatus) {
      payload.status = batchEditForm.status;
    }
    const { data } = await batchUpdateBotUsers(payload);
    message(`已批量修改 ${data?.count || 0} 个机器人`, {
      type: "success"
    });
    editDialogVisible.value = false;
    onSelectionCancel();
    onSearch();
  } catch (error) {
    console.error("批量修改机器人失败", error);
    message("批量修改机器人失败", { type: "error" });
  } finally {
    submitLoading.value = false;
  }
}
</script>

<template>
  <div :class="['flex', 'justify-between', deviceDetection() && 'flex-wrap']">
    <div :class="[deviceDetection() ? ['w-full', 'mt-2'] : 'w-[calc(100%)]']">
      <el-form
        ref="formRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
      >
        <el-form-item label="用户名：" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="昵称：" prop="firstName">
          <el-input
            v-model="form.firstName"
            placeholder="请输入昵称"
            clearable
            class="!w-[180px]"
          />
        </el-form-item>
        <el-form-item label="状态：" prop="status">
          <el-select
            v-model="form.status"
            placeholder="请选择状态"
            clearable
            class="!w-[180px]"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri:search-line')"
            :loading="loading"
            @click="onSearch"
          >
            搜索
          </el-button>
          <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar title="机器人列表" :columns="columns" @refresh="onSearch">
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="openBatchCreateDialog"
          >
            批量添加机器人
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            v-motion-fade
            class="bg-[var(--el-fill-color-light)] w-full h-[46px] mb-2 pl-4 flex items-center"
          >
            <div class="flex-auto">
              <span
                style="font-size: var(--el-font-size-base)"
                class="text-[rgba(42,46,54,0.5)] dark:text-[rgba(220,220,242,0.5)]"
              >
                已选 {{ selectedNum }} 项
              </span>
              <el-button type="primary" text @click="onSelectionCancel">
                取消选择
              </el-button>
            </div>
            <el-button
              type="primary"
              text
              class="mr-1"
              @click="openBatchEditDialog"
            >
              批量修改
            </el-button>
            <el-popconfirm
              title="是否确认批量删除机器人?"
              @confirm="onBatchDel"
            >
              <template #reference>
                <el-button type="danger" text class="mr-1">
                  批量删除
                </el-button>
              </template>
            </el-popconfirm>
          </div>
          <pure-table
            ref="tableRef"
            border
            align-whole="center"
            showOverflowTooltip
            row-key="id"
            table-layout="auto"
            :loading="loading"
            :size="size"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="{
              ...pagination,
              size,
              currentPage: pagination.currentPage + 1
            }"
            :paginationSmall="size === 'small'"
            :header-cell-style="{
              background: 'var(--el-table-row-hover-bg-color)',
              color: 'var(--el-text-color-primary)'
            }"
            @page-size-change="handleSizeChange"
            @page-current-change="handleCurrentChange"
            @selection-change="handleSelectionChange"
          >
            <template #operation="{ row }">
              <el-button
                v-if="row.status !== 1"
                class="reset-margin"
                link
                type="primary"
                :size="size"
                @click="updateStatus(row, 1)"
              >
                启用
              </el-button>
              <el-button
                v-else
                class="reset-margin"
                link
                type="danger"
                :size="size"
                @click="updateStatus(row, 0)"
              >
                禁用
              </el-button>
              <el-popconfirm
                :title="`是否确认删除机器人编号为${row.id}的这条数据`"
                @confirm="handleDelete(row)"
              >
                <template #reference>
                  <el-button
                    class="reset-margin"
                    link
                    type="danger"
                    :size="size"
                    :icon="useRenderIcon(Delete)"
                  >
                    删除
                  </el-button>
                </template>
              </el-popconfirm>
            </template>
          </pure-table>
        </template>
      </PureTableBar>

      <el-dialog
        v-model="createDialogVisible"
        title="批量添加机器人"
        width="900px"
        destroy-on-close
      >
        <el-form :inline="true" :model="batchCreateForm" class="search-form">
          <el-form-item label="生成数量">
            <el-input-number
              v-model="batchCreateForm.num"
              :min="1"
              :max="1000"
            />
          </el-form-item>
          <el-form-item label="随机名称">
            <el-switch v-model="batchCreateForm.randomName" />
          </el-form-item>
        </el-form>

        <div class="grid-wrap">
          <el-card shadow="never">
            <template #header>名称文件</template>
            <div class="mb-3 flex items-center gap-3">
              <el-upload
                :auto-upload="false"
                :show-file-list="false"
                accept=".txt"
                :on-change="file => handleNameFileChange(file, batchCreateForm)"
              >
                <el-button :loading="fileInputLoading" type="primary" plain>
                  上传 txt
                </el-button>
              </el-upload>
              <span class="helper-text">每行一个名字，数量不足会从头循环</span>
            </div>
            <el-input
              v-model="batchCreateForm.nameFile"
              type="textarea"
              :rows="10"
              placeholder="可直接粘贴名字列表，一行一个"
            />
          </el-card>

          <el-card shadow="never">
            <template #header>头像上传</template>
            <div class="helper-text mb-3">
              支持多图上传，也支持选择整个文件夹。上传成功后会自动作为头像列表参与循环使用。
            </div>
            <div class="mb-3 flex items-center gap-3">
              <el-upload
                :action="uploadUrl"
                :headers="getUploadHeaders()"
                :before-upload="
                  file =>
                    handleBeforeAvatarUpload() &&
                    ensureImageFile(file, createAvatarState)
                "
                :show-file-list="false"
                :multiple="true"
                accept="image/*"
                :on-success="
                  response =>
                    handleAvatarUploadSuccess(
                      response,
                      batchCreateForm,
                      createAvatarState
                    )
                "
                :on-error="() => handleAvatarUploadError(createAvatarState)"
                :on-change="handleAvatarUploadChange"
                :on-exceed="handleAvatarUploadExceed"
                :limit="100"
                directory
                webkitdirectory
              >
                <el-button
                  type="primary"
                  plain
                  :loading="createAvatarState.avatarUploading"
                >
                  上传图片/文件夹
                </el-button>
              </el-upload>
              <el-button
                @click="clearAvatarList(batchCreateForm, createAvatarState)"
                >清空头像</el-button
              >
            </div>
            <el-scrollbar max-height="260px">
              <div
                v-if="createAvatarState.avatarUrlList.length"
                class="avatar-grid"
              >
                <div
                  v-for="url in createAvatarState.avatarUrlList"
                  :key="url"
                  class="avatar-item"
                >
                  <el-image
                    :src="url"
                    fit="cover"
                    class="avatar-preview"
                    :preview-src-list="createAvatarState.avatarUrlList"
                    preview-teleported
                  />
                </div>
              </div>
              <el-empty v-else description="暂无已上传头像" :image-size="96" />
            </el-scrollbar>
          </el-card>
        </div>

        <template #footer>
          <div class="flex justify-end gap-3">
            <el-button @click="createDialogVisible = false">取消</el-button>
            <el-button @click="resetBatchCreateForm">重置</el-button>
            <el-button
              type="primary"
              :loading="submitLoading"
              @click="submitBatchCreate"
            >
              立即生成
            </el-button>
          </div>
        </template>
      </el-dialog>

      <el-dialog
        v-model="editDialogVisible"
        title="批量修改机器人"
        width="900px"
        destroy-on-close
      >
        <el-form :inline="true" :model="batchEditForm" class="search-form">
          <el-form-item label="已选数量">
            <el-tag type="primary">{{ selectedNum }} 个机器人</el-tag>
          </el-form-item>
          <el-form-item label="随机名称">
            <el-switch v-model="batchEditForm.randomName" />
          </el-form-item>
          <el-form-item label="批量状态">
            <div class="flex items-center gap-3">
              <el-switch v-model="batchEditForm.statusEnabled" />
              <el-select
                v-model="batchEditForm.status"
                :disabled="!batchEditForm.statusEnabled"
                class="!w-[160px]"
              >
                <el-option
                  v-for="item in statusOptions.filter(
                    item => item.value !== -1
                  )"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </div>
          </el-form-item>
        </el-form>

        <div class="grid-wrap">
          <el-card shadow="never">
            <template #header>批量昵称/用户名</template>
            <div class="mb-3 flex items-center gap-3">
              <el-upload
                :auto-upload="false"
                :show-file-list="false"
                accept=".txt"
                :on-change="file => handleNameFileChange(file, batchEditForm)"
              >
                <el-button :loading="fileInputLoading" type="primary" plain>
                  上传 txt
                </el-button>
              </el-upload>
              <span class="helper-text">
                每行一个名字，按勾选顺序循环覆盖；开启随机名称后会忽略这里的内容
              </span>
            </div>
            <el-input
              v-model="batchEditForm.nameFile"
              type="textarea"
              :rows="10"
              placeholder="留空则不修改昵称和用户名"
            />
          </el-card>

          <el-card shadow="never">
            <template #header>批量头像</template>
            <div class="helper-text mb-3">
              上传后的头像会按勾选顺序循环使用；不上传则保持原头像不变。
            </div>
            <div class="mb-3 flex items-center gap-3">
              <el-upload
                :action="uploadUrl"
                :headers="getUploadHeaders()"
                :before-upload="
                  file =>
                    handleBeforeAvatarUpload() &&
                    ensureImageFile(file, editAvatarState)
                "
                :show-file-list="false"
                :multiple="true"
                accept="image/*"
                :on-success="
                  response =>
                    handleAvatarUploadSuccess(
                      response,
                      batchEditForm,
                      editAvatarState
                    )
                "
                :on-error="() => handleAvatarUploadError(editAvatarState)"
                :on-change="handleAvatarUploadChange"
                :on-exceed="handleAvatarUploadExceed"
                :limit="100"
                directory
                webkitdirectory
              >
                <el-button
                  type="primary"
                  plain
                  :loading="editAvatarState.avatarUploading"
                >
                  上传图片/文件夹
                </el-button>
              </el-upload>
              <el-button
                @click="clearAvatarList(batchEditForm, editAvatarState)"
                >清空头像</el-button
              >
            </div>
            <el-scrollbar max-height="260px">
              <div
                v-if="editAvatarState.avatarUrlList.length"
                class="avatar-grid"
              >
                <div
                  v-for="url in editAvatarState.avatarUrlList"
                  :key="url"
                  class="avatar-item"
                >
                  <el-image
                    :src="url"
                    fit="cover"
                    class="avatar-preview"
                    :preview-src-list="editAvatarState.avatarUrlList"
                    preview-teleported
                  />
                </div>
              </div>
              <el-empty v-else description="暂无待替换头像" :image-size="96" />
            </el-scrollbar>
          </el-card>
        </div>

        <template #footer>
          <div class="flex justify-end gap-3">
            <el-button @click="editDialogVisible = false">取消</el-button>
            <el-button @click="resetBatchEditForm">重置</el-button>
            <el-button
              type="primary"
              :icon="useRenderIcon(Edit)"
              :loading="submitLoading"
              @click="submitBatchUpdate"
            >
              立即修改
            </el-button>
          </div>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<style scoped lang="scss">
.card-title {
  font-size: 15px;
  font-weight: 600;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}

.grid-wrap {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.helper-text {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.avatar-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(88px, 1fr));
  gap: 12px;
}

.avatar-item {
  display: flex;
  justify-content: center;
}

.avatar-preview {
  width: 88px;
  height: 88px;
  border-radius: 10px;
}

@media (max-width: 960px) {
  .grid-wrap {
    grid-template-columns: 1fr;
  }
}
</style>
