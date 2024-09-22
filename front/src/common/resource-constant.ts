export enum ResourceTypeEnum {
  CVM = 'cvm',
  VPC = 'vpc',
  DISK = 'disk',
  SUBNET = 'subnet',
  CLB = 'clb',
  ACCOUNT = 'account',
}

export type ResourcePropertyType = 'string' | 'datetime' | 'enum' | 'number' | 'account' | 'user';

export type ResourceProperty = {
  id: string;
  name: string;
  type: ResourcePropertyType;
  resource?: ResourceTypeEnum;
  option?: Record<string, any>;
  index?: number;
};
