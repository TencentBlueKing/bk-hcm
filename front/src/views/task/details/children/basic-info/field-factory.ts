import { VendorEnum } from '@/common/constant';
import fieldCommon, { type FactoryType } from './field-common';
import fieldTcloud from './field-tcloud';

export default function optionFactory(vendor?: Extract<VendorEnum, VendorEnum.TCLOUD>): FactoryType {
  const optionMap = {
    [VendorEnum.TCLOUD]: fieldTcloud,
  };
  return optionMap[vendor] ?? fieldCommon;
}
