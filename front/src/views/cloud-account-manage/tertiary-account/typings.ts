export type FlagType = 'phone' | 'token' | 'stoken' | 'wechat' | 'custom' | 'mail' | 'u2FToken';

/**
 * 腾讯云 extension 类型定义
 */
export interface ITcloudExtension {
  login_flag?: FlagType;
  action_flag?: FlagType;
  console_login?: number;
  cloud_main_account_id?: string;
  [key: string]: any;
}

/**
 * 账号类型枚举（对应 extension.console_login）
 * 0: 编程账号（无法登录控制台）
 * 1: 控制台账号（可登录控制台）
 */
export type ConsoleLoginType = 0 | 1;
