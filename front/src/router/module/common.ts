/**
 * 公告路由
 */

import type { RouteRecordRaw } from 'vue-router';

const common: RouteRecordRaw[] = [
  {
    path: '/403/:id',
    name: '403',
    component: () => import('@/views/error-pages/403'),
    meta: {
      breadcrumb: ['申请权限'],
      layout: { breadcrumb: { show: false } },
    },
  },
];
export default common;
