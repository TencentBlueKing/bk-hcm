import type { RouteRecordRaw, RouteLocationNormalized } from 'vue-router';
import qs from 'qs';
import Meta from '@/router/meta';
import { getDateRange } from '@/utils/search';
import { AUTH_BIZ_FIND_AUDIT } from '@/constants/auth-symbols';
import {
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
        title: '操作记录',
        activeKey: MENU_BUSINESS_OPERATION_LOG,
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
        activeKey: MENU_BUSINESS_OPERATION_LOG,
        menu: {
          relative: MENU_BUSINESS_OPERATION_LOG,
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
        auth: { view: { type: AUTH_BIZ_FIND_AUDIT } },
      }),
    },
  },
  {
    name: MENU_RESOURCE_OPERATION_LOG_DETAILS,
    path: 'record/details',
    component: () => import('@/views/operation-log/details/flow-task/index'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        title: '操作记录详情',
        activeKey: MENU_RESOURCE_OPERATION_LOG,
        menu: {
          relative: MENU_RESOURCE_OPERATION_LOG,
        },
      }),
    },
  },
];

// 原资源接入「操作记录」页签：accountId 转成搜索条件。entry-rsc 读到地址上的 filter 后不再套默认时间，所以一并写入近 30 天
const toRecordList = (to: RouteLocationNormalized) => {
  const { accountId, ...query } = to.query;
  const id = Array.isArray(accountId) ? accountId[0] : accountId;
  if (id) {
    query.filter = qs.stringify(
      { account_id: id, created_at: getDateRange('last30d') },
      { arrayFormat: 'comma', encode: false },
    );
  }
  return { name: MENU_RESOURCE_OPERATION_LOG, query };
};

const operationLogRscLegacy: RouteRecordRaw[] = [{ path: '/resource/resource/record', redirect: toRecordList }];

export { operationLogBiz, operationLogRsc, operationLogRscLegacy };
