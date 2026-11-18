import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { VENDORS } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { MGMT_TYPE_MAP, SecurityGroupManageType } from '@/constants/security-group';
import { useAccountStore } from '@/store';
import { useRegionStore } from '@/store/region';
import type { FilterType } from '@/typings/resource';
import { QueryRuleOPEnum } from '@/typings';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import { useBusinessGlobalStore } from '@/store/business-global';

const accountStore = useAccountStore();
const cloudAreaStore = useCloudAreaStore();
const businessGlobalStore = useBusinessGlobalStore();
const regionStore = useRegionStore();

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
  {
    name: '固资号',
    id: 'bk_asset_id',
    multiple: true,
    children: [],
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

export const securityGroup: ISearchItem[] = [
  {
    name: '安全组ID',
    id: 'cloud_id',
  },
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
  {
    name: '使用业务',
    id: 'usage_biz_id',
    async: businessGlobalStore.businessFullList.length === 0,
    children: businessGlobalStore.businessFullList as any[],
  },
  {
    name: '管理类型',
    id: 'mgmt_type',
    multiple: true,
    children: [
      { id: SecurityGroupManageType.BIZ, name: MGMT_TYPE_MAP[SecurityGroupManageType.BIZ] },
      { id: SecurityGroupManageType.PLATFORM, name: MGMT_TYPE_MAP[SecurityGroupManageType.PLATFORM] },
      { id: SecurityGroupManageType.UNKNOWN, name: MGMT_TYPE_MAP[SecurityGroupManageType.UNKNOWN] },
    ],
  },
  {
    name: '管理业务',
    id: 'mgmt_biz_id',
    async: businessGlobalStore.businessFullList.length === 0,
    children: businessGlobalStore.businessFullList as any[],
  },
  {
    name: '地域',
    id: 'region',
    async: true,
    children: [],
    placeholder: '请输入地域名',
    onlyRecommendChildren: true,
  },
];

optionMap.set(ResourceTypeEnum.SECURITY_GROUP, securityGroup);

export const gcpFirewall: ISearchItem[] = [
  {
    name: '防火墙ID',
    id: 'cloud_id',
  },
  {
    name: '名称',
    id: 'name',
  },
  {
    name: '云账号ID',
    id: 'account_id',
    async: true,
    multiple: true,
    children: [],
  },
];

optionMap.set(ResourceTypeEnum.GCP_FIREWALL, gcpFirewall);

export const argumentTemplate: ISearchItem[] = [
  {
    name: '模板ID',
    id: 'cloud_id',
  },
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

optionMap.set(ResourceTypeEnum.ARGUMENT_TEMPLATE, argumentTemplate);

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

  if (id === 'bk_biz_id') {
    const list = await businessGlobalStore.getBusinessFullList();
    const options = list.map((biz) => ({ id: String(biz.id), name: biz.name }));
    if (!keyword) {
      return options;
    }
    const kw = keyword.toLowerCase();
    return options.filter((opt) => opt.name.toLowerCase().includes(kw) || opt.id.includes(keyword));
  }

  if (id === 'usage_biz_id' || id === 'mgmt_biz_id') {
    const list = await businessGlobalStore.getBusinessFullList();
    const options = list.map(({ id: bizId, name }) => ({ id: bizId, name }));
    if (!keyword) {
      return options;
    }
    return options.filter(({ name }) => name.includes(keyword));
  }

  if (id === 'region') {
    return regionStore.getAllVendorRegion(keyword);
  }

  return [];
};

const getOptionData = (type: ResourceTypeEnum) => {
  return optionMap.get(type) ?? [];
};

const factory = {
  getOptionData,
  getOptionMenu,
};

export type FactoryType = typeof factory;

export default factory;
