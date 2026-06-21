import { http } from "@/utils/http";

type Result<T = unknown> = {
  code: number;
  message: string;
  success: boolean;
  data: T;
};

/** 修改当前账号密码 */
export const changeTenantPassword = (data: {
  oldPassword: string;
  newPassword: string;
}) => {
  return http.request<Result>(
    "post",
    "/api/v1/tenant/account/changePassword",
    { data }
  );
};

/** Google 验证码绑定状态 */
export const getTenantTwofaStatus = () => {
  return http.request<Result<{ bound: boolean }>>(
    "get",
    "/api/v1/tenant/account/twofa/status"
  );
};

/** 生成待绑定的 Google 验证码密钥 */
export const setupTenantTwofa = () => {
  return http.request<Result<{ secret: string; otpauthUrl: string }>>(
    "get",
    "/api/v1/tenant/account/twofa/setup"
  );
};

/** 校验动态码并绑定 Google 验证码 */
export const bindTenantTwofa = (data: { secret: string; code: string }) => {
  return http.request<Result>("post", "/api/v1/tenant/account/twofa/bind", {
    data
  });
};

/** 校验动态码并解绑 Google 验证码 */
export const unbindTenantTwofa = (data: { code: string }) => {
  return http.request<Result>("post", "/api/v1/tenant/account/twofa/unbind", {
    data
  });
};
