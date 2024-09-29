import type { ResourceTypeEnum } from '@/common/resource-constant';

export type ModelPropertyType = 'string' | 'datetime' | 'enum' | 'number' | 'account' | 'user';

export type ModelProperty = {
  id: string;
  name: string;
  type: ModelPropertyType;
  resource?: ResourceTypeEnum;
  option?: Record<string, any>;
  index?: number;
};
