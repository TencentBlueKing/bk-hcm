/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package pkg

import (
	"math"
	"time"
)

const (
	// HTTPCreate create method
	HTTPCreate = "POST"

	// HTTPSelectPost select method
	HTTPSelectPost = "POST"

	// HTTPSelectGet select method
	HTTPSelectGet = "GET"

	// HTTPUpdate update method
	HTTPUpdate = "PUT"

	// HTTPDelete delete method
	HTTPDelete = "DELETE"

	// BKNoLimit no limit definition
	BKNoLimit = 999999999
	// BKMaxPageSize max limit of a page
	BKMaxPageSize = 1000

	// BKMaxInstanceLimit max limit of instance count
	BKMaxInstanceLimit = 500

	// BKMaxRecordsAtOnce 一次最大操作记录数
	BKMaxRecordsAtOnce = 2000

	// BKDefaultLimit the default limit definition
	BKDefaultLimit = 20

	// BKAuditLogPageLimit the audit log page limit
	BKAuditLogPageLimit = 200

	// BKMaxExportLimit the limit to export
	BKMaxExportLimit = 10000

	// BKInstParentStr the inst parent name
	BKInstParentStr = "bk_parent_id"

	// BKDefaultOwnerID the default owner value
	BKDefaultOwnerID = "0"

	// BKSuperOwnerID the super owner value
	BKSuperOwnerID = "superadmin"

	// BKDefaultDirSubArea the default dir subarea
	BKDefaultDirSubArea = 0

	// BKTimeTypeParseFlag the time flag
	BKTimeTypeParseFlag = "cc_time_type"

	// BKTopoBusinessLevelLimit the mainline topo level limit
	BKTopoBusinessLevelLimit = "level.businessTopoMax"

	// BKTopoBusinessLevelDefault the mainline topo level default level
	BKTopoBusinessLevelDefault = 7
)

const (
	// BKInnerObjIDApp the inner object
	BKInnerObjIDApp = "biz"

	// BKInnerObjIDSet the inner object
	BKInnerObjIDSet = "set"

	// BKInnerObjIDModule the inner object
	BKInnerObjIDModule = "module"

	// BKInnerObjIDHost the inner object
	BKInnerObjIDHost = "host"

	// BKInnerObjIDObject the inner object
	BKInnerObjIDObject = "object"

	// BKInnerObjIDProc the inner object
	BKInnerObjIDProc = "process"

	// BKInnerObjIDConfigTemp the inner object
	BKInnerObjIDConfigTemp = "config_template"

	// BKInnerObjIDTempVersion the inner object
	BKInnerObjIDTempVersion = "template_version"

	// BKInnerObjIDPlat the inner object
	BKInnerObjIDPlat = "plat"

	// BKInnerObjIDSwitch the inner object
	BKInnerObjIDSwitch = "bk_switch"
	// BKInnerObjIDRouter the inner object
	BKInnerObjIDRouter = "bk_router"
	// BKInnerObjIDBlance the inner object
	BKInnerObjIDBlance = "bk_load_balance"
	// BKInnerObjIDFirewall the inner object
	BKInnerObjIDFirewall = "bk_firewall"
	// BKInnerObjIDWeblogic the inner object
	BKInnerObjIDWeblogic = "bk_weblogic"
	// BKInnerObjIDTomcat the inner object
	BKInnerObjIDTomcat = "bk_tomcat"
	// BKInnerObjIDApache the inner object
	BKInnerObjIDApache = "bk_apache"
)

// Revision
const (
	RevisionEnterprise = "enterprise"
	RevisionCommunity  = "community"
	RevisionOpensource = "opensource"
)

const (
	// BKDBMULTIPLELike used only for host search
	BKDBMULTIPLELike = "$multilike"

	// BKDBIN the db operator
	BKDBIN = "$in"

	// BKDBOR the db operator
	BKDBOR = "$or"

	// BKDBAND the db operator
	BKDBAND = "$and"

	// BKDBLIKE the db operator
	BKDBLIKE = "$regex"

	// BKDBOPTIONS the db operator,used with $regex
	// detail to see https://docs.mongodb.com/manual/reference/operator/query/regex/#op._S_options
	BKDBOPTIONS = "$options"

	// BKDBEQ the db operator
	BKDBEQ = "$eq"

	// BKDBNE the db operator
	BKDBNE = "$ne"

	// BKDBNIN the db oeprator
	BKDBNIN = "$nin"

	// BKDBLT the db operator
	BKDBLT = "$lt"

	// BKDBLTE the db operator
	BKDBLTE = "$lte"

	// BKDBGT the db operator
	BKDBGT = "$gt"

	// BKDBGTE the db opeartor
	BKDBGTE = "$gte"

	// BKDBExists the db opeartor
	BKDBExists = "$exists"

	// BKDBNot the db opeartor
	BKDBNot = "$not"

	// BKDBCount the db opeartor
	BKDBCount = "$count"

	// BKDBGroup the db opeartor
	BKDBGroup = "$group"

	// BKDBMatch the db opeartor
	BKDBMatch = "$match"

	// BKDBSum the db opeartor
	BKDBSum = "$sum"

	// BKDBPush the db opeartor
	BKDBPush = "$push"

	// BKDBUNSET the db opeartor
	BKDBUNSET = "$unset"

	// BKDBAddToSet The $addToSet operator adds a value to an array unless the value is already present,
	// in which case $addToSet does nothing to that array.
	BKDBAddToSet = "$addToSet"

	// BKDBPull The $pull operator removes from an existing array all instances of a value or
	// values that match a specified condition.
	BKDBPull = "$pull"

	// BKDBAll matches arrays that contain all elements specified in the query.
	BKDBAll = "$all"

	// BKDBProject passes along the documents with the requested fields to the next stage in the pipeline
	BKDBProject = "$project"

	// BKDBSize counts and returns the total number of items in an array
	BKDBSize = "$size"

	// BKDBType selects documents where the value of the field is an instance of the specified BSON type(s).
	// Querying by data type is useful when dealing with highly unstructured data where data types are not predictable.
	BKDBType = "$type"

	// BKDBSort the db operator
	BKDBSort = "$sort"

	// BKDBReplaceRoot the db operator
	BKDBReplaceRoot = "$replaceRoot"

	// BKDBSkip the db operator to skip return number of doc
	BKDBSkip = "$skip"

	// BKDBLimit the db operator to limit return number of doc
	BKDBLimit = "$limit"

	// BKDBAsc the db operator to sort asc
	BKDBAsc = 1

	// BKDBDesc the db operator to sort desc
	BKDBDesc = -1
)

const (
	// DefaultResModuleName the default idle module name
	DefaultResModuleName string = "空闲机"
	// DefaultFaultModuleName the default fault module name
	DefaultFaultModuleName string = "故障机"
	// DefaultRecycleModuleName the default fault module name
	DefaultRecycleModuleName string = "待回收"
)

const (
	// BKFieldID the id definition
	BKFieldID = "id"
	// BKFieldName the name definition
	BKFieldName = "name"

	// BKDefaultField the default field
	BKDefaultField = "default"

	// BKOwnerIDField the owner field
	BKOwnerIDField = "bk_supplier_account"

	// BKAppIDField the appid field
	BKAppIDField = "bk_biz_id"

	// BKIPArr the ip address
	BKIPArr = "ipArr"

	// BKAssetIDField  the asset id field
	BKAssetIDField = "bk_asset_id"

	// BKSNField  the sn  field
	BKSNField = "bk_sn"

	// BKHostInnerIPField the host innerip field
	BKHostInnerIPField = "bk_host_innerip"

	// BKHostCloudRegionField the host cloud region field
	BKHostCloudRegionField = "bk_cloud_region"

	// BKHostOuterIPField the host outerip field
	BKHostOuterIPField = "bk_host_outerip"

	// BKCloudInstIDField the cloud instance id field
	BKCloudInstIDField = "bk_cloud_inst_id"

	// BKCloudHostStatusField the cloud host status field
	BKCloudHostStatusField = "bk_cloud_host_status"

	// TimeTransferModel the time transferModel field
	TimeTransferModel = "2006-01-02 15:04:05"

	// TimeDayTransferModel the time transferModel field
	TimeDayTransferModel = "2006-01-02"

	// BKCloudTaskID the cloud sync task id
	BKCloudTaskID = "bk_task_id"

	// BKNewAddHost the cloud sync new add hosts
	BKNewAddHost = "new_add"

	// BKImportFrom the host import from field
	BKImportFrom = "import_from"

	// BKHostIDField the host id field
	BKHostIDField = "bk_host_id"

	// BKHostNameField the host name field
	BKHostNameField = "bk_host_name"

	// BKAppNameField the app name field
	BKAppNameField = "bk_biz_name"

	// BKSetIDField the setid field
	BKSetIDField = "bk_set_id"

	// BKSetNameField the set name field
	BKSetNameField = "bk_set_name"

	// BKModuleIDField the module id field
	BKModuleIDField = "bk_module_id"

	// BKModuleNameField the module name field
	BKModuleNameField = "bk_module_name"

	// HostApplyEnabledField the host apply enabled field
	HostApplyEnabledField = "host_apply_enabled"

	// BKOSTypeField the os type field
	BKOSTypeField = "bk_os_type"

	// BKOSNameField the os name field
	BKOSNameField = "bk_os_name"

	// BKHttpGet the http get
	BKHttpGet = "GET"

	// BKTencentCloudTimeOut the tencent cloud timeout
	BKTencentCloudTimeOut = 10

	// TencentCloudUrl the tencent cloud url
	TencentCloudUrl = "cvm.tencentcloudapi.com"

	// TencentCloudSignMethod the tencent cloud sign method
	TencentCloudSignMethod = "HmacSHA1"

	// BKCloudIDField the cloud id field
	BKCloudIDField = "bk_cloud_id"

	// BKCloudNameField the cloud name field
	BKCloudNameField = "bk_cloud_name"

	// BKObjIDField the obj id field
	BKObjIDField = "bk_obj_id"

	// BKObjNameField the obj name field
	BKObjNameField = "bk_obj_name"

	// BKObjIconField the obj icon field
	BKObjIconField = "bk_obj_icon"

	// BKInstIDField the inst id field
	BKInstIDField = "bk_inst_id"

	// BKInstNameField the inst name field
	BKInstNameField = "bk_inst_name"

	// ExportCustomFields the use custom display columns
	ExportCustomFields = "export_custom_fields"

	// BKProcIDField the proc id field
	BKProcIDField = "bk_process_id"

	// BKConfTempIdField is the config template id field
	BKConfTempIdField = "bk_conftemp_id"

	// BKProcNameField the proc name field
	BKProcNameField = "bk_process_name"

	// BKTemlateIDField the process template id field
	BKTemlateIDField = "template_id"

	// BKVersionIDField the version id field
	BKVersionIDField = "version_id"

	// BKTemplateNameField the template name field
	BKTemplateNameField = "template_name"

	// BKFileNameField the file name field
	BKFileNameField = "file_name"

	// BKPropertyIDField the propety id field
	BKPropertyIDField = "bk_property_id"

	// BKPropertyNameField the property name field
	BKPropertyNameField = "bk_property_name"

	// BKPropertyIndexField the property index field
	BKPropertyIndexField = "bk_property_index"

	// BKPropertyTypeField the property type field
	BKPropertyTypeField = "bk_property_type"

	// BKPropertyGroupField the property group field
	BKPropertyGroupField = "bk_property_group"

	// BKPropertyValueField the property value field
	BKPropertyValueField = "bk_property_value"

	// BKObjAttIDField the obj att id field
	BKObjAttIDField = "bk_object_att_id"

	// BKClassificationIDField the classification id field
	BKClassificationIDField = "bk_classification_id"

	// BKClassificationNameField the classification name field
	BKClassificationNameField = "bk_classification_name"

	// BKClassificationIconField the classification icon field
	BKClassificationIconField = "bk_classification_icon"

	// BKPropertyGroupIDField the property group id field
	BKPropertyGroupIDField = "bk_group_id"

	// BKPropertyGroupNameField the property group name field
	BKPropertyGroupNameField = "bk_group_name"

	// BKPropertyGroupIndexField the property group index field
	BKPropertyGroupIndexField = "bk_group_index"

	// BKAsstObjIDField the property obj id field
	BKAsstObjIDField = "bk_asst_obj_id"

	// BKAsstInstIDField the property inst id field
	BKAsstInstIDField = "bk_asst_inst_id"

	// BKOptionField the option field
	BKOptionField = "option"

	// BKAuditTypeField the audit type field
	BKAuditTypeField = "audit_type"

	// BKResourceTypeField the audit resource type field
	BKResourceTypeField = "resource_type"

	// BKOperateFromField the platform where operation from field
	BKOperateFromField = "operate_from"

	// BKOperationDetailField the audit operation detail field
	BKOperationDetailField = "operation_detail"

	// BKOperationTimeField the audit operation time field
	BKOperationTimeField = "operation_time"

	// BKResourceIDField the audit resource ID field
	BKResourceIDField = "resource_id"

	// BKResourceNameField the audit resource name field
	BKResourceNameField = "resource_name"

	// BKLabelField the audit resource name field
	BKLabelField = "label"

	// BKSetEnvField the set env field
	BKSetEnvField = "bk_set_env"

	// BKSetStatusField the set status field
	BKSetStatusField = "bk_service_status"

	// BKSetDescField the set desc field
	BKSetDescField = "bk_set_desc"

	// BKSetCapacityField the set capacity field
	BKSetCapacityField = "bk_capacity"

	// BKPort the port
	BKPort = "port"

	// BKProcPortEnable whether enable port,  enable port use for monitor app. default value
	BKProcPortEnable = "bk_enable_port"

	// BKProcGatewayIP the process gateway ip
	BKProcGatewayIP = "bk_gateway_ip"

	// BKProcGatewayPort the process gateway port
	BKProcGatewayPort = "bk_gateway_port"

	// BKProcGatewayProtocol the process gateway protocol
	BKProcGatewayProtocol = "bk_gateway_protocol"

	// BKProcGatewayCity the process gateway city
	BKProcGatewayCity = "bk_gateway_city"

	// BKProcBindInfo the process bind info
	BKProcBindInfo = "bind_info"

	// BKUser the user
	BKUser = "user"

	// BKProtocol the protocol
	BKProtocol = "protocol"

	// BKIP the ip
	BKIP = "ip"

	// BKEnable the enable
	BKEnable = "enable"

	// BKProcessObjectName the process object name
	BKProcessObjectName = "process"

	// BKProcessIDField the process id field
	BKProcessIDField = "bk_process_id"

	// BKServiceInstanceIDField the service instance id field
	BKServiceInstanceIDField = "service_instance_id"
	// BKServiceTemplateIDField the service template id field
	BKServiceTemplateIDField = "service_template_id"
	// BKProcessTemplateIDField the process template id field
	BKProcessTemplateIDField = "process_template_id"
	// BKServiceCategoryIDField the service category id field
	BKServiceCategoryIDField = "service_category_id"

	// BKSetTemplateIDField the set template id field
	BKSetTemplateIDField = "set_template_id"

	// HostApplyRuleIDField the host apply rule id field
	HostApplyRuleIDField = "host_apply_rule_id"

	// BKParentIDField the parent id field
	BKParentIDField = "bk_parent_id"
	// BKRootIDField the root id field
	BKRootIDField = "bk_root_id"

	// BKProcessNameField the process name field
	BKProcessNameField = "bk_process_name"

	// BKFuncIDField the func id field
	BKFuncIDField = "bk_func_id"

	// BKFuncName the function name
	BKFuncName = "bk_func_name"

	// BKStartParamRegex the start param regex
	BKStartParamRegex = "bk_start_param_regex"

	// BKBindIP the bind ip
	BKBindIP = "bind_ip"

	// BKWorkPath the work path
	BKWorkPath = "work_path"

	// BKIsPre the ispre field
	BKIsPre = "ispre"

	// BKIsIncrementField the isincrement field
	BKIsIncrementField = "is_increment"

	// BKIsCollapseField the iscollapse field
	BKIsCollapseField = "is_collapse"

	// BKProxyListField the proxy list field
	BKProxyListField = "bk_proxy_list"

	// BKIPListField the ip list field
	BKIPListField = "ip_list"

	// BKInvalidIPSField the invalid ips field
	BKInvalidIPSField = "invalid_ips"

	// BKGseProxyField the gse proxy
	BKGseProxyField = "bk_gse_proxy"

	// BKSubAreaField the sub area field
	BKSubAreaField = "bk_cloud_id"

	// BKProcField the proc field
	BKProcField = "bk_process"

	// BKMaintainersField the maintainers field
	BKMaintainersField = "bk_biz_maintainer"

	// BKProductPMField the product pm field
	BKProductPMField = "bk_biz_productor"

	// BKTesterField the tester field
	BKTesterField = "bk_biz_tester"

	// BKOperatorField the operator field
	BKOperatorField = "operator" // the operator of app of module, is means a job position

	// BKLifeCycleField the life cycle field
	BKLifeCycleField = "life_cycle"

	// BKDeveloperField the developer field
	BKDeveloperField = "bk_biz_developer"

	// BKLanguageField the language field
	BKLanguageField = "language"

	// BKBakOperatorField the bak operator field
	BKBakOperatorField = "bk_bak_operator"

	// BKTimeZoneField the time zone field
	BKTimeZoneField = "time_zone"

	// BKIsRequiredField the required field
	BKIsRequiredField = "isrequired"

	// BKModuleTypeField the module type field
	BKModuleTypeField = "bk_module_type"

	// BKOrgIPField the org ip field
	BKOrgIPField = "bk_org_ip"

	// BKDstIPField the dst ip field
	BKDstIPField = "bk_dst_ip"

	// BKDescriptionField the description field
	BKDescriptionField = "description"

	// BKIsOnlyField the isonly name field
	BKIsOnlyField = "isonly"
	// BKGseTaskIDField the gse taskid
	BKGseTaskIDField = "task_id"
	// BKTaskIDField the gse taskid
	BKTaskIDField = "task_id"
	// BKGseOpTaskIDField the gse taskid
	BKGseOpTaskIDField = "gse_task_id"
	// BKProcPidFile the pid file of process
	BKProcPidFile = "pid_file"
	// BKProcStartCmd the start cmd of process
	BKProcStartCmd = "start_cmd"
	// BKProcStopCmd the stop cmd of process
	BKProcStopCmd = "stop_cmd"
	// BKProcReloadCmd the reload cmd of process
	BKProcReloadCmd = "reload_cmd"
	// BKProcRestartCmd the restart cmd of process
	BKProcRestartCmd = "restart_cmd"
	// BKProcTimeOut the timeout of process
	BKProcTimeOut = "timeout"
	// BKProcWorkPath the work path of process
	BKProcWorkPath = "work_path"
	// BKProcInstNum the inst num of process
	BKProcInstNum = "proc_num"

	// BKInstKeyField the inst key field for metric discover
	BKInstKeyField = "bk_inst_key"

	// BKDeviceIDField for net collect device
	BKDeviceIDField = "device_id"
	// BKDeviceNameField for net collect device
	BKDeviceNameField = "device_name"
	// BKDeviceModelField for net collect device
	BKDeviceModelField = "device_model"
	// BKVendorField for net collect device
	BKVendorField = "bk_vendor"

	// BKNetcollectPropertyIDField for net collect property of device
	BKNetcollectPropertyIDField = "netcollect_property_id"
	// BKOIDField for net collect device
	BKOIDField = "oid"
	// BKPeriodField for net collect device
	BKPeriodField = "period"
	// BKActionField for net collect device
	BKActionField = "action"
	// BKProcinstanceID for net collect device
	BKProcinstanceID = "proc_instance_id"

	// BKGseOpProcTaskDetailField gse operate process return detail
	BKGseOpProcTaskDetailField = "detail"
	// BKGroupField for net collect device
	BKGroupField = "group"

	// BKAttributeIDField for net collect device
	BKAttributeIDField = "bk_attribute_id"

	// BKTokenField for net collect device
	BKTokenField = "token"
	// BKCursorField for net collect device
	BKCursorField = "cursor"
	// BKClusterTimeField for net collect device
	BKClusterTimeField = "cluster_time"
	// BKEventTypeField for net collect device
	BKEventTypeField = "type"
	// BKStartAtTimeField for net collect device
	BKStartAtTimeField = "start_at_time"
	// BKSubResourceField for net collect device
	BKSubResourceField = "bk_sub_resource"
)

const (
	// BKRequestField for net collect device
	BKRequestField = "bk_request_id"
	// BKTxnIDField for net collect device
	BKTxnIDField = "bk_txn_id"
)

const (
	// UserDefinedModules user define idle module key.
	UserDefinedModules = "user_modules"

	// SystemSetName user define idle module key.
	SystemSetName = "set_name"

	// SystemIdleModuleKey system idle module key.
	SystemIdleModuleKey = "idle"

	// SystemFaultModuleKey system define fault module name.
	SystemFaultModuleKey = "fault"

	// SystemRecycleModuleKey system define recycle module name.
	SystemRecycleModuleKey = "recycle"
)

// DefaultResSetName the inner module set
const DefaultResSetName string = "空闲机池"

// WhiteListAppName the white list app name
const WhiteListAppName = "蓝鲸"

// WhiteListSetName the white list set name
const WhiteListSetName = "公共组件"

// WhiteListModuleName the white list module name
const WhiteListModuleName = "gitserver"

// the inst record's logging information
const (
	// CreatorField the creator
	CreatorField = "creator"

	// CreateTimeField the create time field
	CreateTimeField = "create_time"

	// ConfirmTimeField the cloud resource confirm time filed
	ConfirmTimeField = "confirm_time"

	// StartTimeFiled the cloud sync start time field
	StartTimeFiled = "start_time"

	// ModifierField the modifier field
	ModifierField = "modifier"

	// LastTimeField the last time field
	LastTimeField = "last_time"
)

const (
	// ValidCreate valid create
	ValidCreate = "create"

	// ValidUpdate valid update
	ValidUpdate = "update"
)

// DefaultResSetFlag the default resource set flat
const DefaultResSetFlag int = 1

// DefaultFlagDefaultValue the default flag
const DefaultFlagDefaultValue int = 0

// DefaultAppFlag the default app flag
const DefaultAppFlag int = 1

// DefaultAppName the default app name
const DefaultAppName string = "资源池"

// DefaultCloudName the default cloud name
const DefaultCloudName string = "default area"

// DefaultInstName the default inst name
const DefaultInstName string = "实例名"

// BKAppName the default app name
const BKAppName string = "蓝鲸"

// BKNetwork bk_classification_id value
const BKNetwork = "bk_network"

const (
	// SNMPActionGet get action
	SNMPActionGet = "get"

	// SNMPActionGetNext getnext action
	SNMPActionGetNext = "getnext"
)

const (
	// DefaultResModuleFlag the default resource module flag
	DefaultResModuleFlag int = 1

	// DefaultFaultModuleFlag the default fault module flag
	DefaultFaultModuleFlag int = 2

	// NormalModuleFlag create module by user , default =0
	NormalModuleFlag int = 0

	// NormalSetDefaultFlag user create set default field value
	NormalSetDefaultFlag int64 = 0

	// DefaultRecycleModuleFlag default recycle module flag
	DefaultRecycleModuleFlag int = 3

	// DefaultResSelfDefinedModuleFlag the default resource self-defined module flag
	DefaultResSelfDefinedModuleFlag int = 4

	// DefaultUserResModuleFlag the default platform self-defined module flag.
	DefaultUserResModuleFlag int = 5
)

const (
	// DefaultModuleType the default module type
	DefaultModuleType string = "1"
)

const (
	// FieldTypeSingleChar the single char filed type
	FieldTypeSingleChar string = "singlechar"

	// FieldTypeLongChar the long char field type
	FieldTypeLongChar string = "longchar"

	// FieldTypeInt the int field type
	FieldTypeInt string = "int"

	// FieldTypeFloat the float field type
	FieldTypeFloat string = "float"

	// FieldTypeEnum the enum field type
	FieldTypeEnum string = "enum"

	// FieldTypeDate the date field type
	FieldTypeDate string = "date"

	// FieldTypeTime the time field type
	FieldTypeTime string = "time"

	// FieldTypeUser the user field type
	FieldTypeUser string = "objuser"

	// FieldTypeTimeZone the timezone field type
	FieldTypeTimeZone string = "timezone"

	// FieldTypeBool the bool type
	FieldTypeBool string = "bool"

	// FieldTypeList the list type
	FieldTypeList string = "list"

	// FieldTypeTable the table type, inner type.
	FieldTypeTable string = "table"

	// FieldTypeOrganization the organization field type
	FieldTypeOrganization string = "organization"

	// FieldTypeSingleLenChar the single char length limit
	FieldTypeSingleLenChar int = 256

	// FieldTypeLongLenChar the long char length limit
	FieldTypeLongLenChar int = 2000

	// FieldTypeUserLenChar the user char length limit
	FieldTypeUserLenChar int = 2000

	// FieldTypeStrictCharRegexp the single char regex expression
	FieldTypeStrictCharRegexp string = `^[a-zA-Z]\w*$`

	// FieldTypeServiceCategoryRegexp the service category regex expression
	FieldTypeServiceCategoryRegexp string = `^([\w\p{Han}]|[:\-\(\)])+$`

	// FieldTypeMainlineRegexp the mainline instance name regex expression
	FieldTypeMainlineRegexp string = `^[^\\\|\/:\*,<>"\?#\s]+$`

	// FieldTypeSingleCharRegexp the single char regex expression
	// FieldTypeSingleCharRegexp string = `^([\w\p{Han}]|[，。？！={}|?<>~～、：＃；％＊——……＆·＄（）‘’“”\[\]『』〔〕｛｝
	//【】￥￡♀‖〖〗《》「」:,;\."'\/\\\+\-\s#@\(\)])+$`
	FieldTypeSingleCharRegexp string = `\S`

	// FieldTypeLongCharRegexp the long char regex expression\
	// FieldTypeLongCharRegexp string = `^([\w\p{Han}]|[，。？！={}|?<>~～、：＃；％＊——……＆·＄（）‘’“”\[\]『』〔〕｛｝【】
	// ￥￡♀‖〖〗《》「」:,;\."'\/\\\+\-\s#@\(\)])+$`
	FieldTypeLongCharRegexp string = `\S`
)

const (
	// HostAddMethodExcel add a host method
	HostAddMethodExcel = "1"

	// HostAddMethodAgent add a  agent method
	HostAddMethodAgent = "2"

	// HostAddMethodAPI add api method
	HostAddMethodAPI = "3"

	// HostAddMethodExcelIndexOffset the height of the table header
	HostAddMethodExcelIndexOffset = 3

	// HostAddMethodExcelAssociationIndexOffset the height of the table header
	HostAddMethodExcelAssociationIndexOffset = 2

	// HostAddMethodExcelDefaultIndex 生成表格数据起始索引，第一列为字段说明
	HostAddMethodExcelDefaultIndex = 1

	/*EXCEL color AARRGGBB :
	AA means Alpha
	RRGGBB means Red, in hex.
	GG means Red, in hex.
	BB means Red, in hex.
	*/

	// ExcelHeaderFirstRowColor cell bg color
	ExcelHeaderFirstRowColor = "FF92D050"
	// ExcelHeaderFirstRowFontColor  font color
	ExcelHeaderFirstRowFontColor = "00000000"
	// ExcelHeaderFirstRowRequireFontColor require font color
	ExcelHeaderFirstRowRequireFontColor = "FFFF0000"
	// ExcelHeaderOtherRowColor cell bg color
	ExcelHeaderOtherRowColor = "FFC6EFCE"
	// ExcelHeaderOtherRowFontColor font color
	ExcelHeaderOtherRowFontColor = "FF000000"
	// ExcelCellDefaultBorderColor black color
	ExcelCellDefaultBorderColor = "FFD4D4D4"
	// ExcelHeaderFirstColumnColor light gray
	ExcelHeaderFirstColumnColor = "fee9da"
	// ExcelFirstColumnCellColor dark gray
	ExcelFirstColumnCellColor = "fabf8f"

	// ExcelAsstPrimaryKeySplitChar split char
	ExcelAsstPrimaryKeySplitChar = ","
	// ExcelAsstPrimaryKeyJoinChar split char
	ExcelAsstPrimaryKeyJoinChar = "="
	// ExcelAsstPrimaryKeyRowChar split char
	ExcelAsstPrimaryKeyRowChar = "\n"

	// ExcelDelAsstObjectRelation delete asst object relation
	ExcelDelAsstObjectRelation = "/"

	// ExcelDataValidationListLen excel dropdown list item count
	ExcelDataValidationListLen = 50

	// ExcelCommentSheetCotentLangPrefixKey excel comment sheet centent language prefixe key
	ExcelCommentSheetCotentLangPrefixKey = "import_comment"

	// ExcelFirstColumnFieldName export excel first column for tips
	ExcelFirstColumnFieldName = "field_name"
	// ExcelFirstColumnFieldType export excel first column for type
	ExcelFirstColumnFieldType = "field_type"
	// ExcelFirstColumnFieldID export excel first column for id
	ExcelFirstColumnFieldID = "field_id"
	// ExcelFirstColumnInstData export excel first column for inst data
	ExcelFirstColumnInstData = "inst_data"

	// ExcelFirstColumnAssociationAttribute export excel first column for association attribute
	ExcelFirstColumnAssociationAttribute = "excel_association_attribute"
	// ExcelFirstColumnFieldDescription export excel first column for field description
	ExcelFirstColumnFieldDescription = "excel_field_description"

	// ExcelCellIgnoreValue the value of ignored excel cell
	ExcelCellIgnoreValue = "--"
)

const (
	// InputTypeExcel  data from excel
	InputTypeExcel = "excel"

	// InputTypeApiNewHostSync data from api for synchronize new host
	InputTypeApiNewHostSync = "api_sync_host"

	// BatchHostAddMaxRow batch sync add host max row
	BatchHostAddMaxRow = 128

	// ExcelImportMaxRow excel import max row
	ExcelImportMaxRow = 1000
)

// KvMap the map definition
type KvMap map[string]interface{}

const (
	// CCSystemOperatorUserName the system user
	CCSystemOperatorUserName = "cc_system"
	// CCSystemCollectorUserName the collector user
	CCSystemCollectorUserName = "cc_collector"
)

// APIRsp the result the http requst
type APIRsp struct {
	HTTPCode int         `json:"-"`
	Result   bool        `json:"result"`
	Code     int         `json:"code"`
	Message  interface{} `json:"message"`
	Data     interface{} `json:"data"`
}

const (
	// BKCacheKeyV3Prefix the prefix definition
	BKCacheKeyV3Prefix = "cc:v3:"
)

// event cache keys
const (
	EventCacheEventIDKey = BKCacheKeyV3Prefix + "event:inst_id"
	RedisSnapKeyPrefix   = BKCacheKeyV3Prefix + "snapshot:"
)

// ApiCacheLimiterRulePrefix api cache keys
const (
	ApiCacheLimiterRulePrefix = BKCacheKeyV3Prefix + "api:limiter_rule:"
)

const (
	// BKHTTPHeaderUser current request http request header fields name for login user
	BKHTTPHeaderUser = "BK_User"
	// BKHTTPLanguage the language key word
	BKHTTPLanguage = "HTTP_BLUEKING_LANGUAGE"
	// BKHTTPOwner the owner
	BKHTTPOwner = "HTTP_BK_SUPPLIER_ACCOUNT"
	// BKHTTPOwnerID the owner id
	BKHTTPOwnerID = "HTTP_BLUEKING_SUPPLIER_ID"
	// BKHTTPCookieLanugageKey the language key word
	BKHTTPCookieLanugageKey = "blueking_language"
	// BKHTTPRequestAppCode the app code
	BKHTTPRequestAppCode = "Bk-App-Code"
	// BKHTTPRequestRealIP the real ip
	BKHTTPRequestRealIP = "X-Real-Ip"

	// BKHTTPCCRequestID cc request id cc_request_id
	BKHTTPCCRequestID = "Cc_Request_Id"
	// BKHTTPOtherRequestID esb request id  X-Bkapi-Request-Id
	BKHTTPOtherRequestID = "X-Bkapi-Request-Id"

	// BKHTTPSecretsToken the secrets token
	BKHTTPSecretsToken = "BK-Secrets-Token"
	// BKHTTPSecretsProject the secrets project
	BKHTTPSecretsProject = "BK-Secrets-Project"
	// BKHTTPSecretsEnv the secrets env
	BKHTTPSecretsEnv = "BK-Secrets-Env"
	// BKHTTPReadReference  query db use secondary node
	BKHTTPReadReference = "Cc_Read_Preference"

	// BKHTTPGatewayName the gateway name
	BKHTTPGatewayName = "HTTP-BK-GATEWAY-NAME"
	// BKHTTPJWTToken the jwt token
	BKHTTPJWTToken = "X-Bkapi-JWT"
	// BKHTTPAPIGWOwnerID the api gw owner id
	BKHTTPAPIGWOwnerID = "HTTP-BLUEKING-SUPPLIER-ID"
	// BKHTTPAPIGWLanguage the api gw language
	BKHTTPAPIGWLanguage = "HTTP-BLUEKING-LANGUAGE"
)

// ReadPreferenceMode read preference mode
type ReadPreferenceMode string

// String get the string of read preference mode.
func (r ReadPreferenceMode) String() string {
	return string(r)
}

// BKHTTPReadRefernceMode constants  这个位置对应的是mongodb 的read preference 的mode，如果driver 没有变化这里是不需要变更的，
// 新增mode 需要修改src/storage/dal/mongo/local/mongo.go 中的getCollectionOption 方法来支持
const (
	// NilMode not set
	NilMode ReadPreferenceMode = ""
	// PrimaryMode indicates that only a primary is
	// considered for reading. This is the default
	// mode.
	PrimaryMode ReadPreferenceMode = "1"
	// PrimaryPreferredMode indicates that if a primary
	// is available, use it; otherwise, eligible
	// secondaries will be considered.
	PrimaryPreferredMode ReadPreferenceMode = "2"
	// SecondaryMode indicates that only secondaries
	// should be considered.
	SecondaryMode ReadPreferenceMode = "3"
	// SecondaryPreferredMode indicates that only secondaries
	// should be considered when one is available. If none
	// are available, then a primary will be considered.
	SecondaryPreferredMode ReadPreferenceMode = "4"
	// NearestMode indicates that all primaries and secondaries
	// will be considered.
	NearestMode ReadPreferenceMode = "5"
)

// transaction related
const (
	TransactionIdHeader      = "cc_transaction_id_string"
	TransactionTimeoutHeader = "cc_transaction_timeout"

	// TransactionDefaultTimeout mongodb default transaction timeout is 1 minute.
	TransactionDefaultTimeout = 2 * time.Minute
)

const (
	// DefaultAppLifeCycleNormal  biz life cycle normal
	DefaultAppLifeCycleNormal = "2"
)

// Host OS type enumeration value
const (
	HostOSTypeEnumLinux   = "1"
	HostOSTypeEnumWindows = "2"
	HostOSTypeEnumAIX     = "3"
	HostOSTypeEnumUNIX    = "4"
	HostOSTypeEnumSolaris = "5"
)

const (
	// MaxProcessPrio max process priority
	MaxProcessPrio = 10000
	// MinProcessPrio min process priority
	MinProcessPrio = -100
)

// integer const
const (
	MaxUint64  = ^uint64(0)
	MinUint64  = 0
	MaxInt64   = int64(MaxUint64 >> 1)
	MinInt64   = -MaxInt64 - 1
	MaxUint    = ^uint(0)
	MinUint    = 0
	MaxInt     = int(MaxUint >> 1)
	MinInt     = -MaxInt - 1
	MaxFloat64 = math.MaxFloat64
	MinFloat64 = -math.MaxFloat64
)

// HostCrossBizField flag
const HostCrossBizField = "hostcrossbiz"

// HostCrossBizValue host cross biz flag
const HostCrossBizValue = "e76fd4d1683d163e4e7e79cef45a74c1"

// config admin
const (
	ConfigAdminID         = "configadmin"
	ConfigAdminValueField = "config"
)

const (
	// APPConfigWaitTime application wait config from zookeeper time (unit sencend)
	APPConfigWaitTime = 15
)

const (
	// URLFilterWhiteListSuffix url filter white list not execute any filter
	// multiple url separated by commas
	URLFilterWhiteListSuffix = "/healthz,/version,/monitor_healthz"

	// URLFilterWhiteListSepareteChar url filter white list not execute any filter
	URLFilterWhiteListSepareteChar = ","
)

// DataStatusFlag data status flag
type DataStatusFlag string

const (
	// DataStatusDisabled data status flag
	DataStatusDisabled DataStatusFlag = "disabled"
	// DataStatusEnable data status flag
	DataStatusEnable DataStatusFlag = "enable"
)

const (
	// BKDataStatusField data status field
	BKDataStatusField = "bk_data_status"
	// BKDataRecoverSuffix data status recover suffix
	BKDataRecoverSuffix = "(recover)"
)

const (
	// Infinite period default value
	Infinite = "∞"
)

// netcollect
const (
	BKNetDevice   = "net_device"
	BKNetProperty = "net_property"
)

const (
	// BKBluekingLoginPluginVersion login type
	BKBluekingLoginPluginVersion = "blueking"
	// BKOpenSourceLoginPluginVersion login type
	BKOpenSourceLoginPluginVersion = "opensource"
	// BKSkipLoginPluginVersion login type
	BKSkipLoginPluginVersion = "skip-login"

	// BKNoopMonitorPlugin monitor plugin type
	BKNoopMonitorPlugin = "noop"
	// BKBluekingMonitorPlugin monitor plugin type
	BKBluekingMonitorPlugin = "blueking"

	// HTTPCookieBKToken cookie key
	HTTPCookieBKToken = "bk_token"

	// WEBSessionUinKey session key
	WEBSessionUinKey = "username"
	// WEBSessionChineseNameKey session key
	WEBSessionChineseNameKey = "chName"
	// WEBSessionPhoneKey session key
	WEBSessionPhoneKey = "phone"
	// WEBSessionEmailKey session key
	WEBSessionEmailKey = "email"
	// WEBSessionRoleKey session key
	WEBSessionRoleKey = "role"
	// WEBSessionOwnerUinKey session key
	WEBSessionOwnerUinKey = "owner_uin"
	// WEBSessionOwnerUinListeKey session key
	WEBSessionOwnerUinListeKey = "owner_uin_list"
	// WEBSessionAvatarUrlKey session key
	WEBSessionAvatarUrlKey = "avatar_url"
	// WEBSessionMultiSupplierKey session key
	WEBSessionMultiSupplierKey = "multisupplier"

	// LoginSystemMultiSupplierTrue multi supplier true
	LoginSystemMultiSupplierTrue = "1"
	// LoginSystemMultiSupplierFalse multi supplier false
	LoginSystemMultiSupplierFalse = "0"

	// LogoutHTTPSchemeCookieKey logout http scheme cookie key
	LogoutHTTPSchemeCookieKey = "http_scheme"
	// LogoutHTTPSchemeHTTP http
	LogoutHTTPSchemeHTTP = "http"
	// LogoutHTTPSchemeHTTPS https
	LogoutHTTPSchemeHTTPS = "https"
)

// BKStatusField status field
const BKStatusField = "status"

const (
	// BKProcInstanceOpUser proc instance op user
	BKProcInstanceOpUser = "proc instance user"
	// BKSynchronizeDataTaskDefaultUser synchronize data task user
	BKSynchronizeDataTaskDefaultUser = "synchronize task user"

	// BKCloudSyncUser cloud sync user
	BKCloudSyncUser = "cloud_sync_user"

	// BKIAMSyncUser IAM sync user
	BKIAMSyncUser = "iam_sync_user"
)

const (
	// RedisProcSrvHostInstanceRefreshModuleKey proc srv host instance refresh module key
	RedisProcSrvHostInstanceRefreshModuleKey = BKCacheKeyV3Prefix + "prochostinstancerefresh:set"
	// RedisProcSrvHostInstanceAllRefreshLockKey proc srv host instance all refresh lock key
	RedisProcSrvHostInstanceAllRefreshLockKey = BKCacheKeyV3Prefix + "lock:prochostinstancerefresh"
	// RedisProcSrvQueryProcOPResultKey proc srv query op result key
	RedisProcSrvQueryProcOPResultKey = BKCacheKeyV3Prefix + "procsrv:query:opresult:set"
	// RedisCloudSyncInstancePendingStart cloud sync instance pending start
	RedisCloudSyncInstancePendingStart = BKCacheKeyV3Prefix + "cloudsyncinstancependingstart:list"
	// RedisCloudSyncInstanceStarted cloud sync instance started
	RedisCloudSyncInstanceStarted = BKCacheKeyV3Prefix + "cloudsyncinstancestarted:list"
	// RedisCloudSyncInstancePendingStop cloud sync instance pending stop
	RedisCloudSyncInstancePendingStop = BKCacheKeyV3Prefix + "cloudsyncinstancependingstop:list"
	// RedisMongoCacheSyncKey mongodb cache sync key
	RedisMongoCacheSyncKey = BKCacheKeyV3Prefix + "mongodb:cache"
)

// association fields
const (
	// AssociationKindIDField the id of the association kind
	AssociationKindIDField    = "bk_asst_id"
	AssociationKindNameField  = "bk_asst_name"
	AssociationObjAsstIDField = "bk_obj_asst_id"
	AssociatedObjectIDField   = "bk_asst_obj_id"
)

// association
const (
	AssociationKindMainline = "bk_mainline"
	AssociationTypeBelong   = "belong"
	AssociationTypeGroup    = "group"
	AssociationTypeRun      = "run"
	AssociationTypeConnect  = "connect"
	AssociationTypeDefault  = "default"
)

const (
	// MetadataField data business key
	MetadataField = "metadata"
)

const (
	// BKBizDefault business default
	BKBizDefault = "bizdefault"
)

const (
	// MetaDataSynchronizeField Synchronous data aggregation field
	MetaDataSynchronizeField = "sync"
	// MetaDataSynchronizeFlagField synchronize flag
	MetaDataSynchronizeFlagField = "flag"
	// MetaDataSynchronizeVersionField synchronize version
	MetaDataSynchronizeVersionField = "version"
	// MetaDataSynchronizeIdentifierField 数据需要同步cmdb系统的身份标识， 值是数组
	MetaDataSynchronizeIdentifierField = "identifier"
	// MetaDataSynchronIdentifierFlagSyncAllValue 数据可以被任何系统同步
	MetaDataSynchronIdentifierFlagSyncAllValue = "__bk_cmdb__"

	// SynchronizeSignPrefix  synchronize sign , Should appear in the configuration file
	SynchronizeSignPrefix = "sync_blueking"

	/* synchronize model description classify*/

	// SynchronizeModelTypeClassification synchroneize model classification
	SynchronizeModelTypeClassification = "model_classification"
	// SynchronizeModelTypeAttribute synchroneize model attribute
	SynchronizeModelTypeAttribute = "model_attribute"
	// SynchronizeModelTypeAttributeGroup synchroneize model attribute group
	SynchronizeModelTypeAttributeGroup = "model_atrribute_group"
	// SynchronizeModelTypeBase synchroneize model attribute
	SynchronizeModelTypeBase = "model"

	/* synchronize instance assoication sign*/

	// SynchronizeAssociationTypeModelHost synchroneize model ggroup
	SynchronizeAssociationTypeModelHost = "module_host"
)

const (
	// AttributePlaceHolderMaxLength attribute place holder max length
	AttributePlaceHolderMaxLength = 2000
	// AttributeOptionMaxLength attribute option max length
	AttributeOptionMaxLength = 2000
	// AttributeIDMaxLength attribute id max length
	AttributeIDMaxLength = 128
	// AttributeNameMaxLength attribute name max length
	AttributeNameMaxLength = 128
	// AttributeUnitMaxLength attribute unit max length
	AttributeUnitMaxLength = 20
	// AttributeOptionValueMaxLength attribute option value max length
	AttributeOptionValueMaxLength = 128
	// AttributeOptionArrayMaxLength attribute option array max length
	AttributeOptionArrayMaxLength = 200
	// ServiceCategoryMaxLength service category max length
	ServiceCategoryMaxLength = 128
)

const (
	// NameFieldMaxLength name field max length
	NameFieldMaxLength = 256
	// MainlineNameFieldMaxLength mainline name field max length
	MainlineNameFieldMaxLength = 256

	// ServiceTemplateIDNotSet 用于表示还未设置服务模板的情况，比如没有绑定服务模板
	ServiceTemplateIDNotSet = 0
	// SetTemplateIDNotSet 用于表示还未设置服务模板的情况，比如还没有绑定服务模板
	SetTemplateIDNotSet = 0

	// MetadataLabelBiz 业务ID
	MetadataLabelBiz = "metadata.label.bk_biz_id"

	// DefaultServiceCategoryName 业务默认分类名称
	DefaultServiceCategoryName = "Default"
)

const (
	// ContextRequestIDField request id
	ContextRequestIDField = "request_id"
	// ContextRequestUserField request user
	ContextRequestUserField = "request_user"
	// ContextRequestOwnerField request owner
	ContextRequestOwnerField = "request_owner"
)

const (
	// OperationCustom custom operation
	OperationCustom = "custom"
	// OperationReportType report type
	OperationReportType = "report_type"
	// OperationConfigID config id
	OperationConfigID = "config_id"
	// BizModuleHostChart biz module host chart
	BizModuleHostChart = "biz_module_host_chart"
	// HostOSChart host os chart
	HostOSChart = "host_os_chart"
	// HostBizChart host biz chart
	HostBizChart = "host_biz_chart"
	// HostCloudChart host cloud chart
	HostCloudChart = "host_cloud_chart"
	// HostChangeBizChart host change biz chart
	HostChangeBizChart = "host_change_biz_chart"
	// ModelAndInstCount model and inst count
	ModelAndInstCount = "model_and_inst_count"
	// ModelInstChart model inst chart
	ModelInstChart = "model_inst_chart"
	// ModelInstChangeChart model inst change chart
	ModelInstChangeChart = "model_inst_change_chart"
	// CreateObject create object
	CreateObject = "create object"
	// DeleteObject delete object
	DeleteObject = "delete object"
	// UpdateObject update object
	UpdateObject = "update object"
	// OperationDescription operation description
	OperationDescription = "op_desc"
	// OptionOther other option
	OptionOther = "其他"
	// TimerPattern the api timer pattern
	TimerPattern = "^[\\d]+\\:[\\d]+$"
	// BKTaskTypeField the api task type field
	BKTaskTypeField = "task_type"
	// SyncSetTaskFlag the api set task flag
	SyncSetTaskFlag = "set_template_sync"
	// SyncModuleTaskFlag the api module task flag
	SyncModuleTaskFlag = "service_template_sync"

	// BKHostState the api host state field
	BKHostState = "bk_state"
)

// LanguageType multiple language support
type LanguageType string

const (
	// Chinese language
	Chinese LanguageType = "zh-cn"
	// English language
	English LanguageType = "en"
)

// cloud sync const
const (
	BKCloudAccountID             = "bk_account_id"
	BKCloudAccountName           = "bk_account_name"
	BKCloudVendor                = "bk_cloud_vendor"
	BKCloudSyncTaskName          = "bk_task_name"
	BKCloudSyncTaskID            = "bk_task_id"
	BKCloudSyncStatus            = "bk_sync_status"
	BKCloudSyncStatusDescription = "bk_status_description"
	BKCloudLastSyncTime          = "bk_last_sync_time"
	BKCreator                    = "bk_creator"
	BKStatus                     = "bk_status"
	BKStatusDetail               = "bk_status_detail"
	BKLastEditor                 = "bk_last_editor"
	BKSecretID                   = "bk_secret_id"
	BKVpcID                      = "bk_vpc_id"
	BKVpcName                    = "bk_vpc_name"
	BKRegion                     = "bk_region"
	BKCloudSyncVpcs              = "bk_sync_vpcs"

	// IsDestroyedCloudHost 是否为被销毁的云主机
	IsDestroyedCloudHost = "is_destroyed_cloud_host"
)

const (
	// BKCloudHostStatusUnknown the cloud host status is unknown
	BKCloudHostStatusUnknown = "1"
	// BKCloudHostStatusStarting the cloud host is starting
	BKCloudHostStatusStarting = "2"
	// BKCloudHostStatusRunning the cloud host is running
	BKCloudHostStatusRunning = "3"
	// BKCloudHostStatusStopping the cloud host is stopping
	BKCloudHostStatusStopping = "4"
	// BKCloudHostStatusStopped the cloud host is stopped
	BKCloudHostStatusStopped = "5"
	// BKCloudHostStatusDestroyed the cloud host is destroyed
	BKCloudHostStatusDestroyed = "6"
)

const (
	// BKCloudAreaStatusNormal the cloud area status is normal
	BKCloudAreaStatusNormal = "1"
	// BKCloudAreaStatusAbnormal the cloud area status is abnormal
	BKCloudAreaStatusAbnormal = "2"
)

// BKDefaultConfigCenter configcenter
const (
	BKDefaultConfigCenter = "zookeeper"
)

const (
	// CCLogicUniqueIdxNamePrefix unique index name prefix
	CCLogicUniqueIdxNamePrefix = "bkcc_unique_"
	// CCLogicIndexNamePrefix index name prefix
	CCLogicIndexNamePrefix = "bkcc_idx_"
)
