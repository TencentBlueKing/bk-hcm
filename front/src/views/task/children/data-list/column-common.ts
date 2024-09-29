import { ModelProperty } from '@/model/typings';
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

const clbFieldIds = [...baseFieldIds];

columnIds.set(ResourceTypeEnum.CLB, clbFieldIds);

const taskViewProperties: ModelProperty[] = [
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
  return columnIds.map((id) => taskViewProperties.find((item) => item.id === id));
};

const factory = {
  getColumns,
};

export type FactoryType = typeof factory;

export default factory;
