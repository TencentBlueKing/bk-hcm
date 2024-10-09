import { VendorEnum } from '@/common/constant';
import columnCommon, { type FactoryType } from './column-common';
import columnTcloud from './column-tcloud';

export default function optionFactory(vendor?: Extract<VendorEnum, VendorEnum.TCLOUD>): FactoryType {
  const optionMap = {
    [VendorEnum.TCLOUD]: columnTcloud,
  };
  return optionMap[vendor] ?? columnCommon;
}
