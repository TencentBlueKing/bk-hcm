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

package gcp

import (
	"fmt"

	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/gcp"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncCvmWithRelRes ....
func (svc *service) SyncCvmWithRelRes(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &cvmHandler{cli: svc.syncCli})
}

// cvmHandler cvm sync handler.
type cvmHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request   *sync.GcpCvmSyncReq
	syncCli   gcp.Interface
	pageToken string
}

var _ handler.Handler = new(cvmHandler)

// Prepare ...
func (hd *cvmHandler) Prepare(cts *rest.Contexts) error {
	req := new(sync.GcpCvmSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	hd.request = req
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *cvmHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &typecvm.GcpListOption{
		Zone: hd.request.Zone,
		Page: &typecore.GcpPage{
			PageToken: hd.pageToken,
			PageSize:  constant.CloudResourceSyncMaxLimit,
		},
	}

	cvmResult, token, err := hd.syncCli.CloudCli().ListCvm(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list gcp cvm failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(cvmResult) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(cvmResult))
	for _, one := range cvmResult {
		cloudIDs = append(cloudIDs, fmt.Sprint(one.Id))
	}

	hd.pageToken = token
	return cloudIDs, nil
}

// Sync ...
func (hd *cvmHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &gcp.SyncBaseParams{
		AccountID: hd.request.AccountID,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.CvmWithRelRes(kt, params, &gcp.SyncCvmWithRelResOption{
		Region: hd.request.Region,
		Zone:   hd.request.Zone,
	}); err != nil {
		logs.Errorf("sync gcp cvm failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *cvmHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveCvmDeleteFromCloud(kt, hd.request.AccountID, hd.request.Zone); err != nil {
		logs.Errorf("remove cvm delete from cloud failed, err: %v, accountID: %s, rid: %s", err,
			hd.request.AccountID, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *cvmHandler) Name() enumor.CloudResourceType {
	return enumor.CvmCloudResType
}
