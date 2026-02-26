import type { RouteRecordRaw } from 'vue-router';

import host from '@/views/host/route-config';
import drive from '@/views/drive/route-config';
import image from '@/views/image/route-config';
import vpc from '@/views/vpc/route-config';
import subnet from '@/views/subnet/route-config';
import eip from '@/views/eip/route-config';
import networkInterface from '@/views/network-interface/route-config';
import routeTable from '@/views/route-table/route-config';
import securityGroup from '@/views/security-group/route-config';
import { loadBalancerBiz } from '@/views/load-balancer/route-config';
import cert from '@/views/cert/route-config';
import { operationLogBiz, operationLogRsc } from '@/views/operation-log/route-config';
import task from '@/views/task/route-config';
import { recyclebinBiz, recyclebinRsc } from '@/views/recyclebin/route-config';

import { ticketRoutes } from '@/views/ticket/route-config';
import scheme from '@/views/scheme/route-config';
import accountManage from '@/views/account-manage/route-config';

import resourceManage from '@/views/resource-manage/route-config';
import billRoutes from '@/views/bill/route-config';

import statusNotFound from '@/views/status/404.vue';
import statusError from '@/views/status/error.vue';
import statusBusiness from '@/views/status/business.vue';
import statusPermission from '@/views/status/permission.vue';
import { MENU_BUSINESS } from '@/constants/menu-symbol';

/**
 * 为路由注入状态组件（named views）
 * - notFound: 404 页面（available=false 或功能未开放时）
 * - error: 500 错误页面（路由守卫异常时）
 * - permission: 权限申请页面（业务→business，资源→permission）
 */
const injectStatusComponents = (routes: RouteRecordRaw[]) => {
  routes.forEach((route) => {
    route.components = {
      default: route.component,
      notFound: statusNotFound,
      error: statusError,
    };

    if (route.meta?.owner === MENU_BUSINESS) {
      route.components.permission = statusBusiness;
    } else {
      route.components.permission = statusPermission;
    }

    if (route.children) {
      injectStatusComponents(route.children);
    }
  });

  return routes;
};

export const businessViews = injectStatusComponents([
  ...host,
  ...drive,
  ...image,
  ...vpc,
  ...subnet,
  ...eip,
  ...networkInterface,
  ...routeTable,
  ...securityGroup,
  ...loadBalancerBiz,
  ...cert,
  ...operationLogBiz,
  ...task,
  ...recyclebinBiz,
]);

export const serviceViews = injectStatusComponents([...ticketRoutes, ...scheme, ...accountManage]);

export const resourceViews = injectStatusComponents([
  ...resourceManage,
  ...recyclebinRsc,
  ...operationLogRsc,
  ...billRoutes,
]);

export default {
  businessViews,
  serviceViews,
  resourceViews,
};
