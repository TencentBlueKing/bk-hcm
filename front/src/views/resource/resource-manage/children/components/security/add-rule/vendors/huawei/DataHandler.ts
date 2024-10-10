import { HuaweiSecurityGroupRule } from '.';

export enum HuaweiProtocolEnum {
  ALL = 'huaweiAll',
  ICMP = 'icmp',
}

export const huaweiHandler = (data: HuaweiSecurityGroupRule) => {
  if ([HuaweiProtocolEnum.ALL, HuaweiProtocolEnum.ICMP].includes(data.protocol as HuaweiProtocolEnum)) delete data.port;
  return data;
};

export const huaweiPreHandler = (data: HuaweiSecurityGroupRule) => {
  return data;
};