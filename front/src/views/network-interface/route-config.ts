import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
  MENU_BUSINESS_NIF_DETAILS,
} from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
    path: 'network-interface',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
        menu: {
          i18n: '网络接口',
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
    name: MENU_BUSINESS_NIF_DETAILS,
    path: 'network-interface/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
        menu: {
          i18n: '网络接口详情',
          relative: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
