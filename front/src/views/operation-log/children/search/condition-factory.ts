import { ResourceTypeEnum } from '@/common/resource-constant';
import { getModel } from '@/model/manager';
import { SearchConditionAll } from './condition-all';
import { SearchConditionSecurityGroup } from './condition-security-group';
import { SearchConditionClb } from './condition-clb';

export class SearchConditionFactory {
  static createModel(resourceType: ResourceTypeEnum | 'all') {
    switch (resourceType) {
      case ResourceTypeEnum.CLB:
        return getModel(SearchConditionClb);
      case 'all':
        return getModel(SearchConditionAll);
      default:
        return getModel(SearchConditionSecurityGroup); // 除了CLB 与 全部，其余搜索列暂时跟安全组一样
    }
  }
}
