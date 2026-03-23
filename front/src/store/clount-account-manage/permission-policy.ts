import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';

export interface IPermissionPolicyItem {
  id: string;
  // TODO 待补充
}

export const usePermissionPolicyStore = defineStore('permissionPolicy', () => {
  const permissionPolicyListLoading = ref(false);

  /**
   * 创建权限策略库
   * @param bk_biz_id 业务ID
   * @param params 账号参数
   */
  const createPermissionPolicy = async (bk_biz_id: number, params: IAccountCreateParams) => {
    try {
      const res = await http.post(`/api/v1/cloud/bizs/${bk_biz_id}/applications/types/add_account`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 更新权限策略库
   * @param bk_biz_id 业务ID
   * @param account_id 账号ID
   * @param params 更新参数
   */
  const updatePermissionPolicy = async (bk_biz_id: number, account_id: string, params: IAccountUpdateParams) => {
    try {
      const res = await http.patch(`/api/v1/cloud/bizs/${bk_biz_id}/accounts/${account_id}`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 获取权限策略库列表
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 查询参数
   */
  const getPermissionPolicyList = async (
    bk_biz_id: number,
    vendor: string,
    params: { filter?: any; page: any } & Record<string, any>,
  ): Promise<{ list: IPermissionPolicyItem[]; count: number }> => {
    permissionPolicyListLoading.value = true;

    // 使用真实接口
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/list`;
    try {
      // 构建请求参数
      const requestData = { ...params };

      // 获取列表数据
      const listRes = await http.post(api, {
        ...requestData,
        page: { ...requestData.page, count: false },
      });

      // 获取总数
      const countRes = await http.post(api, {
        ...requestData,
        page: { count: true, start: 0, limit: 0 },
      });

      const list = listRes?.data?.details || [];
      const count = countRes?.data?.count || 0;

      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      permissionPolicyListLoading.value = false;
    }
  };

  return {
    permissionPolicyListLoading,
    createPermissionPolicy,
    updatePermissionPolicy,
    getPermissionPolicyList,
  };
});
