import {
  createRouter,
  RouteRecordRaw,
  NavigationGuardNext,
  createWebHashHistory,
  RouteLocationNormalized,
} from 'vue-router';
import {
  MENU_BILL,
  MENU_BUSINESS,
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_RESOURCE,
  MENU_RESOURCE_RESOURCE_MANAGEMENT,
  MENU_SCHEME,
  MENU_SERVICE,
  MENU_SERVICE_TICKET_MANAGEMENT,
} from '@/constants/menu-symbol';
import { billViews, businessViews, resourceAccessViews, schemeViews, serviceViews } from '@/views';
import common from './module/common';
import { useCommonStore } from '@/store';
import { useVerify } from '@/hooks';
import { isArray, isRegExp, isString } from 'lodash';

// 状态页面路由（不属于任何一级菜单，不显示侧栏和面包屑）
const statusRouters: RouteRecordRaw[] = [
  {
    name: '404',
    path: '/404',
    component: () => import('@/views/status/404.vue'),
    meta: { layout: { breadcrumb: { show: false } } },
  },
  {
    name: 'error',
    path: '/error',
    component: () => import('@/views/status/error.vue'),
    meta: { layout: { breadcrumb: { show: false } } },
  },
];

// 顶层记录的 name 即一级菜单 id，顶栏高亮与侧栏按 route.matched[0].name 判定
const routes: RouteRecordRaw[] = [
  ...statusRouters,
  ...common,
  {
    name: MENU_BUSINESS,
    path: '/business',
    redirect: { name: MENU_BUSINESS_HOST_MANAGEMENT },
    children: businessViews,
  },
  {
    name: MENU_SERVICE,
    path: '/service',
    redirect: { name: MENU_SERVICE_TICKET_MANAGEMENT },
    children: serviceViews,
  },
  // TODO(菜单整合二期·组装)：资源接入迁入资源运营后删除
  {
    name: MENU_RESOURCE,
    path: '/resource',
    redirect: { name: MENU_RESOURCE_RESOURCE_MANAGEMENT },
    children: resourceAccessViews,
  },
  {
    name: MENU_SCHEME,
    path: '/scheme',
    redirect: { name: 'scheme-recommendation' },
    children: schemeViews,
  },
  // TODO(菜单整合二期·组装)：资源运营改为最终结构后删除
  {
    name: MENU_BILL,
    path: '/bill',
    redirect: { path: '/bill/account-manage' },
    children: billViews,
  },
  {
    path: '/403',
    redirect: '/403',
  },
  {
    path: '/',
    redirect: { name: MENU_BUSINESS },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: () => ({ name: '404', query: {} }),
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

// 进入目标页面
const toCurrentPage = (
  authVerifyData: {
    permissionAction: Record<string, boolean>;
    urlParams: {
      system_id: string;
      actions: Array<{
        id: string;
        name: string;
        related_resource_types: Array<any>;
      }>;
    };
  },
  currentFindAuthData: {
    action: string;
    id: string;
    path: string;
    type: string;
  },
  next: NavigationGuardNext,
  to?: RouteLocationNormalized,
) => {
  // 是否需要鉴权
  const needAuth = !!currentFindAuthData?.id;
  // 是否有权限
  const hasAuth = !!authVerifyData?.permissionAction?.[currentFindAuthData?.id];

  if (!needAuth) {
    if (to?.name === '403') next(!!authVerifyData?.permissionAction?.biz_access ? { path: '/' } : undefined);
    else next();
    return;
  }

  if (hasAuth) next();
  else next({ name: '403', params: { id: currentFindAuthData?.id } });
};

router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const commonStore = useCommonStore();
  const { pageAuthData, authVerifyData } = commonStore; // 所有需要检验的查看权限数据
  const currentFindAuthData = pageAuthData.find((e: any) => {
    const { path } = e;
    if (isString(path)) return path === to.path;
    if (isArray(path)) return path.includes(to.path);
    if (isRegExp(path)) return path.test(to.path);
    return undefined;
  });

  // if (to.path === '/service/my-approval') {
  //   window.open(`${BK_ITSM_URL}/#/workbench/ticket/approval`);
  //   window.location.reload();
  // }
  if (from.path === '/') {
    // 刷新或者首次进入请求权限接口
    const { getAuthVerifyData } = useVerify(); // 权限中心权限
    getAuthVerifyData(pageAuthData).then(() => {
      const { authVerifyData } = commonStore;
      toCurrentPage(authVerifyData, currentFindAuthData as any, next, to);
    });
  } else if (['/scheme/recommendation', '/scheme/deployment/list'].includes(to.path)) {
    next();
  } else {
    toCurrentPage(authVerifyData, currentFindAuthData as any, next);
  }
});

export default router;
