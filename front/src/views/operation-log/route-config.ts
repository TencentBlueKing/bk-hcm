import type { RouteRecordRaw, RouteLocationNormalized } from 'vue-router';
import Meta from '@/router/meta';
import {
  MENU_RESOURCE,
  MENU_BUSINESS_OPERATION_LOG,
  MENU_BUSINESS_OPERATION_LOG_DETAILS,
  MENU_RESOURCE_OPERATION_LOG,
  MENU_RESOURCE_OPERATION_LOG_DETAILS,
} from '@/constants/menu-symbol';

const removeQueryParams = (to: RouteLocationNormalized) => {
  if (Object.keys(to.query).length) return { path: to.path, query: {} };
};

const operationLogBiz: RouteRecordRaw[] = [
  {
    name: MENU_BUSINESS_OPERATION_LOG,
    path: 'record',
    component: () => import('@/views/operation-log/entry-biz.vue'),
    meta: {
      ...new Meta({
        title: '操作记录',
        activeKey: MENU_BUSINESS_OPERATION_LOG,
        isShowBreadcrumb: true,
        menu: {},
        icon: 'hcm-icon bkhcm-icon-operation-record',
      }),
    },
  },
  {
    name: MENU_BUSINESS_OPERATION_LOG_DETAILS,
    path: 'record/details',
    component: () => import('@/views/operation-log/details/flow-task/index'),
    meta: {
      ...new Meta({
        title: '操作记录详情',
        notMenu: true,
        activeKey: MENU_BUSINESS_OPERATION_LOG_DETAILS,
        isShowBreadcrumb: false,
        menu: {
          relative: MENU_BUSINESS_OPERATION_LOG_DETAILS,
        },
      }),
    },
  },
];

const operationLogRsc: RouteRecordRaw[] = [
  {
    name: MENU_RESOURCE_OPERATION_LOG,
    path: 'record',
    component: () => import('@/views/operation-log/entry-rsc.vue'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        title: '操作记录',
        activeKey: MENU_RESOURCE_OPERATION_LOG,
        menu: {},
      }),
    },
    beforeEnter: removeQueryParams,
  },
  {
    name: MENU_RESOURCE_OPERATION_LOG_DETAILS,
    path: 'record/details',
    component: () => import('@/views/operation-log/details/flow-task/index'),
    meta: {
      ...new Meta({
        title: '操作记录详情',
        notMenu: true,
        activeKey: MENU_RESOURCE_OPERATION_LOG,
        isShowBreadcrumb: false,
        menu: {
          relative: MENU_RESOURCE_OPERATION_LOG,
        },
      }),
    },
  },
];

export { operationLogBiz, operationLogRsc };
