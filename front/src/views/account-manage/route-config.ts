import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_ACCOUNT } from '@/constants/auth-symbols';
import {
  MENU_SERVICE,
  MENU_SERVICE_ACCOUNT_MANAGEMENT,
  MENU_SERVICE_ACCOUNT_DETAIL,
  MENU_SERVICE_ACCOUNT_DETAIL_BASIC,
} from '@/constants/menu-symbol';

/**
 * 账号管理模块 —— 放置在"工作台"一级菜单下
 * 路由结构：
 *   service/account                              → 账号列表
 *   service/account/details/:accountId            → 详情容器（redirect → basic）
 *   service/account/details/:accountId/basic      → 基本信息
 */
const accountManageRoutes: RouteRecordRaw[] = [
  // 账号列表
  {
    name: MENU_SERVICE_ACCOUNT_MANAGEMENT,
    path: 'account',
    component: () => import('@/views/account-manage/accountmanage/index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_SERVICE,
        activeKey: MENU_SERVICE_ACCOUNT_MANAGEMENT,
        auth: {
          view: { type: AUTH_FIND_ACCOUNT },
        },
        menu: {
          i18n: '账号管理',
        },
      }),
    },
  },
  // 账号详情（Tab 容器）
  {
    name: MENU_SERVICE_ACCOUNT_DETAIL,
    path: 'account/details/:accountId',
    redirect: { name: MENU_SERVICE_ACCOUNT_DETAIL_BASIC },
    component: () => import('@/views/account-manage/accountInfo/index'),
    meta: {
      ...new Meta({
        owner: MENU_SERVICE,
        activeKey: MENU_SERVICE_ACCOUNT_MANAGEMENT,
        menu: {
          i18n: '账号详情',
          relative: MENU_SERVICE_ACCOUNT_MANAGEMENT,
        },
      }),
    },
    children: [
      {
        name: MENU_SERVICE_ACCOUNT_DETAIL_BASIC,
        path: 'basic',
        component: () => import('@/views/account-manage/accountmanage/account-detail'),
        meta: {
          ...new Meta({
            owner: MENU_SERVICE,
            activeKey: MENU_SERVICE_ACCOUNT_MANAGEMENT,
            menu: {
              i18n: '基本信息',
              relative: MENU_SERVICE_ACCOUNT_MANAGEMENT,
            },
          }),
        },
      },
    ],
  },
];

export default accountManageRoutes;
