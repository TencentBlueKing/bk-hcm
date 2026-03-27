import { getModel } from '@/model/manager';
import { TableColumn } from './column';

export class TableColumnFactory {
  static createModel() {
    return getModel(TableColumn);
  }
}
