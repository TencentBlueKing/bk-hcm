import { getModel } from '@/model/manager';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { ListView } from '@/model/task/list.view';

const columnIds = new Map<ResourceTypeEnum, string[]>();

export const getColumnIds = (resourceType: ResourceTypeEnum) => {
  return columnIds.get(resourceType);
};

const getColumns = (_type: ResourceTypeEnum) => {
  return getModel(ListView).getProperties();
};

const factory = {
  getColumns,
};

export type FactoryType = typeof factory;

export default factory;
