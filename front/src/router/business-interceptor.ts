import type { RouteLocationNormalized } from 'vue-router';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { AUTH_ACCESS_BIZ } from '@/constants/auth-symbols';
import { VIEW_AUTH_BIZ_ACCESS } from '@/constants/view-auth';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useAuthStore } from '@/store/auth';

export type BusinessStatus = 'noBiz' | 'bizNotFound' | 'bizUnauthed';

/**
 * 业务路由拦截器：检查业务访问权限，命中时把目标路由切到业务无权限页
 * 当前业务取自 ?bizs=，缺省时放行，由业务选择器挂载后补参数再进一次守卫；地址与业务 ID 存储都不在这里改
 *
 * @returns true 表示可以继续后续处理，false 表示已切到业务无权限页
 */
export const before = async (to: RouteLocationNormalized): Promise<boolean> => {
  // 非业务路由，可以继续
  const isBusinessRoute = to.matched.some((record) => record.name === MENU_BUSINESS);
  if (!isBusinessRoute) {
    return true;
  }

  const authStore = useAuthStore();
  const businessGlobalStore = useBusinessGlobalStore();

  // 1. 没有任何业务的访问权限
  const bizAccess = await authStore.checkViewPermission(VIEW_AUTH_BIZ_ACCESS, { type: AUTH_ACCESS_BIZ });
  if (!bizAccess.authorized) {
    to.meta.permissionData = bizAccess.permissionData;
    showBusinessStatus(to, 'noBiz');
    return false;
  }

  const bizId = Number(to.query[GLOBAL_BIZS_KEY]);
  if (!bizId) {
    return true;
  }

  // 2. 业务不存在
  const allBizIds = businessGlobalStore.businessFullList.map((biz) => biz.id);
  if (allBizIds.length > 0 && !allBizIds.includes(bizId)) {
    showBusinessStatus(to, 'bizNotFound');
    return false;
  }

  // 3. 当前业务无权限
  const authorizedBizIds = businessGlobalStore.businessAuthorizedList.map((biz) => biz.id);
  if (authorizedBizIds.includes(bizId)) {
    return true;
  }

  const authSign = { type: AUTH_ACCESS_BIZ, relation: [bizId] };
  const { authorized, permissionData } = await authStore.checkViewPermission(Symbol.for(`biz_${bizId}`), authSign);
  if (authorized) {
    return true;
  }

  to.meta.permissionData = permissionData;
  showBusinessStatus(to, 'bizUnauthed');
  return false;
};

function showBusinessStatus(to: RouteLocationNormalized, status: BusinessStatus) {
  to.meta.view = 'permission';
  to.meta.extra = { ...((to.meta.extra as object) || {}), businessStatus: status };
}
