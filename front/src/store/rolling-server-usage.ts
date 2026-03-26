import { ref } from 'vue';
import { defineStore } from 'pinia';
import rollRequest from '@blueking/roll-request';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { enableCount } from '@/utils/search';
import { IListResData, IQueryResData, QueryBuilderType, QueryFilterType } from '@/typings';

export enum AppliedType {
  NORMAL = 'normal',
  RESOURCE_POOL = 'resource_pool',
  CVM_PRODUCT = 'cvm_product',
}

export enum ReturnedWay {
  CRP = 'crp',
  RESOURCE_POOL = 'resource_pool',
}

interface IRollingServerBaseRecordItem {
  id: string;
  bk_biz_id: number;
  order_id: string;
  suborder_id: string;
  year: string;
  month: string;
  day: string;
  creator: string;
  created_at: string;
  [key: string]: any;
}

export interface IRollingServerAppliedRecordItem extends IRollingServerBaseRecordItem {
  applied_type: AppliedType;
  applied_core: number;
  delivered_core: number;
  exempted_returned_core?: number; // 减免退还核心数
}

export interface IRollingServerReturnedRecordItem extends IRollingServerBaseRecordItem {
  applied_record_id: string;
  match_applied_core: number;
  returned_way: ReturnedWay;
}

export type RollingServerRecordItem = IRollingServerAppliedRecordItem & {
  returned_records?: IRollingServerReturnedRecordItem[];
  returned_core?: number;
  not_returned_core?: number;
  not_returned_core_after_exempted?: number; // 减免后的未退还核心数
  exec_rate?: string;
};

export interface IRollingServerCpuCoreSummary {
  sum_delivered_core: number;
  sum_returned_applied_core: number;
}

export const useRollingServerUsageStore = defineStore('rolling-server-usage', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const appliedRecordsListLoading = ref(false);
  const getAppliedRecordList = async (data: QueryBuilderType) => {
    appliedRecordsListLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}rolling_servers/applied_records/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [
          Promise<IListResData<IRollingServerAppliedRecordItem[]>>,
          Promise<IListResData<IRollingServerAppliedRecordItem[]>>,
        ]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      appliedRecordsListLoading.value = false;
    }
  };

  const returnedRecordsListLoading = ref(false);
  const getReturnedRecordList = async (params: { filter: QueryFilterType }) => {
    returnedRecordsListLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}rolling_servers/returned_records/list`;
    try {
      const list = (await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<IRollingServerReturnedRecordItem>(api, params, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => res.data.details,
      })) as IRollingServerReturnedRecordItem[];
      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      returnedRecordsListLoading.value = false;
    }
  };

  const cpuCoreSummaryLoading = ref(false);
  const getCpuCoreSummary = async (data: {
    start: { year: number; month: number; day: number };
    end: { year: number; month: number; day: number };
    bk_biz_ids?: number[];
    order_ids?: string[];
    suborder_ids?: string[];
    returned_way?: ReturnedWay;
    applied_type?: AppliedType;
    require_type?: number;
  }) => {
    cpuCoreSummaryLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}rolling_servers/cpu_core/summary`;
    try {
      const res: IQueryResData<IRollingServerCpuCoreSummary> = await http.post(api, data);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      cpuCoreSummaryLoading.value = false;
    }
  };

  const updateAppliedRecordsNoticeDisabledLoading = ref(false);
  const updateAppliedRecordsNoticeDisabled = async (ids: string[]) => {
    updateAppliedRecordsNoticeDisabledLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}rolling_servers/applied_records/notice/disabled/update`;
    try {
      const res = await http.post(api, { ids });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateAppliedRecordsNoticeDisabledLoading.value = false;
    }
  };

  // 更新减免退还核心数
  const updateExemptedReturnedCoreLoading = ref(false);
  const updateExemptedReturnedCore = async (ids: string[], exempted_returned_core: number) => {
    updateExemptedReturnedCoreLoading.value = true;
    const api = '/api/v1/woa/rolling_servers/applied_records/exempted_returned_core';
    try {
      const res = await http.patch(api, { ids, exempted_returned_core });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateExemptedReturnedCoreLoading.value = false;
    }
  };

  return {
    appliedRecordsListLoading,
    getAppliedRecordList,
    returnedRecordsListLoading,
    getReturnedRecordList,
    cpuCoreSummaryLoading,
    getCpuCoreSummary,
    updateAppliedRecordsNoticeDisabledLoading,
    updateAppliedRecordsNoticeDisabled,
    updateExemptedReturnedCoreLoading,
    updateExemptedReturnedCore,
  };
});
