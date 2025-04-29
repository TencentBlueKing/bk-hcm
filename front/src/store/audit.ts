import { ref } from 'vue';
import { defineStore } from 'pinia';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { IListResData, QueryBuilderType } from '@/typings';
import { enableCount } from '@/utils/search';
import http from '@/http';

export interface IAuditItem {
  id: number;
  res_id: string;
  cloud_res_id: string;
  res_name: string;
  res_type: string;
  associated_res_id: string;
  associated_cloud_res_id: string;
  associated_res_name: string;
  associated_res_type: string;
  action: string;
  bk_biz_id: number;
  vendor: string;
  account_id: string;
  operator: string;
  source: string;
  rid: string;
  app_code: string;
  detail: {
    data: any;
  };
  created_at: string;
  [key: string]: any;
}

export const useAuditStore = defineStore('audit', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const isAuditListLoading = ref(false);
  const getAuditList = async (params: QueryBuilderType) => {
    isAuditListLoading.value = true;
    const api = `/api/v1/cloud/${getBusinessApiPath()}audits/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IAuditItem[]>>, Promise<IListResData<IAuditItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isAuditListLoading.value = false;
    }
  };

  return {
    isAuditListLoading,
    getAuditList,
  };
});
