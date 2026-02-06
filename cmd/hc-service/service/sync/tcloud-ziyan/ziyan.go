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

// Package ziyan ...
package ziyan

import (
	"fmt"

	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/ziyan"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/cvmapi"
)

// InitService initial tcloud sync service
func InitService(cap *capability.Capability) {
	v := &service{
		ad:      cap.CloudAdaptor,
		cs:      cap.ClientSet,
		dataCli: cap.ClientSet.DataService(),
		syncCli: cap.ResSyncCli,
		crpCli:  cap.CrpCli,
	}

	h := rest.NewHandler()
	h.Path("/vendors/tcloud-ziyan")

	h.Add("SyncSecurityGroup", "POST", "/security_groups/sync", v.SyncSecurityGroup)
	h.Add("SyncZone", "POST", "/zones/sync", v.SyncZone)
	h.Add("SyncRegion", "POST", "/regions/sync", v.SyncRegion)
	h.Add("SyncArgsTpl", "POST", "/argument_templates/sync", v.SyncArgsTpl)
	h.Add("SyncCert", "POST", "/certs/sync", v.SyncCert)
	h.Add("SyncLoadBalancer", "POST", "/load_balancers/sync", v.SyncLoadBalancer)
	h.Add("SyncVpc", "POST", "/vpcs/sync", v.SyncVpc)
	h.Add("SyncSubnet", "POST", "/subnets/sync", v.SyncSubnet)
	h.Add("SyncHostWithRelRes", "POST", "/hosts/with/relation_resources/sync", v.SyncHostWithRelRes)
	h.Add("SyncHostWithRelResByCond", "POST", "/hosts/with/relation_resources/by_condition/sync",
		v.SyncHostWithRelResByCond)
	h.Add("DeleteHost", "DELETE", "/hosts/by_condition/delete", v.DeleteHostByCond)
	h.Add("SyncSecurityGroupUsageBiz", "POST", "/security_groups/usage_biz_rels/sync", v.SyncSecurityGroupUsageBiz)
	h.Add("SyncImage", "POST", "/images/sync", v.SyncImage)
	h.Add("SyncDeviceType", "POST", "/device_types/sync", v.SyncDeviceType)

	h.Load(cap.WebService)
}

type service struct {
	ad      *cloudadaptor.CloudAdaptorClient
	cs      *client.ClientSet
	dataCli *dataservice.Client
	syncCli ressync.Interface
	crpCli  cvmapi.CVMClientInterface
}

func defaultPrepare(cts *rest.Contexts, cli ressync.Interface) (*sync.TCloudSyncReq, ziyan.Interface, error) {
	req := new(sync.TCloudSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := cli.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, nil, err
	}

	return req, syncCli, nil
}

// baseHandler ...
type baseHandler struct {
	resType enumor.CloudResourceType
	request *sync.TCloudSyncReq
	cli     ressync.Interface

	syncCli ziyan.Interface
}

// Describe load_balancer
func (hd *baseHandler) Describe() string {
	if hd.request == nil {
		return fmt.Sprintf("ziyan %s(-)", hd.Resource())
	}
	return fmt.Sprintf("ziyan %s(region=%s,account=%s)", hd.Resource(), hd.request.Region, hd.request.AccountID)
}

// SyncConcurrent use request specified or 1
func (hd *baseHandler) SyncConcurrent() uint {
	if hd.request != nil && hd.request.Concurrent != 0 {
		return hd.request.Concurrent
	}
	// read from config file
	_, syncing := cc.HCService().SyncConfig.GetSyncConcurrent(enumor.Ziyan, hd.resType, hd.request.Region)
	return max(syncing, 1)
}

// Resource return resource type of handler
func (hd *baseHandler) Resource() enumor.CloudResourceType {
	return hd.resType
}

// Prepare ...
func (hd *baseHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}
