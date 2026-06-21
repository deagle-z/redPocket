import { http } from "@/utils/http";

type Result<T = unknown> = {
  code: number;
  message: string;
  success: boolean;
  data: T;
};

/** 修改当前账号密码（校验旧密码） */
export const changeOwnPass = (data: {
  oldPassword: string;
  newPassword: string;
}) => {
  return http.request<Result>("put", "/api/v1/outside/user/changeOwnPass", {
    data
  });
};

/** 谷歌验证绑定状态 */
export const getTwofaStatus = () => {
  return http.request<Result<{ bound: boolean }>>(
    "get",
    "/api/v1/outside/user/twofa/status"
  );
};

/** 生成待绑定的谷歌验证密钥 */
export const setupTwofa = () => {
  return http.request<Result<{ secret: string; otpauthUrl: string }>>(
    "get",
    "/api/v1/outside/user/twofa/setup"
  );
};

/** 校验动态码并绑定谷歌验证 */
export const bindTwofa = (data: { secret: string; code: string }) => {
  return http.request<Result>("post", "/api/v1/outside/user/twofa/bind", {
    data
  });
};

/** 解绑谷歌验证 */
export const unbindTwofa = () => {
  return http.request<Result>("get", "/api/v1/outside/user/unbind/gauth");
};
