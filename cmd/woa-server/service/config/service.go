/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

// Package config service
package config

import (
	"net/http"

	"hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/plan"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		authorizer: c.Authorizer,
		logics:     c.ConfigLogics,
		planLogics: c.PlanController,
		client:     c.Client,
	}
	h := rest.NewHandler()

	s.initCommonRestrict(h)
	s.initCvmImage(h)
	s.initCvmRestrict(h)
	s.initDevice(h)
	s.initDeviceRestrict(h)
	s.initDvmRestrict(h)
	s.initLeftIP(h)
	s.initPlan(h)
	s.initPmRestrict(h)
	s.initRegion(h)
	s.initRequirement(h)
	s.initSubnet(h)
	s.initVpc(h)
	s.initZone(h)
	s.initCapacity(h)
	s.initSg(h)

	h.Load(c.WebService)
}

type service struct {
	logics     config.Logics
	planLogics plan.Logics
	authorizer auth.Authorizer
	client     *client.ClientSet
}

func (s *service) initCommonRestrict(h *rest.Handler) {
	h.Add("GetAffinity", http.MethodPost, "/config/find/config/affinity", s.GetAffinity)
	h.Add("GetApplyStage", http.MethodGet, "/config/find/config/apply/stage", s.GetApplyStage)
}

func (s *service) initCvmImage(h *rest.Handler) {
	h.Add("GetCvmImage", http.MethodPost, "/config/findmany/config/cvm/image", s.GetCvmImage)
	h.Add("BatchEnableImageToApplyCVM", http.MethodPost, "/config/images/enable_cvm/batch",
		s.BatchEnableImageToApplyCVM)
	h.Add("BatchDisableImageToApplyCVM", http.MethodPost, "/config/images/disable_cvm/batch",
		s.BatchDisableImageToApplyCVM)
}

func (s *service) initCvmRestrict(h *rest.Handler) {
	h.Add("GetCvmDiskType", http.MethodGet, "/config/find/config/cvm/disktype", s.GetCvmDiskType)
}

func (s *service) initDevice(h *rest.Handler) {
	h.Add("GetDeviceWithCapacity", http.MethodPost, "/config/findmany/config/cvm/device/detail",
		s.GetDeviceWithCapacity)
	h.Add("GetDevice", http.MethodPost, "/config/findmany/config/cvm/device/detail/avail", s.GetDevice)
	h.Add("GetDeviceType", http.MethodPost, "/config/findmany/config/cvm/devicetype", s.GetDeviceType)
	h.Add("GetDeviceTypeDetail", http.MethodPost, "/config/findmany/config/cvm/devicetype/detail",
		s.GetDeviceTypeDetail)
	h.Add("GetCvmDeviceDetail", http.MethodPost, "/config/findmany/config/cvm/device", s.GetCvmDeviceDetail)
	h.Add("CreateDevice", http.MethodPost, "/config/create/config/cvm/device", s.CreateDevice)
	h.Add("CreateManyDevice", http.MethodPost, "/config/createmany/config/cvm/device", s.CreateManyDevice)
	h.Add("UpdateDevice", http.MethodPut, "/config/update/config/cvm/device/{id}", s.UpdateDevice)
	h.Add("UpdateDeviceProperty", http.MethodPut, "/config/updatemany/config/cvm/device/property",
		s.UpdateDeviceProperty)
	h.Add("DeleteDevice", http.MethodDelete, "/config/delete/config/cvm/device/{id}", s.DeleteDevice)
	h.Add("GetDvmDeviceType", http.MethodPost, "/config/findmany/config/dvm/devicetype", s.GetDvmDeviceType)
	h.Add("CreateDvmDevice", http.MethodPost, "/config/create/config/dvm/device", s.CreateDvmDevice)
	h.Add("GetPmDeviceType", http.MethodPost, "/config/findmany/config/idcpm/devicetype", s.GetPmDeviceType)
	h.Add("CreatePmDevice", http.MethodPost, "/config/create/config/idcpm/device", s.CreatePmDevice)
}

func (s *service) initDeviceRestrict(h *rest.Handler) {
	h.Add("GetDeviceRestrict", http.MethodGet, "/config/find/config/cvm/devicerestrict", s.GetDeviceRestrict)
	h.Add("CreateDeviceRestrict", http.MethodPost, "/config/create/config/cvm/devicerestrict", s.CreateDeviceRestrict)
	h.Add("UpdateDeviceRestrict", http.MethodPut, "/config/update/config/cvm/devicerestrict/{id}",
		s.UpdateDeviceRestrict)
	h.Add("DeleteDeviceRestrict", http.MethodDelete, "/config/delete/config/cvm/devicerestrict/{id}",
		s.DeleteDeviceRestrict)
}

func (s *service) initDvmRestrict(h *rest.Handler) {
	h.Add("GetDvmImage", http.MethodGet, "/config/find/config/dvm/image", s.GetDvmImage)
	h.Add("GetDvmKernel", http.MethodGet, "/config/find/config/dvm/kernel", s.GetDvmKernel)
	h.Add("GetDvmMountPath", http.MethodGet, "/config/find/config/dvm/mountpath", s.GetDvmMountPath)
	h.Add("GetDvmIdcDeviceGroup", http.MethodGet, "/config/find/config/dvm/idc/devicegroup", s.GetDvmIdcDeviceGroup)
	h.Add("GetDvmQcloudDeviceGroup", http.MethodGet, "/config/find/config/dvm/qcloud/devicegroup",
		s.GetDvmQcloudDeviceGroup)
}

func (s *service) initLeftIP(h *rest.Handler) {
	h.Add("GetLeftIP", http.MethodPost, "/config/findmany/config/cvm/leftip", s.GetLeftIP)
	h.Add("CreateLeftIP", http.MethodPost, "/config/create/config/cvm/leftip", s.CreateLeftIP)
	h.Add("UpdateLeftIPProperty", http.MethodPut, "/config/updatemany/config/cvm/leftip/property",
		s.UpdateLeftIPProperty)
	h.Add("SyncLeftIP", http.MethodPost, "/config/sync/config/cvm/leftip", s.SyncLeftIP)
}

func (s *service) initPlan(h *rest.Handler) {
	h.Add("GetPlanCoreType", http.MethodGet, "/config/find/config/plan/coretype", s.GetPlanCoreType)
	h.Add("GetPlanDiskType", http.MethodGet, "/config/find/config/plan/disktype", s.GetPlanDiskType)
	h.Add("GetPlanOrderType", http.MethodGet, "/config/find/config/plan/ordertype", s.GetPlanOrderType)
	h.Add("GetPlanDeviceGroup", http.MethodGet, "/config/find/config/plan/devicegroup", s.GetPlanDeviceGroup)
}

func (s *service) initPmRestrict(h *rest.Handler) {
	h.Add("GetPmOstype", http.MethodGet, "/config/find/config/idcpm/ostype", s.GetPmOstype)
	h.Add("GetPmIsp", http.MethodGet, "/config/find/config/idcpm/isp", s.GetPmIsp)
	h.Add("GetPmRaidtype", http.MethodGet, "/config/find/config/idcpm/raidtype", s.GetPmRaidtype)
}

func (s *service) initRegion(h *rest.Handler) {
	h.Add("GetQcloudRegion", http.MethodGet, "/config/find/config/qcloud/region", s.GetQcloudRegion)
	h.Add("GetIdcRegion", http.MethodGet, "/config/find/config/idc/region", s.GetIdcRegion)
}

func (s *service) initRequirement(h *rest.Handler) {
	h.Add("GetRequirement", http.MethodGet, "/config/find/config/requirement", s.GetRequirement)
	h.Add("CreateRequirement", http.MethodPost, "/config/create/config/requirement", s.CreateRequirement)
	h.Add("UpdateRequirement", http.MethodPut, "/config/update/config/requirement/{id}", s.UpdateRequirement)
	h.Add("DeleteRequirement", http.MethodDelete, "/config/delete/config/requirement/{id}", s.DeleteRequirement)
}

func (s *service) initSubnet(h *rest.Handler) {
	h.Add("GetSubnet", http.MethodPost, "/config/findmany/config/cvm/subnet", s.GetSubnet)
	h.Add("GetSubnetList", http.MethodPost, "/config/findmany/config/cvm/subnet/list", s.GetSubnetList)
	h.Add("UpdateSubnetProperty", http.MethodPut, "/config/updatemany/config/cvm/subnet/property",
		s.UpdateSubnetProperty)
}

func (s *service) initVpc(h *rest.Handler) {
	h.Add("GetVpc", http.MethodPost, "/config/findmany/config/cvm/vpc", s.GetVpc)
	h.Add("GetVpcList", http.MethodPost, "/config/findmany/config/cvm/vpclist", s.GetVpcList)
	h.Add("UpsertRegionDftVpc", http.MethodPost, "/config/region/default_vpc/upsert", s.UpsertRegionDftVpc)
}

func (s *service) initZone(h *rest.Handler) {
	h.Add("GetQcloudZone", http.MethodPost, "/config/findmany/config/qcloud/zone", s.GetQcloudZone)
	h.Add("GetIdcZone", http.MethodPost, "/config/findmany/config/idc/zone", s.GetIdcZone)
	h.Add("CreateIdcZone", http.MethodPost, "/config/create/config/idc/zone", s.CreateIdcZone)
}

func (s *service) initCapacity(h *rest.Handler) {
	h.Add("GetCapacity", http.MethodPost, "/config/find/cvm/capacity", s.GetCapacity)
	h.Add("UpsertCapacity", http.MethodPost, "/config/sync/cvm/capacity", s.UpsertCapacity)
	h.Add("BatchGetCapacity", http.MethodPost, "/config/findmany/cvm/capacity", s.BatchGetCapacity)
	h.Add("ListCapacityWithDeviceInfo", http.MethodPost, "/config/capacity/list_with_device_info",
		s.ListCapacityWithDeviceInfo)

}

func (s *service) initSg(h *rest.Handler) {
	h.Add("UpsertRegionDftSg", http.MethodPost, "/config/region/default_sg/upsert", s.UpsertRegionDftSg)
}
