import { VendorEnum } from '@/common/constant';
import optionCommon, { type FactoryType } from './option-common';
import optionAws from './option-aws';

export default function optionFactory(vendor?: Extract<VendorEnum, VendorEnum.AWS>): FactoryType {
  const optionMap = {
    [VendorEnum.AWS]: optionAws,
  };
  return optionMap[vendor] ?? optionCommon;
}
