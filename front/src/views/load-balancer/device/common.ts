import { LB_NETWORK_TYPE_MAP } from '@/constants';
import { TARGET_GROUP_PROTOCOLS, VendorEnum } from '@/common/constant';

export enum DeviceTabEnum {
  LISTENER = 'listener',
  URL = 'url',
  RS = 'rs',
}

export interface ICount {
  listenerCount: number;
  urlCount: number;
  rsCount: number;
}

// 条件类型
export interface ILoadBalanceDeviceCondition {
  vendor: VendorEnum;
  account_id: string;
  lb_regions?: string[];
  lb_network_types?: string[];
  lb_ip_versions?: string[];
  cloud_lb_ids?: string[];
  lb_vips?: string[];
  lb_domains?: string[];
  lbl_protocols?: string[];
  lbl_ports?: number[];
  rule_domains?: string[];
  rule_urls?: string[];
  target_ips?: string[];
  target_ports?: number[];
  [key: string]: any;
}
export const numberField = ['lbl_ports', 'target_ports'];

export const selectField = [
  {
    id: 'lb_network_types',
    name: '网络类型',
    list: Object.keys(LB_NETWORK_TYPE_MAP).map((lbType) => ({
      value: lbType,
      label: LB_NETWORK_TYPE_MAP[lbType as keyof typeof LB_NETWORK_TYPE_MAP],
    })),
  },
  {
    id: 'ip_version',
    name: 'IP版本',
    list: [
      { value: 'ipv4', label: 'IPv4' },
      { value: 'ipv6', label: 'IPv6' },
      { value: 'ipv6_dual_stack', label: 'IPv6DualStack' },
      { value: 'ipv6_nat64', label: 'IPv6Nat64' },
    ],
  },
  {
    id: 'lbl_protocols',
    name: '监听器协议',
    list: TARGET_GROUP_PROTOCOLS.map((item) => ({ value: item, label: item })),
  },
];

export const inputField = [
  {
    id: 'lbl_ports',
    name: '监听器端口',
  },
  {
    id: 'target_ips',
    name: 'RS IP',
  },
  {
    id: 'lb_vips',
    name: '负载均衡 VIP',
  },
  {
    id: 'cloud_lb_ids',
    name: '负载均衡 ID',
  },
  {
    id: 'lb_domains',
    name: '负载均衡域名',
  },
  {
    id: 'rule_domains',
    name: 'HTTP/HTTPS监听器域名',
  },
  {
    id: 'rule_urls',
    name: 'URL路径',
  },
];
