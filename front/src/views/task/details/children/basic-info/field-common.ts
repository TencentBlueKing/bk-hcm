import { ResourceTypeEnum } from '@/common/resource-constant';
import accountProperties from '@/model/account/properties';
import taskProperties from '@/model/task/properties';

const conditionFieldIds = new Map<ResourceTypeEnum, string[]>();
const baseFieldIds = ['vendor', 'account_id', 'creator', 'operations', 'created_at', 'state'];
const clbFieldIds = [...baseFieldIds];
conditionFieldIds.set(ResourceTypeEnum.CLB, clbFieldIds);

const taskViewProperties = [...accountProperties, ...taskProperties];

export const getFieldIds = (resourceType: ResourceTypeEnum) => {
  return conditionFieldIds.get(resourceType);
};

const getFields = (type: ResourceTypeEnum) => {
  const fieldIds = getFieldIds(type);
  const fields = fieldIds.map((id) => taskViewProperties.find((item) => item.id === id));
  return fields;
};

const factory = {
  getFields,
};

export type FactoryType = typeof factory;

export default factory;
