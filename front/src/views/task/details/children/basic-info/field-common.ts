import { ResourceTypeEnum } from '@/common/resource-constant';
import { getModel } from '@/model/manager';
import { Properties } from '@/model/task/properties';

const getFields = (_type: ResourceTypeEnum) => {
  return getModel(Properties).getProperties();
};

const factory = {
  getFields,
};

export type FactoryType = typeof factory;

export default factory;
