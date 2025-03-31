import { VendorEnum } from '@/common/constant';

export const showClone = (vendor: string | string[]) => {
  return vendor === VendorEnum.TCLOUD;
};
