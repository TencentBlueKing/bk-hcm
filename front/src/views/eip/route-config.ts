import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS, MENU_BUSINESS_EIP_MANAGEMENT, MENU_BUSINESS_EIP_DETAILS } from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_EIP_MANAGEMENT,
    path: 'ip',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_EIP_MANAGEMENT,
        menu: {
          i18n: '弹性IP',
        },
        layout: {
          breadcrumb: {
            show: true,
          },
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_EIP_DETAILS,
    path: 'ip/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_EIP_MANAGEMENT,
        menu: {
          i18n: '弹性IP详情',
          relative: MENU_BUSINESS_EIP_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
