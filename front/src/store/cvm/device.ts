import { ref } from 'vue';
import { defineStore } from 'pinia';
import rollRequest from '@blueking/roll-request';
import http, { type HttpRequestConfig } from '@/http';
import { IListResData, IQueryResData, QueryBuilderType } from '@/typings';
import { enableCount, onePageParams } from '@/utils/search';
import { RequirementType } from '@/store/config/requirement';

export interface ICvmDeviceItem {
  id: number;
  require_type: RequirementType;
  region: string;
  zone: string;
  device_type: string;
  cpu: number;
  mem: number;
  disk: number;
  remark: string;
  label: {
    device_group: string;
    device_size: string;
  };
  capacity_flag: number;
  enable_capacity: boolean;
  enable_apply: boolean;
  score: number;
  comment: string;
  device_type_class: 'SpecialType' | 'CommonType';
  [k: string]: any;
}

export interface ICvmDevicetypeItem {
  device_type: string;
  device_type_class: string;
  device_family: string; // 原 device_group
  cpu_core: number; // 原 cpu_amount
  memory: number; // 原 ram_amount
  core_type: string; // 核心类型，枚举值：小核心、中核心、大核心
  device_class: string;
  technical_class?: string;
  [k: string]: any;
}

export interface ICvmChargeTypDevicetypeItem {
  available: boolean;
  charge_type: string;
  device_types: Array<{
    device_type: string;
    available: boolean;
    remain_core: number;
  }>;
}

export interface IRollingServerCvm {
  device_type: string;
  instance_charge_type: string;
  charge_months: number;
  billing_start_time: string;
  old_billing_expire_time: string;
  bk_cloud_inst_id: string;
  device_group: string;
}

export interface IManyCvmCapacityItem {
  device_type: string;
  region: string;
  zone: string;
  vpc: string;
  subnet: string;
  max_num: number;
  max_info: { key: string; value: number }[];
}

export const CoreTypeMap = {
  1: '小核心',
  2: '中核心',
  3: '大核心',
} as const;

export const useCvmDeviceStore = defineStore('cvm-device', () => {
  const deviceListLoading = ref(false);

  const rollingServerCvmLoading = ref(false);
  const cvmCapacityLoading = ref(false);

  const getDeviceList = async (params: QueryBuilderType) => {
    deviceListLoading.value = true;
    const api = '/api/v1/woa/config/capacity/list_with_device_info';
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ICvmDeviceItem[]>>, Promise<IListResData<ICvmDeviceItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      deviceListLoading.value = false;
    }
  };

  const devicetypeListLoading = ref(false);
  const getOneDevicetype = async (params: QueryBuilderType) => {
    devicetypeListLoading.value = true;
    const api = '/api/v1/woa/config/findmany/config/cvm/devicetype';
    try {
      const res = await http.post(api, enableCount({ ...params, page: onePageParams() }, false));
      return res?.data?.details?.[0];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      devicetypeListLoading.value = false;
    }
  };

  const deviceTypeFullListLoading = ref(false);
  const getDeviceTypeFullList = async (params: QueryBuilderType) => {
    deviceTypeFullListLoading.value = true;
    try {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<ICvmDevicetypeItem>('/api/v1/woa/config/findmany/config/cvm/devicetype', params, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => res.data.details,
      });
      return { list, count: list.length };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      deviceTypeFullListLoading.value = false;
    }
  };

  const chargeTypeDeviceTypeListLoading = ref(false);
  const getChargeTypeDeviceTypeList = async (params: {
    bk_biz_id: number;
    require_type: RequirementType;
    region: string;
    zone?: string;
  }) => {
    chargeTypeDeviceTypeListLoading.value = true;
    try {
      const res: IListResData<ICvmChargeTypDevicetypeItem[]> = await http.post(
        '/api/v1/woa/config/findmany/config/cvm/charge_type/device_type',
        params,
      );
      const { info: list = [], count = 0 } = res.data ?? {};
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      chargeTypeDeviceTypeListLoading.value = false;
    }
  };

  const getRollingServerCvm = async (
    params: {
      bk_biz_id: number;
      bk_asset_id: string;
      region: string;
    },
    config?: { globalError?: boolean },
  ) => {
    rollingServerCvmLoading.value = true;
    try {
      const res: IQueryResData<IRollingServerCvm> = await http.post(
        '/api/v1/woa/task/check/rolling_server/host',
        params,
        config,
      );
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      rollingServerCvmLoading.value = false;
    }
  };

  const getManyCvmCapacity = async (
    params: {
      device_types: string[];
      require_type: number;
      region: string;
      zones: string[];
      vpc?: string;
      subnet?: string;
      charge_type?: string;
    },
    config?: HttpRequestConfig,
  ) => {
    cvmCapacityLoading.value = true;
    try {
      const res: IListResData<IManyCvmCapacityItem[]> = await http.post(
        '/api/v1/woa/config/findmany/cvm/capacity',
        params,
        config,
      );
      return res.data.info;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      cvmCapacityLoading.value = false;
    }
  };

  return {
    deviceListLoading,
    getDeviceList,
    devicetypeListLoading,
    getOneDevicetype,
    deviceTypeFullListLoading,
    getDeviceTypeFullList,
    chargeTypeDeviceTypeListLoading,
    getChargeTypeDeviceTypeList,
    rollingServerCvmLoading,
    getRollingServerCvm,
    cvmCapacityLoading,
    getManyCvmCapacity,
  };
});
