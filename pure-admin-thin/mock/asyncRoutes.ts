// // 模拟后端动态生成路由
// import { defineFakeRoute } from "vite-plugin-fake-server/client";
//
// /**
//  * roles：页面级别权限，这里模拟二种 "admin"、"common"
//  * admin：管理员角色
//  * common：普通角色
//  */
// const permissionRouter = {
//   id: 1,
//   parent_id: 0,
//   menuType: 1,
//   path: "/sys",
//   name: "SystemSetting",
//   queue: 0,
//   meta: {
//     title: "系统管理",
//     icon: "solar:settings-linear",
//     showLink: true,
//     showParent: false,
//     keepAlive: false,
//     frameLoading: false,
//     hiddenTag: false,
//     transition: {
//       enterTransition: "animate__fadeIn animate__faster",
//       leaveTransition: "animate__fadeOut animate__faster"
//     }
//   },
//   children: [
//     {
//       id: 2,
//       parent_id: 1,
//       menuType: 1,
//       path: "/sys/menu/index",
//       name: "SysMenuManagement",
//       queue: 0,
//       meta: {
//         title: "菜单管理",
//         showLink: true,
//         showParent: true,
//         keepAlive: true,
//         frameLoading: false,
//         hiddenTag: false,
//         transition: {
//           enterTransition: "animate__fadeIn animate__faster",
//           leaveTransition: "animate__fadeOut animate__faster"
//         }
//       },
//       children: [
//         {
//           id: 9,
//           parent_id: 2,
//           menuType: 1,
//           path: "path/sys/menu/text",
//           name: "text",
//           queue: 0,
//           meta: {
//             title: "测试",
//             showLink: true,
//             showParent: true,
//             keepAlive: false,
//             frameLoading: false,
//             hiddenTag: false,
//             transition: {
//               enterTransition: "animate__fadeIn animate__faster",
//               leaveTransition: "animate__fadeOut animate__faster"
//             }
//           }
//         }
//       ]
//     }
//   ]
// };
//
// export default defineFakeRoute([
//   {
//     url: "/api/v1/outside/menus",
//     method: "get",
//     response: () => {
//       let result = {
//         msg: "OK",
//         code: 0,
//         data: [permissionRouter]
//       };
//       // console.info("get menu back.", JSON.stringify(result));
//       return result;
//     }
//   }
// ]);
