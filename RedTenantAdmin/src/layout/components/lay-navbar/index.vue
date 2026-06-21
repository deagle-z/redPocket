<script setup lang="ts">
import { useNav } from "@/layout/hooks/useNav";
import LaySearch from "../lay-search/index.vue";
import LayNotice from "../lay-notice/index.vue";
import LayNavMix from "../lay-sidebar/NavMix.vue";
import { useTranslationLang } from "@/layout/hooks/useTranslationLang";
import LaySidebarFullScreen from "../lay-sidebar/components/SidebarFullScreen.vue";
import LaySidebarBreadCrumb from "../lay-sidebar/components/SidebarBreadCrumb.vue";
import LaySidebarTopCollapse from "../lay-sidebar/components/SidebarTopCollapse.vue";

import { reactive, ref } from "vue";
import QRCode from "qrcode";
import { message } from "@/utils/message";
import {
  bindTenantTwofa,
  changeTenantPassword,
  getTenantTwofaStatus,
  setupTenantTwofa,
  unbindTenantTwofa
} from "@/api/account";

import GlobalizationIcon from "@/assets/svg/globalization.svg?component";
import AccountSettingsIcon from "@iconify-icons/ri/user-settings-line";
import LogoutCircleRLine from "@iconify-icons/ri/logout-circle-r-line";
import LockPasswordLine from "@iconify-icons/ri/lock-password-line";
import ShieldKeyholeLine from "@iconify-icons/ri/shield-keyhole-line";
import Setting from "@iconify-icons/ri/settings-3-line";
import Check from "@iconify-icons/ep/check";

const {
  layout,
  device,
  logout,
  onPanel,
  pureApp,
  username,
  userAvatar,
  avatarsStyle,
  toggleSideBar,
  getDropdownItemStyle,
  getDropdownItemClass
} = useNav();

const { t, locale, translationCh, translationEn } = useTranslationLang();

// ===== 修改密码 =====
const pwdDialogVisible = ref(false);
const pwdSaving = ref(false);
const pwdForm = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: ""
});
function openPwdDialog() {
  pwdForm.oldPassword = "";
  pwdForm.newPassword = "";
  pwdForm.confirmPassword = "";
  pwdDialogVisible.value = true;
}
async function submitPwd() {
  if (!pwdForm.oldPassword) {
    message("请输入原密码", { type: "warning" });
    return;
  }
  if (pwdForm.newPassword.length < 6 || pwdForm.newPassword.length > 64) {
    message("新密码长度需 6-64 位", { type: "warning" });
    return;
  }
  if (pwdForm.newPassword !== pwdForm.confirmPassword) {
    message("两次输入的新密码不一致", { type: "warning" });
    return;
  }
  pwdSaving.value = true;
  try {
    await changeTenantPassword({
      oldPassword: pwdForm.oldPassword,
      newPassword: pwdForm.newPassword
    });
    message("密码修改成功，请重新登录", { type: "success" });
    pwdDialogVisible.value = false;
    setTimeout(() => logout(), 800);
  } catch {
    // http 拦截器已弹出错误提示
  } finally {
    pwdSaving.value = false;
  }
}

// ===== 绑定 Google 验证码 =====
const twofaDialogVisible = ref(false);
const twofaLoading = ref(false);
const twofaSaving = ref(false);
const twofaBound = ref(false);
const twofaSecret = ref("");
const twofaOtpauthUrl = ref("");
const twofaQrDataUrl = ref("");
const twofaCode = ref("");
async function openTwofaDialog() {
  twofaDialogVisible.value = true;
  twofaLoading.value = true;
  twofaSecret.value = "";
  twofaOtpauthUrl.value = "";
  twofaQrDataUrl.value = "";
  twofaCode.value = "";
  try {
    const statusRes = await getTenantTwofaStatus();
    twofaBound.value = statusRes.data?.bound === true;
    if (!twofaBound.value) {
      const setupRes = await setupTenantTwofa();
      twofaSecret.value = setupRes.data?.secret || "";
      twofaOtpauthUrl.value = setupRes.data?.otpauthUrl || "";
      if (twofaOtpauthUrl.value) {
        try {
          twofaQrDataUrl.value = await QRCode.toDataURL(twofaOtpauthUrl.value, {
            width: 200,
            margin: 1
          });
        } catch {
          twofaQrDataUrl.value = "";
        }
      }
    }
  } catch {
    // ignore
  } finally {
    twofaLoading.value = false;
  }
}
async function copyTwofaSecret() {
  if (!twofaSecret.value) return;
  try {
    await navigator.clipboard.writeText(twofaSecret.value);
    message("密钥已复制", { type: "success" });
  } catch {
    message("复制失败，请手动复制", { type: "warning" });
  }
}
async function submitTwofaBind() {
  const code = twofaCode.value.trim();
  if (!/^\d{6}$/.test(code)) {
    message("请输入 6 位动态码", { type: "warning" });
    return;
  }
  twofaSaving.value = true;
  try {
    await bindTenantTwofa({ secret: twofaSecret.value, code });
    message("绑定成功，下次登录需输入验证码", { type: "success" });
    twofaBound.value = true;
    twofaDialogVisible.value = false;
  } catch {
    // ignore
  } finally {
    twofaSaving.value = false;
  }
}
async function submitTwofaUnbind() {
  const code = twofaCode.value.trim();
  if (!/^\d{6}$/.test(code)) {
    message("请输入当前 6 位动态码", { type: "warning" });
    return;
  }
  twofaSaving.value = true;
  try {
    await unbindTenantTwofa({ code });
    message("已解绑", { type: "success" });
    twofaBound.value = false;
    twofaDialogVisible.value = false;
  } catch {
    // ignore
  } finally {
    twofaSaving.value = false;
  }
}
</script>

<template>
  <div class="navbar bg-[#fff] shadow-sm shadow-[rgba(0,21,41,0.08)]">
    <LaySidebarTopCollapse
      v-if="device === 'mobile'"
      class="hamburger-container"
      :is-active="pureApp.sidebar.opened"
      @toggleClick="toggleSideBar"
    />

    <LaySidebarBreadCrumb
      v-if="layout !== 'mix' && device !== 'mobile'"
      class="breadcrumb-container"
    />

    <LayNavMix v-if="layout === 'mix'" />

    <div v-if="layout === 'vertical'" class="vertical-header-right">
      <!-- 菜单搜索 -->
      <LaySearch id="header-search" />
      <!-- 国际化 -->
      <el-dropdown id="header-translation" trigger="click">
        <GlobalizationIcon
          class="navbar-bg-hover w-[40px] h-[48px] p-[11px] cursor-pointer outline-none"
        />
        <template #dropdown>
          <el-dropdown-menu class="translation">
            <el-dropdown-item
              :style="getDropdownItemStyle(locale, 'zh')"
              :class="['dark:!text-white', getDropdownItemClass(locale, 'zh')]"
              @click="translationCh"
            >
              <IconifyIconOffline
                v-show="locale === 'zh'"
                class="check-zh"
                :icon="Check"
              />
              简体中文
            </el-dropdown-item>
            <el-dropdown-item
              :style="getDropdownItemStyle(locale, 'en')"
              :class="['dark:!text-white', getDropdownItemClass(locale, 'en')]"
              @click="translationEn"
            >
              <span v-show="locale === 'en'" class="check-en">
                <IconifyIconOffline :icon="Check" />
              </span>
              English
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <!-- 全屏 -->
      <LaySidebarFullScreen id="full-screen" />
      <!-- 消息通知 -->
      <LayNotice id="header-notice" />
      <!-- 退出登录 -->
      <el-dropdown trigger="click">
        <span class="el-dropdown-link navbar-bg-hover select-none">
          <img :src="userAvatar" :style="avatarsStyle" />
          <p v-if="username" class="dark:text-white">{{ username }}</p>
        </span>
        <template #dropdown>
          <el-dropdown-menu class="logout">
            <el-dropdown-item @click="openPwdDialog">
              <IconifyIconOffline :icon="LockPasswordLine" style="margin: 5px" />
              修改密码
            </el-dropdown-item>
            <el-dropdown-item @click="openTwofaDialog">
              <IconifyIconOffline
                :icon="ShieldKeyholeLine"
                style="margin: 5px"
              />
              绑定Google验证码
            </el-dropdown-item>
            <el-dropdown-item divided @click="logout">
              <IconifyIconOffline
                :icon="LogoutCircleRLine"
                style="margin: 5px"
              />
              {{ t("buttons.pureLoginOut") }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <span
        class="set-icon navbar-bg-hover"
        :title="t('buttons.pureOpenSystemSet')"
        @click="onPanel"
      >
        <IconifyIconOffline :icon="Setting" />
      </span>
    </div>

    <!-- 修改密码 -->
    <el-dialog
      v-model="pwdDialogVisible"
      title="修改密码"
      width="420px"
      append-to-body
      destroy-on-close
    >
      <el-form :model="pwdForm" label-width="90px">
        <el-form-item label="原密码">
          <el-input
            v-model="pwdForm.oldPassword"
            type="password"
            show-password
            placeholder="请输入原密码"
          />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input
            v-model="pwdForm.newPassword"
            type="password"
            show-password
            placeholder="6-64 位"
          />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input
            v-model="pwdForm.confirmPassword"
            type="password"
            show-password
            placeholder="再次输入新密码"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="pwdSaving" @click="submitPwd">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 绑定 Google 验证码 -->
    <el-dialog
      v-model="twofaDialogVisible"
      title="Google 验证码"
      width="460px"
      append-to-body
      destroy-on-close
    >
      <div v-loading="twofaLoading">
        <template v-if="twofaBound">
          <p class="twofa-tip">
            当前账号已绑定 Google 验证码。解绑请输入当前 6 位动态码。
          </p>
        </template>
        <template v-else>
          <p class="twofa-tip">
            用 Google Authenticator 扫描下方二维码（或「手动输入密钥」添加），再输入
            6 位动态码完成绑定（绑定后下次登录需输入验证码）。
          </p>
          <div v-if="twofaQrDataUrl" class="twofa-qr">
            <img :src="twofaQrDataUrl" alt="Google Authenticator QR" />
          </div>
          <div class="twofa-secret">
            <span class="twofa-secret__key">{{ twofaSecret || "-" }}</span>
            <el-button link type="primary" @click="copyTwofaSecret">
              复制
            </el-button>
          </div>
        </template>
        <el-input
          v-model="twofaCode"
          maxlength="6"
          placeholder="请输入 6 位动态码"
          class="mt-3"
        />
      </div>
      <template #footer>
        <el-button @click="twofaDialogVisible = false">关闭</el-button>
        <el-button
          v-if="twofaBound"
          type="danger"
          :loading="twofaSaving"
          @click="submitTwofaUnbind"
        >
          解绑
        </el-button>
        <el-button
          v-else
          type="primary"
          :loading="twofaSaving"
          @click="submitTwofaBind"
        >
          绑定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.navbar {
  width: 100%;
  height: 48px;
  overflow: hidden;

  .hamburger-container {
    float: left;
    height: 100%;
    line-height: 48px;
    cursor: pointer;
  }

  .vertical-header-right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    min-width: 280px;
    height: 48px;
    color: #000000d9;

    .el-dropdown-link {
      display: flex;
      align-items: center;
      justify-content: space-around;
      height: 48px;
      padding: 10px;
      color: #000000d9;
      cursor: pointer;

      p {
        font-size: 14px;
      }

      img {
        width: 22px;
        height: 22px;
        border-radius: 50%;
      }
    }
  }

  .breadcrumb-container {
    float: left;
    margin-left: 16px;
  }
}

.translation {
  ::v-deep(.el-dropdown-menu__item) {
    padding: 5px 40px;
  }

  .check-zh {
    position: absolute;
    left: 20px;
  }

  .check-en {
    position: absolute;
    left: 20px;
  }
}

.logout {
  width: 180px;

  ::v-deep(.el-dropdown-menu__item) {
    display: inline-flex;
    flex-wrap: wrap;
    min-width: 100%;
  }
}

.twofa-tip {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--el-text-color-regular);
}

.twofa-qr {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;

  img {
    width: 200px;
    height: 200px;
    border: 1px solid var(--el-border-color);
    border-radius: 6px;
  }
}

.twofa-secret {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 6px;

  &__key {
    font-family: monospace;
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 1px;
    word-break: break-all;
    user-select: all;
  }
}
</style>
