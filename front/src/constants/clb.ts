// 网络类型
export const LOAD_BALANCER_TYPE = [
  {
    label: '公网',
    value: 'OPEN',
  },
  {
    label: '内网',
    value: 'INTERNAL',
  },
];
// IP版本
export const ADDRESS_IP_VERSION = [
  {
    label: 'IPV4',
    value: 'IPV4',
  },
  {
    label: 'IPV6',
    value: 'IPv6FullChain',
  },
  {
    label: 'IPV6 NAT64',
    value: 'IPV6',
    isDisabled: (region: string) => !WHITE_LIST_REGION_IPV6_NAT64.includes(region),
  },
];
// 可用区类型
export const ZONE_TYPE = [
  {
    label: '单可用区',
    value: 'single',
  },
  {
    label: '主备可用区',
    value: 'primaryStand',
    isDisabled: (region: string) => !WHITE_LIST_REGION_PRIMARY_STAND_ZONE.includes(region),
  },
];
// 网络计费模式
export const INTERNET_CHARGE_TYPE = [
  {
    label: '包月',
    value: undefined,
  },
  {
    label: '按流量',
    value: 'TRAFFIC_POSTPAID_BY_HOUR',
  },
  {
    label: '按带宽',
    value: 'BANDWIDTH_POSTPAID_BY_HOUR',
  },
  // {
  //   label: '共享带宽包',
  //   value: 'BANDWIDTH_PACKAGE',
  // },
];

// 支持IPv6 NAT64的地域
export const WHITE_LIST_REGION_IPV6_NAT64 = ['ap-beijing', 'ap-shanghai', 'ap-guangzhou'];
// 支持主备可用区的地域
export const WHITE_LIST_REGION_PRIMARY_STAND_ZONE = [
  'ap-guangzhou',
  'ap-shanghai',
  'ap-nanjing',
  'ap-beijing',
  'ap-hongkong',
  'ap-seoul',
];
