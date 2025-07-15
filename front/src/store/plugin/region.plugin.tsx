import { VendorEnum } from '@/common/constant';
import { isChinese } from '@/language/i18n';

export const otherCloud: string[] = [];

export const getRegionIDName = (vendor: string) => {
  let dataIdKey = 'region_id';
  let dataNameKey = 'region_name';
  switch (vendor) {
    case VendorEnum.AZURE:
      dataIdKey = 'name';
      dataNameKey = 'display_name';
      break;
    case VendorEnum.HUAWEI: {
      dataNameKey = isChinese ? 'locales_zh_cn' : 'region_id';
      break;
    }
    case VendorEnum.TCLOUD: {
      dataNameKey = isChinese ? 'region_name' : 'display_name';
      break;
    }
  }

  return {
    dataIdKey,
    dataNameKey,
  };
};
