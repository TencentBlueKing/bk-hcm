import { VendorEnum } from '@/common/constant';

export const COMMON_STATUS = [
  {
    label: '关',
    value: false,
  },
  {
    label: '开',
    value: 'INGRESS',
  },
];

export const POLICY_STATUS = [
  {
    id: 'reject',
    name: '拒绝',
  },
  {
    id: 'agree',
    name: '同意',
  },
];

export const ACTION_STATUS = [
  {
    id: 'ACCEPT',
    name: '接受',
  },
  {
    id: 'DROP',
    name: '拒绝',
  },
];

export const HUAWEI_ACTION_STATUS = [
  {
    id: 'allow',
    name: '允许',
  },
  {
    id: 'deny',
    name: '拒绝',
  },
];

export const AZURE_ACTION_STATUS = [
  {
    id: 'Allow',
    name: '允许',
  },
  {
    id: 'Deny',
    name: '拒绝',
  },
];

export const GCP_TYPE_STATUS = [
  {
    label: '出站',
    value: 'EGRESS',
  },
  {
    label: '入站',
    value: 'INGRESS',
  },
];

export const GCP_MATCH_STATUS = [
  {
    label: '允许',
    value: 'allowed',
  },
  {
    label: '拒绝',
    value: 'denied',
  },
];

export const GCP_EXECUTION_STATUS = [
  {
    label: '已启用',
    value: true,
  },
  {
    label: '已停用',
    value: false,
  },
];

export const GCP_TARGET_LIST = [
  {
    id: 'destination_ranges',
    name: '网络中的所有实例',
  },
  {
    id: 'target_tags',
    name: '指定的目标标记',
  },
  {
    id: 'target_service_accounts',
    name: '指定的服务账号',
  },
];

export const GCP_SOURCE_LIST = [
  {
    id: 'source_ranges',
    name: 'IPv6 CIDR/IPv4 CIDR',
  },
  {
    id: 'source_tags',
    name: '来源标记',
  },
  {
    id: 'source_service_accounts',
    name: '服务账号',
  },
];

export const GCP_PROTOCOL_LIST = [
  {
    id: 'tcp',
    name: 'TCP',
  },
  {
    id: 'udp',
    name: 'UDP',
  },
  {
    id: 'icmp',
    name: 'ICMP',
  },
  {
    id: 'esp',
    name: 'ESP',
  },
  {
    id: 'ah',
    name: 'AH',
  },
  {
    id: 'ipip',
    name: 'IPIP',
  },
  {
    id: 'sctp',
    name: 'SCTP',
  },
];

export const AZURE_PROTOCOL_LIST = [
  {
    id: '*',
    name: 'ALL',
  },
  {
    id: 'Ah',
    name: 'Ah',
  },
  {
    id: 'Esp',
    name: 'Esp',
  },
  {
    id: 'Icmp',
    name: 'Icmp',
  },
  {
    id: 'Tcp',
    name: 'Tcp',
  },
  {
    id: 'Udp',
    name: 'Udp',
  },
];

export const IP_TYPE_LIST = [
  {
    id: 'ipv4_cidr',
    name: 'IPv4',
  },
  {
    id: 'ipv6_cidr',
    name: 'IPv6',
  },
];

export const TCLOUD_SOURCE_IP_TYPE_LIST = [
  {
    id: 'cloud_address_id',
    name: '参数模板-IP地址',
  },
  {
    id: 'cloud_address_group_id',
    name: '参数模板-IP地址组',
  },
];

export const HUAWEI_TYPE_LIST = [
  {
    id: 'ipv4',
    name: 'IPv4',
  },
  {
    id: 'ipv6',
    name: 'IPv6',
  },
];
export const DISTRIBUTE_STATUS_LIST = [
  {
    label: '全部',
    value: 'all',
  },
  {
    label: '未分配',
    value: -1,
  },
  {
    label: '已分配',
    value: 1,
  },
];

export const RECYCLE_BIN_ITEM_STATUS = {
  wait_recycle: '等待回收',
  recycled: '已回收',
  recovered: '已恢复',
  failed: '回收失败',
};

export const CLOUD_VENDOR = {
  tcloud: 'tcloud', // 腾讯云
  aws: 'aws', // 亚马逊云
  huawei: 'huawei', // 华为云
  azure: 'azure', // 微软云
  gcp: 'gcp', // 谷歌云
};

export const TCLOUD_SECURITY_MESSAGE = {
  protocol: '协议',
  port: '端口',
  sourceAddress: '源地址类型',
  ipv4_cidr: '源地址',
  action: '策略',
};

export const HUAWEI_SECURITY_MESSAGE = {
  priority: '优先级',
  ethertype: '类型',
  protocol: '协议',
  port: '端口',
  sourceAddress: '源地址类型',
  ipv4_cidr: '源地址',
  action: '策略',
};

export const AWS_SECURITY_MESSAGE = {
  protocol: '协议',
  port: '端口',
  sourceAddress: '源地址类型',
  ipv4_cidr: '源地址',
};

export const AZURE_SECURITY_MESSAGE = {
  name: '名称',
  priority: '优先级',
  sourceAddress: '源地址类型',
  ipv4_cidr: '源地址',
  source_port_range: '源端口',
  targetAddress: '目标地址类型',
  destination_address_prefix: '目标地址',
  protocol: '目标协议',
  access: '策略',
};

export const TCLOUD_SECURITY_RULE_PROTOCALS = [
  {
    id: 'tcp',
    name: 'TCP',
  },
  {
    id: 'udp',
    name: 'UDP',
  },
  {
    id: 'icmp',
    name: 'ICMP',
  },
  {
    id: 'icmpv6',
    name: 'ICMPv6',
  },
  {
    id: 'gre',
    name: 'GRE',
  },
  {
    id: 'cloud_service_id',
    name: '参数模板-端口',
  },
  {
    id: 'cloud_service_group_id',
    name: '参数模板-端口组',
  },
];

export const AWS_SECURITY_RULE_PEOTOCALS = [
  {
    id: 'tcp',
    name: 'TCP',
  },
  {
    id: 'udp',
    name: 'UDP',
  },
  {
    id: 'icmp',
    name: 'ICMP',
  },
  {
    id: 'icmpv6',
    name: 'ICMPv6',
  },
];

export const HUAWEI_SECURITY_RULE_PEOTOCALS = [
  {
    id: 'tcp',
    name: 'TCP',
  },
  {
    id: 'udp',
    name: 'UDP',
  },
  {
    id: 'icmp',
    name: 'ICMP',
  },
  // {
  //   id: 'gre',
  //   name: 'GRE',
  // },
];

export const AZURE_SECURITY_RULE_PEOTOCALS = [
  {
    id: 'tcp',
    name: 'TCP',
  },
  {
    id: 'udp',
    name: 'UDP',
  },
  {
    id: 'icmp',
    name: 'ICMP',
  },
  {
    id: 'any',
    name: 'ANY',
  },
];

export const SECURITY_RULES_MAP = {
  [VendorEnum.TCLOUD]: TCLOUD_SECURITY_RULE_PROTOCALS,
  [VendorEnum.AWS]: AWS_SECURITY_RULE_PEOTOCALS,
  [VendorEnum.HUAWEI]: HUAWEI_SECURITY_RULE_PEOTOCALS,
  [VendorEnum.AZURE]: AZURE_SECURITY_RULE_PEOTOCALS,
};
