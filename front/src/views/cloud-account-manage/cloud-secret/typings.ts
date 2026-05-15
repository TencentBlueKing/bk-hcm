export interface ICloudSecretItem {
  id: string;
  vendor: string;
  status: 'enabled' | 'disabled';
  account_id: string;
  sub_account_id: string;
  extension: {
    cloud_secret_id: string;
    cloud_main_account_id: string;
    cloud_sub_account_id: string;
    console_login?: number; // 0: 编程账号, 1: 控制台账号
  };
  tenant_id?: string;
  cloud_created_at: string;
  disabled_time?: string;
  last_used_time?: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  sub_account_managers?: string[];
  account_managers?: string[];
  account_name?: string;
  sub_account_name?: string;
  operable?: boolean;
  // 前端扩展字段 - 从 extension 中提取便于展示
  cloud_secret_id?: string;
  cloud_main_account_id?: string;
  cloud_sub_account_id?: string;
  console_login?: number;
}

export type ISearchCondition = Record<string, any>;

export type SecretActionType = 'enable' | 'disable' | 'delete';

export interface ISecretActionConfig {
  type: SecretActionType;
  title: string;
  alertType: 'warning' | 'error';
  alertMessage: string;
  alertDescription?: string;
  confirmText: string;
  confirmTheme: 'primary' | 'danger';
}
