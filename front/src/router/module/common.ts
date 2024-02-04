/**
 * 公告路由
 */

import type { RouteRecordRaw } from 'vue-router';

const common: RouteRecordRaw[] = [
  {
    path: '/:pathMatch(.*)',
    redirect: '/',
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
  {
    path: '/*',
    redirect: '/',
  },
  {
    path: '/403/:id',
    name: '403',
    component: () => import('@/views/error-pages/403'),
    meta: {
      breadcrumb: ['申请权限'],
    },
  },
  //   {
  //     path: '/root',
  //     name: 'root',
  //     alias: '/',
  //     component: import('@/views/home/RootPath'),
  //   },
  //   {
  //     path: '/test',
  //     name: 'test',
  //     component: () => import('@/views/test/index'),
  //   },
  //   {
  //     path: '/exception',
  //     name: 'exception',
  //     component: () => import('@/views/exception'),
  //     meta: {
  //       isHideNav: true,
  //     },
  //   },
  //   {
  //     path: '/403',
  //     name: '403',
  //     component: () => import('@/views/403'),
  //   },
];
export default common;
