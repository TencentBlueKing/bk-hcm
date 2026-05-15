import { FlagType } from './typings';

export const ACCOUNT_TYPE_OPTIONS: Record<number, string> = {
  1: '控制台账号',
  0: '编程账号',
};

export const FLAG_OPTIONS: Record<FlagType, string> = {
  phone: '安全手机',
  token: '硬token',
  stoken: 'MFA',
  wechat: '微信',
  custom: '自定义',
  mail: '邮箱',
  u2FToken: 'u2硬件token',
};

export const SECRET_STATUS_MAP: Record<string, string> = {
  enabled: '已启用',
  disabled: '已禁用',
};
