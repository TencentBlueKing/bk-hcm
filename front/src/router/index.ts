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
  MENU_SERVICE_TICKET_MANAGEMENT,
} from '@/constants/menu-symbol';
import { businessViews, serviceViews, resourceViews } from '@/views';
import { useAuthStore } from '@/store/auth';
import { before as businessBeforeInterceptor } from './business-interceptor';

// 状态页面路由（独立页面，不显示侧边栏和面包屑）
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

// 旧路由兼容重定向
const legacyRedirects: RouteRecordRaw[] = [
  // 旧: /resource/detail/:resourceType?id=xxx&type=yyy → 新: /resource/manage/:resourceType/details/:id?type=yyy
  {
    path: '/resource/detail/:resourceType',
    redirect: (to) => {
      const { id, type } = to.query;
      return {
        path: `/resource/manage/${to.params.resourceType}/details/${id}`,
        query: type ? { type: type as string } : {},
      };
    },
  },
  // 旧: /resource/account/detail/?accountId=xxx → 新: /service/account/details/:accountId
  {
    path: '/resource/account/detail',
    redirect: (to) => ({
      path: `/service/account/details/${to.query.accountId || to.query.id}`,
    }),
  },
];

// 重定向路由
const redirectRouters: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: { name: MENU_BUSINESS },
  },
  ...legacyRedirects,
  // catch-all 放在最后
  {
    path: '/:pathMatch(.*)*',
    redirect: () => ({ name: '404', query: {} }),
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
    redirect: { name: MENU_SERVICE_TICKET_MANAGEMENT },
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
  try {
    const authStore = useAuthStore();

    // 1. 业务路由拦截器（bizId 兼容性、默认值、业务权限检查）
    const canContinue = await businessBeforeInterceptor(to, from, next);
    if (!canContinue) return;

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

    // 4. 检查视图权限
    const viewAuthConfig = to.meta?.auth?.view;
    if (viewAuthConfig) {
      const viewId = to.name as symbol;
      const authSign = typeof viewAuthConfig === 'function' ? viewAuthConfig(to) : viewAuthConfig;
      const { authorized, permissionData } = await authStore.checkViewPermission(viewId, authSign);

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
