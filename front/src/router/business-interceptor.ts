import type { RouteLocationNormalized, NavigationGuardNext } from 'vue-router';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { AUTH_ACCESS_BIZ } from '@/constants/auth-symbols';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { localStorageActions } from '@/common/util';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useAccountStore } from '@/store/account';
import { useAuthStore } from '@/store/auth';

export type BusinessStatus = 'noBiz' | 'bizNotFound' | 'bizUnauthed';

/**
 * 业务路由拦截器
 * 负责：
 * 1. 向后兼容查询参数 ?bizs=xxx 到路径参数 /business/:bizId
 * 2. 如果没有 bizId，尝试从 localStorage 或有权限的业务中获取默认值
 * 3. 检查业务权限（当前 bizId 是否在有权限的业务列表中）
 *
 * @returns true 表示可以继续后续处理，false 表示已拦截（已调用 next）
 */
export const before = async (
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext,
): Promise<boolean> => {
  // 非业务路由，可以继续
  const isBusinessRoute = to.matched.some((record) => record.name === MENU_BUSINESS);
  if (!isBusinessRoute) {
    return true;
  }

  const businessGlobalStore = useBusinessGlobalStore();

  // 1. 向后兼容：处理查询参数 ?bizs=xxx
  const queryBizId = to.query[GLOBAL_BIZS_KEY];
  if (queryBizId && !to.params.bizId) {
    const newQuery = { ...to.query };
    delete newQuery[GLOBAL_BIZS_KEY];
    next({
      name: to.name as any,
      params: { ...to.params, bizId: String(queryBizId) },
      query: newQuery,
      replace: true,
    });
    return false;
  }

  // 2. 获取当前 bizId
  let bizId = Number(to.params.bizId);

  if (!bizId) {
    const storedBizId = localStorageActions.get(GLOBAL_BIZS_KEY, (value) => value);
    if (storedBizId) {
      bizId = Number(storedBizId);
    }

    if (!bizId && businessGlobalStore.businessAuthorizedList.length > 0) {
      bizId = businessGlobalStore.businessAuthorizedList[0].id;
    }

    if (!bizId && businessGlobalStore.businessFullList.length > 0) {
      bizId = businessGlobalStore.businessFullList[0].id;
    }

    if (bizId) {
      next({
        name: to.name as any,
        params: { ...to.params, bizId: String(bizId) },
        query: to.query,
        replace: true,
      });
      return false;
    }

    // 所有业务列表为空，没有任何可用业务
    showBusinessStatus(to, 'noBiz');
    next();
    return false;
  }

  // 3. 检查业务权限
  const allBizIds = businessGlobalStore.businessFullList.map((biz) => biz.id);
  const authorizedBizIds = businessGlobalStore.businessAuthorizedList.map((biz) => biz.id);

  if (allBizIds.length > 0 && !allBizIds.includes(bizId)) {
    // bizId 不在全量列表中 → 业务不存在
    showBusinessStatus(to, 'bizNotFound');
    next();
    return false;
  }

  if (authorizedBizIds.length > 0 && !authorizedBizIds.includes(bizId)) {
    // bizId 存在但不在授权列表中 → 无权限
    const authStore = useAuthStore();
    const authSign = { type: AUTH_ACCESS_BIZ, relation: [bizId] };
    const { permissionData } = await authStore.checkViewPermission(Symbol.for(`biz_${bizId}`), authSign);
    to.meta.permissionData = permissionData;
    showBusinessStatus(to, 'bizUnauthed');
    next();
    return false;
  }

  // 4. 有权限时确保 view 恢复为 default
  to.meta.view = 'default';

  // 5. 保存 bizId 到 localStorage 并更新 accountStore
  localStorageActions.set(GLOBAL_BIZS_KEY, bizId);
  // TODO: [待重构] accountStore.bizs 是老的设计，应该使用 route.params.bizId 获取当前业务 ID
  const accountStore = useAccountStore();
  accountStore.updateBizsId(bizId);

  return true;
};

function showBusinessStatus(to: RouteLocationNormalized, status: BusinessStatus) {
  to.meta.view = 'permission';
  to.meta.extra = { ...((to.meta.extra as object) || {}), businessStatus: status };
}
