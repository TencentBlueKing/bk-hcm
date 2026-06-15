import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, IQueryResData, QueryParamsType } from '@/typings';
import { enableCount } from '@/utils/search';
import { VendorEnum } from '@/common/constant';
import rollRequest from '@blueking/roll-request';
import { ListGeneratorFactory } from '@/components/form/list.vue';

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
  const getPermissionTemplateSubAccountIds = async (
    bizId: number,
    vendor: VendorEnum,
    id: string,
  ): Promise<{ id: string; cloud_id: string }[]> => {
    subAccountIdsLoading.value = true;
    try {
      const res = await http.get(
        `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/permission_templates/${id}/sub_account_ids`,
      );
      return res?.data?.sub_accounts ?? [];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      subAccountIdsLoading.value = false;
    }
  };

  /**
   * 获取云权限模板详情（通过列表接口按 id 查询单条）
   */
  const getPermissionTemplateDetail = async (
    bizId: number,
    vendor: VendorEnum,
    id: string,
  ): Promise<IPermissionTemplateItem | null> => {
    const api = `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/permission_templates/list`;
    try {
      const res = await http.post(api, {
        ids: [id],
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
   * 创建云权限模板列表迭代器（用于选择器列表组件）
   * @param bizId 业务ID
   * @param vendor 云厂商
   * @param defaultParams 默认参数
   */
  const createPermissionTemplateListGenerator = (
    bizId: number,
    vendor: VendorEnum,
    defaultParams?: Record<string, any>,
  ): ListGeneratorFactory => {
    return async function* (keywordOrOptions) {
      const api = `/api/v1/cloud/bizs/${bizId}/vendors/${vendor}/permission_templates/list`;
      let params: Record<string, any> = { ...defaultParams };
      const keyword = typeof keywordOrOptions === 'string' ? keywordOrOptions : undefined;
      const options = typeof keywordOrOptions === 'object' ? keywordOrOptions : undefined;
      if (keyword) params.names = [keyword];

      // 如果传递ids将丢弃其它条件只保留ids条件，在真实的使用场景中存在ids指定的值与其它条件不兼容导致查询不到数据
      // 但业务逻辑允许这种情况，所以采用这种策略
      if (options?.ids?.length) {
        params = {
          ids: options.ids,
        };
      }

      const gen = await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
        IListResData<IPermissionTemplateItem[]>
      >(
        api,
        { ...params },
        { limit: 500, countGetter: (res) => res.data.count, listGetter: (res) => res.data.details, generator: true },
        true,
      );

      for (const promise of gen) {
        const res = await promise;
        yield res?.data?.details ?? [];
      }
    };
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
    getPermissionTemplateDetail,
    createPermissionTemplateListGenerator,
  };
});
