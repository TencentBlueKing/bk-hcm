export type FlagType = 'phone' | 'token' | 'stoken' | 'wechat' | 'custom' | 'mail' | 'u2FToken';

export interface ITcloudExtension {
  login_flag?: FlagType;
  action_flag?: FlagType;
  console_login?: number;
  cloud_main_account_id?: string;
  [key: string]: any;
}

export type ConsoleLoginType = 0 | 1;
