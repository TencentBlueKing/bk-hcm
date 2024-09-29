import { ResourceTypeEnum } from '@/common/resource-constant';
import accountProperties from '@/model/account/properties';
import taskProperties from '@/model/task/properties';

const conditionFieldIds = new Map<ResourceTypeEnum, string[]>();
const baseFieldIds = ['account_id', 'operations', 'state', 'source', 'created_at', 'creator'];
const clbFieldIds = [...baseFieldIds];
conditionFieldIds.set(ResourceTypeEnum.CLB, clbFieldIds);

const taskViewProperties = [...accountProperties, ...taskProperties];

export const getConditionFieldIds = (resourceType: ResourceTypeEnum) => {
  return conditionFieldIds.get(resourceType);
};

const getConditionField = (type: ResourceTypeEnum) => {
  const fieldIds = getConditionFieldIds(type);
  const fields = fieldIds.map((id) => taskViewProperties.find((item) => item.id === id));
  return fields;
};

const factory = {
  getConditionField,
};

export type FactoryType = typeof factory;

export default factory;
