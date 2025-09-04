import { VendorEnum } from '@/common/constant';

export const VendorAccountNameMap = {
  [VendorEnum.AWS]: 'cloud_main_account_name',
  [VendorEnum.GCP]: 'cloud_project_name',
  [VendorEnum.AZURE]: 'cloud_subscription_name',
  [VendorEnum.HUAWEI]: 'cloud_main_account_name',
  [VendorEnum.ZENLAYER]: 'cloud_main_account_name',
  [VendorEnum.KAOPU]: 'cloud_main_account_name',
};
