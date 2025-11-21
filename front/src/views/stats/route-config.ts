import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_STATS_DELIVERY, MENU_PLATFORM_MANAGEMENT } from '@/constants/menu-symbol';

export default [
  {
    name: MENU_STATS_DELIVERY,
    path: 'stats/delivery',
    component: () => import('./delivery/delivery.vue'),
    meta: {
      ...new Meta({
        title: '交付分析',
        icon: 'hcm-icon bkhcm-icon-stats-menu',
        activeKey: MENU_STATS_DELIVERY,
        checkAuth: 'ziyan_resource_deliver_analyze',
        menu: {
          relative: MENU_PLATFORM_MANAGEMENT,
        },
        layout: {
          breadcrumbs: {
            show: true,
            back: false,
          },
        },
      }),
    },
  },
] as RouteRecordRaw[];
