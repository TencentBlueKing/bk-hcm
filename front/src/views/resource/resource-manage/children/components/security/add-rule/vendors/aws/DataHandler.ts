import { cloneDeep } from 'lodash';
import { AwsSecurityGroupRule, AwsSourceAddressType, AwsSourceTypeArr } from '.';

const SPLIT_SIGN = '-';

export const AWS_PORT_ALL = 'ALL';
export enum AWS_PROTOCOL {
  ALL = '-1',
  ICMP = 'icmp',
  ICMPv6 = 'icmpv6',
}

export const awsHandler = (
  data: AwsSecurityGroupRule & {
    port: string;
  },
): AwsSecurityGroupRule => {
  let res = {} as typeof data;

  if (data.port.includes(SPLIT_SIGN)) {
    const [from_port, to_port] = data.port.split(SPLIT_SIGN).map((v) => +v);
    res = Object.assign(data, res, { from_port, to_port });
  } else if (data.port === AWS_PORT_ALL) {
    res.from_port = -1;
    res.to_port = -1;
    res = Object.assign(data, res);
  } else {
    res.from_port = +data.port;
    res.to_port = +data.port;
    res = Object.assign(data, res);
  }

  delete res.port;
  delete res.key;

  return res;
};

export const awsPreHandler = (data: AwsSecurityGroupRule & { sourceAddress: AwsSourceAddressType; port: string }) => {
  const res = cloneDeep(data);
  AwsSourceTypeArr.forEach((type) => res[type] && (res.sourceAddress = type));
  if (res.from_port) res.port = String(res.from_port);
  if (res.protocol === AWS_PROTOCOL.ALL) res.port = AWS_PORT_ALL;
  else if (res.port === AWS_PROTOCOL.ALL) res.port = AWS_PORT_ALL;
  return res;
};
