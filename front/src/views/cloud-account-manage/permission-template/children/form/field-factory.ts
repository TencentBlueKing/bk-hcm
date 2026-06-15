import { VendorEnum } from '@/common/constant';
import { getModel } from '@/model/manager';
import { FieldTcloud } from './field-tcloud';

export class FieldFactory {
  static createModel(vendor: VendorEnum) {
    switch (vendor) {
      case VendorEnum.TCLOUD:
        return getModel(FieldTcloud);
      default:
        throw new Error(`Unsupported vendor: ${vendor}`);
    }
  }
}
