export const COMMON_STATUS = [
  {
    label: '关',
    value: false,
  }, {
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

export const GCP_TYPE_STATUS = [
  {
    label: '出站',
    value: 'EGRESS',
  }, {
    label: '入站',
    value: 'INGRESS',
  },
];

export const GCP_MATCH_STATUS = [
  {
    label: '允许',
    value: 'allowed',
  }, {
    label: '拒绝',
    value: 'denied',
  },
];

export const GCP_EXECUTION_STATUS = [
  {
    label: '已启用',
    value: true,
  }, {
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
    label: '未分配',
    value: -1,
  },
  {
    label: '已分配',
    value: 1,
  },
  {
    label: '全部',
    value: 0,
  },
];
