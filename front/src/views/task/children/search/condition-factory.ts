import { VendorEnum } from '@/common/constant';
import conditionCommon, { type FactoryType } from './condition-common';
import conditonTcloud from './condition-tcloud';

export default function optionFactory(vendor?: Extract<VendorEnum, VendorEnum.TCLOUD>): FactoryType {
  const optionMap = {
    [VendorEnum.TCLOUD]: conditonTcloud,
  };
  return optionMap[vendor] ?? conditionCommon;
}
