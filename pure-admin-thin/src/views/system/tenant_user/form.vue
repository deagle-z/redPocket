<script setup lang="ts">
import { ref } from "vue";
import ReCol from "@/components/ReCol";
import { formRules } from "./utils/rule";
import { FormProps } from "./utils/types";
import { roleOptions, statusOptions } from "./utils/enums";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    title: "",
    id: 0,
    tenantId: 0,
    username: "",
    passwordHash: "",
    passwordAlgo: "bcrypt",
    email: "",
    mobile: "",
    roleCode: "member",
    isOwner: false,
    status: 1,
    require2fa: false,
    twofaSecret: "",
    remark: ""
  })
});

const tenantUserFormRef = ref();
const newFormInline = ref(props.formInline);

function getRef() {
  return tenantUserFormRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="tenantUserFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="110px"
  >
    <el-row :gutter="20">
      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="租户ID" prop="tenantId">
          <el-input-number
            v-model="newFormInline.tenantId"
            class="!w-full"
            :min="1"
            controls-position="right"
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="账号" prop="username">
          <el-input
            v-model="newFormInline.username"
            placeholder="登录账号"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="密码哈希" prop="passwordHash">
          <el-input
            v-model="newFormInline.passwordHash"
            placeholder="bcrypt/argon2 哈希"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="密码算法" prop="passwordAlgo">
          <el-input
            v-model="newFormInline.passwordAlgo"
            placeholder="bcrypt/argon2"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="newFormInline.email" clearable />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="newFormInline.mobile" clearable />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="角色" prop="roleCode">
          <el-select v-model="newFormInline.roleCode" class="w-full">
            <el-option
              v-for="item in roleOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="状态" prop="status">
          <el-select v-model="newFormInline.status" class="w-full">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="租户Owner">
          <el-switch v-model="newFormInline.isOwner" />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="强制2FA">
          <el-switch v-model="newFormInline.require2fa" />
        </el-form-item>
      </re-col>

      <re-col :value="12" :xs="24" :sm="24">
        <el-form-item label="2FA密钥" prop="twofaSecret">
          <el-input
            v-model="newFormInline.twofaSecret"
            placeholder="可选"
            clearable
          />
        </el-form-item>
      </re-col>

      <re-col :value="24" :xs="24" :sm="24">
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="newFormInline.remark"
            placeholder="备注"
            clearable
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>
