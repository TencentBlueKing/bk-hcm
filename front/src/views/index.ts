import type { Component } from 'vue';
import type { RouteRecordRaw } from 'vue-router';
import business from '@/router/module/business';
import service from '@/router/module/service';
import serviceInside from '@/router/module/service-inside';
import resource from '@/router/module/resource';
import resourceInside from '@/router/module/resource-inside';
import scheme from '@/router/module/scheme';
import { billAccountManage, billAccountManageLegacy, billManage } from '@/views/bill/route-config';

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
  business.flatMap((group) => group.children),
  statusBusiness,
);
export const serviceViews = injectStatusComponents([...service.flatMap((group) => group.children), ...serviceInside]);
// TODO(菜单整合二期·组装)：资源接入迁入资源运营后删除
export const resourceAccessViews = injectStatusComponents([...resource, ...resourceInside]);
export const schemeViews = injectStatusComponents(scheme);
export const billViews = injectStatusComponents([...billAccountManage, ...billManage, ...billAccountManageLegacy]);
