import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_BUSINESS_RESOURCE_PLAN_GPU,
  MENU_BUSINESS_RESOURCE_PLAN_GPU_DETAIL,
  MENU_SERVICE_RESOURCE_PLAN_GPU,
  MENU_SERVICE_RESOURCE_PLAN_GPU_DETAIL,
} from '@/constants/menu-symbol';

const gpuDemandBiz: RouteRecordRaw[] = [
  {
    name: MENU_BUSINESS_RESOURCE_PLAN_GPU,
    path: 'resource-plan/gpu',
    component: () => import('./entry-biz.vue'),
    meta: {
      ...new Meta({
        activeKey: 'bizResourcePlan',
        notMenu: true,
        title: '资源预测',
        layout: {
          breadcrumbs: {
            show: true,
          },
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_RESOURCE_PLAN_GPU_DETAIL,
    path: 'resource-plan/gpu/detail',
    component: () => import('./gpu/detail/index.vue'),
    meta: {
      ...new Meta({
        activeKey: 'bizResourcePlan',
        notMenu: true,
      }),
    },
  },
];

const gpuDemandSrv: RouteRecordRaw[] = [
  {
    name: MENU_SERVICE_RESOURCE_PLAN_GPU,
    path: 'resource-plan/gpu',
    component: () => import('./entry-srv.vue'),
    meta: {
      ...new Meta({
        activeKey: 'opResourcePlan',
        notMenu: true,
        title: '资源预测',
        checkAuth: 'ziyan_resource_plan_manage',
        layout: {
          breadcrumbs: {
            show: true,
          },
        },
      }),
    },
  },
  {
    name: MENU_SERVICE_RESOURCE_PLAN_GPU_DETAIL,
    path: 'resource-plan/gpu/detail',
    component: () => import('./gpu/detail/index.vue'),
    meta: {
      ...new Meta({
        activeKey: 'opResourcePlan',
        notMenu: true,
        checkAuth: 'ziyan_resource_plan_manage',
      }),
    },
  },
];

export { gpuDemandBiz, gpuDemandSrv };
