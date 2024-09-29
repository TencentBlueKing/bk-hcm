import { cloneDeep, isArray } from 'lodash';
import {
  AzureSecurityGroupRule,
  AzureSourceAddressType,
  AzureSourceTypeArr,
  AzureTargetAddressType,
  AzureTargetTypeArr,
} from '.';

export const AZURE_PROTOCOL_ALL = '*';
export const AZURE_PROTOCOL_ICMP = 'Icmp';
export const AZURE_PORT_ALL = '*';

export const azureHandler = (data: AzureSecurityGroupRule & { port: string }): AzureSecurityGroupRule => {
  delete data.key;
  if ([AZURE_PROTOCOL_ALL, AZURE_PROTOCOL_ICMP].includes(data.protocol)) data.port = AZURE_PORT_ALL;
  if (data.destination_port_range === 'ALL') data.destination_port_range = AZURE_PORT_ALL;
  return data;
};

export const azurePreHandler = (
  data: AzureSecurityGroupRule & { sourceAddress: AzureSourceAddressType; targetAddress: AzureTargetAddressType },
) => {
  const res: any = cloneDeep(data);
  AzureSourceTypeArr.forEach((type) => res[type] && (res.sourceAddress = type));
  AzureTargetTypeArr.forEach((type) => res[type] && (res.targetAddress = type));
  if (res.destination_port_range === AZURE_PORT_ALL) res.destination_port_range = 'ALL';
  return res;
};
