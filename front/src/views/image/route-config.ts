import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS, MENU_BUSINESS_IMAGE_MANAGEMENT, MENU_BUSINESS_IMAGE_DETAILS } from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_IMAGE_MANAGEMENT,
    path: 'image',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_IMAGE_MANAGEMENT,
        menu: {
          i18n: '镜像',
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
    name: MENU_BUSINESS_IMAGE_DETAILS,
    path: 'image/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_IMAGE_MANAGEMENT,
        menu: {
          i18n: '镜像详情',
          relative: MENU_BUSINESS_IMAGE_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
