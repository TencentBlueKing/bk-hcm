import { QueryRuleOPEnum } from './common';

// define
export type PlainObject = {
  [k: string]: string | boolean | number
};

export type DoublePlainObject = {
  [k: string]: PlainObject
};

export type FilterType = {
  op: 'and' | 'or' | QueryRuleOPEnum;
  rules: {
    field: string;
    op: QueryRuleOPEnum;
    value: string | number | string[] | any;
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
export enum ImageTypeEnum {
  gold = '公共镜像',
  private = '私有镜像',
  shared = '共享镜像',
}

export enum HostCloudEnum {
  PENDING = '创建中',
  LAUNCH_FAILED = '创建失败',
  RUNNING = '运行中',
  STOPPED = '关机',
  stopped = '关机',
  STARTING = '开机中',
  STOPPING = '关机中',
  REBOOTING = '重启中',
  SHUTDOWN = '停止待销毁',
  TERMINATING = '销毁中',
  running = '运行中'
}
