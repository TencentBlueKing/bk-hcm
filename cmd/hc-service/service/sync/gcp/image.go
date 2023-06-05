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
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/gcp"
	"hcm/cmd/hc-service/service/sync/handler"
	adaptorgcp "hcm/pkg/adaptor/gcp"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/image"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncImage ....
func (svc *service) SyncImage(cts *rest.Contexts) (interface{}, error) {
	imageHandler := &imageHandler{cli: svc.syncCli}
	for index, projectID := range adaptorgcp.PublicImagePlatforms {
		imageHandler.index = index
		imageHandler.projectID = projectID
		err := handler.ResourceSync(cts, imageHandler)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// imageHandler image sync handler.
type imageHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request   *sync.GcpSyncReq
	syncCli   gcp.Interface
	index     int
	projectID string
	pageToken string
}

var _ handler.Handler = new(imageHandler)

// Prepare ...
func (hd *imageHandler) Prepare(cts *rest.Contexts) error {
	if hd.index != 0 {
		return nil
	}

	req := new(sync.GcpSyncReq)
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
func (hd *imageHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &image.GcpImageListOption{
		ProjectID: hd.projectID,
		Page: &typecore.GcpPage{
			PageToken: hd.pageToken,
			PageSize:  constant.CloudResourceSyncMaxLimit,
		},
	}
	imageResult, token, err := hd.syncCli.CloudCli().ListImage(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list gcp disk failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(imageResult.Details) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(imageResult.Details))
	for _, one := range imageResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	hd.pageToken = token
	return cloudIDs, nil
}

// Sync ...
func (hd *imageHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &gcp.SyncBaseParams{
		AccountID: hd.request.AccountID,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.Image(kt, params, &gcp.SyncImageOption{Region: hd.request.Region,
		ProjectID: hd.projectID}); err != nil {
		logs.Errorf("sync gcp image failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *imageHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveImageDeleteFromCloud(kt, hd.request.AccountID, hd.projectID); err != nil {
		logs.Errorf("remove image delete from cloud failed, err: %v, accountID: %s, rid: %s", err,
			hd.request.AccountID, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *imageHandler) Name() enumor.CloudResourceType {
	return enumor.ImageCloudResType
}
