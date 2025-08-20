import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { VENDORS } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { useAccountStore } from '@/store';
import type { FilterType } from '@/typings/resource';
import { QueryRuleOPEnum } from '@/typings';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';

const accountStore = useAccountStore();
const cloudAreaStore = useCloudAreaStore();

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
  {
    name: '管控区域',
    id: 'bk_cloud_id',
    multiple: true,
    // 兼容home组件中的全量拉取。避免后期移除home中全量拉取管控区域的操作导致这里没有数据
    async: cloudAreaStore.cloudAreaList.length === 0,
    children: cloudAreaStore.cloudAreaList as any[],
  },
  {
    name: '操作系统',
    id: 'os_name',
  },
  {
    name: '所属VPC',
    id: 'cloud_vpc_ids',
  },
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
    page: { start: 0, limit: 50 },
  };
  const res = await accountStore.getAccountList(params);
  return res?.data?.details;
};

const getOptionMenu = async (item: ISearchItem, keyword: string): Promise<any[]> => {
  const { id, async, children = [] } = item;

  if (!async) {
    return children;
  }

  if (id === 'account_id') {
    return getAccountList(keyword);
  }

  if (id === 'bk_cloud_id') {
    return cloudAreaStore.fetchAllCloudAreas();
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
