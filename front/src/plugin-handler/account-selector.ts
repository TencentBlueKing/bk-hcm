import { VendorEnum } from '@/common/constant';

export const filterAccountList = (accountList: any[]) => {
  return accountList.filter((item) => item.vendor === VendorEnum.TCLOUD);
};
