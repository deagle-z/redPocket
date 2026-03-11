const Layout = () => import("@/layout/index.vue");

export default [
  {
    path: "/sys-bot-user",
    name: "SysBotUserMenu",
    component: Layout,
    redirect: "/sys-bot-user/index",
    meta: {
      icon: "ri:robot-2-fill",
      title: "机器人管理",
      rank: 20
    },
    children: [
      {
        path: "/sys-bot-user/index",
        name: "SysBotUser",
        component: () => import("@/views/sys/bot_user/index.vue"),
        meta: {
          title: "批量添加机器人",
          showLink: true
        }
      }
    ]
  }
] satisfies RouteConfigsTable[];
