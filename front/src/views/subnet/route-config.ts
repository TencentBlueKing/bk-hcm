import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_SUBNET_MANAGEMENT,
  MENU_BUSINESS_SUBNET_DETAILS,
  MENU_BUSINESS_APPLY_SUBNET,
} from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_SUBNET_MANAGEMENT,
    path: 'subnet',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_SUBNET_MANAGEMENT,
        menu: {
          i18n: '子网',
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
    name: MENU_BUSINESS_SUBNET_DETAILS,
    path: 'subnet/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_SUBNET_MANAGEMENT,
        menu: {
          i18n: '子网详情',
          relative: MENU_BUSINESS_SUBNET_MANAGEMENT,
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_APPLY_SUBNET,
    path: 'service/service-apply/subnet',
    component: () => import('@/views/service/service-apply/subnet'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_SUBNET_MANAGEMENT,
        menu: {
          i18n: '申请子网',
          relative: MENU_BUSINESS_SUBNET_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
