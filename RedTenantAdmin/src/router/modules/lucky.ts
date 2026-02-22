const Layout = () => import("@/layout/index.vue");

export default [
  {
    path: "/auth-group",
    name: "AuthGroupMenu",
    component: Layout,
    redirect: "/auth-group/index",
    meta: {
      icon: "ri:group-fill",
      title: "授权群组",
      rank: 1
    },
    children: [
      {
        path: "/auth-group/index",
        name: "LuckyAuthGroup",
        component: () => import("@/views/lucky/auth_group/index.vue"),
        meta: {
          title: "授权群组",
          showLink: true
        }
      }
    ]
  },
  {
    path: "/lucky-history",
    name: "LuckyHistoryMenu",
    component: Layout,
    redirect: "/lucky-history/index",
    meta: {
      icon: "ri:history-fill",
      title: "红包历史",
      rank: 2
    },
    children: [
      {
        path: "/lucky-history/index",
        name: "LuckyHistory",
        component: () => import("@/views/lucky/lucky_history/index.vue"),
        meta: {
          title: "红包历史",
          showLink: true
        }
      }
    ]
  },
  {
    path: "/lucky-money",
    name: "LuckyMoneyMenu",
    component: Layout,
    redirect: "/lucky-money/index",
    meta: {
      icon: "ri:money-dollar-circle-fill",
      title: "红包管理",
      rank: 3
    },
    children: [
      {
        path: "/lucky-money/index",
        name: "LuckyMoney",
        component: () => import("@/views/lucky/lucky_money/index.vue"),
        meta: {
          title: "红包管理",
          showLink: true
        }
      }
    ]
  },
  {
    path: "/recharge-order",
    name: "RechargeOrderMenu",
    component: Layout,
    redirect: "/recharge-order/index",
    meta: {
      icon: "ri:wallet-3-fill",
      title: "充值订单",
      rank: 4
    },
    children: [
      {
        path: "/recharge-order/index",
        name: "RechargeOrder",
        component: () => import("@/views/lucky/recharge_order/index.vue"),
        meta: {
          title: "充值订单",
          showLink: true
        }
      }
    ]
  },
  {
    path: "/tg-user",
    name: "TgUserMenu",
    component: Layout,
    redirect: "/tg-user/index",
    meta: {
      icon: "ri:telegram-fill",
      title: "TG用户",
      rank: 5
    },
    children: [
      {
        path: "/tg-user/index",
        name: "TgUser",
        component: () => import("@/views/lucky/tg_user/index.vue"),
        meta: {
          title: "TG用户",
          showLink: true
        }
      }
    ]
  },
  {
    path: "/tg-user-rebate-record",
    name: "TgUserRebateRecordMenu",
    component: Layout,
    redirect: "/tg-user-rebate-record/index",
    meta: {
      icon: "ri:exchange-dollar-fill",
      title: "返水记录",
      rank: 6
    },
    children: [
      {
        path: "/tg-user-rebate-record/index",
        name: "TgUserRebateRecord",
        component: () => import("@/views/lucky/rebate_record/index.vue"),
        meta: {
          title: "返水记录",
          showLink: true
        }
      }
    ]
  },
  {
    path: "/withdraw-order-br",
    name: "WithdrawOrderBrMenu",
    component: Layout,
    redirect: "/withdraw-order-br/index",
    meta: {
      icon: "ri:bank-card-fill",
      title: "提现订单",
      rank: 7
    },
    children: [
      {
        path: "/withdraw-order-br/index",
        name: "WithdrawOrderBr",
        component: () => import("@/views/lucky/withdraw_order_br/index.vue"),
        meta: {
          title: "提现订单",
          showLink: true
        }
      }
    ]
  }
] satisfies RouteConfigsTable[];
