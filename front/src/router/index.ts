import {
  createRouter,
  RouteRecordRaw,
  NavigationGuardNext,
  createWebHashHistory,
  RouteLocationNormalized,
} from 'vue-router';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_SERVICE,
  MENU_RESOURCE,
  MENU_RESOURCE_MANAGE,
} from '@/constants/menu-symbol';
import { businessViews, serviceViews, resourceViews } from '@/views';
import { useAuthStore } from '@/store/auth';
import { before as businessBeforeInterceptor } from './business-interceptor';

// 状态页面路由
const statusRouters: RouteRecordRaw[] = [
  {
    name: '404',
    path: '/404',
    component: () => import('@/views/status/404.vue'),
  },
  {
    name: 'error',
    path: '/error',
    component: () => import('@/views/status/error.vue'),
  },
];

// 重定向路由
const redirectRouters: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: { name: MENU_BUSINESS },
  },
  // catch-all 放在最后
  {
    path: '/:pathMatch(.*)*',
    redirect: { name: '404' },
  },
];

const routes: RouteRecordRaw[] = [
  ...statusRouters,
  {
    name: MENU_BUSINESS,
    path: '/business/:bizId(\\d+)?',
    redirect: { name: MENU_BUSINESS_HOST_MANAGEMENT },
    children: businessViews,
  },
  {
    name: MENU_SERVICE,
    path: '/service',
    children: serviceViews,
  },
  {
    name: MENU_RESOURCE,
    path: '/resource',
    redirect: { name: MENU_RESOURCE_MANAGE },
    children: resourceViews,
  },
  ...redirectRouters,
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

router.beforeEach(async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const authStore = useAuthStore();

  // 1. 业务路由拦截器（bizId 兼容性、默认值、业务权限检查）
  const canContinue = await businessBeforeInterceptor(to, from, next);
  if (!canContinue) {
    // 已拦截（已调用 next），直接返回
    return;
  }

  // 2. 默认恢复为 default 视图
  to.meta.view = 'default';

  // 3. 检查视图权限
  // - 如果已在 preload 中预获取，直接使用缓存
  // - 否则根据 meta.auth.view 动态请求权限
  const viewAuthConfig = to.meta?.auth?.view;
  if (viewAuthConfig) {
    const viewId = to.name as symbol;
    // 支持函数形式的权限配置（用于动态 relation，如 bizId）
    const authSign = typeof viewAuthConfig === 'function' ? viewAuthConfig(to) : viewAuthConfig;
    const { authorized, permissionData } = await authStore.checkViewPermission(viewId, authSign);

    if (!authorized) {
      to.meta.view = 'permission';
      to.meta.permissionData = permissionData;
    }
  }

  next();
});

export default router;
