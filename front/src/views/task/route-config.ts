import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS_TASK_MANAGEMENT, MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';

export default [
  {
    name: MENU_BUSINESS_TASK_MANAGEMENT,
    path: 'task/:resourceType?',
    component: () => import('./index.vue'),
    meta: {
      ...new Meta({
        title: '任务管理',
        activeKey: MENU_BUSINESS_TASK_MANAGEMENT,
        // 没有业务访问权限不会展示侧边栏导航，这里只是做一个权限优化的占位提示
        auth: {
          view: { type: 'biz_access' },
        },
        menu: {
          relative: MENU_BUSINESS_TASK_MANAGEMENT,
        },
        icon: 'hcm-icon bkhcm-icon-bushu',
      }),
    },
  },
  {
    name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
    path: 'task/:resourceType?/details/:id',
    component: () => import('./details/index.vue'),
    meta: {
      ...new Meta({
        title: '任务详情',
        notMenu: true,
        activeKey: MENU_BUSINESS_TASK_MANAGEMENT,
        isShowBreadcrumb: true,
        menu: {
          relative: MENU_BUSINESS_TASK_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
