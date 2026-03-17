export type FlagType = 'phone' | 'token' | 'stoken' | 'wechat' | 'custom' | 'mail' | 'u2FToken';

/**
 * 云厂商枚举值
 * vendor: 供应商
 */
export type VendorType = 'tcloud' | 'aws' | 'azure' | 'gcp' | 'huawei';

/**
 * 腾讯云 extension 类型定义
 */
export interface ITcloudExtension {
  login_flag?: FlagType;
  action_flag?: FlagType;
  console_login?: any;
  [key: string]: any;
}

/**
 * AWS extension 类型定义
 */
export interface IAwsExtension {
  [key: string]: any;
}

/**
 * Azure extension 类型定义
 */
export interface IAzureExtension {
  [key: string]: any;
}

/**
 * GCP extension 类型定义
 */
export interface IGcpExtension {
  [key: string]: any;
}

/**
 * 华为云 extension 类型定义
 */
export interface IHuaweiExtension {
  [key: string]: any;
}

/**
 * Extensions 类型定义 - 按云厂商分类
 * 格式: extensions[vendor]
 */
export interface IExtensions {
  tcloud?: ITcloudExtension;
  aws?: IAwsExtension;
  azure?: IAzureExtension;
  gcp?: IGcpExtension;
  huawei?: IHuaweiExtension;
}
