import { getModel } from '@/model/manager';
import { SearchConditionClb } from './condition-clb';
import { SearchConditionListener } from './condition-listener';
import { SearchConditionUrl } from './condition-url';
import { SearchConditionRs } from './condition-rs';

export enum ConditionKeyType {
  CLB = 'clb',
  LISTENER = 'listener',
  URL = 'url',
  RS = 'rs',
}

export class SearchConditionFactory {
  static createModel(key: ConditionKeyType) {
    switch (key) {
      case ConditionKeyType.CLB:
        return getModel(SearchConditionClb);
      case ConditionKeyType.LISTENER:
        return getModel(SearchConditionListener);
      case ConditionKeyType.URL:
        return getModel(SearchConditionUrl);
      case ConditionKeyType.RS:
        return getModel(SearchConditionRs);
    }
  }
}
