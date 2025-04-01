import { VendorEnum } from '@/common/constant';

export const showClone = (vendor: string) => {
  return vendor === VendorEnum.TCLOUD;
};
