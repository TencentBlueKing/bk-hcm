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

package huawei

import (
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/huawei"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncSubnet ....
func (svc *service) SyncSubnet(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &subnetHandler{cli: svc.syncCli})
}

// subnetHandler subnet sync handler.
type subnetHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.HuaWeiSubnetSyncReq
	syncCli huawei.Interface
	// marker 分页查询起始的资源ID，为空时查询第一页
	marker *string
}

var _ handler.Handler = new(subnetHandler)

// Prepare ...
func (hd *subnetHandler) Prepare(cts *rest.Contexts) error {
	req := new(sync.HuaWeiSubnetSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	hd.request = req
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *subnetHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &adtysubnet.HuaWeiSubnetListOption{
		CloudVpcID: hd.request.CloudVpcID,
		Region:     hd.request.Region,
		Page: &typecore.HuaWeiPage{
			Limit:  converter.ValToPtr(int32(constant.CloudResourceSyncMaxLimit)),
			Marker: hd.marker,
		},
	}

	subnetResult, err := hd.syncCli.CloudCli().ListSubnet(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list huawei subnet failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(subnetResult.Details) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(subnetResult.Details))
	for _, one := range subnetResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	hd.marker = converter.ValToPtr(subnetResult.Details[len(subnetResult.Details)-1].CloudID)
	return cloudIDs, nil
}

// Sync ...
func (hd *subnetHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &huawei.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	opt := &huawei.SyncSubnetOption{
		CloudVpcID: hd.request.CloudVpcID,
	}
	if _, err := hd.syncCli.Subnet(kt, params, opt); err != nil {
		logs.Errorf("sync huawei subnet failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *subnetHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	err := hd.syncCli.RemoveSubnetDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region, hd.request.CloudVpcID)
	if err != nil {
		logs.Errorf("remove subnet delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *subnetHandler) Name() enumor.CloudResourceType {
	return enumor.SubnetCloudResType
}
