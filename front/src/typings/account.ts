import { VendorEnum } from '@/common/constant';

export interface ProjectModel {
  id: number;
  type: string;
  name: string;
  vendor: VendorEnum;
  managers: string[];
  departmentId?: string[] | number[];
  memo: string; // 备注
  account?: number | string;
  subAccountId?: number | string;
  subAccountName?: number | string;
  usage_biz_ids?: number[]; // 使用业务
  bk_biz_id?: number; // 管理业务
  mainAccount?: number | string; // 主账号
  subAccount?: number | string; // 子账号
  secretId?: string; // secretId
  secretKey?: string; // secretKey
  accountId?: string; // 账号id
  accountName?: string; // 账号名称
  iamUsername?: string; // IAM用户名称
  iamUserId?: string; // IAM用户Id
  price?: number; // 价格
  price_unit?: string; // 单位
  creator?: string; // 创建者
  reviser?: string;
  created_at?: string; // 创建时间
  updated_at?: string; // 修改时间
  extension?: any; // 根据每种云返回值不一定
  site?: string; // 站点
  projectId?: string; // 项目ID
  projectName?: string; // 项目名称
  tenantId?: string; // 租户id
  subScriptionId?: string; // 订阅id
  subScriptionName?: string; // 订阅名称
  applicationId?: string; // 应用程序ID
  applicationName?: string; // 应用程序名称
}

export enum StaffType {
  RTX = 'rtx',
  MAIL = 'email',
  ALL = 'all',
}
export interface Staff {
  english_name: string;
  chinese_name: string;
  username: string;
  display_name: string;
}

export interface Department {
  id: number;
  name: string;
  full_name: string;
  has_children: boolean;
  parent?: number;
  children?: Department[];
  checked?: boolean;
  indeterminate?: boolean;
  isOpen: boolean;
  loaded: boolean;
  loading: boolean;
}

export interface FormItems {
  label?: string;
  required?: boolean;
  property?: string;
  content?: Function;
  component?: Function;
  formName?: string;
  noBorBottom?: boolean;
  type?: string;
  description?: string;
  rules?: Object;
}

export interface SecretModel {
  secretId: string;
  secretKey: string;
  subAccountId?: string;
  iamUserName?: string;
  iamUserId?: string;
  accountId?: string;
  accountName?: string;
  applicationId?: string;
  applicationName?: string;
}

export enum CloudType {
  tcloud = '腾讯云',
  aws = '亚马逊',
  azure = '微软云',
  gcp = '谷歌云',
  huawei = '华为云',
}

export enum AccountType {
  resource = '资源账号',
  registration = '登记账号',
  security_audit = '安全审计账号',
}

export enum SiteType {
  china = '中国站',
  international = '国际站',
}

export interface IAccountItem {
  id: string;
  vendor: VendorEnum;
  name: string;
  managers: string[];
  type: 'resource' | 'registration' | 'security_audit';
  site: 'china' | 'international';
  price: string;
  price_unit: string;
  memo: string;
  bk_biz_ids: number[];
  sync_status: string;
  sync_failed_reason: string;
  recycle_reserve_time: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}
