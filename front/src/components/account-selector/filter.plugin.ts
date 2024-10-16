import { IAccountItem } from '@/typings';
import { VendorEnum } from '@/common/constant';
import Filter from './filter.class';

class FilterPlugin extends Filter {
  filterfn(value: IAccountItem): boolean {
    return value.vendor === VendorEnum.TCLOUD;
  }
}

export default new FilterPlugin();
