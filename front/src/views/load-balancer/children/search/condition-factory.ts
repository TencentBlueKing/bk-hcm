import { getModel } from '@/model/manager';
import { SearchConditionClb } from './condition-clb';
import { SearchConditionListener } from './condition-listener';

export enum ConditionKeyType {
  CLB = 'clb',
  LISTENER = 'listener',
}

export class SearchConditionFactory {
  static createModel(key: ConditionKeyType) {
    switch (key) {
      case ConditionKeyType.CLB:
        return getModel(SearchConditionClb);
      case ConditionKeyType.LISTENER:
        return getModel(SearchConditionListener);
    }
  }
}
