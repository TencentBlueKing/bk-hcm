import { ResourceTypeEnum } from '@/common/resource-constant';
import { COMMON_PROPERTIES, TASK_BASE_PROPERIES } from '@/views/task/constants';

const conditionFieldIds = new Map<ResourceTypeEnum, string[]>();
const baseFieldIds = ['account_id', 'operations', 'state', 'source', 'created_at', 'creator'];
const clbFieldIds = [...baseFieldIds];
conditionFieldIds.set(ResourceTypeEnum.CLB, clbFieldIds);

const taskColumns = [...COMMON_PROPERTIES, ...TASK_BASE_PROPERIES];

export const getConditionFieldIds = (resourceType: ResourceTypeEnum) => {
  return conditionFieldIds.get(resourceType);
};

const getConditionField = (type: ResourceTypeEnum) => {
  const fieldIds = getConditionFieldIds(type);
  const fields = fieldIds.map((id) => taskColumns.find((item) => item.id === id));
  return fields;
};

const factory = {
  getConditionField,
};

export type FactoryType = typeof factory;

export default factory;
