import type { Component } from 'vue';
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
import { ticketRoutesBiz } from '@/views/ticket/route-config';
import { operationLogBiz, operationLogRsc } from '@/views/operation-log/route-config';
import task from '@/views/task/route-config';
import cloudAccountManage from '@/views/cloud-account-manage/route-config';
import service from '@/router/module/service';
import scheme from '@/views/scheme/route-config';
import resourceManage from '@/views/resource-manage/route-config';
import { recyclebinBiz, recyclebinRsc } from '@/views/recyclebin/route-config';
import { accountManageRsc } from '@/views/account-manage/route-config';
import { billAccountManage, billManage } from '@/views/bill/route-config';

import statusNotFound from '@/views/status/404.vue';
import statusError from '@/views/status/error.vue';
import statusBusiness from '@/views/status/business.vue';
import statusPermission from '@/views/status/permission.vue';

/**
 * 为路由注入状态组件（named views），由守卫按 meta.view 选择渲染
 * - notFound: 404 页面（available=false 时）
 * - error: 错误页面（路由守卫异常时）
 * - permission: 无权限页面（业务资源用业务无权限页，其余用视图无权限页）
 * 没有 component 的记录只递归不注入：挂上 components 后 RouterView 不再跳过它，会整页白屏
 */
export const injectStatusComponents = (routes: RouteRecordRaw[], permission: Component = statusPermission) => {
  routes.forEach((route) => {
    if (route.component) {
      Object.assign(route, {
        components: {
          default: route.component,
          notFound: statusNotFound,
          error: statusError,
          permission,
        },
      });
    }

    if (route.children) {
      injectStatusComponents(route.children, permission);
    }
  });

  return routes;
};

export const businessViews = injectStatusComponents(
  [
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
    ...ticketRoutesBiz,
    ...operationLogBiz,
    ...task,
    ...cloudAccountManage,
    ...recyclebinBiz,
  ],
  statusBusiness,
);
export const serviceViews = injectStatusComponents([
  ...service.flatMap((group) => group.children),
  ...scheme,
  ...accountManageRsc,
]);
export const resourceViews = injectStatusComponents([
  ...resourceManage,
  ...recyclebinRsc,
  ...operationLogRsc,
  ...billAccountManage,
  ...billManage,
]);
