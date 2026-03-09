import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_DISK_MANAGEMENT,
  MENU_BUSINESS_DRIVE_DETAILS,
  MENU_BUSINESS_APPLY_DISK,
} from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_DISK_MANAGEMENT,
    path: 'drive',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_DISK_MANAGEMENT,
        menu: {
          i18n: '硬盘',
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
    name: MENU_BUSINESS_DRIVE_DETAILS,
    path: 'drive/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_DISK_MANAGEMENT,
        menu: {
          i18n: '云硬盘详情',
          relative: MENU_BUSINESS_DISK_MANAGEMENT,
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_APPLY_DISK,
    path: 'service/service-apply/disk',
    component: () => import('@/views/service/service-apply/disk'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_DISK_MANAGEMENT,

        menu: {
          i18n: '申请硬盘',
          relative: MENU_BUSINESS_DISK_MANAGEMENT,
        },
      }),
    },
  },
] as RouteRecordRaw[];
