import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
  MENU_BUSINESS_ROUTE_TABLE_DETAILS,
} from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
    path: 'routing',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
        menu: {
          i18n: '路由表',
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
    name: MENU_BUSINESS_ROUTE_TABLE_DETAILS,
    path: 'routing/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
        menu: {
          i18n: '路由表详情',
          relative: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
