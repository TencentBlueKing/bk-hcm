/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package task ...
package task

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/thirdparty/cvmapi"
)

// CVM create cvm request param
type CVM struct {
	AppId             string                 `json:"appId"`
	ApplyType         int64                  `json:"applyType"`
	AppModuleId       int64                  `json:"appModuleId"`
	Operator          string                 `json:"operator"`
	ApplyNumber       uint                   `json:"applyNumber"`
	NoteInfo          string                 `json:"noteInfo"`
	VPCId             string                 `json:"vpcId"`
	SubnetId          string                 `json:"subnetId"`
	Area              string                 `json:"area"`
	Zone              string                 `json:"zone"`
	ImageId           string                 `json:"image_id"`
	ImageName         string                 `json:"image_name"`
	InstanceType      string                 `json:"instanceType"`
	DiskType          enumor.DiskType        `json:"disk_type"`
	DiskSize          int64                  `json:"disk_size"`
	SecurityGroupId   string                 `json:"securityGroupId"`
	SecurityGroupName string                 `json:"securityGroupName"`
	SecurityGroupDesc string                 `json:"securityGroupDesc"`
	ChargeType        cvmapi.ChargeType      `json:"chargeType"`
	ChargeMonths      uint                   `json:"chargeMonths"`
	InheritInstanceId string                 `json:"inherit_instance_id"`
	BkProductID       int64                  `json:"bk_product_id"`
	BkProductName     string                 `json:"bk_product_name"`
	VirtualDeptID     int64                  `json:"virtual_dept_id"`
	VirtualDeptName   string                 `json:"virtual_dept_name"`
	SystemDisk        enumor.DiskSpec        `json:"system_disk"`
	DataDisk          []enumor.DiskSpec      `json:"data_disk"`
	FuzzyZone         []cvmapi.FuzzyZoneItem `json:"fuzzyZone"` // 可用区模糊申领，传入多个可用区+vpc+子网
	CPUCore           int64                  `json:"cpu_core"`  // CPU核心数
	// CPU超线程(0:默认 1:关闭 2:开启)
	CPUThreadSwitch enumor.CPUThreadSwitch `json:"cpu_thread_switch"`
}

// DeliveredCVMKey delivered cvm key
type DeliveredCVMKey struct {
	DeviceType string          `json:"device_type"`
	Region     string          `json:"region"`
	Zone       string          `json:"zone"`
	DiskType   enumor.DiskType `json:"disk_type"`
}

// PlanExpendGroup plan expend group
type PlanExpendGroup struct {
	DeviceType string          `json:"device_type" bson:"device_type"`
	Region     string          `json:"region" bson:"region"`
	Zone       string          `json:"zone" bson:"zone"`
	DiskType   enumor.DiskType `json:"disk_type" bson:"disk_type"`
	CPUCore    int64           `json:"cpu_core" bson:"cpu_core"`
}
