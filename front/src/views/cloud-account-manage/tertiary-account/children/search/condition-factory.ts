import { getModel } from '@/model/manager';
import { SearchCondition } from './condition';

export class SearchConditionFactory {
  static createModel() {
    return getModel(SearchCondition);
  }
}
