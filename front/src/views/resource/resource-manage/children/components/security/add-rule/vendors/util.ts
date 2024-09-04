import { parse, parseCIDR, IPv4, isValid } from 'ipaddr.js';
import { random as _random } from 'lodash-es'

/**
 * 检查端口号是否合法
 * @param val 端口号、端口范围、多个端口
 * @returns boolean
 */
export const isPortAvailable = (val: string | number) => {
  const port = String(val).trim();
  const isPortValid = /^(ALL|([1-9]\d*|[1-9]\d*-\d+)(,([1-9]\d*|[1-9]\d*-\d+))*)$/.test(port);
  if (!isPortValid) return false;
  if (/^ALL$/.test(port)) return true;

  const rangesAndPorts = port.split(/,/);
  for (const rangeOrPort of rangesAndPorts) {
    if (/-/.test(rangeOrPort)) {
      const [start, end] = rangeOrPort.split(/-/).map(Number);
      if (start < 1 || end < 1 || start > 65535 || end > 65535 || start >= end) {
        return false;
      }
    } else {
      const num = Number(rangeOrPort);
      if (num < 1 || num > 65535) {
        return false;
      }
    }
  }

  return true;
};

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

/**
 * 生成随机id
 * @returns 随机id
 */
export const random = () => `${_random(0, 999999)}_${Date.now()}_${_random(0, 999999)}`;

/**
 * 清除对象中的空值
 * @param obj 对象
 * @returns 清除空值后的对象
 */
export const cleanObject = <T extends object>(obj: T): Partial<T> => {
  return Object.keys(obj).reduce((acc, key) => {
      const value = obj[key as keyof T];
      if (value !== "" && value !== null && value !== undefined) {
          acc[key as keyof T] = value;
      }
      return acc;
  }, {} as Partial<T>);
};
