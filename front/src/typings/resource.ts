// define
export type PlainObject = {
  [k: string]: string | boolean | number
};

export type DoublePlainObject = {
  [k: string]: PlainObject
};

export type FilterType = {
  op: 'and' | 'or';
  rules: {
    field: string;
    op: 'eq';
    value: string | number | string[];
  }[]
};

export enum GcpTypeEnum {
  EGRESS = '出站',
  INGRESS = '入站',
}

export enum SecurityRuleEnum {
  ACCEPT = '接受',
  DROP = '拒绝',
}

export enum HuaweiSecurityRuleEnum {
  allow = '允许',
  deny = '拒绝',
}

export enum AzureSecurityRuleEnum {
  Allow = '允许',
  Deny = '拒绝',
}
