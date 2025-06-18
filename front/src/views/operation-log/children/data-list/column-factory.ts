import { ResourceTypeEnum } from '@/common/resource-constant';
import { getModel } from '@/model/manager';
import { TableColumnAll } from './column-all';

export class TableColumnFactory {
  static createModel(_resourceType: ResourceTypeEnum) {
    return getModel(TableColumnAll);
  }
}
