import { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS_LOAD_BALANCER,
  MENU_BUSINESS_LOAD_BALANCER_APPLY,
  MENU_BUSINESS_LOAD_BALANCER_DETAILS,
  MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
  MENU_BUSINESS_TARGET_GROUP_DETAILS,
  MENU_BUSINESS_TARGET_GROUP_OVERVIEW,
} from '@/constants/menu-symbol';

const loadBalancerBiz: RouteRecordRaw[] = [
  {
    name: MENU_BUSINESS_LOAD_BALANCER,
    path: '/business/load-balancer',
    redirect: {
      name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
    },
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
        path: 'clb',
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
        path: 'clb/details/:id',
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
      {
        // 默认展示全部目标组（概览）
        name: MENU_BUSINESS_TARGET_GROUP_OVERVIEW,
        path: 'target-group',
        component: () => import('@/views/business/load-balancer/group-view/all-groups-manager/index'),
        meta: {
          ...new Meta({
            title: '全部目标组',
            activeKey: MENU_BUSINESS_LOAD_BALANCER,
            menu: {},
          }),
        },
      },
      {
        name: MENU_BUSINESS_TARGET_GROUP_DETAILS,
        path: 'target-group/details/:id',
        component: () => import('@/views/business/load-balancer/group-view/specific-target-group-manager/index'),
        meta: {
          ...new Meta({
            title: '目标组详情',
            activeKey: MENU_BUSINESS_LOAD_BALANCER,
            menu: {
              relative: MENU_BUSINESS_TARGET_GROUP_OVERVIEW,
            },
          }),
        },
      },
    ],
  },
  {
    name: MENU_BUSINESS_LOAD_BALANCER_APPLY,
    path: '/business/load-balancer/apply',
    component: () => import('@/views/load-balancer/clb/apply/index.vue'),
    meta: {
      ...new Meta({
        title: '购买负载均衡',
        activeKey: MENU_BUSINESS_LOAD_BALANCER,
        menu: {
          relative: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
        },
        notMenu: true,
      }),
    },
  },
];

export { loadBalancerBiz };
