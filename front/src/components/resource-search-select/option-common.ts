import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { VENDORS } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { useAccountStore } from '@/store';
import type { FilterType } from '@/typings/resource';
import { QueryRuleOPEnum } from '@/typings';

const accountStore = useAccountStore();

const optionMap = new Map<ResourceTypeEnum, ISearchItem[]>();

export const base: ISearchItem[] = [
  {
    name: '名称',
    id: 'name',
  },
  {
    name: '云厂商',
    id: 'vendor',
    multiple: true,
    children: VENDORS,
  },
  {
    name: '云账号ID',
    id: 'account_id',
    async: true,
    multiple: true,
    children: [],
  },
];

export const cvm: ISearchItem[] = [
  {
    name: '内网IP',
    id: 'private_ip',
  },
  {
    name: '公网IP',
    id: 'public_ip',
  },
  {
    name: '主机ID',
    id: 'cloud_id',
  },
  ...base,
  ...[
    {
      name: '管控区域',
      id: 'bk_cloud_id',
    },
    {
      name: '操作系统',
      id: 'os_name',
    },
    {
      name: '所属VPC',
      id: 'cloud_vpc_ids',
    },
  ],
];

optionMap.set(ResourceTypeEnum.CVM, cvm);

export const getAccountList = async (keyword: string) => {
  const query: FilterType = {
    op: 'and',
    rules: [{ field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' }],
  };
  if (keyword) {
    query.rules.push({ field: 'name', op: QueryRuleOPEnum.CS, value: keyword });
  }
  const params = {
    filter: query,
    page: {
      start: 0,
      limit: 50,
    },
  };
  const res = await accountStore.getAccountList(params);
  return res?.data?.details;
};

const getOptionMenu = async (item: ISearchItem, keyword: string): Promise<ISearchItem[]> => {
  const { id, async, children = [] } = item;

  if (!async) {
    return children;
  }

  if (id === 'account_id') {
    return getAccountList(keyword);
  }
};

const getOptionData = (type: ResourceTypeEnum) => {
  return optionMap.get(type);
};

const factory = {
  getOptionData,
  getOptionMenu,
};

export type FactoryType = typeof factory;

export default factory;
