import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS_CLOUD_ACCOUNT, MENU_BUSINESS_CLOUD_ACCOUNT_DETAILS } from '@/constants/menu-symbol';

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
  {
    name: MENU_BUSINESS_CLOUD_ACCOUNT_DETAILS,
    path: 'cloud-account-manage/detail/:id',
    component: () => import('./index.vue'), // 后续替换为详情页面组件
    meta: {
      ...new Meta({
        title: '云账号详情',
        notMenu: true,
        activeKey: MENU_BUSINESS_CLOUD_ACCOUNT,
        isShowBreadcrumb: true,
        menu: {
          relative: MENU_BUSINESS_CLOUD_ACCOUNT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
