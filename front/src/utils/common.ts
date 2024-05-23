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

export { getInstVip };
