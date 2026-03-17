import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
// import { IListResData, QueryBuilderType, QueryFilterType } from '@/typings';
import { VendorEnum } from '@/common/constant';
// import { enableCount } from '@/utils/search';
// import rollRequest from '@blueking/roll-request';

// 二级账号项接口定义
export interface ISecondaryAccountItem {
  id: string;
  vendor: VendorEnum;
  name: string;
  managers: string[];
  security_managers: string[];
  type: string;
  site: string;
  price: string;
  price_unit: string;
  memo: string;
  bk_biz_id: number;
  usage_biz_ids: number[];
  email: string;
  cloud_created_at: string;
  sync_status: string;
  sync_failed_reason: string;
  sub_account_count: number;
  account_secret_count: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  extension: {
    login_flag?: string;
    action_flag?: string;
    console_login?: number;
    [k: string]: any;
  };
  [k: string]: any;
}

// 账号密钥项接口定义
export interface IAccountSecretItem {
  id: string;
  vendor: string;
  type: string; // 密钥类型：resource(资源管理)、security(安全管理)
  status: string; // 密钥状态：normal(正常)、invalid(失效)
  account_id: string;
  extension: {
    cloud_secret_id: string;
    cloud_main_account_id?: string;
    cloud_sub_account_id?: string;
  };
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

// 密钥校验响应接口定义
export interface ISecretCheckResult {
  cloud_main_account_id: string;
  cloud_sub_account_id: string;
}

// 创建/更新密钥参数接口定义
export interface ISecretCreateParams {
  account_id: string;
  type: string;
  extension: {
    cloud_secret_id: string;
    cloud_secret_key: string;
  };
}

export interface ISecretUpdateParams {
  type?: string;
  extension?: {
    cloud_secret_id: string;
    cloud_secret_key: string;
  };
}

export interface ISecretCheckParams {
  account_id: string;
  type: string;
  extension: {
    cloud_secret_id: string;
    cloud_secret_key: string;
  };
}

// 创建二级账号参数接口定义
export interface IAccountCreateParams {
  vendor: string;
  name: string;
  managers: string[];
  security_managers?: string[];
  type: string;
  site: string;
  bk_biz_id?: number;
  usage_biz_ids: number[];
  memo?: string;
  extension: Record<string, any>;
  remark?: string;
}

// 更新二级账号参数接口定义
export interface IAccountUpdateParams {
  name?: string;
  managers?: string[];
  security_managers?: string[];
  bk_biz_id?: number;
  usage_biz_ids?: number[];
  memo?: string;
  extension?: Record<string, any>;
}

// 三级账号密钥项接口定义
export interface ISubAccountSecretItem {
  id: string;
  vendor: string;
  status: 'enabled' | 'disabled';
  account_id: string;
  sub_account_id: string;
  extension: {
    cloud_secret_id: string;
    cloud_main_account_id: string;
    cloud_sub_account_id: string;
    console_login?: number;
  };
  tenant_id?: string;
  cloud_created_at: string;
  disabled_time?: string;
  last_used_time?: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  sub_account_manager?: string;
  account_manager?: string;
}

// 更新密钥状态参数
export interface IUpdateSecretStatusParams {
  id: string;
  status: 'enabled' | 'disabled';
}

export const usePermissionPolicyStore = defineStore('permissionPolicy', () => {
  const accountListLoading = ref(false);
  const secretListLoading = ref(false);
  const secretCheckLoading = ref(false);
  const subAccountSecretListLoading = ref(false);

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
  ): Promise<{ list: ISubAccountSecretItem[]; count: number }> => {
    subAccountSecretListLoading.value = true;

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
      subAccountSecretListLoading.value = false;
    }
  };

  return {
    accountListLoading,
    secretListLoading,
    secretCheckLoading,
    subAccountSecretListLoading,

    createPermissionPolicy,
    updatePermissionPolicy,
    getPermissionPolicyList,
  };
});
