import { http } from "@/utils/http";

export type UserResult = {
  code: number;
  message: string;
  data: {
    /** 用户ID */
    userId: number;
    /** 租户ID */
    tenantId: number;
    /** 用户名 */
    username: string;
    /** 当前登录用户角色 */
    roleCode: string;
    /** 当前登录用户角色数组（前端权限兼容） */
    roles: Array<string>;
    /** `token` */
    accessToken: string;
    /** 兼容前端token结构 */
    refreshToken: string;
    /** `accessToken`过期时间 */
    expires: Date;
  };
};

export type RefreshTokenResult = {
  success: boolean;
  data: {
    /** `token` */
    accessToken: string;
    /** 用于调用刷新`accessToken`的接口时所需的`token` */
    refreshToken: string;
    /** `accessToken`的过期时间（格式'xxxx/xx/xx xx:xx:xx'） */
    expires: Date;
  };
};

/** 登录 */
export const getLogin = (data?: object) => {
  return http
    .request<UserResult>("post", "/api/v1/tenant/login", { data })
    .then(res => {
      const roleCode = res?.data?.roleCode || "";
      const accessToken = res?.data?.accessToken || "";
      return {
        ...res,
        data: {
          ...res.data,
          roles: roleCode ? [roleCode] : [],
          refreshToken: accessToken,
          expires: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
        }
      };
    });
};

/** 刷新`token` */
export const refreshTokenApi = (data?: object) => {
  return http.request<RefreshTokenResult>("post", "/refresh-token", { data });
};
