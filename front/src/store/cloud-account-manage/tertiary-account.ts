import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryFilterType } from '@/typings';
import { enableCount } from '@/utils/search';
import rollRequest from '@blueking/roll-request';

// 三级账号项接口定义
export interface ISubAccountItem {
  id: string;
  cloud_id: string;
  name: string;
  vendor: string;
  site: string;
  account_id: string;
  managers: string[];
  bk_biz_ids: number[];
  memo: string;
  email: string;
  phone_num: string;
  country_code: string;
  cloud_created_at: string;
  sub_account_secret_count: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  extension: {
    cloud_main_account_id: string;
    uin: number;
    nick_name: string;
    create_time: string;
    login_flag: string;
    action_flag: string;
    console_login: number;
    [k: string]: any;
  };
  operable?: boolean;
  [k: string]: any;
}

// 创建三级账号参数
export interface ISubAccountCreateParams {
  account_id: string;
  name: string;
  receive_email: string;
  permission_template_ids: string[];
  email?: string;
  phone_num?: string;
  country_code?: string;
  managers?: string[];
  memo?: string;
  extension: {
    console_login: number; // 0=编程账号，1=控制台账号
  };
}

// 更新三级账号参数
export interface ISubAccountUpdateParams {
  id: string;
  name?: string;
  email?: string;
  phone_num?: string;
  bk_biz_id?: number;
  country_code?: string;
  managers?: string[];
  memo?: string;
  permission_template_ids?: string[];
}

export const useTertiaryAccountStore = defineStore('tertiaryAccount', () => {
  const subAccountListLoading = ref(false);

  /**
   * 获取三级账号详情（通过列表接口按 id 查询单条）
   */
  const getSubAccountDetail = async (
    bk_biz_id: number,
    vendor: string,
    id: string,
  ): Promise<ISubAccountItem | null> => {
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_accounts/list`;
    try {
      const res = await http.post(api, {
        filter: { rules: [{ field: 'id', op: 'eq', value: id }], op: 'and' },
        page: { count: false, start: 0, limit: 1 },
      });
      const list = res?.data?.details ?? [];
      return list.length > 0 ? list[0] : null;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 获取三级账号全量列表（用于前端分页）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param filter 过滤条件
   * @param onProgress 进度回调
   */
  const getSubAccountFullList = async (
    bk_biz_id: number,
    vendor: string,
    filter: QueryFilterType,
    onProgress?: (list: ISubAccountItem[], count: number) => void,
  ): Promise<ISubAccountItem[]> => {
    subAccountListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_accounts/list`;
    const allList: ISubAccountItem[] = [];

    try {
      const listGen = await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
        IListResData<ISubAccountItem[]>
      >(
        api,
        { filter },
        {
          limit: 500,
          countGetter: (res) => res.data.count,
          listGetter: (res) => res.data.details,
          generator: true,
        },
        true,
      );

      for await (const res of listGen) {
        const details = res.data?.details || [];
        allList.push(...details);
        onProgress?.(allList, res.data?.count || allList.length);
        subAccountListLoading.value = false;
      }

      return allList;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      subAccountListLoading.value = false;
    }
  };

  /**
   * 获取三级账号数量（纯计数查询）
   * @param bk_biz_id 业务ID
   * @param vendor 云厂商
   * @param filter 过滤条件
   */
  const getSubAccountCount = async (bk_biz_id: number, vendor: string, filter: QueryFilterType): Promise<number> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/sub_accounts/list`;
      const res = await http.post(api, enableCount({ filter }, true));
      return res?.data?.count ?? 0;
    } catch (error) {
      console.error(error);
      return 0;
    }
  };

  /**
   * 创建三级账号（提交申请）
   */
  const createSubAccount = async (
    bk_biz_id: number,
    vendor: string,
    subAccounts: ISubAccountCreateParams[],
  ): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/add_sub_account`;
      const res = await http.post(api, { sub_accounts: subAccounts });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 更新三级账号（提交申请）
   */
  const updateSubAccount = async (
    bk_biz_id: number,
    vendor: string,
    subAccounts: ISubAccountUpdateParams[],
  ): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/update_sub_account`;
      const res = await http.post(api, { sub_accounts: subAccounts });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  /**
   * 删除三级账号（提交申请）
   */
  const deleteSubAccount = async (bk_biz_id: number, vendor: string, ids: string[]): Promise<{ ids: string[] }> => {
    try {
      const api = `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/applications/types/delete_sub_account`;
      const res = await http.post(api, { ids });
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    subAccountListLoading,
    getSubAccountDetail,
    getSubAccountFullList,
    getSubAccountCount,
    createSubAccount,
    updateSubAccount,
    deleteSubAccount,
  };
});
