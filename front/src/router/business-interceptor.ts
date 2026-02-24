import type { RouteLocationNormalized, NavigationGuardNext } from 'vue-router';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { localStorageActions } from '@/common/util';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useAccountStore } from '@/store/account';

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
    // 重定向到路径参数形式
    const newQuery = { ...to.query };
    delete newQuery[GLOBAL_BIZS_KEY];
    next({
      name: to.name as any,
      params: { ...to.params, bizId: String(queryBizId) },
      query: newQuery,
      replace: true,
    });
    return false; // 已拦截
  }

  // 2. 获取当前 bizId
  let bizId = Number(to.params.bizId);

  // 如果没有 bizId，尝试获取默认值
  if (!bizId) {
    // 优先从 localStorage 获取
    const storedBizId = localStorageActions.get(GLOBAL_BIZS_KEY, (value) => value);
    if (storedBizId) {
      bizId = Number(storedBizId);
    }

    // 如果 localStorage 没有，从有权限的业务列表获取第一个
    if (!bizId && businessGlobalStore.businessAuthorizedList.length > 0) {
      bizId = businessGlobalStore.businessAuthorizedList[0].id;
    }

    // 如果有权限的业务列表为空，从全部业务列表获取第一个
    if (!bizId && businessGlobalStore.businessFullList.length > 0) {
      bizId = businessGlobalStore.businessFullList[0].id;
    }

    // 如果找到了 bizId，重定向到带 bizId 的路由
    if (bizId) {
      next({
        name: to.name as any,
        params: { ...to.params, bizId: String(bizId) },
        query: to.query,
        replace: true,
      });
      return false; // 已拦截
    }

    // 如果仍然没有 bizId，显示业务状态页面
    to.meta.view = 'permission';
    next();
    return false; // 已拦截
  }

  // 3. 检查业务权限：当前 bizId 是否在有权限的业务列表中
  const authorizedBizIds = businessGlobalStore.businessAuthorizedList.map((biz) => biz.id);
  if (authorizedBizIds.length > 0 && !authorizedBizIds.includes(bizId)) {
    // 当前业务不在有权限列表中，显示无权限视图
    to.meta.view = 'permission';
    next();
    return false; // 已拦截
  }

  // 4. 有权限时确保 view 恢复为 default
  to.meta.view = 'default';

  // 5. 保存 bizId 到 localStorage 并更新 accountStore
  localStorageActions.set(GLOBAL_BIZS_KEY, bizId);
  // TODO: [待重构] accountStore.bizs 是老的设计，应该使用 route.params.bizId 获取当前业务 ID
  // 后续需要将所有依赖 accountStore.bizs 的组件改为使用 useWhereAmI().getBizsId() 或 route.params.bizId
  const accountStore = useAccountStore();
  accountStore.updateBizsId(bizId);

  // 可以继续后续处理
  return true;
};
