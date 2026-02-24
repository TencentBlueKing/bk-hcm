import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import {
  type IAuthSign,
  type IPermission,
  type IVerifyResourceInstance,
  getVerifyParams,
  extractSignPermission,
} from '@/common/auth-service';
import { type IQueryResData } from '@/typings/common';
import { type IViewAuthConfig } from '@/constants/view-auth';

// 重新导出以保持兼容性
export type { IVerifyResourceInstance };

export interface IVerifyResult {
  results: {
    authorized: boolean;
  }[];
  permission?: IPermission;
}

// 视图权限数据：[是否有权限, 对应的 permission 数据]
export type ViewPermissionValue = [boolean, IPermission | null];

export const useAuthStore = defineStore('auth', () => {
  const applyPermUrlLoading = ref(false);
  const viewPermissions = ref<Map<symbol, ViewPermissionValue>>(new Map());

  const verify = async (authSign: IAuthSign | IAuthSign[]) => {
    try {
      const params = getVerifyParams(authSign);
      const res: IQueryResData<IVerifyResult> = await http.post('/api/v1/web/auth/verify', params);
      return {
        results: res.data?.results ?? [],
        permission: res.data?.permission,
      };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const getApplyPermUrl = async (params: IPermission) => {
    try {
      applyPermUrlLoading.value = true;
      const res = await http.post('/api/v1/web/auth/find/apply_perm_url', params);
      return res.data;
    } finally {
      applyPermUrlLoading.value = false;
    }
  };

  /**
   * 预获取视图权限（在 preload 中调用，之后直接使用）
   * @param configs 视图权限配置列表
   */
  const fetchViewPermissions = async (configs: IViewAuthConfig[]) => {
    if (configs.length === 0) {
      return;
    }

    const authSigns: IAuthSign[] = configs.map((config) => config.authSign);

    try {
      const result = await verify(authSigns);

      configs.forEach((config, index) => {
        const authorized = result.results[index]?.authorized ?? false;
        // 提取当前 sign 对应的 permission 数据
        const signPermission = result.permission ? extractSignPermission(config.authSign, result.permission) : null;
        viewPermissions.value.set(config.viewId, [authorized, signPermission]);
      });
    } catch (error) {
      console.error('Failed to fetch view permissions:', error);
      // 失败时设置所有为无权限
      configs.forEach((config) => {
        viewPermissions.value.set(config.viewId, [false, null]);
      });
    }
  };

  /**
   * 检查是否有视图权限（同步，仅检查缓存）
   * @param viewId 视图ID（Symbol）
   */
  const hasViewPermission = (viewId: symbol): boolean => {
    const value = viewPermissions.value.get(viewId);
    return value ? value[0] : true;
  };

  /**
   * 获取指定视图的 permission 数据（同步，仅检查缓存）
   * @param viewId 视图ID（Symbol）
   */
  const getViewPermissionData = (viewId: symbol): IPermission | null => {
    const value = viewPermissions.value.get(viewId);
    return value ? value[1] : null;
  };

  /**
   * 检查视图权限是否已缓存
   * @param viewId 视图ID（Symbol）
   */
  const isViewPermissionCached = (viewId: symbol): boolean => {
    return viewPermissions.value.has(viewId);
  };

  /**
   * 检查视图权限（支持动态获取）
   * - 如果已在 preload 中预获取，直接使用缓存
   * - 否则根据 authSign 动态请求权限
   *
   * @param viewId 视图ID（Symbol）
   * @param authSign 权限配置，用于动态获取
   * @returns { authorized: boolean, permissionData: IPermission | null }
   */
  const checkViewPermission = async (
    viewId: symbol,
    authSign?: IAuthSign,
  ): Promise<{ authorized: boolean; permissionData: IPermission | null }> => {
    // 1. 先检查缓存（preload 中预获取的权限）
    if (isViewPermissionCached(viewId)) {
      return {
        authorized: hasViewPermission(viewId),
        permissionData: getViewPermissionData(viewId),
      };
    }

    // 2. 如果没有缓存且有 authSign，动态获取
    if (authSign) {
      try {
        const result = await verify(authSign);
        const authorized = result.results[0]?.authorized ?? false;
        const signPermission = result.permission ? extractSignPermission(authSign, result.permission) : null;
        // 缓存结果，避免重复请求
        viewPermissions.value.set(viewId, [authorized, signPermission]);
        return { authorized, permissionData: signPermission };
      } catch (error) {
        console.error('Failed to check view permission:', error);
        // 请求失败时设置为无权限
        viewPermissions.value.set(viewId, [false, null]);
        return { authorized: false, permissionData: null };
      }
    }

    // 3. 没有缓存也没有配置，默认有权限
    return { authorized: true, permissionData: null };
  };

  /**
   * 重置视图权限
   */
  const resetViewPermissions = () => {
    viewPermissions.value.clear();
  };

  return {
    verify,
    getApplyPermUrl,
    applyPermUrlLoading,
    viewPermissions,
    fetchViewPermissions,
    hasViewPermission,
    getViewPermissionData,
    isViewPermissionCached,
    checkViewPermission,
    resetViewPermissions,
  };
});
