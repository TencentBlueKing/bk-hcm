import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
  MENU_BUSINESS_SECURITY_GROUP_DETAILS,
  MENU_BUSINESS_GCP_DETAILS,
  MENU_BUSINESS_TEMPLATE_DETAILS,
} from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
    path: 'security',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        menu: {
          i18n: '安全组',
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
    name: MENU_BUSINESS_SECURITY_GROUP_DETAILS,
    path: 'security/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        menu: {
          i18n: '安全组详情',
          relative: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_GCP_DETAILS,
    path: 'gcp/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        menu: {
          i18n: 'GCP详情',
          relative: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_TEMPLATE_DETAILS,
    path: 'template/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        menu: {
          i18n: '模板详情',
          relative: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
