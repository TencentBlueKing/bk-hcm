import { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS_LOAD_BALANCER,
  MENU_BUSINESS_LOAD_BALANCER_DETAILS,
  MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
} from '@/constants/menu-symbol';

const loadBalancerBiz: RouteRecordRaw[] = [
  {
    name: MENU_BUSINESS_LOAD_BALANCER,
    path: '/business/load-balancer',
    component: () => import('@/views/load-balancer/entry-biz.vue'),
    meta: {
      ...new Meta({
        title: '负载均衡',
        activeKey: MENU_BUSINESS_LOAD_BALANCER,
        menu: {},
        icon: 'hcm-icon bkhcm-icon-loadbalancer',
      }),
    },
    children: [
      {
        // 默认展示全部负载均衡（概览）
        name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
        path: '',
        component: () => import('@/views/load-balancer/clb/load-balancer-table.vue'),
        meta: {
          ...new Meta({
            title: '全部负载均衡',
            activeKey: MENU_BUSINESS_LOAD_BALANCER,
            menu: {},
          }),
        },
      },
      {
        name: MENU_BUSINESS_LOAD_BALANCER_DETAILS,
        path: 'details/:id',
        component: () => import('@/views/load-balancer/clb/details.vue'),
        meta: {
          ...new Meta({
            title: '负载均衡详情',
            activeKey: MENU_BUSINESS_LOAD_BALANCER,
            menu: {
              relative: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
            },
          }),
        },
      },
    ],
  },
];

export { loadBalancerBiz };
