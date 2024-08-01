import { VendorEnum } from '@/common/constant';
import { type FactoryType } from './option-common';

export default async function optionFactory(vendor?: VendorEnum): Promise<FactoryType> {
  const { factory } = await import(`./option-${vendor ?? 'common'}.ts`);
  return factory;
}
