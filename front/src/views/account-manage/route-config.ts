import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_SERVICE,
  MENU_SERVICE_ACCOUNT_MANAGE,
  MENU_SERVICE_ACCOUNT_DETAIL,
  MENU_SERVICE_ACCOUNT_BASIC,
  MENU_SERVICE_ACCOUNT_RESOURCE,
  MENU_SERVICE_ACCOUNT_USERS,
} from '@/constants/menu-symbol';

/**
 * 账号管理模块 —— 放置在"工作台"一级菜单下
 * 路由结构：
 *   service/account                         → 账号列表
 *   service/account/details/:accountId      → 详情容器（redirect → basic）
 *   service/account/details/:accountId/basic    → 基本信息
 *   service/account/details/:accountId/resource → 资源状态
 *   service/account/details/:accountId/user     → 用户列表
 */
const accountManageRoutes: RouteRecordRaw[] = [
  // 账号列表
  {
    name: MENU_SERVICE_ACCOUNT_MANAGE,
    path: 'account',
    component: () => import('@/views/account-manage/accountmanage/index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_SERVICE,
        activeKey: MENU_SERVICE_ACCOUNT_MANAGE,
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
    redirect: { name: MENU_SERVICE_ACCOUNT_BASIC },
    component: () => import('@/views/account-manage/accountInfo/index'),
    meta: {
      ...new Meta({
        owner: MENU_SERVICE,
        activeKey: MENU_SERVICE_ACCOUNT_MANAGE,
        menu: {
          i18n: '账号详情',
          relative: MENU_SERVICE_ACCOUNT_MANAGE,
        },
      }),
    },
    children: [
      {
        name: MENU_SERVICE_ACCOUNT_BASIC,
        path: 'basic',
        component: () => import('@/views/account-manage/accountmanage/account-detail'),
        meta: {
          ...new Meta({
            owner: MENU_SERVICE,
            activeKey: MENU_SERVICE_ACCOUNT_MANAGE,

            menu: {
              i18n: '基本信息',
              relative: MENU_SERVICE_ACCOUNT_DETAIL,
            },
          }),
        },
      },
      {
        name: MENU_SERVICE_ACCOUNT_RESOURCE,
        path: 'resource',
        component: () => import('@/views/account-manage/accountInfo/component/resourceStatus/index'),
        meta: {
          ...new Meta({
            owner: MENU_SERVICE,
            activeKey: MENU_SERVICE_ACCOUNT_MANAGE,

            menu: {
              i18n: '资源状态',
              relative: MENU_SERVICE_ACCOUNT_DETAIL,
            },
          }),
        },
      },
      {
        name: MENU_SERVICE_ACCOUNT_USERS,
        path: 'user',
        component: () => import('@/views/account-manage/accountInfo/component/usersList/index'),
        meta: {
          ...new Meta({
            owner: MENU_SERVICE,
            activeKey: MENU_SERVICE_ACCOUNT_MANAGE,

            menu: {
              i18n: '用户列表',
              relative: MENU_SERVICE_ACCOUNT_DETAIL,
            },
          }),
        },
      },
    ],
  },
];

export default accountManageRoutes;
