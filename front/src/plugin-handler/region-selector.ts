import { CLOUD_AREA_REGION_AWS, CLOUD_AREA_REGION_GCP, VendorEnum } from '@/common/constant';

export const getRegionName = (isChinese: boolean, vendor: VendorEnum, key: string, name: string) => {
  switch (vendor) {
    case VendorEnum.AWS:
      return isChinese ? CLOUD_AREA_REGION_AWS[key] : key;
    case VendorEnum.GCP:
      return isChinese ? CLOUD_AREA_REGION_GCP[key] : key;
    case VendorEnum.TCLOUD:
    case VendorEnum.HUAWEI:
      return isChinese ? name : key;
    case VendorEnum.AZURE:
      return name;
    default:
      return '--';
  }
};
