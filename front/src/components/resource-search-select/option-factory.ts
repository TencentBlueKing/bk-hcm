import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { ResourceTypeEnum } from '@/common/resource-constant';
import optionCommon from './option-common';
import optionSecurity, { type SecuritySubType } from './option-security';
import optionClb from './option-clb';

export type GetMenuListFunc = (item: ISearchItem, keyword: string) => Promise<any[]>;

export interface OptionModule {
  getOptionData: (...args: any[]) => ISearchItem[];
  getOptionMenu: GetMenuListFunc;
}

/**
 * 根据 resourceType + subType 分派到对应的 option 模块
 */
export default function optionFactory(resourceType: ResourceTypeEnum, subType?: string): OptionModule {
  switch (resourceType) {
    case ResourceTypeEnum.SECURITY_GROUP:
      return {
        getOptionData: () => optionSecurity.getOptionData((subType as SecuritySubType) || 'group'),
        getOptionMenu: optionSecurity.getOptionMenu,
      };
    case ResourceTypeEnum.CLB:
      return {
        getOptionData: () => optionClb.getOptionData(),
        getOptionMenu: optionClb.getOptionMenu,
      };
    default:
      return {
        getOptionData: () => optionCommon.getOptionData(resourceType) || [],
        getOptionMenu: optionCommon.getOptionMenu,
      };
  }
}
