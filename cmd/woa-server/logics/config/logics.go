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

package config

import (
	"hcm/pkg/client"
	"hcm/pkg/dal/dao"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/ziyan"
)

// Logics provides management interface for operations of model and instance and related resources like association
type Logics interface {
	Requirement() RequirementIf
	Region() RegionIf
	Zone() ZoneIf
	Vpc() VpcIf
	Subnet() SubnetIf
	DeviceRestrict() DeviceRestrictIf
	CvmImage() CvmImageIf
	Device() DeviceIf
	Capacity() CapacityIf
	LeftIP() LeftIPIf
	Sg() ziyan.SgIf
	ApplyOrderStatistics() ApplyOrderStatisticsIf
}

type logics struct {
	requirement          RequirementIf
	region               RegionIf
	zone                 ZoneIf
	vpc                  VpcIf
	subnet               SubnetIf
	deviceRestrict       DeviceRestrictIf
	cvmImage             CvmImageIf
	device               DeviceIf
	capacity             CapacityIf
	leftIP               LeftIPIf
	sg                   ziyan.SgIf
	applyOrderStatistics ApplyOrderStatisticsIf
}

// New create a logics manager
func New(client *client.ClientSet, thirdCli *thirdparty.Client, cmdbCli cmdb.Client, daoSet dao.Set) Logics {
	vpcOp := NewVpcOp(client, thirdCli)
	subnetOp := NewSubnetOp(client, thirdCli)
	return &logics{
		requirement:          NewRequirementOp(),
		region:               NewRegionOp(client),
		zone:                 NewZoneOp(client),
		vpc:                  vpcOp,
		subnet:               subnetOp,
		deviceRestrict:       NewDeviceRestrictOp(),
		cvmImage:             NewCvmImageOp(client),
		capacity:             NewCapacityOp(client, subnetOp, vpcOp, thirdCli, cmdbCli),
		device:               NewDeviceOp(thirdCli, NewZoneOp(client), client),
		leftIP:               NewLeftIPOp(vpcOp, subnetOp, thirdCli),
		sg:                   ziyan.NewSgOp(client),
		applyOrderStatistics: NewApplyOrderStatisticsOp(daoSet),
	}
}

// Requirement requirement interface
func (l *logics) Requirement() RequirementIf {
	return l.requirement
}

// Region region interface
func (l *logics) Region() RegionIf {
	return l.region
}

// Zone zone interface
func (l *logics) Zone() ZoneIf {
	return l.zone
}

// Vpc vpc interface
func (l *logics) Vpc() VpcIf {
	return l.vpc
}

// Subnet subnet interface
func (l *logics) Subnet() SubnetIf {
	return l.subnet
}

// DeviceRestrict device restrict interface
func (l *logics) DeviceRestrict() DeviceRestrictIf {
	return l.deviceRestrict
}

// CvmImage cvm image interface
func (l *logics) CvmImage() CvmImageIf {
	return l.cvmImage
}

// Device device interface
func (l *logics) Device() DeviceIf {
	return l.device
}

// Capacity capacity interface
func (l *logics) Capacity() CapacityIf {
	return l.capacity
}

// LeftIP left ip interface
func (l *logics) LeftIP() LeftIPIf {
	return l.leftIP
}

// Sg security group interface
func (l *logics) Sg() ziyan.SgIf {
	return l.sg
}

// ApplyOrderStatistics apply order statistics interface
func (l *logics) ApplyOrderStatistics() ApplyOrderStatisticsIf {
	return l.applyOrderStatistics
}
