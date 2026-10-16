import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS, MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';

export default [
  {
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    path: 'cloud-account-manage',
    component: () => import('./index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_CLOUD_ACCOUNT,
        menu: { i18n: '云账号管理' },
      }),
    },
  },
] as RouteRecordRaw[];
