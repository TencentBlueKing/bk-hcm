import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { useAuthStore } from '@/store/auth';
import { AUTH_FIND_ROOT_ACCOUNT, AUTH_FIND_MAIN_ACCOUNT } from '@/constants/auth-symbols';
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

// 列表页的查看权限由 meta.auth.view 拦截；侧栏显隐由 constants/view-auth.ts 的 hideMenu 决定
// 二级账号面向普通用户：无权限也要显示菜单，点进去由无权限页引导自行申请，所以不配 hideMenu
const billAccountManage: RouteRecordRaw[] = [
  {
    name: MENU_BILL_ROOT_ACCOUNT,
    path: 'root-account',
    component: () => import('./account/account-manage/root-account-list.vue'),
    meta: {
      ...new Meta({
        title: '一级账号',
        activeKey: MENU_BILL_ROOT_ACCOUNT,
        auth: { view: { type: AUTH_FIND_ROOT_ACCOUNT } },
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
        auth: { view: { type: AUTH_FIND_MAIN_ACCOUNT } },
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
// 分流读 preload 预鉴权的一级账号视图权限；未取到（接口失败）时按无权限去二级账号
const billAccountManageLegacy: RouteRecordRaw[] = [
  {
    path: 'account-manage',
    // 类型要求 component/children/redirect 之一；带 redirect 的记录不会执行 beforeEnter
    children: [],
    beforeEnter: (to) => {
      const authStore = useAuthStore();
      const hasRootAccount =
        authStore.isViewPermissionCached(MENU_BILL_ROOT_ACCOUNT) && authStore.hasViewPermission(MENU_BILL_ROOT_ACCOUNT);
      const name = hasRootAccount ? MENU_BILL_ROOT_ACCOUNT : MENU_BILL_MAIN_ACCOUNT;
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
        layout: { breadcrumb: { show: false } },
      }),
    },
  },
];

export { billAccountManage, billAccountManageLegacy, billManage };
