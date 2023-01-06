export interface ProjectModel {
  id: number,
  type: string
  name: string,
  vendor: string,
  managers: string[]
  departmentId?: string[] | number[],
  memo: string,             // 备注
  account?: number | string,
  subAccountId?: number | string,
  subAccountName?: number | string,
  bizIds?: string | string[] | number[] // 使用业务
  mainAccount?: number | string,        // 主账号
  subAccount?: number | string,         // 子账号
  secretId?: string,             // secretId
  secretKey?: string,            // secretKey
  accountId?: string            // 账号id
  iamUsername?: string          // IAM用户名称
  price?: number      // 价格
  price_unit?: string   // 单位
  creator?: string    // 创建者
  reviser?: string
  created_at?: string // 创建时间
  updated_at?: string // 修改时间
  extension?: any     // 根据每种云返回值不一定
  site?: string     //
}

export enum StaffType {
  RTX = 'rtx',
  MAIL = 'email',
  ALL = 'all',
}
export interface Staff {
  english_name: string
  chinese_name: string
  username: string
  display_name: string
}

export interface Department {
  id: number
  name: string
  full_name: string
  has_children: boolean
  parent?: number
  children?: Department[]
  checked?: boolean
  indeterminate?: boolean
  isOpen: boolean
  loaded: boolean
  loading: boolean
}

export interface FormItems {
  label?: string
  required?: boolean
  property?: string,
  content?: Function,
  component?: Function,
}

export interface SecretModel {
  secretId: string,
  secretKey: string,
  subAccountId?: string,
  iamUserName?: string
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
  register = '登记账号',
}

export enum SiteType {
  china = '中国站',
  international = '国际站',
}

