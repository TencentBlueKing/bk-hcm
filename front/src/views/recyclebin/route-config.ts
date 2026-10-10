import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_RECYCLE_BIN } from '@/constants/auth-symbols';
import { MENU_RESOURCE, MENU_RESOURCE_RECYCLEBIN } from '@/constants/menu-symbol';

const recyclebinRsc: RouteRecordRaw[] = [
  {
    name: MENU_RESOURCE_RECYCLEBIN,
    path: 'recyclebin',
    component: () => import('@/views/recyclebin/index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        title: '回收管理',
        activeKey: MENU_RESOURCE_RECYCLEBIN,
        auth: { view: { type: AUTH_FIND_RECYCLE_BIN } },
      }),
    },
  },
];

// 原资源接入「回收站」页签；accountId / type / cvm 原样带过去
const recyclebinRscLegacy: RouteRecordRaw[] = [
  { path: '/resource/resource/recycle', redirect: { name: MENU_RESOURCE_RECYCLEBIN } },
];

export { recyclebinRsc, recyclebinRscLegacy };
