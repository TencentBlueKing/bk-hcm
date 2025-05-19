/**
 * 云资源选型
 */
export const AUTH_CREATE_CLOUD_SELECTION_SCHEME = Symbol.for('auth_create_cloud_selection_scheme');
export const AUTH_FIND_CLOUD_SELECTION_SCHEME = Symbol.for('auth_find_cloud_selection_scheme');
export const AUTH_UPDATE_CLOUD_SELECTION_SCHEME = Symbol.for('auth_update_cloud_selection_scheme');
export const AUTH_DELETE_CLOUD_SELECTION_SCHEME = Symbol.for('auth_delete_cloud_selection_scheme');

/**
 * 账号
 */
export const AUTH_FIND_ACCOUNT = Symbol.for('auth_find_account');
export const AUTH_IMPORT_ACCOUNT = Symbol.for('auth_import_account');
export const AUTH_UPDATE_ACCOUNT = Symbol.for('auth_update_account');

/**
 * 业务访问
 */
export const AUTH_ACCESS_BIZ = Symbol.for('auth_access_biz');

/**
 * 账号下IaaS资源（主机、vpc、子网、安全组、云硬盘、网络接口、弹性IP、路由表、镜像）
 */
export const AUTH_FIND_IAAS_RESOURCE = Symbol.for('auth_find_iaas_resource');
export const AUTH_CREATE_IAAS_RESOURCE = Symbol.for('auth_create_iaas_resource');
export const AUTH_UPDATE_IAAS_RESOURCE = Symbol.for('auth_update_iaas_resource');
export const AUTH_DELETE_IAAS_RESOURCE = Symbol.for('auth_delete_iaas_resource');

/**
 * 业务下IaaS资源（主机、vpc、子网、安全组、云硬盘、网络接口、弹性IP、路由表、镜像）
 */
export const AUTH_BIZ_FIND_IAAS_RESOURCE = Symbol.for('auth_biz_find_iaas_resource');
export const AUTH_BIZ_CREATE_IAAS_RESOURCE = Symbol.for('auth_biz_create_iaas_resource');
export const AUTH_BIZ_UPDATE_IAAS_RESOURCE = Symbol.for('auth_biz_update_iaas_resource');
export const AUTH_BIZ_DELETE_IAAS_RESOURCE = Symbol.for('auth_biz_delete_iaas_resource');

/**
 * 审计查看
 */
export const AUTH_BIZ_FIND_AUDIT = Symbol.for('auth_biz_find_audit');

/**
 * 回收站
 */
export const AUTH_FIND_RECYCLE_BIN = Symbol.for('auth_find_recycle_bin');
export const AUTH_MANAGE_RECYCLE_BIN = Symbol.for('auth_manage_recycle_bin');

/**
 * 证书
 */
export const AUTH_CREATE_CERT = Symbol.for('auth_create_cert');
export const AUTH_BIZ_CREATE_CERT = Symbol.for('auth_biz_create_cert');
export const AUTH_DELETE_CERT = Symbol.for('auth_delete_cert');
export const AUTH_BIZ_DELETE_CERT = Symbol.for('auth_biz_delete_cert');

/**
 * 账号管理
 */
export const AUTH_FIND_ROOT_ACCOUNT = Symbol.for('auth_find_root_account');
export const AUTH_FIND_MAIN_ACCOUNT = Symbol.for('auth_find_main_account');
export const AUTH_UPDATE_MAIN_ACCOUNT = Symbol.for('auth_update_main_account');
export const AUTH_FIND_ACCOUNT_BILL = Symbol.for('auth_find_account_bill');

/**
 * 负载均衡
 */
export const AUTH_CREATE_CLB = Symbol.for('auth_create_clb');
export const AUTH_BIZ_CREATE_CLB = Symbol.for('auth_biz_create_clb');
export const AUTH_UPDATE_CLB = Symbol.for('auth_update_clb');
export const AUTH_BIZ_UPDATE_CLB = Symbol.for('auth_biz_update_clb');
export const AUTH_DELETE_CLB = Symbol.for('auth_delete_clb');
export const AUTH_BIZ_DELETE_CLB = Symbol.for('auth_biz_delete_clb');
