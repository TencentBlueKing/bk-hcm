import { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_LOAD_BALANCER,
  MENU_BUSINESS_APPLY_CLB,
  MENU_BUSINESS_LOAD_BALANCER_DETAILS,
  MENU_BUSINESS_LOAD_BALANCER_LB_VIEW,
  MENU_BUSINESS_TARGET_GROUP_DETAILS,
  MENU_BUSINESS_LOAD_BALANCER_TG_VIEW,
  MENU_BUSINESS_LOAD_BALANCE_DEVICE_SEARCH,
} from '@/constants/menu-symbol';

const loadBalancerBiz: RouteRecordRaw[] = [
  {
    name: MENU_BUSINESS_LOAD_BALANCER,
    path: 'load-balancer',
    redirect: { name: MENU_BUSINESS_LOAD_BALANCER_LB_VIEW },
    component: () => import('@/views/load-balancer/entry-biz.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_LOAD_BALANCER,
        menu: {
          i18n: '负载均衡',
        },
      }),
    },
    children: [
      {
        path: 'resource',
        redirect: { name: MENU_BUSINESS_LOAD_BALANCER_LB_VIEW },
        component: () => import('@/views/load-balancer/clb/index.vue'),
        children: [
          {
            name: MENU_BUSINESS_LOAD_BALANCER_LB_VIEW,
            path: 'clb',
            component: () => import('@/views/load-balancer/clb/load-balancer-table.vue'),
            meta: {
              ...new Meta({
                owner: MENU_BUSINESS,
                activeKey: MENU_BUSINESS_LOAD_BALANCER,
                layout: { breadcrumb: { show: false } },
                menu: {
                  i18n: '全部负载均衡',
                  relative: MENU_BUSINESS_LOAD_BALANCER,
                },
                extra: { type: 'clb' },
              }),
            },
          },
          {
            name: MENU_BUSINESS_LOAD_BALANCER_DETAILS,
            path: 'clb/details/:id',
            component: () => import('@/views/load-balancer/clb/details.vue'),
            meta: {
              ...new Meta({
                owner: MENU_BUSINESS,
                activeKey: MENU_BUSINESS_LOAD_BALANCER,
                layout: { breadcrumb: { show: false } },
                menu: {
                  i18n: '负载均衡详情',
                  relative: MENU_BUSINESS_LOAD_BALANCER_LB_VIEW,
                },
                extra: { type: 'clb' },
              }),
            },
          },
          {
            name: MENU_BUSINESS_LOAD_BALANCER_TG_VIEW,
            path: 'target-group',
            component: () => import('@/views/business/load-balancer/group-view/all-groups-manager/index'),
            meta: {
              ...new Meta({
                owner: MENU_BUSINESS,
                activeKey: MENU_BUSINESS_LOAD_BALANCER,
                layout: { breadcrumb: { show: false } },
                menu: {
                  i18n: '全部目标组',
                  relative: MENU_BUSINESS_LOAD_BALANCER,
                },
                extra: { type: 'target_group' },
              }),
            },
          },
          {
            name: MENU_BUSINESS_TARGET_GROUP_DETAILS,
            path: 'target-group/details/:id',
            component: () => import('@/views/business/load-balancer/group-view/specific-target-group-manager/index'),
            meta: {
              ...new Meta({
                owner: MENU_BUSINESS,
                activeKey: MENU_BUSINESS_LOAD_BALANCER,
                layout: { breadcrumb: { show: false } },
                menu: {
                  i18n: '目标组详情',
                  relative: MENU_BUSINESS_LOAD_BALANCER_TG_VIEW,
                },
                extra: { type: 'target_group' },
              }),
            },
          },
        ],
      },
      {
        name: MENU_BUSINESS_LOAD_BALANCE_DEVICE_SEARCH,
        path: 'device',
        component: () => import('@/views/load-balancer/device/index.vue'),
        meta: {
          ...new Meta({
            owner: MENU_BUSINESS,
            activeKey: MENU_BUSINESS_LOAD_BALANCER,
            layout: { breadcrumb: { show: false } },
            menu: {
              i18n: '配置检索',
              relative: MENU_BUSINESS_LOAD_BALANCER,
            },
          }),
        },
      },
    ],
  },
  {
    name: MENU_BUSINESS_APPLY_CLB,
    path: 'load-balancer/apply',
    component: () => import('@/views/load-balancer/clb/apply/index.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_LOAD_BALANCER,
        menu: {
          i18n: '购买负载均衡',
          relative: MENU_BUSINESS_LOAD_BALANCER_LB_VIEW,
        },
      }),
    },
  },
];

export { loadBalancerBiz };
