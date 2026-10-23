import type { NavigationGuardNext, RouteLocationNormalized } from 'vue-router';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { AUTH_ACCESS_BIZ } from '@/constants/auth-symbols';
import { VIEW_AUTH_BIZ_ACCESS } from '@/constants/view-auth';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { localStorageActions } from '@/common/util';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useAccountStore } from '@/store/account';
import { useAuthStore } from '@/store/auth';

export type BusinessStatus = 'noBiz' | 'bizNotFound' | 'bizUnauthed';

/**
 * 业务路由拦截器（对齐 #1897）
 * 1. 路径上没有业务 ID 时，把旧地址 ?bizs=xxx 转成路径参数。路径上已有业务 ID 时以路径为准
 * 2. 仍然没有时，依次取 localStorage、第一个有权限业务、第一个业务
 * 3. 业务 ID 定下来之后再查：不存在、无权限。没有任何可用业务才是 noBiz
 * 转址带 replace：缺业务 ID 的地址是中间态，不留在历史记录里
 *
 * @returns true 表示可以继续后续处理；false 表示已调用 next（转址，或已切到业务状态页）
 */
export const before = async (
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext,
): Promise<boolean> => {
  const isBusinessRoute = to.matched.some((record) => record.name === MENU_BUSINESS);
  if (!isBusinessRoute) {
    return true;
  }

  const businessGlobalStore = useBusinessGlobalStore();

  // 1. 旧地址兼容。路径已有业务 ID 时不看 ?bizs=
  const queryBizId = to.query[GLOBAL_BIZS_KEY];
  if (queryBizId && !to.params.bizId) {
    const newQuery = { ...to.query };
    delete newQuery[GLOBAL_BIZS_KEY];
    next({
      name: to.name,
      params: { ...to.params, bizId: String(queryBizId) },
      query: newQuery,
      replace: true,
    });
    return false;
  }

  // 2. 路径没有业务 ID 时补默认业务。localStorage 按用户共享，多标签页读到同一份是预期行为
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
        name: to.name,
        params: { ...to.params, bizId: String(bizId) },
        query: to.query,
        replace: true,
      });
      return false;
    }

    // 全量列表、授权列表、localStorage 都没有业务。permissionData 供 noBiz「去申请」（#1897 的 noBiz 不可申请）
    const bizAccess = await useAuthStore().checkViewPermission(VIEW_AUTH_BIZ_ACCESS, { type: AUTH_ACCESS_BIZ });
    to.meta.permissionData = bizAccess.permissionData;
    showBusinessStatus(to, 'noBiz');
    next();
    return false;
  }

  // 3. 按已确定的业务 ID 检查
  const allBizIds = businessGlobalStore.businessFullList.map((biz) => biz.id);
  const authorizedBizIds = businessGlobalStore.businessAuthorizedList.map((biz) => biz.id);

  if (allBizIds.length > 0 && !allBizIds.includes(bizId)) {
    showBusinessStatus(to, 'bizNotFound');
    next();
    return false;
  }

  if (authorizedBizIds.length > 0 && !authorizedBizIds.includes(bizId)) {
    const authSign = { type: AUTH_ACCESS_BIZ, relation: [bizId] };
    const { permissionData } = await useAuthStore().checkViewPermission(Symbol.for(`biz_${bizId}`), authSign);
    to.meta.permissionData = permissionData;
    showBusinessStatus(to, 'bizUnauthed');
    next();
    return false;
  }

  to.meta.view = 'default';
  localStorageActions.set(GLOBAL_BIZS_KEY, bizId);
  useAccountStore().updateBizsId(bizId);

  return true;
};

function showBusinessStatus(to: RouteLocationNormalized, status: BusinessStatus) {
  to.meta.view = 'permission';
  to.meta.extra = { ...((to.meta.extra as object) || {}), businessStatus: status };
}
