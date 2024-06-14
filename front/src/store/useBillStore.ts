import http from '@/http';
import { defineStore } from 'pinia';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineStore('billStore', () => {
  /**
   * 一级账号列表
   * @param param0
   * @returns
   */
  const root_accounts_list = ({ filter, page }: ListProp) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/root_accounts/list`, { filter, page });
  };

  /**
   * 一级账号详情获取
   * @param data
   * @returns
   */
  const root_accounts_add = (data: {
    name: string; // 名字
    vendor: string; // 云厂商
    email: string; // 邮箱
    managers: string[]; // 负责人，最大5个
    bak_managers: string[]; // 备份负责人
    site: string; // 站点
    dept_id: number; // 组织架构ID
    memo: string; // 备忘录
    extension: Extension; // Extension对象
  }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/root_accounts/add`, data);
  };

  /**
   * 获取单个一级账号详情
   * @param id
   * @returns
   */
  const root_account_detail = (id: string) => {
    return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/root_accounts/${id}`);
  };

  /**
   * 修改一级账号
   * @param id 账号id
   * @param data
   * @returns
   */
  const root_account_update = (
    id: string,
    data: {
      name: string; // 名字
      managers: string[]; // 负责人，最大5个
      bak_managers: string[]; // 备份负责人
      dept_id: number; // 组织架构ID
      memo: string; // 备忘录
      extension: Extension; // Extension对象
    },
  ) => {
    return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/root_accounts/${id}`, data);
  };

  /**
   * 二级账号查询
   * @param data
   * @returns
   */
  const main_accounts_list = (data: ListProp) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/main_accounts/list`, data);
  };

  /**
   * 获取二级账号详情
   * @param id 账号id
   * @returns
   */
  const main_account_detail = (id: string) => {
    return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/main_accounts/${id}`);
  };

  /**
   * 创建二级账号
   * @param data
   * @returns
   */
  const create_main_account = (data: {
    vendor: 'aws' | 'gcp' | 'azure' | 'huawei' | 'zenlayer' | 'kaopu'; // 云厂商
    site: 'international' | 'china'; // 站点
    name: string; // 账号名
    email: string; // 邮箱
    project_id: string; // 项目ID
    project_type: 'international' | 'china'; // 项目类型
    managers: string[]; // 负责人，最大5个
    bak_managers: string[]; // 备份负责人
    dept_id: number; // 组织架构ID
    bk_biz_id: number[]; // 关联业务ID
    memo: string; // 备注
    extension: CreateMainExtension; // 扩展信息
  }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/types/create_main_account`, data);
  };

  /**
   * 二级账号管理员单据信息填写
   * @param data
   * @returns
   */
  const complete_main_account = (data: {
    sn: string; // ITSM单据编号
    id: string; // 海曼单据编号
    vendor: 'aws' | 'gcp' | 'azure' | 'huawei' | 'zenlayer' | 'kaopu'; // 云厂商
    root_account_id: string; // 一级账号的id
    extension?: CompleteExtension; // 扩展信息，根据云厂商传递的参数不一致
  }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/types/complete_main_account`, data);
  };

  /**
   * 修改二级账号
   * @param data
   * @returns
   */
  const update_main_account = (data: {
    id: string; // 必填，要变更的账号的id，不可修改
    vendor: 'aws' | 'gcp' | 'azure' | 'huawei' | 'zenlayer' | 'kaopu'; // 必填，要变更的账号vendor, 不可修改
    managers?: string[]; // 选填，要变更成为的管理员列表
    bak_managers?: string[]; // 选填，要变更成为的备份负责人列表
    dept_id?: number; // 选填，要变更成为的部门id
    op_product_id?: number; // 选填，要变更成为的运营产品id
    bk_biz_id?: number; // 选填，要变更成为的业务id
  }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/types/update_main_account`, data);
  };

  /**
   * 查询单据列表
   * @param data
   * @returns
   */
  const list_applications = (data: ListProp) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/list`, data);
  };

  /**
   * 获取单据详情
   * @param application_id 海曼申请单ID
   * @returns
   */
  const get_application = (application_id: string) => {
    return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/applications/${application_id}`);
  };

  return {
    // 一级账号
    root_accounts_add,
    root_accounts_list,
    root_account_detail,
    root_account_update,
    // 二级账号
    main_accounts_list,
    main_account_detail,
    create_main_account,
    complete_main_account,
    update_main_account,
    // 单据
    list_applications,
    get_application,
  };
});

export interface FilterRule {
  field: string; // 字段名
  op: string; // 操作符，可以是 "=", "!=" 等等
  value: string | number | boolean; // 值，可以是字符串、数值、布尔值等
}

export interface Filter {
  op: 'and' | 'or'; // 操作符，可以是 "and" 或 "or"
  rules: FilterRule[]; // 规则数组，包含多个条件
}

export interface Page {
  count: boolean; // 是否计数
  start: number; // 起始位置
  limit: number; // 每页最多项目数
  sort: string; // 排序字段
  order: 'ASC' | 'DESC'; // 排序方式，可以是 "ASC" 或 "DESC"
}

export interface ListProp {
  filter: Filter; // 过滤条件
  page: Page; // 分页信息
}

// AWS
export interface AwsExtension {
  cloud_account_id: string;
  cloud_iam_username: string;
  cloud_secret_id: string;
  cloud_secret_key: string;
}

// GCP
export interface GcpExtension {
  email: string;
  cloud_project_id: string;
  cloud_project_name: string;
  cloud_service_account_id: string;
  cloud_service_account_name: string;
  cloud_service_secret_id: string;
  cloud_service_secret_key: string;
}

// Azure
export interface AzureExtension {
  display_name_name: string;
  cloud_tenant_id: string;
  cloud_subscription_id: string;
  cloud_subscription_name: string;
  cloud_application_id: string;
  cloud_application_name: string;
  cloud_client_secret_id: string;
  cloud_client_secret_key: string;
}

// Huawei
export interface HuaweiExtension {
  cloud_main_account_name: string;
  cloud_sub_account_id: string;
  cloud_sub_account_name: string;
  cloud_secret_id: string;
  cloud_secret_key: string;
  cloud_iam_user_id: string;
  cloud_iam_username: string;
}

// Zenlayer/Kaopu
export interface ZenlayerKaopuExtension {
  cloud_account_id: string;
  cloud_iam_username: string;
  cloud_secret_id: string;
  cloud_secret_key: string;
}

// Union Type for all possible extensions
export type Extension = AwsExtension | GcpExtension | AzureExtension | HuaweiExtension | ZenlayerKaopuExtension;

// AWS 二级账号扩展类型
export interface AwsMainExtension {
  cloud_main_account_name: string;
  cloud_main_account_id: string;
}

// GCP 二级账号扩展类型
export interface GcpMainExtension {
  cloud_project_name: string;
  cloud_project_id: string;
}

// Azure 二级账号扩展类型
export interface AzureMainExtension {
  cloud_subscription_name: string;
  cloud_subscription_id: string;
}

// Huawei 二级账号扩展类型
export interface HuaweiMainExtension {
  cloud_main_account_name: string;
  cloud_main_account_id: string;
}

// Zenlayer/Kaopu 二级账号扩展类型
export interface ZenlayerKaopuMainExtension {
  cloud_main_account_name: string;
  cloud_main_account_id: string;
}

// Union Type for all possible main extensions
export type MainExtension =
  | AwsMainExtension
  | GcpMainExtension
  | AzureMainExtension
  | HuaweiMainExtension
  | ZenlayerKaopuMainExtension;

// AWS 二级账号扩展类型
interface CreateAwsMainExtension {
  cloud_main_account_name: string;
}

// GCP 二级账号扩展类型
interface CreateGcpMainExtension {
  cloud_project_name: string;
}

// Azure 二级账号扩展类型
interface CreateAzureMainExtension {
  cloud_subscription_name: string;
}

// Huawei 二级账号扩展类型
interface CreateHuaweiMainExtension {
  cloud_main_account_name: string;
}

// Zenlayer/Kaopu 二级账号扩展类型
interface CreateZenlayerKaopuMainExtension {
  cloud_main_account_name: string;
}

// Union Type for all possible main extensions
type CreateMainExtension =
  | CreateAwsMainExtension
  | CreateGcpMainExtension
  | CreateAzureMainExtension
  | CreateHuaweiMainExtension
  | CreateZenlayerKaopuMainExtension;

// Azure 扩展类型
interface AzureCompleteExtension {
  cloud_subscription_name: string;
  cloud_subscription_id: string;
  cloud_init_password: string;
}

// Huawei 扩展类型
interface HuaweiCompleteExtension {
  cloud_main_account_name: string;
  cloud_main_account_id: string;
  cloud_init_password: string;
}

// Zenlayer/Kaopu 扩展类型
interface ZenlayerKaopuCompleteExtension {
  cloud_main_account_name: string;
  cloud_main_account_id: string;
  cloud_init_password: string;
}

// Union Type for all possible complete extensions (单据信息填写)
type CompleteExtension = AzureCompleteExtension | HuaweiCompleteExtension | ZenlayerKaopuCompleteExtension;
