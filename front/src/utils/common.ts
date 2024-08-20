import isIP from 'validator/es/lib/isIP';

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
  } = inst;
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

export { getInstVip, splitIP, parseIP, encodeValueByBtoa, decodeValueByAtob };
