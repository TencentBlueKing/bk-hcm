import type { RouteRecordRaw, RouteLocationNormalized } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_BIZ_FIND_AUDIT } from '@/constants/auth-symbols';
import {
  MENU_BUSINESS,
  MENU_RESOURCE,
  MENU_BUSINESS_OPERATION_LOG,
  MENU_BUSINESS_OPERATION_LOG_DETAILS,
  MENU_RESOURCE_OPERATION_LOG,
  MENU_RESOURCE_OPERATION_LOG_DETAILS,
} from '@/constants/menu-symbol';

const _removeQueryParams = (to: RouteLocationNormalized) => {
  if (Object.keys(to.query).length) return { path: to.path, query: {} };
};

const operationLogBiz: RouteRecordRaw[] = [
  {
    name: MENU_BUSINESS_OPERATION_LOG,
    path: 'record',
    component: () => import('@/views/operation-log/entry-biz.vue'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        activeKey: MENU_BUSINESS_OPERATION_LOG,
        menu: {
          i18n: '操作记录',
        },
      }),
    },
  },
  {
    name: MENU_BUSINESS_OPERATION_LOG_DETAILS,
    path: 'record/details',
    component: () => import('@/views/operation-log/details/flow-task/index'),
    meta: {
      ...new Meta({
        owner: MENU_BUSINESS,
        menu: {
          i18n: '操作记录详情',
          relative: MENU_BUSINESS_OPERATION_LOG,
        },
        activeKey: MENU_BUSINESS_OPERATION_LOG,
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
        activeKey: MENU_RESOURCE_OPERATION_LOG,
        auth: {
          view: { type: AUTH_BIZ_FIND_AUDIT },
        },
        menu: {
          i18n: '操作记录',
        },
      }),
    },
    // beforeEnter: removeQueryParams,
  },
  {
    name: MENU_RESOURCE_OPERATION_LOG_DETAILS,
    path: 'record/details',
    component: () => import('@/views/operation-log/details/flow-task/index'),
    meta: {
      ...new Meta({
        activeKey: MENU_RESOURCE_OPERATION_LOG,
        menu: {
          i18n: '操作记录详情',
          relative: MENU_RESOURCE_OPERATION_LOG,
        },
      }),
    },
  },
];

export { operationLogBiz, operationLogRsc };
