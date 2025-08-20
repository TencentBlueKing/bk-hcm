import { parse, parseCIDR, IPv4, isValid } from 'ipaddr.js';

/**
 * 检查是否合法的 IP CIDR
 * @param ip IP CIDR
 * @returns ip 类型，不合法则返回 'invalid'
 */
export const validateIpCidr = (ip: string): IpType => {
  ip = ip?.trim();
  if (isValid(ip)) {
    const type = parse(ip).kind();
    if (type === IpType.ipv4 && IPv4.isValidFourPartDecimal(ip)) return IpType.ipv4;
    if (type === IpType.ipv6) return IpType.ipv6;
    return IpType.invalid;
  }
  try {
    const [host, _mask] = parseCIDR(ip);
    const host_type = host.kind();
    if (host_type === IpType.ipv4) return IpType.ipv4_cidr;
    if (host_type === IpType.ipv6) return IpType.ipv6_cidr;
  } catch (err) {
    return IpType.invalid;
  }
  return IpType.cidr;
};

export enum IpType {
  invalid = 'invalid',
  ipv4 = 'ipv4',
  ipv6 = 'ipv6',
  cidr = 'cidr',
  ipv4_cidr = 'ipv4_cidr',
  ipv6_cidr = 'ipv6_cidr',
}
