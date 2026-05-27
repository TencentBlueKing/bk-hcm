import { VendorEnum } from '@/common/constant';
import { getModel } from '@/model/manager';
import { SearchConditionTcloud } from './condition-tcloud';

export class SearchConditionFactory {
  static createModel(vendor: VendorEnum) {
    switch (vendor) {
      case VendorEnum.TCLOUD:
        return getModel(SearchConditionTcloud);
      default:
        throw new Error(`Unsupported vendor: ${vendor}`);
    }
  }
}
