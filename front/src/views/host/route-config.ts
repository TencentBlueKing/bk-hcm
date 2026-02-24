import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_BUSINESS_HOST_DETAILS,
  MENU_BUSINESS_APPLY_CVM,
} from '@/constants/menu-symbol';
import { AUTH_ACCESS_BIZ } from '@/constants/auth-symbols';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_HOST_MANAGEMENT,
    path: 'host',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_HOST_MANAGEMENT,
        auth: {
          view: (to) => ({ type: AUTH_ACCESS_BIZ, relation: [Number(to.params.bizId)] }),
        },
        menu: {
          i18n: '主机',
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
    name: MENU_BUSINESS_HOST_DETAILS,
    path: 'host/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_HOST_MANAGEMENT,
        menu: {
          i18n: '主机详情',
          relative: MENU_BUSINESS_HOST_MANAGEMENT,
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_APPLY_CVM,
    path: 'service/service-apply/cvm',
    component: () => import('@/views/service/service-apply/cvm'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_HOST_MANAGEMENT,

        menu: {
          i18n: '申请主机',
          relative: MENU_BUSINESS_HOST_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
