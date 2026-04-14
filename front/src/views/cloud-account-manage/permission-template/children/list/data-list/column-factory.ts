import { VendorEnum } from '@/common/constant';
import { getModel } from '@/model/manager';
import { TableColumnTcloud } from './column-tcloud';

export class TableColumnFactory {
  static createModel(vendor: VendorEnum) {
    switch (vendor) {
      case VendorEnum.TCLOUD:
        return getModel(TableColumnTcloud);
      default:
        throw new Error(`Unsupported vendor: ${vendor}`);
    }
  }
}
