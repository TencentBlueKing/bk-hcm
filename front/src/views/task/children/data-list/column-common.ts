import { PropertyColumnConfig, ModelPropertyColumn } from '@/model/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import accountProperties from '@/model/account/properties';
import taskProperties from '@/model/task/properties';

const columnIds = new Map<ResourceTypeEnum, string[]>();

const baseFieldIds = [
  'created_at',
  'source',
  'operations',
  'vendor',
  'account_id',
  'creator',
  'count_total',
  'count_success',
  'count_failed',
  'state',
];

// TODO: 可以与baseFieldIds合并到一起
const columnConfig: Record<string, PropertyColumnConfig> = {
  created_at: {
    sort: true,
  },
  source: {
    sort: true,
  },
  operations: {
    sort: true,
  },
  vendor: {
    sort: true,
    defaultHidden: true,
  },
  account_id: {
    sort: true,
    defaultHidden: true,
  },
  creator: {
    sort: true,
  },
  state: {
    sort: true,
  },
};

const clbFieldIds = [...baseFieldIds];

columnIds.set(ResourceTypeEnum.CLB, clbFieldIds);

// TODO: 可以放到model中定义一个view
const taskViewProperties: ModelPropertyColumn[] = [
  ...accountProperties,
  ...taskProperties,
  { id: 'count_total', name: '总数', type: 'number' },
  { id: 'count_success', name: '成功数', type: 'number' },
  { id: 'count_failed', name: '失败数', type: 'number' },
];

export const getColumnIds = (resourceType: ResourceTypeEnum) => {
  return columnIds.get(resourceType);
};

const getColumns = (type: ResourceTypeEnum) => {
  const columnIds = getColumnIds(type);
  return columnIds.map((id) => ({
    ...taskViewProperties.find((item) => item.id === id),
    ...columnConfig[id],
  }));
};

const factory = {
  getColumns,
};

export type FactoryType = typeof factory;

export default factory;
