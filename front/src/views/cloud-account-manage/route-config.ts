import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';

export default [
  {
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    path: 'cloud-account-manage',
    component: () => import('./index.vue'),
    meta: {
      ...new Meta({
        title: '云账号管理',
        activeKey: MENU_BUSINESS_CLOUD_ACCOUNT,
        isShowBreadcrumb: true,
        icon: 'hcm-icon bkhcm-icon-account-manage',
      }),
    },
  },
] as RouteRecordRaw[];
