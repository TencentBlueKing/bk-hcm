import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_VPC_MANAGEMENT,
  MENU_BUSINESS_VPC_DETAILS,
  MENU_BUSINESS_APPLY_VPC,
} from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_VPC_MANAGEMENT,
    path: 'vpc',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_VPC_MANAGEMENT,
        menu: {
          i18n: 'VPC',
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
    name: MENU_BUSINESS_VPC_DETAILS,
    path: 'vpc/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_VPC_MANAGEMENT,
        menu: {
          i18n: 'VPC详情',
          relative: MENU_BUSINESS_VPC_MANAGEMENT,
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_APPLY_VPC,
    path: 'service/service-apply/vpc',
    component: () => import('@/views/service/service-apply/vpc'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_VPC_MANAGEMENT,

        menu: {
          i18n: '申请VPC',
          relative: MENU_BUSINESS_VPC_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
