/**
 * @deprecated 此文件已废弃。bill 模块路由已迁移到 views/bill/route-config.ts，
 * 并从一级菜单调整为"资源运营"下的二级路由。
 * 原"云账号管理"（一级/二级账号 tab 形式）已拆分为独立路由页面（MENU_BILL_ROOT_ACCOUNT、MENU_BILL_MAIN_ACCOUNT）。
 * 原"云账单管理"已更名为"账单管理"（MENU_BILL_MANAGE）。
 */
import type { RouteRecordRaw } from 'vue-router';

const bill: RouteRecordRaw[] = [
  {
    path: '/bill/account-manage',
    name: 'account-manage',
    component: () => import('@/views/bill/account/account-manage/index'),
    meta: {
      title: '云账号管理',
      activeKey: 'account-manage',
      icon: 'hcm-icon bkhcm-icon-account-manage',
    },
  },
  {
    path: '/bill/bill-manage',
    name: 'bill-manage',
    component: () => import('@/views/bill/bill/index'),
    redirect: '/bill/bill-manage/summary',
    children: [
      {
        path: 'summary',
        name: 'billSummary',
        component: () => import('@/views/bill/bill/summary'),
        redirect: '/bill/bill-manage/summary/manage',
        children: [
          {
            path: 'manage',
            name: 'billSummaryManage',
            component: () => import('@/views/bill/bill/summary/manage'),
          },
          {
            path: 'operation-record',
            name: 'billSummaryOperationRecord',
            component: () => import('@/views/bill/bill/summary/operation-record'),
          },
        ],
      },
      {
        path: 'detail',
        name: 'billDetail',
        component: () => import('@/views/bill/bill/detail'),
      },
      {
        path: 'adjust',
        name: 'billAdjust',
        component: () => import('@/views/bill/bill/adjust'),
      },
    ],
    meta: {
      title: '云账单管理',
      activeKey: 'bill-manage',
      icon: 'hcm-icon bkhcm-icon-bill-manage',
      hasPageRoute: true,
      checkAuth: 'account_bill_find',
    },
  },
  {
    path: '/bill/account-manage/first-account',
    name: '录入一级账号',
    component: () => import('@/views/bill/account/create-account/create-first-account'),
    meta: {
      notMenu: true,
      activeKey: 'account-manage',
    },
  },
  {
    path: '/bill/account-manage/second-account',
    name: '创建二级账号',
    component: () => import('@/views/bill/account/create-account/create-second-account'),
    meta: {
      notMenu: true,
      activeKey: 'account-manage',
    },
  },
];

export default bill;
