import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_ROOT_ACCOUNT, AUTH_FIND_MAIN_ACCOUNT, AUTH_FIND_ACCOUNT_BILL } from '@/constants/auth-symbols';
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

/**
 * 账单模块 —— 放置在"资源运营"一级菜单下，二级菜单名"云账号管理"
 *
 * 路由结构（均为 resource 的相对路径）：
 *   bill/root-account          → 一级账号列表
 *   bill/root-account/create   → 录入一级账号
 *   bill/main-account          → 二级账号列表
 *   bill/main-account/create   → 创建二级账号
 *   bill/manage                → 账单管理（带子路由 tab）
 */
const billRoutes: RouteRecordRaw[] = [
  // ===== 一级账号 =====
  {
    name: MENU_BILL_ROOT_ACCOUNT,
    path: 'bill/root-account',
    component: () => import('@/views/bill/account/account-manage/root-account-list'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_ROOT_ACCOUNT,
        auth: {
          view: { type: AUTH_FIND_ROOT_ACCOUNT },
        },
        menu: {
          i18n: '一级账号',
        },
      }),
    },
  },
  {
    name: MENU_BILL_ROOT_ACCOUNT_CREATE,
    path: 'bill/root-account/create',
    component: () => import('@/views/bill/account/create-account/create-first-account'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_ROOT_ACCOUNT,
        menu: {
          i18n: '录入一级账号',
          relative: MENU_BILL_ROOT_ACCOUNT,
        },
      }),
    },
  },
  // ===== 二级账号 =====
  {
    name: MENU_BILL_MAIN_ACCOUNT,
    path: 'bill/main-account',
    component: () => import('@/views/bill/account/account-manage/main-account-list'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_MAIN_ACCOUNT,
        auth: {
          view: { type: AUTH_FIND_MAIN_ACCOUNT },
        },
        menu: {
          i18n: '二级账号',
        },
      }),
    },
  },
  {
    name: MENU_BILL_MAIN_ACCOUNT_CREATE,
    path: 'bill/main-account/create',
    component: () => import('@/views/bill/account/create-account/create-second-account'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_MAIN_ACCOUNT,
        menu: {
          i18n: '创建二级账号',
          relative: MENU_BILL_MAIN_ACCOUNT,
        },
      }),
    },
  },
  // ===== 账单管理 =====
  {
    name: MENU_BILL_MANAGE,
    path: 'bill/manage',
    redirect: { name: MENU_BILL_MANAGE_SUMMARY },
    component: () => import('@/views/bill/bill/index'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_BILL_MANAGE,
        auth: {
          view: { type: AUTH_FIND_ACCOUNT_BILL },
        },
        menu: {
          i18n: '账单管理',
        },
      }),
    },
    children: [
      {
        name: MENU_BILL_MANAGE_SUMMARY,
        path: 'summary',
        redirect: { name: MENU_BILL_MANAGE_SUMMARY_MANAGE },
        component: () => import('@/views/bill/bill/summary'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_BILL_MANAGE,
          }),
        },
        children: [
          {
            name: MENU_BILL_MANAGE_SUMMARY_MANAGE,
            path: 'manage',
            component: () => import('@/views/bill/bill/summary/manage'),
            meta: {
              ...new Meta({
                owner: MENU_RESOURCE,
                activeKey: MENU_BILL_MANAGE,
              }),
            },
          },
          {
            name: MENU_BILL_MANAGE_SUMMARY_OPERATION_RECORD,
            path: 'operation-record',
            component: () => import('@/views/bill/bill/summary/operation-record'),
            meta: {
              ...new Meta({
                owner: MENU_RESOURCE,
                activeKey: MENU_BILL_MANAGE,
              }),
            },
          },
        ],
      },
      {
        name: MENU_BILL_MANAGE_DETAIL,
        path: 'detail',
        component: () => import('@/views/bill/bill/detail'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_BILL_MANAGE,
          }),
        },
      },
      {
        name: MENU_BILL_MANAGE_ADJUST,
        path: 'adjust',
        component: () => import('@/views/bill/bill/adjust'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_BILL_MANAGE,
          }),
        },
      },
    ],
  },
];

export default billRoutes;
