import isIP from 'validator/es/lib/isIP';
import { AddressDescription } from '@/typings';
import { IAuthSign } from '@/common/auth-service';

const getAuthSignByBusinessId = (
  businessId: number,
  rscAuthSymbol: symbol,
  bizAuthSymbol: symbol,
): IAuthSign | IAuthSign[] => {
  if (businessId) return { type: bizAuthSymbol, relation: [businessId] };
  return { type: rscAuthSymbol };
};

/**
 * 获取实例的ip地址
 * @param inst 实例
 * @returns 实例的ip地址
 */
const getInstVip = (inst: any) => {
  const {
    private_ipv4_addresses,
    private_ipv6_addresses,
    public_ipv4_addresses,
    public_ipv6_addresses,
    private_ip_address,
    public_ip_address,
  } = inst ?? {};
  if (private_ipv4_addresses || private_ipv6_addresses || public_ipv4_addresses || public_ipv6_addresses) {
    if (public_ipv4_addresses.length > 0) return public_ipv4_addresses.join(',');
    if (public_ipv6_addresses.length > 0) return public_ipv6_addresses.join(',');
    if (private_ipv4_addresses.length > 0) return private_ipv4_addresses.join(',');
    if (private_ipv6_addresses.length > 0) return private_ipv6_addresses.join(',');
  }
  if (private_ip_address || public_ip_address) {
    if (private_ip_address.length > 0) return private_ip_address.join(',');
    if (public_ip_address.length > 0) return public_ip_address.join(',');
  }

  return '--';
};

const getPrivateIPs = (data: any) => {
  return [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--';
};
const getPublicIPs = (data: any) => {
  return [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--';
};

/**
 * 按内置分隔符切割IP文本
 * @param raw 原始文本
 * @returns 切割后的列表
 */
const splitIP = (raw: string): string[] => {
  const list: string[] = [];
  raw
    .trim()
    .split(/\n|;|；|,|，|\|/)
    .forEach((text) => {
      const ip = text.trim();
      ip.length && list.push(ip);
    });
  return list;
};

/**
 * 从文本中解析出IP地址
 * @param text IP文本
 * @returns IPv4与IPv6地址列表
 */
const parseIP = (text: string) => {
  const list = splitIP(text);
  const IPv4List: string[] = [];
  const IPv6List: string[] = [];

  list.forEach((text) => {
    if (isIP(text, 4)) {
      IPv4List.push(text);
    } else if (isIP(text, 6)) {
      IPv6List.push(text);
    }
  });

  return {
    IPv4List,
    IPv6List,
  };
};

// 将值进行btoa编码
const encodeValueByBtoa = (v: any) => btoa(JSON.stringify(v));
// 获取atob解码后的值
const decodeValueByAtob = (v: string) => JSON.parse(atob(v));

/**
 * 从文本（单个IP、CIDR 网段、连续地址段）中解析出IP地址和备注
 * @param text 单个IP、CIDR 网段、连续地址段的IP文本
 * @returns IP地址列表
 */
const analysisIP = (text: string): AddressDescription[] => {
  const list: AddressDescription[] = [];
  // 通过换行符来分割字符串
  const lines = text.split('\n');
  // 判断每一行的情况（单个IP、CIDR 网段、连续地址段）
  lines.forEach((text) => {
    // 剔除备注
    const parts = text.split(/\s+/);
    const description = parts.length >= 2 ? parts.slice(1).join(' ') : '';
    if (isSingleIP(parts[0]) || isCIDR(parts[0]) || isRange(parts[0])) {
      // 1. 单个IP    2. CIDR 网段     // 3. 连续地址段
      list.push({ address: parts[0], description });
    }
  });
  return list;
};

const isIpsValid = (text: string) => {
  // 全部行数
  const lines = text.split('\n').filter((element) => element !== '');
  if (lines.length > analysisIP(text).length) {
    return false;
  }
  return true;
};
// 判断是否为单个IP
const isSingleIP = (ip: string) => {
  return isIP(ip, 4) || isIP(ip, 6);
};
// 判断是否为CIDR网段
const isCIDR = (cidr: string) => {
  const parts = cidr.split('/');
  if (parts.length !== 2) {
    return false;
  }
  const [ip, prefix] = parts;
  if (isIP(ip, 4)) {
    const prefixNum = parseInt(prefix, 10);
    return prefixNum >= 0 && prefixNum <= 32;
  }
  if (isIP(ip, 6)) {
    const prefixNum = parseInt(prefix, 10);
    return prefixNum >= 0 && prefixNum <= 128;
  }
  return false;
};
// 判断是否为连续地址段
const isRange = (range: string) => {
  const parts = range.split('-');
  if (parts.length !== 2) {
    return false;
  }
  const [startIP, endIP] = parts;
  return (isIP(startIP, 4) && isIP(endIP, 4)) || (isIP(startIP, 6) && isIP(endIP, 6));
};

/**
 * 从文本（单个端口、多个离散端口、连续端口、所有端口）中解析出协议端口和备注
 * @param text 单个端口、多个离散端口、连续端口、所有端口的协议端口文本
 * @returns 协议端口列表
 */
const analysisPort = (port: string) => {
  // 判断是否为合法端口
  function isPortNumber(port: string) {
    // 使用正则表达式检查字符串是否只包含数字
    const isNumeric = /^\d+$/.test(port);
    if (!isNumeric) {
      return false;
    }
    const portNumber = parseInt(port, 10);
    return !isNaN(portNumber) && portNumber > 0 && portNumber <= 65535;
  }
  // 判断是否为多个离散端口方案
  function isDispersedPort(port: string) {
    if (!port.includes(',')) return false;
    const ports = port.split(',');
    return ports.every(isPortNumber);
  }
  // 判断是否为连续端口方案
  function isContinuityPort(port: string) {
    const rangeParts = port.split('-');
    if (rangeParts.length !== 2) {
      // 端口范围应该只有两个部分
      return false;
    }
    return rangeParts.every(isPortNumber);
  }

  const list: AddressDescription[] = [];
  const protocolArray = ['tcp', 'TCP', 'UDP', 'udp'];
  const protocolSpecial = ['ICMP', 'icmp', 'GRE', 'gre'];
  // 通过换行符来分割字符串
  const lines = port.split('\n');
  lines.forEach((text) => {
    // 剔除备注
    const parts = text.split(/\s+/);
    const description = parts.length >= 2 ? parts.slice(1).join(' ') : '';
    const portArr = parts[0].trim().split(':');
    if (portArr.length === 2) {
      const [protocol, port] = portArr;
      if (protocolArray.includes(protocol)) {
        if (isPortNumber(port) || ['all', 'ALL'].includes(port) || isDispersedPort(port) || isContinuityPort(port)) {
          // 1. 单个端口   // 2. 多个离散端口  // 3. 连续端口  // 4. 所有端口
          list.push({
            address: parts[0],
            description,
          });
        }
      }
    } else {
      const [protocol] = portArr;
      if (protocolSpecial.includes(protocol)) {
        list.push({
          address: parts[0],
          description,
        });
      }
    }
  });
  return list;
};
const isPortValid = (text: string) => {
  // 全部行数
  const lines = text.split('\n').filter((element) => element !== '');

  if (lines.length > analysisPort(text).length) {
    return false;
  }
  return true;
};
export {
  getAuthSignByBusinessId,
  getInstVip,
  getPrivateIPs,
  getPublicIPs,
  splitIP,
  parseIP,
  encodeValueByBtoa,
  decodeValueByAtob,
  analysisIP,
  analysisPort,
  isIpsValid,
  isPortValid,
};
