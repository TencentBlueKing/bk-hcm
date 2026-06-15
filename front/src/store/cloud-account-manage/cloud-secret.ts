import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import type { ICloudSecretItem } from '@/views/cloud-account-manage/cloud-secret/typings';

// 更新密钥状态参数
export interface IUpdateSecretStatusParams {
  id: string;
  status: 'enabled' | 'disabled';
}

export interface ISubAccountSecretParams {
  status?: string;
  account_ids?: string[];
  sub_account_ids?: string[];
  account_managers?: string[];
  sub_account_managers?: string[];
  extension?: {
    cloud_secret_ids?: string[];
    cloud_main_account_ids?: string[];
    cloud_sub_account_ids?: string[];
  };
  page: any;
}

// ISubAccountSecretItem 已统一为 ICloudSecretItem（来自 cloud-secret/typings.ts），两者共用同一套接口
// 保留别名以保持向后兼容
export { type ICloudSecretItem as ISubAccountSecretItem } from '@/views/cloud-account-manage/cloud-secret/typings';

export const useCloudSecretStore = defineStore('cloudSecret', () => {
  const subAccountSecretListLoading = ref(false);

  /**
   * 获取三级账号密钥列表
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 查询参数
   */
  const getSubAccountSecretList = async (
    bk_biz_id: number,
    vendor: string,
    params: ISubAccountSecretParams,
  ): Promise<{ list: ICloudSecretItem[]; count: number }> => {
    subAccountSecretListLoading.value = true;

    // 使用真实接口
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/list`;
    try {
      // 构建请求参数（去除 page 后的查询条件）
      const { page, ...queryParams } = params;

      // 获取列表数据
      const listRes = await http.post(api, {
        ...queryParams,
        page: { ...page, count: false },
      });

      // 获取总数
      const countRes = await http.post(api, {
        ...queryParams,
        page: { count: true, start: 0, limit: 0 },
      });

      const list = listRes?.data?.details || [];
      const count = countRes?.data?.count || 0;

      // 处理数据，将 extension 中的字段提取到顶层便于展示
      const processedList = list.map((item: ICloudSecretItem) => ({
        ...item,
        ...item.extension,
      }));

      return { list: processedList, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      subAccountSecretListLoading.value = false;
    }
  };

  /**
   * 获取三级账号密钥详情（通过 sub_account_secrets/list 接口按 id 查询单条）
   */
  const getSubAccountSecretDetail = async (
    bk_biz_id: number,
    vendor: string,
    id: string,
  ): Promise<ICloudSecretItem | null> => {
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/list`;
    try {
      const res = await http.post(api, {
        ids: [id],
        page: { count: false, start: 0, limit: 1 },
      });
      const list = (res?.data as { details: ICloudSecretItem[] })?.details ?? [];
      if (list.length === 0) return null;
      const item = list[0];
      // 将 extension 中的字段提取到顶层，与列表数据处理保持一致
      return { ...item, ...item.extension };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 启用或禁用三级账号密钥（创建申请）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param params 密钥状态更新参数列表
   */
  const updateSubAccountSecretStatus = async (
    bk_biz_id: number,
    vendor: string,
    params: IUpdateSecretStatusParams[],
  ): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/update_sub_account_secret_status`;
      const res = await http.post(api, {
        sub_account_secrets: params,
      });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 删除三级账号密钥（创建申请）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param ids 密钥ID列表
   */
  const deleteSubAccountSecret = async (
    bk_biz_id: number,
    vendor: string,
    ids: string[],
  ): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/delete_sub_account_secret`;
      const res = await http.post(api, { ids });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 新增三级账号密钥
   */
  const createSubAccountSecret = async (
    bk_biz_id: number,
    vendor: string,
    sub_account_id: string,
  ): Promise<{ id: string; extension: { cloud_secret_id: string; cloud_secret_key: string } }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_account_secrets/create`;
      const res = await http.post(api, { sub_account_id });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    subAccountSecretListLoading,
    getSubAccountSecretList,
    getSubAccountSecretDetail,
    updateSubAccountSecretStatus,
    deleteSubAccountSecret,
    createSubAccountSecret,
  };
});
