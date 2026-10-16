import {
  createRouter,
  RouteRecordRaw,
  RouteRecordName,
  NavigationGuardNext,
  createWebHashHistory,
  RouteLocationNormalized,
} from 'vue-router';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_RESOURCE,
  MENU_RESOURCE_MANAGE,
  MENU_SCHEME,
  MENU_SERVICE,
  MENU_SERVICE_TICKET_MANAGEMENT,
} from '@/constants/menu-symbol';
import { businessViews, resourceViews, schemeViews, serviceViews } from '@/views';
import { resourceManageLegacy, resourceDetailLegacy } from '@/views/resource-manage/route-config';
import { recyclebinRscLegacy } from '@/views/recyclebin/route-config';
import { accountManageLegacy } from '@/views/account-manage/route-config';
import { operationLogRscLegacy } from '@/views/operation-log/route-config';
import { billLegacy } from '@/views/bill/route-config';
import { before as businessBeforeInterceptor } from './business-interceptor';
import { useAccountStore } from '@/store';
import { useAuthStore } from '@/store/auth';

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
  {
    name: MENU_SCHEME,
    path: '/scheme',
    redirect: { name: 'scheme-recommendation' },
    children: schemeViews,
  },
  ...resourceManageLegacy,
  ...resourceDetailLegacy,
  ...recyclebinRscLegacy,
  ...operationLogRscLegacy,
  ...accountManageLegacy,
  ...billLegacy,
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

router.beforeEach(async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  try {
    const authStore = useAuthStore();

    // 1. 业务路由拦截器（业务 ID 兼容与默认值、业务访问权限检查）。false 表示拦截器已调用 next
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

// accountStore.bizs 是业务 ID 的镜像（业务页由业务拦截器写入）；资源运营、资源选型与无归属页面清零，旧代码按它为空判断「非业务」
router.afterEach((to) => {
  const topMenuId = to.matched[0]?.name;
  if (!topMenuId || topMenuId === MENU_RESOURCE || topMenuId === MENU_SCHEME) {
    useAccountStore().updateBizsId(0);
  }
});

export default router;
