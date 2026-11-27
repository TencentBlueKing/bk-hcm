import type { RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_ROOT_ACCOUNT, AUTH_FIND_MAIN_ACCOUNT } from '@/constants/auth-symbols';
import {
  MENU_RESOURCE,
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
    path: 'bill/root-account',
    component: () => import('./account/account-manage/root-account-list.vue'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_ROOT_ACCOUNT,
        auth: { view: { type: AUTH_FIND_ROOT_ACCOUNT } },
        menu: { i18n: '一级账号' },
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
  {
    name: MENU_BILL_ROOT_ACCOUNT_CREATE,
    path: 'bill/root-account/create',
    component: () => import('./account/create-account/create-first-account'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_ROOT_ACCOUNT,
        menu: { i18n: '录入一级账号', relative: MENU_BILL_ROOT_ACCOUNT },
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
  {
    name: MENU_BILL_MAIN_ACCOUNT,
    path: 'bill/main-account',
    component: () => import('./account/account-manage/main-account-list.vue'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_MAIN_ACCOUNT,
        auth: { view: { type: AUTH_FIND_MAIN_ACCOUNT } },
        menu: { i18n: '二级账号' },
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
  {
    name: MENU_BILL_MAIN_ACCOUNT_CREATE,
    path: 'bill/main-account/create',
    component: () => import('./account/create-account/create-second-account'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_MAIN_ACCOUNT,
        menu: { i18n: '创建二级账号', relative: MENU_BILL_MAIN_ACCOUNT },
        layout: { breadcrumb: { show: true } },
      }),
    },
  },
];

// 旧地址 /bill/*（账单迁到资源运营 /resource/bill/* 之前），query 原样保留；裸 /bill 不再注册，落 404
// /bill/account-manage 统一默认跳二级账号（不再按一级账号权限分流）
const toName = (name: symbol) => (to: RouteLocationNormalized) => ({ name, query: to.query });
const billLegacy: RouteRecordRaw[] = [
  { path: '/bill/account-manage', redirect: toName(MENU_BILL_MAIN_ACCOUNT) },
  { path: '/bill/account-manage/first-account', redirect: toName(MENU_BILL_ROOT_ACCOUNT_CREATE) },
  { path: '/bill/account-manage/second-account', redirect: toName(MENU_BILL_MAIN_ACCOUNT_CREATE) },
  { path: '/bill/root-account', redirect: toName(MENU_BILL_ROOT_ACCOUNT) },
  { path: '/bill/root-account/create', redirect: toName(MENU_BILL_ROOT_ACCOUNT_CREATE) },
  { path: '/bill/main-account', redirect: toName(MENU_BILL_MAIN_ACCOUNT) },
  { path: '/bill/main-account/create', redirect: toName(MENU_BILL_MAIN_ACCOUNT_CREATE) },
  { path: '/bill/bill-manage', redirect: toName(MENU_BILL_MANAGE_SUMMARY_MANAGE) },
  { path: '/bill/bill-manage/summary', redirect: toName(MENU_BILL_MANAGE_SUMMARY_MANAGE) },
  { path: '/bill/bill-manage/summary/manage', redirect: toName(MENU_BILL_MANAGE_SUMMARY_MANAGE) },
  { path: '/bill/bill-manage/summary/operation-record', redirect: toName(MENU_BILL_MANAGE_SUMMARY_OPERATION_RECORD) },
  { path: '/bill/bill-manage/detail', redirect: toName(MENU_BILL_MANAGE_DETAIL) },
  { path: '/bill/bill-manage/adjust', redirect: toName(MENU_BILL_MANAGE_ADJUST) },
];

// 云账单管理是布局型页面：./bill/index 提供页签栏与 currentMonth/bill_year/bill_month 上下文，
// 子路由是页面内部的页签而非侧栏菜单，因此保留嵌套，不打平。
const billManage: RouteRecordRaw[] = [
  {
    name: MENU_BILL_MANAGE,
    path: 'bill/manage',
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
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_MANAGE,
        menu: { i18n: '账单管理' },
        layout: { breadcrumb: { show: false } },
      }),
    },
  },
];

export { billAccountManage, billLegacy, billManage };
