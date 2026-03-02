// 三级账号密钥项接口定义
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
  sub_account_manager?: string;
  account_manager?: string;
  // 前端扩展字段 - 从 extension 中提取便于展示
  cloud_secret_id?: string;
  cloud_main_account_id?: string;
  cloud_sub_account_id?: string;
  console_login?: number;
}

// 搜索条件类型
export type ISearchCondition = Record<string, any>;

// 操作类型
export type SecretActionType = 'enable' | 'disable' | 'delete';

// 操作弹窗配置
export interface ISecretActionConfig {
  type: SecretActionType;
  title: string;
  alertType: 'warning' | 'error';
  alertMessage: string;
  alertDescription?: string;
  confirmText: string;
  confirmTheme: 'primary' | 'danger';
}
