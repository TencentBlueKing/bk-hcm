import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_TASK_MANAGEMENT,
  MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
} from '@/constants/menu-symbol';

export default [
  {
    name: MENU_BUSINESS_TASK_MANAGEMENT,
    path: 'task/:resourceType?',
    component: () => import('./index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_TASK_MANAGEMENT,
        layout: { breadcrumb: { show: false } },
        menu: {
          i18n: '任务管理',
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
    path: 'task/:resourceType?/details/:id',
    component: () => import('./details/index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_TASK_MANAGEMENT,
        layout: { breadcrumb: { show: false } },
        menu: {
          i18n: '任务详情',
          relative: MENU_BUSINESS_TASK_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
