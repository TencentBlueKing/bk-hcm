import { ResourceTypeEnum } from '@/common/resource-constant';
import { getModel } from '@/model/manager';
import { SearchConditionAll } from './condition-all';
import { SearchConditionSecurityGroup } from './condition-security-group';
import { SearchConditionClb } from './condition-clb';

export class SearchConditionFactory {
  static createModel(resourceType: ResourceTypeEnum) {
    switch (resourceType) {
      case ResourceTypeEnum.SECURITY_GROUP:
        return getModel(SearchConditionSecurityGroup);
      case ResourceTypeEnum.CLB:
        return getModel(SearchConditionClb);
      default:
        return getModel(SearchConditionAll);
    }
  }
}
