import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import Meta from '@/router/meta';
import { MENU_SERVICE_APPLY_MANAGEMENT_DETAILS, MENU_SERVICE_APPLY_MANAGEMENT } from '@/constants/menu-symbol';

const { t } = i18n.global;

export const ticketRoutes: RouteRecordRaw[] = [
  // 兼容老路由
  {
    path: '/service/my-apply',
    redirect: '/service/ticket',
    meta: { ...new Meta({}) },
  },
  {
    path: '/service/my-apply/detail',
    redirect: '/service/ticket/detail',
    meta: { ...new Meta({}) },
  },
  {
    path: 'ticket',
    name: MENU_SERVICE_APPLY_MANAGEMENT,
    component: () => import('@/views/ticket/entry-srv.vue'),
    meta: {
      ...new Meta({
        activeKey: MENU_SERVICE_APPLY_MANAGEMENT,
        menu: {
          i18n: t('单据管理'),
        },
      }),
    },
  },
  {
    path: 'ticket/detail',
    name: MENU_SERVICE_APPLY_MANAGEMENT_DETAILS,
    component: () => import('@/views/ticket/children/apply-detail'),
    meta: {
      ...new Meta({
        activeKey: MENU_SERVICE_APPLY_MANAGEMENT,
        menu: {
          i18n: t('单据详情'),
        },
      }),
    },
  },
];
