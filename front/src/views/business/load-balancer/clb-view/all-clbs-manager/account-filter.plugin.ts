import { IAccountItem } from '@/typings';
import { VendorEnum } from '@/common/constant';

export const accountFilter = (list: IAccountItem[]) => {
  return list.filter((item) => item.vendor === VendorEnum.TCLOUD);
};
