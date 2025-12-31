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
	"fmt"
	"strings"
)

// table names
const (
	// BKTableNameApplyTicket the table name of resource apply ticket
	BKTableNameApplyTicket = "cr_ApplyTicket"

	// BKTableNameApplyOrder the table name of resource apply order
	BKTableNameApplyOrder = "cr_ApplyOrder"

	// BKTableNameApplyStep the table name of resource apply order step info
	BKTableNameApplyStep = "cr_ApplyStep"

	// BKTableNameGenerateRecord the table name of device generate record
	BKTableNameGenerateRecord = "cr_GenerateRecord"

	// BKTableNameInitRecord the table name of device init record
	BKTableNameInitRecord = "cr_InitRecord"

	// BKTableNameDiskCheckRecord the table name of device disk check record
	BKTableNameDiskCheckRecord = "cr_DiskCheckRecord"

	// BKTableNameDeliverRecord the table name of device deliver record
	BKTableNameDeliverRecord = "cr_DeliverRecord"

	// BKTableNameDeviceInfo the table name of device info
	BKTableNameDeviceInfo = "cr_DeviceInfo"

	// BKTableNameNoticeInfo the table name of event info
	BKTableNameNoticeInfo = "cr_NoticeInfo"

	// BKTableNameRecycleRecord the table name of resource recycle record
	BKTableNameRecycleRecord = "cr_RecycleRecord"

	// BKTableNameCfgRequirement the table name of resource requirement config
	BKTableNameCfgRequirement = "cr_CfgRequirement"

	// BKTableNameCfgQcloudRegion the table name of resource region config
	BKTableNameCfgQcloudRegion = "cr_CfgRegion"

	// BKTableNameCfgQcloudZone the table name of resource zone config
	BKTableNameCfgQcloudZone = "cr_CfgZone"

	// BKTableNameCfgIdcZone the table name of resource region config
	BKTableNameCfgIdcZone = "cr_CfgIdcZone"

	// BKTableNameCfgCvmImage the table name of cvm image config
	BKTableNameCfgCvmImage = "cr_CfgCvmImage"

	// BKTableNameCfgDeviceRestrict the table name of device restriction config
	BKTableNameCfgDeviceRestrict = "cr_CfgDeviceRestrict"

	// BKTableNameCfgDevice the table name of cvm device config
	BKTableNameCfgDevice = "cr_CfgDevice"

	// BKTableNameCfgDvmDevice the table name of dvm device config
	BKTableNameCfgDvmDevice = "cr_CfgDvmDevice"

	// BKTableNameCfgPmDevice the table name of physical machine device config
	BKTableNameCfgPmDevice = "cr_CfgPmDevice"

	// BKTableNamePlanInfo the table name of cvm&cbs plan info
	BKTableNamePlanInfo = "cr_CvmCbsPlan"

	// BKTableNameCvmApplyOrder the table name of cvm apply order
	BKTableNameCvmApplyOrder = "cr_CvmApplyOrder"

	// BKTableNameCvmInfo the table name of cvm info
	BKTableNameCvmInfo = "cr_CvmInfo"

	// BKTableNameInstAsst the table name of the inst association
	BKTableNameInstAsst = "cc_InstAsst"

	BKTableNameBaseApp     = "cc_ApplicationBase"
	BKTableNameBaseHost    = "cc_HostBase"
	BKTableNameBaseModule  = "cc_ModuleBase"
	BKTableNameBaseInst    = "cc_ObjectBase"
	BKTableNameBasePlat    = "cc_PlatBase"
	BKTableNameBaseSet     = "cc_SetBase"
	BKTableNameBaseProcess = "cc_Process"
	BKTableNameDelArchive  = "cc_DelArchive"

	BKTableNameModuleHostConfig = "cc_ModuleHostConfig"
	BKTableNameObjAsst          = "cc_ObjAsst"

	BKTableNameNetcollectReport = "cc_NetcollectReport"

	BKTableNameServiceTemplate         = "cc_ServiceTemplate"
	BKTableNameServiceInstance         = "cc_ServiceInstance"
	BKTableNameProcessTemplate         = "cc_ProcessTemplate"
	BKTableNameProcessInstanceRelation = "cc_ProcessInstanceRelation"

	BKTableNameSetTemplate = "cc_SetTemplate"

	BKTableNameCloudAccount = "cc_CloudAccount"

	// BKTableNameWatchToken the table to store the latest watch token for collections
	BKTableNameWatchToken = "cc_WatchToken"

	// BKTableNameMainlineInstance is a virtual collection name which represent for mainline instance events
)

// TableSpecifier is table specifier type which describes the metadata
// access or classification level.
type TableSpecifier string

const (
	// TableSpecifierPublic is public specifier for table.
	TableSpecifierPublic TableSpecifier = "pub"
)

const (
	// BKObjectInstShardingTablePrefix is prefix of object instance sharding table.
	BKObjectInstShardingTablePrefix = BKTableNameBaseInst + "_"

	// BKObjectInstAsstShardingTablePrefix is prefix of object instance association sharding table.
	BKObjectInstAsstShardingTablePrefix = BKTableNameInstAsst + "_"
)

// GetObjectInstTableName return the object instance table name in sharding mode base on
// the object ID. Format: cc_ObjectBase_{supplierAccount}_{Specifier}_{ObjectID}, such as 'cc_ObjectBase_0_pub_switch'.
func GetObjectInstTableName(objID, supplierAccount string) string {
	return fmt.Sprintf("%s%s_%s_%s", BKObjectInstShardingTablePrefix, supplierAccount, TableSpecifierPublic, objID)
}

// GetObjectInstAsstTableName return the object instance association table name in sharding mode base on
// the object ID. Format: cc_InstAsst_{supplierAccount}_{Specifier}_{ObjectID}, such as 'cc_InstAsst_0_pub_switch'.
func GetObjectInstAsstTableName(objID, supplierAccount string) string {
	return fmt.Sprintf("%s%s_%s_%s", BKObjectInstAsstShardingTablePrefix, supplierAccount, TableSpecifierPublic, objID)
}

// IsObjectShardingTable returns if the target table is an object sharding table, include
// object instance and association.
func IsObjectShardingTable(tableName string) bool {
	if IsObjectInstShardingTable(tableName) {
		return true
	}
	return IsObjectInstAsstShardingTable(tableName)
}

// IsObjectInstShardingTable returns if the target table is an object instance sharding table.
func IsObjectInstShardingTable(tableName string) bool {
	// check object instance table, cc_ObjectBase_{Specifier}_{ObjectID}
	return strings.HasPrefix(tableName, BKObjectInstShardingTablePrefix)
}

// IsObjectInstAsstShardingTable returns if the target table is an object instance association sharding table.
func IsObjectInstAsstShardingTable(tableName string) bool {
	// check object instance association table, cc_InstAsst_{Specifier}_{ObjectID}
	return strings.HasPrefix(tableName, BKObjectInstAsstShardingTablePrefix)
}

// GetInstTableName returns inst data table name
func GetInstTableName(objID, supplierAccount string) string {
	switch objID {
	case BKInnerObjIDApp:
		return BKTableNameBaseApp
	case BKInnerObjIDSet:
		return BKTableNameBaseSet
	case BKInnerObjIDModule:
		return BKTableNameBaseModule
	case BKInnerObjIDHost:
		return BKTableNameBaseHost
	case BKInnerObjIDProc:
		return BKTableNameBaseProcess
	case BKInnerObjIDPlat:
		return BKTableNameBasePlat
	default:
		return GetObjectInstTableName(objID, supplierAccount)
	}
}
