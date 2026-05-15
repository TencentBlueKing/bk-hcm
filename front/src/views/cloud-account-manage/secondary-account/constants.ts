// 二级账号管理相关常量

import { FlagType } from './typings';

// 资源纳管状态
export const RESOURCE_MANAGE_STATUS_MAP: Record<string, string> = {
  managed: '已纳管',
  unmanaged: '未纳管',
};

// 站点类型
export const SITE_TYPE_MAP: Record<string, string> = {
  china: '中国站',
  international: '国际站',
};

// 登录保护状态
export const LOGIN_PROTECTION_STATUS_MAP: Record<string, string> = {
  enabled: '已开启',
  disabled: '未开启',
};

// 操作保护状态
export const OPERATION_PROTECTION_STATUS_MAP: Record<string, string> = {
  enabled: '已开启',
  disabled: '未开启',
};

// MFA设备绑定状态
export const MFA_BINDDING_STATUS_MAP: Record<string, string> = {
  bound: '已绑定',
  unbound: '未绑定',
};

// 状态颜色配置
export const STATUS_COLOR_MAP: Record<string, string> = {
  managed: '#2dcb56', // 绿色
  unmanaged: '#ff9c01', // 橙色
  enabled: '#2dcb56', // 绿色
  disabled: '#c4c6cc', // 灰色
  bound: '#2dcb56', // 绿色
  unbound: '#c4c6cc', // 灰色
};

// 保护设置
export const FLAG_OPTIONS: Record<FlagType, string> = {
  phone: '安全手机',
  token: '硬token',
  stoken: 'MFA',
  wechat: '微信',
  custom: '自定义',
  mail: '邮箱',
  u2FToken: 'u2硬件token',
};
