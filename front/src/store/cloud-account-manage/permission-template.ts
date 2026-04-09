import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, IQueryResData, QueryParamsType } from '@/typings';
import { enableCount } from '@/utils/search';
import { VendorEnum } from '@/common/constant';

export interface IPermissionTemplateItem {
  id: string;
  cloud_id: string;
  name: string;
  vendor: string;
  account_id: string;
  cloud_account_id: string;
  policy_library_id: string;
  policy_library_name: string;
  policy_library_version: number;
  policy_library_sync_time: string;
  policy_document: string;
  memo: string;
  associated_sub_account_count: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  extension: {
    cloud_type?: number;
    [k: string]: any;
  };
}

export interface ICreatePermissionTemplateParams {
  account_id: string;
  policy_library_id: string;
  name: string;
  memo?: string;
}

export interface IUpdatePermissionTemplateParams {
  id: string;
  policy_library_id?: string;
  name?: string;
  memo?: string;
}

export interface IDeletePermissionTemplateParams {
  id: string;
}

export const usePermissionTemplateStore = defineStore('permission-template', () => {
  const listLoading = ref(false);
  const createLoading = ref(false);
  const updateLoading = ref(false);
  const deleteLoading = ref(false);
  const subAccountIdsLoading = ref(false);

  /**
   * 查询云权限模板列表
   */
  const getPermissionTemplateList = async (bizId: number, vendor: VendorEnum, params: QueryParamsType) => {
    listLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/permission_templates/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IPermissionTemplateItem[]>>, Promise<IListResData<IPermissionTemplateItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      listLoading.value = false;
    }
  };

  /**
   * 创建云权限模板（提交审批单）
   */
  const createPermissionTemplate = async (
    bizId: number,
    vendor: VendorEnum,
    params: ICreatePermissionTemplateParams,
  ) => {
    createLoading.value = true;
    try {
      const res: IQueryResData<{ id: string }> = await http.post(
        `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/applications/types/create_permission_template`,
        params,
      );
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      createLoading.value = false;
    }
  };

  /**
   * 编辑云权限模板（提交审批单）
   */
  const updatePermissionTemplate = async (
    bizId: number,
    vendor: VendorEnum,
    params: IUpdatePermissionTemplateParams,
  ) => {
    updateLoading.value = true;
    try {
      const res: IQueryResData<{ id: string }> = await http.post(
        `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/applications/types/update_permission_template`,
        params,
      );
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateLoading.value = false;
    }
  };

  /**
   * 删除云权限模板（提交审批单）
   */
  const deletePermissionTemplate = async (
    bizId: number,
    vendor: VendorEnum,
    params: IDeletePermissionTemplateParams,
  ) => {
    deleteLoading.value = true;
    try {
      const res: IQueryResData<{ id: string }> = await http.post(
        `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/applications/types/delete_permission_template`,
        params,
      );
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      deleteLoading.value = false;
    }
  };

  /**
   * 查询云权限模板关联的三级账号ID列表（全量返回，不分页）
   */
  const getPermissionTemplateSubAccountIds = async (bizId: number, vendor: VendorEnum, id: string) => {
    subAccountIdsLoading.value = true;
    try {
      const res: IQueryResData<{ sub_account_ids: string[] }> = await http.get(
        `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/permission_templates/${id}/sub_account_ids`,
      );
      return res?.data?.sub_account_ids ?? [];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      subAccountIdsLoading.value = false;
    }
  };

  return {
    listLoading,
    createLoading,
    updateLoading,
    deleteLoading,
    subAccountIdsLoading,
    getPermissionTemplateList,
    createPermissionTemplate,
    updatePermissionTemplate,
    deletePermissionTemplate,
    getPermissionTemplateSubAccountIds,
  };
});
