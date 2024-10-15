import { ColumnConfig, ModelPropertyColumn } from '@/model/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import accountProperties from '@/model/account/properties';
import taskProperties from '@/model/task/properties';

const columnIds = new Map<ResourceTypeEnum, string[]>();

const baseFieldIds = [
  'created_at',
  'vendor',
  'account_id',
  'source',
  'operations',
  'creator',
  'count_total',
  'count_success',
  'count_failed',
  'state',
];

const columnConfig: Record<string, ColumnConfig> = {
  created_at: {
    sort: true,
  },
  vendor: {
    sort: true,
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
