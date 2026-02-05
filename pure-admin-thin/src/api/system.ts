/*
 * @Author: yujiang yujiang@qq.com
 * @Date: 2025-08-20 14:31:26
 * @LastEditors: yujiang yujiang@gmail.com
 * @LastEditTime: 2025-10-11 12:36:18
 * @FilePath: \pure-admin-thin\src\api\system.ts
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
import { http } from "@/utils/http";

export type Result = {
  code: number;
  message: string;
  success: boolean;
  data?: Array<any>;
};

export type PageResult = {
  code: number;
  message: string;
  success: boolean;
  data?: {
    /** 列表数据 */
    list: Array<any>;
    /** 总条目数 */e
    total?: number;
    /** 每页显示条目个数 */
    pageSize?: number;
    /** 当前页数 */
    currentPage?: number;
  };
};

/** 获取系统管理-用户管理列表 */
export const getUserList = (data?: object) => {
  return http.request<PageResult>("post", "/api/v1/outside/user", { data });
};

/** 获取系统管理-设置用户 */
export const setUser = (data?: object) => {
  return http.request<Result>("post", "/api/v1/manager/setUser", { data });
};

/** 获取系统管理-设置用户 */
export const delUsers = (data?: object) => {
  return http.request<Result>("post", "/api/v1/manager/delUsers", { data });
};


/** 获取系统管理-角色列表 */
export const getRoleList = (data?: object) => {
  return http.request<PageResult>("post", "/api/v1/outside/roles", { data });
};

/** 获取系统管理-菜单管理列表 */
export const getMenuList = (data?: object) => {
  return http.request<Result>("post", "/api/v1/outside/menus", { data });
};

/** 获取系统管理-菜单更新 */
export const setMenuList = (data?: object) => {
  return http.request<Result>("put", "/api/v1/manager/setMenus", { data });
};

/** 获取角色管理-删除菜单 */
export const delMenu = (menuId?: number) => {
  return http.request<Result>("delete", "/api/v1/manager/menu/" + menuId);
};

/** 获取角色管理-权限-菜单权限 */
export const getRoleMenu = (data?: object) => {
  return http.request<Result>("post", "/api/v1/outside/role-menu", { data });
};

/** 获取角色管理-权限-菜单权限-根据角色 id 查对应菜单 */
export const getRoleMenuIds = (roleId?: number) => {
  return http.request<Result>("get", "/api/v1/outside/role-menu-ids/" + roleId);
};

/** 获取角色管理-删除权限 */
export const delRole = (roleId?: number) => {
  return http.request<Result>("delete", "/api/v1/manager/role/" + roleId);
};

/** 获取角色管理-设置权限 */
export const setRole = (data?: object) => {
  return http.request<Result>("post", "/api/v1/manager/setRole", { data });
};


/** 获取角色管理-用户余额管理 */
export const setBalance = (data?: object) => {
  return http.request<Result>("post", "/api/v1/manager/user/award", { data });
};
/** 获取角色管理-用户余额管理 */
export const checkCashHistory = (data?: object) => {
  return http.request<Result>("post", "/api/v1/manager/cashHistory", { data });
};

/** 获取词典管理-获取一级词典类型 */
export const getDictTypes = (data?: object) => {
  return http.request<Result>("post", "/api/v1/outside/dictTypes",{data});
};

/** 获取词典管理-根据类型查找二级词典 */
export const getDictItems = (data?: object) => {
  return http.request<Result>("post", "/api/v1/outside/dictItems",{data});
}

/** 获取词典管理-删除词典类型 */
export const delDictTypes = (data?: object) => {
  return http.request<Result>("delete", "/api/v1/outside/dictTypes",{data});
};
/** 获取词典管理-删除字典项 */
export const delDictItems = (data?: object) => {
  return http.request<Result>("delete", "/api/v1/outside/dictItems",{data});
};
/** 获取词典管理-修改字典类型 */
export const setDictType = (data?: object) => {
  return http.request<Result>("post", "/api/v1/outside/setDictType",{data});
};
/** 获取词典管理-修改字典项 */
export const setDictItem = (data?: object) => {
  return http.request<Result>("post", "/api/v1/outside/setDictItem",{data});
};






