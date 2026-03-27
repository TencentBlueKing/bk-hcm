// 三级账号管理相关常量

import { FlagType } from './typings';

// 账号类型选项（对应 extension.console_login 字段：1=控制台账号，0=编程账号）
export const ACCOUNT_TYPE_OPTIONS: Record<number, string> = {
  1: '控制台账号',
  0: '编程账号',
};

// 保护设置
export const FLAG_OPTIONS: Record<FlagType, string> = {
  phone: '安全手机',
  token: '硬token',
  stoken: 'MFA字段',
  wechat: '微信',
  custom: '自定义',
  mail: '邮箱',
  u2FToken: 'u2硬件token',
};

// 密钥状态
export const SECRET_STATUS_MAP: Record<string, string> = {
  enabled: '已启用',
  disabled: '已禁用',
};
