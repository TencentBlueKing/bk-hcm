import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useResourceStore } from './resource';
import { isChinese } from '@/language/i18n';
import {
  CLOUD_AREA_REGION_GCP,
  CLOUD_AREA_REGION_AWS,
  CLOUD_AREA_REGION_GCP_EN,
  CLOUD_AREA_REGION_AWS_EN,
  VendorEnum,
} from '@/common/constant';
import { swapMapKeysAndValuesToObj } from '@/common/util';

export const useRegionsStore = defineStore('useRegions', () => {
  const tcloud = ref<Map<string, string>>(new Map());
  const huawei = ref<Map<string, string>>(new Map());
  const vendor = ref('' as VendorEnum);

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
    let regionName;
    switch (vendor) {
      case VendorEnum.AWS:
        regionName = CLOUD_AREA_REGION_AWS[id] || id;
        break;
      case VendorEnum.GCP:
        regionName = CLOUD_AREA_REGION_GCP[id] || id;
        break;
      case VendorEnum.HUAWEI:
        regionName = huawei.value.get(id) || id;
        break;
      case VendorEnum.TCLOUD:
        regionName = tcloud.value.get(id) || id;
        break;
      default:
        regionName = id;
    }
    return regionName || '--';
  };

  const getRegionNameEN = (id: string) => {
    if (!isChinese) return id;
    const CLOUD_AREA_REGION_TCLOUD_EN = swapMapKeysAndValuesToObj(tcloud.value);
    const CLOUD_AREA_REGION_HUAWEI_EN = swapMapKeysAndValuesToObj(huawei.value);
    if (CLOUD_AREA_REGION_TCLOUD_EN[id]) {
      vendor.value = VendorEnum.TCLOUD;
      return CLOUD_AREA_REGION_TCLOUD_EN[id];
    }
    if (CLOUD_AREA_REGION_HUAWEI_EN[id]) {
      vendor.value = VendorEnum.HUAWEI;
      return CLOUD_AREA_REGION_HUAWEI_EN[id];
    }
    if (CLOUD_AREA_REGION_AWS_EN[id]) {
      vendor.value = VendorEnum.AWS;
      return CLOUD_AREA_REGION_AWS_EN[id];
    }
    if (CLOUD_AREA_REGION_GCP_EN[id]) {
      vendor.value = VendorEnum.GCP;
      return CLOUD_AREA_REGION_GCP_EN[id];
    }
    return id;
  };

  const getZoneName = (zone: string, vendor: VendorEnum) => {
    const idx = zone.lastIndexOf('-');
    return getRegionName(vendor, zone.substring(0, idx)) + zone.substring(idx);
  };

  return {
    getRegionName,
    fetchRegions,
    getRegionNameEN,
    getZoneName,
    vendor,
  };
});
