import type { RouteRecordRaw } from 'vue-router';
import { AUTH_CREATE_CLOUD_SELECTION_SCHEME, AUTH_FIND_CLOUD_SELECTION_SCHEME } from '@/constants/auth-symbols';

// 资源选型模块路由
const scheme: RouteRecordRaw[] = [
  {
    path: '/scheme',
    name: 'resource-scheme',
    component: () => import('@/views/scheme/index'),
    meta: {
      layout: { breadcrumb: { show: false } },
    },
    children: [
      {
        path: 'recommendation',
        name: 'scheme-recommendation',
        component: () => import('@/views/scheme/scheme-recommendation/index'),
        meta: {
          notMenu: true,
          auth: { view: { type: AUTH_CREATE_CLOUD_SELECTION_SCHEME } },
        },
      },
      {
        path: 'deployment/list',
        name: 'scheme-list',
        component: () => import('@/views/scheme/scheme-list/index'),
        meta: {
          notMenu: true,
          auth: { view: { type: AUTH_FIND_CLOUD_SELECTION_SCHEME } },
        },
      },
      {
        path: 'deployment/detail',
        name: 'scheme-detail',
        component: () => import('@/views/scheme/scheme-detail/index'),
        meta: {
          notMenu: true,
        },
      },
    ],
  },
];

export default scheme;
