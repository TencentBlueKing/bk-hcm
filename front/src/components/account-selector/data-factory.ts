import { VendorEnum } from '@/common/constant';
import dataCommon, { type FactoryType } from './data-common';

export default function optionFactory(vendor?: VendorEnum): FactoryType {
  const optionMap: { [K in VendorEnum]?: FactoryType } = {};
  return optionMap[vendor] ?? dataCommon;
}
