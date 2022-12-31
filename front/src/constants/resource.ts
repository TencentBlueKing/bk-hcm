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

export const GCP_TYPE_STATUS = [
  {
    label: '出站',
    value: 'egress',
  }, {
    label: '入站',
    value: 'ingress',
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
