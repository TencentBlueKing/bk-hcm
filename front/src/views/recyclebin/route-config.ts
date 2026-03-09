import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_RECYCLE_BIN } from '@/constants/auth-symbols';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_RECYCLEBIN,
  MENU_RESOURCE,
  MENU_RESOURCE_RECYCLEBIN,
} from '@/constants/menu-symbol';

/**
 * 回收站模块 —— 业务视图与资源视图共用同一组件
 * 参考 operation-log 的双导出模式
 */

const recyclebinBiz: RouteRecordRaw[] = [
  {
    name: MENU_BUSINESS_RECYCLEBIN,
    path: 'recyclebin',
    component: () => import('@/views/recyclebin/index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_RECYCLEBIN,
        menu: {
          i18n: '回收站',
        },
        layout: {
          breadcrumb: {
            show: true,
          },
        },
      }),
    },
  },
];

const recyclebinRsc: RouteRecordRaw[] = [
  {
    name: MENU_RESOURCE_RECYCLEBIN,
    path: 'recyclebin',
    component: () => import('@/views/recyclebin/index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_RESOURCE_RECYCLEBIN,
        auth: {
          view: { type: AUTH_FIND_RECYCLE_BIN },
        },
        menu: {
          i18n: '回收管理',
        },
      }),
    },
  },
];

export { recyclebinBiz, recyclebinRsc };
