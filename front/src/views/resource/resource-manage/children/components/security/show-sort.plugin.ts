import { VendorEnum } from '@/common/constant';

export const showSort = (vendor: string | string[]) => {
  return vendor === VendorEnum.TCLOUD;
};
