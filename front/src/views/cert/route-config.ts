import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS, MENU_BUSINESS_CERT_MANAGEMENT } from '@/constants/menu-symbol';

export default [
  {
    name: MENU_BUSINESS_CERT_MANAGEMENT,
    path: 'cert',
    component: () => import('@/views/business/cert-manager/index'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_CERT_MANAGEMENT,
        menu: {
          i18n: '证书托管',
        },
      }),
      // 账号选择器按 route.meta.isFilterAccount 读取
      isFilterAccount: true,
    },
  },
] as RouteRecordRaw[];
