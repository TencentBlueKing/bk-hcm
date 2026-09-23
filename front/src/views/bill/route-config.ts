import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BILL_ACCOUNT_MANAGE,
  MENU_BILL_ROOT_ACCOUNT_CREATE,
  MENU_BILL_MAIN_ACCOUNT_CREATE,
  MENU_BILL_MANAGE,
  MENU_BILL_MANAGE_SUMMARY,
  MENU_BILL_MANAGE_SUMMARY_MANAGE,
  MENU_BILL_MANAGE_SUMMARY_OPERATION_RECORD,
  MENU_BILL_MANAGE_DETAIL,
  MENU_BILL_MANAGE_ADJUST,
} from '@/constants/menu-symbol';

const billAccountManage: RouteRecordRaw[] = [
  {
    name: MENU_BILL_ACCOUNT_MANAGE,
    path: 'account-manage',
    component: () => import('./account/account-manage/index'),
    meta: {
      ...new Meta({
        title: '云账号管理',
        activeKey: MENU_BILL_ACCOUNT_MANAGE,
      }),
    },
  },
  {
    name: MENU_BILL_ROOT_ACCOUNT_CREATE,
    path: 'account-manage/first-account',
    component: () => import('./account/create-account/create-first-account'),
    meta: {
      ...new Meta({
        activeKey: MENU_BILL_ACCOUNT_MANAGE,
      }),
    },
  },
  {
    name: MENU_BILL_MAIN_ACCOUNT_CREATE,
    path: 'account-manage/second-account',
    component: () => import('./account/create-account/create-second-account'),
    meta: {
      ...new Meta({
        activeKey: MENU_BILL_ACCOUNT_MANAGE,
      }),
    },
  },
];

// 云账单管理是布局型页面：./bill/index 提供页签栏与 currentMonth/bill_year/bill_month 上下文，
// 子路由是页面内部的页签而非侧栏菜单，因此保留嵌套，不打平。
const billManage: RouteRecordRaw[] = [
  {
    name: MENU_BILL_MANAGE,
    path: 'bill-manage',
    component: () => import('./bill/index'),
    redirect: { name: MENU_BILL_MANAGE_SUMMARY },
    children: [
      {
        path: 'summary',
        name: MENU_BILL_MANAGE_SUMMARY,
        component: () => import('./bill/summary'),
        redirect: { name: MENU_BILL_MANAGE_SUMMARY_MANAGE },
        children: [
          {
            path: 'manage',
            name: MENU_BILL_MANAGE_SUMMARY_MANAGE,
            component: () => import('./bill/summary/manage'),
          },
          {
            path: 'operation-record',
            name: MENU_BILL_MANAGE_SUMMARY_OPERATION_RECORD,
            component: () => import('./bill/summary/operation-record'),
          },
        ],
      },
      {
        path: 'detail',
        name: MENU_BILL_MANAGE_DETAIL,
        component: () => import('./bill/detail'),
      },
      {
        path: 'adjust',
        name: MENU_BILL_MANAGE_ADJUST,
        component: () => import('./bill/adjust'),
      },
    ],
    meta: {
      ...new Meta({
        title: '云账单管理',
        activeKey: MENU_BILL_MANAGE,
        checkAuth: 'account_bill_find',
      }),
    },
  },
];

export { billAccountManage, billManage };
