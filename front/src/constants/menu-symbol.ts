/**
 * 业务（资源管理）
 */
export const MENU_BUSINESS = Symbol('menu_business');
export const MENU_BUSINESS_TASK_MANAGEMENT = Symbol('menu_business_task_management');
export const MENU_BUSINESS_TASK_MANAGEMENT_DETAILS = Symbol('menu_business_task_management_details');
export const MENU_BUSINESS_OPERATION_LOG = Symbol('menu_business_operation_log');
export const MENU_BUSINESS_OPERATION_LOG_DETAILS = Symbol('menu_business_operation_log_details');
export const MENU_BUSINESS_HOST_MANAGEMENT = Symbol('menu_business_host_management');
export const MENU_BUSINESS_DISK_MANAGEMENT = Symbol('menu_business_disk_management');
export const MENU_BUSINESS_IMAGE_MANAGEMENT = Symbol('menu_business_image_management');
export const MENU_BUSINESS_VPC_MANAGEMENT = Symbol('menu_business_vpc_management');
export const MENU_BUSINESS_SUBNET_MANAGEMENT = Symbol('menu_business_subnet_management');
export const MENU_BUSINESS_EIP_MANAGEMENT = Symbol('menu_business_eip_management');
export const MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT = Symbol('menu_business_network_interface_management');
export const MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT = Symbol('menu_business_routeing_table_management');
export const MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT = Symbol('menu_business_security_group_management');
export const MENU_BUSINESS_HOST_DETAILS = Symbol('menu_business_host_details');
export const MENU_BUSINESS_DRIVE_DETAILS = Symbol('menu_business_drive_details');
export const MENU_BUSINESS_IMAGE_DETAILS = Symbol('menu_business_image_details');
export const MENU_BUSINESS_VPC_DETAILS = Symbol('menu_business_vpc_details');
export const MENU_BUSINESS_SUBNET_DETAILS = Symbol('menu_business_subnet_details');
export const MENU_BUSINESS_EIP_DETAILS = Symbol('menu_business_eip_details');
export const MENU_BUSINESS_NIF_DETAILS = Symbol('menu_business_nif_details');
export const MENU_BUSINESS_ROUTE_TABLE_DETAILS = Symbol('menu_business_route_table_details');
export const MENU_BUSINESS_SECURITY_GROUP_DETAILS = Symbol('menu_business_security_group_details');
export const MENU_BUSINESS_GCP_DETAILS = Symbol('menu_business_gcp_details');
export const MENU_BUSINESS_TEMPLATE_DETAILS = Symbol('menu_business_template_details');
export const MENU_BUSINESS_LOAD_BALANCER = Symbol('menu_business_load_balancer');
export const MENU_BUSINESS_LOAD_BALANCER_LB_VIEW = Symbol('menu_business_load_balancer_lb_view');
export const MENU_BUSINESS_LOAD_BALANCER_TG_VIEW = Symbol('menu_business_load_balancer_tg_view');
export const MENU_BUSINESS_LOAD_BALANCER_DETAILS = Symbol('menu_business_load_balancer_details');
export const MENU_BUSINESS_TARGET_GROUP_DETAILS = Symbol('menu_business_target_group_details');
export const MENU_BUSINESS_LOAD_BALANCE_DEVICE_SEARCH = Symbol('menu_business_device_search');

export const MENU_BUSINESS_CERT_MANAGEMENT = Symbol('menu_business_cert_management');

export const MENU_BUSINESS_RECYCLEBIN = Symbol('menu_business_recyclebin');

/**
 * 资源申请
 */
export const MENU_BUSINESS_APPLY_CVM = Symbol('menu_business_apply_cvm');
export const MENU_BUSINESS_APPLY_VPC = Symbol('menu_business_apply_vpc');
export const MENU_BUSINESS_APPLY_DISK = Symbol('menu_business_apply_disk');
export const MENU_BUSINESS_APPLY_SUBNET = Symbol('menu_business_apply_subnet');
export const MENU_BUSINESS_APPLY_CLB = Symbol('menu_business_apply_clb');

/**
 * 工作台（复用 service 视图）
 */
export const MENU_SERVICE = Symbol('menu_service');
// 顶栏「服务请求」过渡入口，四期删除服务请求时一并删除
export const MENU_SERVICE_REQUEST = Symbol('menu_service_request');
export const MENU_SERVICE_HOST_APPLICATION = Symbol('menu_service_host_application');
export const MENU_SERVICE_HOST_RECYCLE_ENTRY = Symbol('menu_service_host_recycle_entry');
export const MENU_SERVICE_HOST_RECYCLE = Symbol('menu_service_host_recycle');

/**
 * 资源运营
 */
export const MENU_RESOURCE = Symbol('menu_resource');
export const MENU_RESOURCE_MANAGE = Symbol('menu_resource_manage');
export const MENU_RESOURCE_RESOURCE_LIST = Symbol('menu_resource_resource_list');
export const MENU_RESOURCE_DETAIL = Symbol('menu_resource_detail');
// 顶栏「资源接入」入口，资源接入下线时删除
export const MENU_RESOURCE_ACCESS = Symbol('menu_resource_access');
export const MENU_RESOURCE_ACCOUNT_MANAGEMENT = Symbol('menu_resource_account_management');
export const MENU_RESOURCE_ACCOUNT_CHECK_IN = Symbol('menu_resource_account_check_in');
export const MENU_RESOURCE_ACCOUNT_DETAIL = Symbol('menu_resource_account_detail');
export const MENU_RESOURCE_ACCOUNT_DETAIL_BASIC = Symbol('menu_resource_account_detail_basic');
export const MENU_RESOURCE_RECYCLEBIN = Symbol('menu_resource_recyclebin');
export const MENU_RESOURCE_OPERATION_LOG = Symbol('menu_resource_operation_log');
export const MENU_RESOURCE_OPERATION_LOG_DETAILS = Symbol('menu_resource_operation_log_details');
export const MENU_RESOURCE_HOST_APPLY = Symbol('menu_resource_host_apply');
export const MENU_RESOURCE_VPC_APPLY = Symbol('menu_resource_vpc_apply');
export const MENU_RESOURCE_DISK_APPLY = Symbol('menu_resource_disk_apply');
export const MENU_RESOURCE_SUBNET_APPLY = Symbol('menu_resource_subnet_apply');
export const MENU_RESOURCE_LOAD_BALANCER_APPLY = Symbol('menu_resource_load_balancer_apply');
export const MENU_RESOURCE_LOAD_BALANCER = Symbol('menu_resource_load_balancer');

/**
 * 资源选型
 */
export const MENU_SCHEME = Symbol('menu_scheme');

/**
 * 单据管理
 */
export const MENU_SERVICE_TICKET_MANAGEMENT = Symbol('menu_service_ticket_management');
export const MENU_SERVICE_TICKET_DETAILS = Symbol('menu_service_ticket_details');
export const MENU_BUSINESS_TICKET_MANAGEMENT = Symbol('menu_business_ticket_management');
export const MENU_BUSINESS_TICKET_DETAILS = Symbol('menu_business_ticket_details');

/**
 * 云账号管理
 */
export const MENU_BUSINESS_CLOUD_ACCOUNT = Symbol('menu_business_cloud_account');
export const MENU_BUSINESS_CLOUD_ACCOUNT_DETAILS = Symbol('menu_business_cloud_account_details');

/**
 * 权限策略库管理
 */
export const MENU_SERVICE_PERMISSION_POLICY = Symbol('menu_service_permission_policy');

/**
 * 账单与账号（一级菜单「资源运营」）
 * 命名与取值对齐 refactor-router-menu 分支的 menu-symbol.ts，便于后续合流
 */
export const MENU_BILL_ROOT_ACCOUNT = Symbol('menu_bill_root_account');
export const MENU_BILL_ROOT_ACCOUNT_CREATE = Symbol('menu_bill_root_account_create');
export const MENU_BILL_MAIN_ACCOUNT = Symbol('menu_bill_main_account');
export const MENU_BILL_MAIN_ACCOUNT_CREATE = Symbol('menu_bill_main_account_create');
export const MENU_BILL_MANAGE = Symbol('menu_bill_manage');
export const MENU_BILL_MANAGE_SUMMARY = Symbol('menu_bill_manage_summary');
export const MENU_BILL_MANAGE_SUMMARY_MANAGE = Symbol('menu_bill_manage_summary_manage');
export const MENU_BILL_MANAGE_SUMMARY_OPERATION_RECORD = Symbol('menu_bill_manage_summary_operation_record');
export const MENU_BILL_MANAGE_DETAIL = Symbol('menu_bill_manage_detail');
export const MENU_BILL_MANAGE_ADJUST = Symbol('menu_bill_manage_adjust');
