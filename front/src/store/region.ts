import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { QueryRuleOPEnum, IListResData, QueryFilterType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { isChinese } from '@/language/i18n';
import { getRegionName } from '@pluginHandler/region-selector';

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

  const getRegionList = async (params: IRegionListParams) => {
    const { vendor, resourceType } = params;
    const filter: QueryFilterType = {
      op: 'and',
      rules: [],
    };
    let dataIdKey = 'region_id';
    let dataNameKey = 'region_name';
    switch (vendor) {
      case VendorEnum.AZURE:
        filter.rules = [
          {
            field: 'type',
            op: QueryRuleOPEnum.EQ,
            value: 'Region',
          },
        ];
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
        filter.rules = [
          {
            field: 'type',
            op: QueryRuleOPEnum.EQ,
            value: 'public',
          },
          {
            field: 'service',
            op: QueryRuleOPEnum.EQ,
            value: services[resourceType],
          },
        ];
        dataNameKey = isChinese ? 'locales_zh_cn' : 'region_id';
        break;
      }
      case VendorEnum.TCLOUD: {
        filter.rules = [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: vendor,
          },
          {
            field: 'status',
            op: QueryRuleOPEnum.EQ,
            value: 'AVAILABLE',
          },
        ];
        dataNameKey = isChinese ? 'region_name' : 'display_name';
        break;
      }
      case VendorEnum.AWS: {
        filter.rules = [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: vendor,
          },
          {
            field: 'status',
            op: QueryRuleOPEnum.EQ,
            value: 'opt-in-not-required',
          },
        ];
        break;
      }
      case VendorEnum.GCP:
        filter.rules = [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: vendor,
          },
          {
            field: 'status',
            op: QueryRuleOPEnum.EQ,
            value: 'UP',
          },
        ];
        break;
    }

    regionListLoading.value = true;

    try {
      const result: IListResData<IRegionItem[]> = await http.post(`/api/v1/cloud/vendors/${vendor}/regions/list`, {
        filter,
        page: {
          count: false,
          start: 0,
          // TODO: 滚动获取
          limit: 500,
        },
      });

      const details = result?.data?.details ?? [];
      const list = details.map((item: any) => ({
        id: item[dataIdKey],
        name: getRegionName(isChinese, vendor as VendorEnum, item[dataIdKey], item[dataNameKey]) || item[dataIdKey],
      }));

      return list;
    } catch (error) {
      return Promise.reject(error);
    } finally {
      regionListLoading.value = false;
    }
  };

  return {
    regionListLoading,
    getRegionList,
  };
});
