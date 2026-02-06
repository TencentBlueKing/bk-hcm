import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import Meta from '@/router/meta';
import {
  MENU_SERVICE_TICKET_DETAILS,
  MENU_BUSINESS_TICKET_DETAILS,
  MENU_SERVICE_TICKET_MANAGEMENT,
  MENU_BUSINESS_TICKET_MANAGEMENT,
  MENU_SERVICE_TICKET_RESOURCE_PLAN_DETAILS,
  MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS,
} from '@/constants/menu-symbol';

const { t } = i18n.global;

export const ticketRoutes: RouteRecordRaw[] = [
  // 兼容老路由
  {
    path: '/service/my-apply',
    redirect: '/service/ticket',
    meta: { ...new Meta({ notMenu: true }) },
  },
  {
    path: '/service/my-apply/detail',
    redirect: '/service/ticket/detail',
    meta: { ...new Meta({ notMenu: true }) },
  },
  {
    path: 'ticket',
    name: MENU_SERVICE_TICKET_MANAGEMENT,
    component: () => import('@/views/ticket/entry-srv.vue'),
    meta: {
      ...new Meta({
        activeKey: MENU_SERVICE_TICKET_MANAGEMENT,
        title: t('单据管理'),
        isShowBreadcrumb: true,
        icon: 'hcm-icon bkhcm-icon-my-apply',
      }),
    },
  },
  {
    path: 'ticket/detail',
    name: MENU_SERVICE_TICKET_DETAILS,
    component: () => import('@/views/ticket/children/apply-detail'),
    meta: {
      ...new Meta({
        activeKey: MENU_SERVICE_TICKET_MANAGEMENT,
        notMenu: true,
      }),
    },
  },
  // 服务请求下 单据管理 tab 资源预测详情
  {
    path: 'ticket/resource-plan/detail',
    name: MENU_SERVICE_TICKET_RESOURCE_PLAN_DETAILS,
    component: () => import('@/views/ticket/children/resource-plan/detail/index.vue'),
    meta: {
      ...new Meta({
        activeKey: MENU_SERVICE_TICKET_MANAGEMENT,
        notMenu: true,
      }),
    },
  },
];

export const ticketRoutesBiz: RouteRecordRaw[] = [
  // 重定向兼容老路由
  {
    path: '/business/applications/detail',
    redirect: '/business/ticket/detail',
    meta: { ...new Meta({ notMenu: true }) },
  },
  {
    path: '/business/applications/resource-plan/detail',
    redirect: '/business/ticket/resource-plan/detail',
    meta: { ...new Meta({ notMenu: true }) },
  },
  {
    path: 'ticket',
    name: MENU_BUSINESS_TICKET_MANAGEMENT,
    component: () => import('@/views/ticket/entry-biz.vue'),
    meta: {
      ...new Meta({
        activeKey: MENU_BUSINESS_TICKET_MANAGEMENT,
        title: t('单据管理'),
        isShowBreadcrumb: true,
        icon: 'hcm-icon bkhcm-icon-my-apply',
      }),
    },
  },
  {
    path: 'ticket/detail',
    name: MENU_BUSINESS_TICKET_DETAILS,
    component: () => import('@/views/ticket/children/apply-detail'),
    meta: {
      ...new Meta({
        activeKey: MENU_BUSINESS_TICKET_MANAGEMENT,
        notMenu: true,
      }),
    },
  },
  // 资源管理下 单据管理 tab 资源预测详情
  {
    path: 'ticket/resource-plan/detail',
    name: MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS,
    component: () => import('@/views/ticket/children/resource-plan/detail/index.vue'),
    meta: {
      ...new Meta({
        activeKey: MENU_BUSINESS_TICKET_MANAGEMENT,
        notMenu: true,
      }),
    },
  },
];

export default ticketRoutes;
