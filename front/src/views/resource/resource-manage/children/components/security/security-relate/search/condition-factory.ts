import { VendorEnum } from '@/common/constant';
import conditionCommon, { type FactoryType } from './condition-common';
import conditionTcloud from './condition-tcloud';

export default function optionFactory(vendor?: Extract<VendorEnum, VendorEnum.TCLOUD>): FactoryType {
  const optionMap = {
    [VendorEnum.TCLOUD]: conditionTcloud,
  };
  return optionMap[vendor] ?? conditionCommon;
}
