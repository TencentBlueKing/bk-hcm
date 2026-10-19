import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { useCommonStore } from '@/store';
import {
  MENU_BILL_ROOT_ACCOUNT,
  MENU_BILL_ROOT_ACCOUNT_CREATE,
  MENU_BILL_MAIN_ACCOUNT,
  MENU_BILL_MAIN_ACCOUNT_CREATE,
  MENU_BILL_MANAGE,
  MENU_BILL_MANAGE_SUMMARY,
  MENU_BILL_MANAGE_SUMMARY_MANAGE,
  MENU_BILL_MANAGE_SUMMARY_OPERATION_RECORD,
  MENU_BILL_MANAGE_DETAIL,
  MENU_BILL_MANAGE_ADJUST,
} from '@/constants/menu-symbol';

// 直接访问列表地址时的查看权限拦截在 store/common.ts 的 pageAuthData（按 path 匹配）
// 二级账号面向普通用户：无权限也要显示菜单，点进去由 403 页引导自行申请，所以不配 checkAuth
const billAccountManage: RouteRecordRaw[] = [
  {
    name: MENU_BILL_ROOT_ACCOUNT,
    path: 'root-account',
    component: () => import('./account/account-manage/root-account-list.vue'),
    meta: {
      ...new Meta({
        title: '一级账号',
        activeKey: MENU_BILL_ROOT_ACCOUNT,
        checkAuth: 'root_account_find',
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
  {
    name: MENU_BILL_ROOT_ACCOUNT_CREATE,
    path: 'root-account/create',
    component: () => import('./account/create-account/create-first-account'),
    meta: {
      ...new Meta({
        title: '录入一级账号',
        activeKey: MENU_BILL_ROOT_ACCOUNT,
        menu: { relative: MENU_BILL_ROOT_ACCOUNT },
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
  {
    name: MENU_BILL_MAIN_ACCOUNT,
    path: 'main-account',
    component: () => import('./account/account-manage/main-account-list.vue'),
    meta: {
      ...new Meta({
        title: '二级账号',
        activeKey: MENU_BILL_MAIN_ACCOUNT,
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
  {
    name: MENU_BILL_MAIN_ACCOUNT_CREATE,
    path: 'main-account/create',
    component: () => import('./account/create-account/create-second-account'),
    meta: {
      ...new Meta({
        title: '创建二级账号',
        activeKey: MENU_BILL_MAIN_ACCOUNT,
        menu: { relative: MENU_BILL_MAIN_ACCOUNT },
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
];

// 旧地址兼容。account-manage 同时是顶栏「资源运营」的入口地址。
// 分流依赖权限数据：redirect 在全局守卫加载权限之前求值，拿不到，所以放在 beforeEnter（全局守卫放行之后才执行）
const billAccountManageLegacy: RouteRecordRaw[] = [
  {
    path: 'account-manage',
    // 类型要求 component/children/redirect 之一；带 redirect 的记录不会执行 beforeEnter
    children: [],
    beforeEnter: (to) => {
      const { authVerifyData } = useCommonStore();
      const name = authVerifyData?.permissionAction?.root_account_find
        ? MENU_BILL_ROOT_ACCOUNT
        : MENU_BILL_MAIN_ACCOUNT;
      return { name, query: to.query, replace: true };
    },
  },
  {
    path: 'account-manage/first-account',
    redirect: { name: MENU_BILL_ROOT_ACCOUNT_CREATE },
  },
  {
    path: 'account-manage/second-account',
    redirect: { name: MENU_BILL_MAIN_ACCOUNT_CREATE },
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

export { billAccountManage, billAccountManageLegacy, billManage };
