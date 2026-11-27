import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_BUSINESS, MENU_BUSINESS_EIP_MANAGEMENT, MENU_BUSINESS_EIP_DETAILS } from '@/constants/menu-symbol';
import { normalizeDetailId } from '@/router/utils/normalize-detail-id';

export default [
  {
    name: MENU_BUSINESS_EIP_MANAGEMENT,
    path: 'eip',
    component: () => import('@/views/business/business-manage.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_EIP_MANAGEMENT,
        menu: {
          i18n: '弹性IP',
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_EIP_DETAILS,
    path: 'eip/detail/:id?',
    beforeEnter: normalizeDetailId,
    component: () => import('@/views/business/business-detail.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_EIP_MANAGEMENT,
        menu: {
          i18n: '弹性IP详情',
          relative: MENU_BUSINESS_EIP_MANAGEMENT,
        },
      }),
    },
  },
  // 旧地址 ip，query 原样带过去（详情的 ?id= 由详情路由的 normalizeDetailId 转成路径参数）
  {
    path: 'ip',
    redirect: (to) => ({ name: MENU_BUSINESS_EIP_MANAGEMENT, params: to.params, query: to.query }),
  },
  {
    path: 'ip/detail/:id?',
    redirect: (to) => ({ name: MENU_BUSINESS_EIP_DETAILS, params: to.params, query: to.query }),
  },
] as RouteRecordRaw[];
