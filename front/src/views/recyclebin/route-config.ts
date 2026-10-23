import type { RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_RECYCLE_BIN } from '@/constants/auth-symbols';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_RECYCLEBIN,
  MENU_RESOURCE,
  MENU_RESOURCE_RECYCLEBIN,
} from '@/constants/menu-symbol';

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
      }),
    },
  },
  // 旧地址：主机、硬盘下的回收站，:type 即页签
  ...['host', 'drive'].map((resource) => ({
    path: `${resource}/recyclebin/:type`,
    redirect: (to: RouteLocationNormalized) => ({
      name: MENU_BUSINESS_RECYCLEBIN,
      params: { bizId: to.params.bizId },
      query: { ...to.query, type: to.params.type },
    }),
  })),
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
        auth: { view: { type: AUTH_FIND_RECYCLE_BIN } },
        menu: { i18n: '回收管理' },
      }),
    },
  },
];

// 原资源接入「回收站」页签；accountId / type / cvm 原样带过去
const recyclebinRscLegacy: RouteRecordRaw[] = [
  { path: '/resource/resource/recycle', redirect: { name: MENU_RESOURCE_RECYCLEBIN } },
];

export { recyclebinBiz, recyclebinRsc, recyclebinRscLegacy };
