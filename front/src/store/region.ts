import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { QueryRuleOPEnum, QueryFilterType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { isChinese } from '@/language/i18n';
import { getRegionName } from '@pluginHandler/region-selector';
import rollRequest from '@blueking/roll-request';

export interface IRegionItem {
  id: string;
  region_id: string;
  region_name: string;
  vendor: string;
}

export interface IRegionListParams {
  vendor: string;
  resourceType?: ResourceTypeEnum.CVM | ResourceTypeEnum.VPC | ResourceTypeEnum.DISK | ResourceTypeEnum.SUBNET;
}

export const useRegionStore = defineStore('region', () => {
  const regionListLoading = ref(false);
  const cache = new Map();
  const requestQueue = new Map();

  const getRegionList = async (params: IRegionListParams) => {
    const { vendor, resourceType } = params;
    const key = JSON.stringify(params);

    // 检查缓存
    if (cache.has(key)) {
      return cache.get(key);
    }

    // 如果已经有一个请求在进行中，则返回该请求的 Promise
    if (requestQueue.has(key)) {
      return requestQueue.get(key);
    }

    const filter: QueryFilterType = { op: 'and', rules: [] };
    let dataIdKey = 'region_id';
    let dataNameKey = 'region_name';
    switch (vendor) {
      case VendorEnum.AZURE:
        filter.rules = [{ field: 'type', op: QueryRuleOPEnum.EQ, value: 'Region' }];
        dataIdKey = 'name';
        dataNameKey = 'display_name';
        break;
      case VendorEnum.HUAWEI: {
        const services = {
          [ResourceTypeEnum.CVM]: 'ecs',
          [ResourceTypeEnum.VPC]: 'vpc',
          [ResourceTypeEnum.DISK]: 'ecs',
          [ResourceTypeEnum.SUBNET]: 'vpc',
        };
        filter.rules = [{ field: 'type', op: QueryRuleOPEnum.EQ, value: 'public' }];
        // TODO：临时解决CLB资源-华为云拉取region的问题
        services[resourceType] &&
          filter.rules.push({ field: 'service', op: QueryRuleOPEnum.EQ, value: services[resourceType] });
        dataNameKey = isChinese ? 'locales_zh_cn' : 'region_id';
        break;
      }
      case VendorEnum.TCLOUD: {
        filter.rules = [
          { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
          { field: 'status', op: QueryRuleOPEnum.EQ, value: 'AVAILABLE' },
        ];
        dataNameKey = isChinese ? 'region_name' : 'display_name';
        break;
      }
      case VendorEnum.AWS: {
        filter.rules = [
          { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
          { field: 'status', op: QueryRuleOPEnum.EQ, value: 'opt-in-not-required' },
        ];
        break;
      }
      case VendorEnum.GCP:
        filter.rules = [
          { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
          { field: 'status', op: QueryRuleOPEnum.EQ, value: 'UP' },
        ];
        break;
    }

    // 创建一个新的请求并加入队列
    regionListLoading.value = true;
    const requestPromise = new Promise(async (resolve, reject) => {
      try {
        const list = (
          await rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<IRegionItem>(
            `/api/v1/cloud/vendors/${vendor}/regions/list`,
            { filter },
            { limit: 500, listGetter: (res) => res.data.details, countGetter: (res) => res.data.count },
          )
        ).map((item: any) => ({
          id: item[dataIdKey],
          name: getRegionName(isChinese, vendor as VendorEnum, item[dataIdKey], item[dataNameKey]) || item[dataIdKey],
        }));

        // 更新缓存
        cache.set(key, list);
        resolve(list);
      } catch (error) {
        reject(error);
      } finally {
        requestQueue.delete(key);
        regionListLoading.value = false;
      }
    });

    requestQueue.set(key, requestPromise);

    return requestPromise;
  };

  return {
    regionListLoading,
    getRegionList,
  };
});
