import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useResourceStore } from './resource';
import { isChinese } from '@/language/i18n';
import {
  CLOUD_AREA_REGION_GCP,
  CLOUD_AREA_REGION_AWS,
  VendorEnum } from '@/common/constant';

export const useRegionsStore = defineStore('useRegions', () => {
  const tcloud = ref<Map<string, string>>(new Map());
  const huawei = ref<Map<string, string>>(new Map());

  const ressourceStore = useResourceStore();

  const REQUIRED_MAX_SIZE = 500;

  const fetchRegions = async (
    vendor: VendorEnum.TCLOUD | VendorEnum.HUAWEI,
    payload: Object = {
      filter: {
        op: 'and',
        rules: [],
      },
      page: {
        count: false,
        start: 0,
        limit: REQUIRED_MAX_SIZE,
      },
    },
  ) => {
    const res = await ressourceStore.getCloudRegion(vendor, payload);
    const details = res?.data?.details || [];
    if (vendor === VendorEnum.TCLOUD) {
      details.forEach((v: { region_id: string; region_name: string }) => {
        tcloud.value.set(v.region_id, v.region_name);
      });
    }
    if (vendor === VendorEnum.HUAWEI) {
      details.forEach((v: { region_id: string; locales_zh_cn: string }) => {
        huawei.value.set(v.region_id, v.locales_zh_cn);
      });
    }
  };

  const getRegionName = (vendor: VendorEnum, id: string) => {
    if (!isChinese) return id;
    switch (vendor) {
      case VendorEnum.AWS:
        return CLOUD_AREA_REGION_AWS[id] || id;
      case VendorEnum.GCP:
        return CLOUD_AREA_REGION_GCP[id] || id;
      case VendorEnum.HUAWEI:
        return huawei.value.get(id) || id;
      case VendorEnum.TCLOUD:
        return tcloud.value.get(id) || id;
    }
    return id;
  };

  return {
    getRegionName,
    fetchRegions,
  };
});
