import {
  createRouter,
  RouteRecordRaw,
  RouteRecordName,
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
import {
  billViews,
  businessViews,
  injectStatusComponents,
  resourceAccessViews,
  schemeViews,
  serviceViews,
} from '@/views';
import common from './module/common';
import { before as businessBeforeInterceptor } from './business-interceptor';
import { useCommonStore } from '@/store';
import { useAuthStore } from '@/store/auth';
import { useVerify } from '@/hooks';

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
  ...injectStatusComponents(statusRouters),
  ...injectStatusComponents(common),
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

let isFirstNavigation = true;

router.beforeEach(async (to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const commonStore = useCommonStore();
  const isFirst = isFirstNavigation;
  isFirstNavigation = false;

  // 旧按钮鉴权数据预取：只供页面按钮级的 useVerify().authVerifyData 使用，不参与路由拦截，失败不阻断导航
  if (isFirst) {
    try {
      await useVerify().getAuthVerifyData(commonStore.pageAuthData);
    } catch (err) {
      console.error('[router] getAuthVerifyData error:', err);
    }
  }

  // 旧 403 页（/403/:id）仅保留可打开：首次进入且有业务访问权限时回首页，与改造前一致
  if (to.name === '403') {
    next(isFirst && commonStore.authVerifyData?.permissionAction?.biz_access ? { path: '/' } : undefined);
    return;
  }

  try {
    const authStore = useAuthStore();

    // 1. 业务路由拦截器（业务访问权限检查）
    const canContinue = await businessBeforeInterceptor(to);
    if (!canContinue) {
      next();
      return;
    }

    // 2. 默认恢复为 default 视图
    to.meta.view = 'default';
    to.meta.errorMessage = '';

    // 3. 检查 available（功能是否开放）
    const { available } = to.meta;
    const isAvailable = typeof available === 'function' ? available(to) : available;
    if (isAvailable === false) {
      to.meta.view = 'notFound';
      next();
      return;
    }

    // 4. 检查视图权限：路由 name 原样作缓存键（Symbol 不能转字符串）
    const viewAuthConfig = to.meta?.auth?.view;
    if (viewAuthConfig) {
      const authSign = typeof viewAuthConfig === 'function' ? viewAuthConfig(to) : viewAuthConfig;
      const { authorized, permissionData } = await authStore.checkViewPermission(to.name as RouteRecordName, authSign);

      if (!authorized) {
        to.meta.view = 'permission';
        to.meta.permissionData = permissionData;
      }
    }

    next();
  } catch (err) {
    console.error('[router] beforeEach error:', err);
    to.meta.view = 'error';
    to.meta.errorMessage = err instanceof Error ? err.message : String(err);
    next();
  }
});

export default router;
