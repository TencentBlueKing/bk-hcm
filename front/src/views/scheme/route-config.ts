import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_SERVICE,
  MENU_SCHEME,
  MENU_SCHEME_RECOMMENDATION,
  MENU_SCHEME_LIST,
  MENU_SCHEME_DETAIL,
} from '@/constants/menu-symbol';
import { AUTH_CREATE_CLOUD_SELECTION_SCHEME, AUTH_FIND_CLOUD_SELECTION_SCHEME } from '@/constants/auth-symbols';

export default [
  {
    name: MENU_SCHEME,
    path: 'scheme',
    component: () => import('@/views/scheme/index'),
    redirect: { name: MENU_SCHEME_RECOMMENDATION },
    meta: {
      ...new Meta({
        owner: MENU_SERVICE,
        activeKey: MENU_SCHEME,
        menu: {
          i18n: '资源选型',
        },
      }),
    },
    children: [
      {
        name: MENU_SCHEME_RECOMMENDATION,
        path: 'recommendation',
        component: () => import('@/views/scheme/scheme-recommendation/index'),
        meta: {
          ...new Meta({
            owner: MENU_SERVICE,
            activeKey: MENU_SCHEME,
            auth: { view: { type: AUTH_CREATE_CLOUD_SELECTION_SCHEME } },
            menu: {
              i18n: '资源推荐',
              relative: MENU_SCHEME,
            },
            layout: {
              breadcrumb: {
                show: false,
              },
            },
          }),
        },
      },
      {
        name: MENU_SCHEME_LIST,
        path: 'deployment/list',
        component: () => import('@/views/scheme/scheme-list/index'),
        meta: {
          ...new Meta({
            owner: MENU_SERVICE,
            activeKey: MENU_SCHEME,
            auth: { view: { type: AUTH_FIND_CLOUD_SELECTION_SCHEME } },
            menu: {
              i18n: '方案列表',
              relative: MENU_SCHEME,
            },
          }),
        },
      },
      {
        name: MENU_SCHEME_DETAIL,
        path: 'deployment/detail',
        component: () => import('@/views/scheme/scheme-detail/index'),
        meta: {
          ...new Meta({
            owner: MENU_SERVICE,
            activeKey: MENU_SCHEME,

            menu: {
              i18n: '方案详情',
              relative: MENU_SCHEME_LIST,
            },
          }),
        },
      },
    ],
  },
] as RouteRecordRaw[];
