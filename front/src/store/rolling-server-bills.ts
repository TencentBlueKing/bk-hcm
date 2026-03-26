import { ref } from 'vue';
import { defineStore } from 'pinia';
import { IListResData, QueryBuilderType } from '@/typings';
import { enableCount } from '@/utils/search';
import http from '@/http';

export interface IRollingServerBillItem {
  id: string;
  bk_biz_id: number;
  offset_config_id: string;
  product_id: number;
  delivered_core: number;
  returned_core: number;
  not_returned_core: number;
  year: number;
  month: number;
  day: number;
  creator: string;
  created_at: string;
}

export interface IFineDetailsItem {
  id: string;
  bk_biz_id: number;
  order_id: string;
  suborder_id: string;
  year: number;
  month: number;
  day: number;
  delivered_core: number;
  returned_core: number;
  exempted_returned_core?: number; // 减免退还核心数
  creator: string;
  created_at: string;
}

export const useRollingServerBillsStore = defineStore('rolling-server-bills', () => {
  const billListLoading = ref(false);
  const billFineDetailsListLoading = ref(false);

  const getBillList = async (params: QueryBuilderType) => {
    billListLoading.value = true;
    const api = '/api/v1/woa/rolling_servers/bills/list';
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IRollingServerBillItem[]>>, Promise<IListResData<IRollingServerBillItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list: list || [], count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      billListLoading.value = false;
    }
  };

  const getBillFineDetailsList = async (params: QueryBuilderType) => {
    billFineDetailsListLoading.value = true;
    const api = '/api/v1/woa/rolling_servers/fine_details/list';
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IFineDetailsItem[]>>, Promise<IListResData<IFineDetailsItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list: list || [], count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      billFineDetailsListLoading.value = false;
    }
  };

  return {
    billListLoading,
    billFineDetailsListLoading,
    getBillList,
    getBillFineDetailsList,
  };
});
