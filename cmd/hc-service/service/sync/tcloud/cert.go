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

package tcloud

import (
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	"hcm/pkg/adaptor/types/cert"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncCert ....
func (svc *service) SyncCert(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &certHandler{cli: svc.syncCli})
}

// certHandler sync handler.
type certHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request        *sync.TCloudSyncReq
	syncCli        tcloud.Interface
	offset         uint64
	cachedCertList []cert.TCloudCert
}

var _ handler.Handler = new(certHandler)

// Prepare ...
func (hd *certHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *certHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &cert.TCloudListOption{
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	result, err := hd.syncCli.CloudCli().ListCert(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud cert failed, opt: %v, err: %v, rid: %s", listOpt, err, kt.Rid)
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(result))
	for _, one := range result {
		cloudIDs = append(cloudIDs, converter.PtrToVal(one.CertificateId))
	}

	hd.offset += uint64(len(result))
	hd.cachedCertList = result
	return cloudIDs, nil
}

// Sync ...
func (hd *certHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	// 腾讯云证书api 不支持按id批量获取证书，因此将Next步骤中获取的证书直接传入
	opt := &tcloud.SyncCertOption{
		BkBizID:           constant.UnassignedBiz,
		PreCachedCertList: hd.cachedCertList,
	}
	if _, err := hd.syncCli.Cert(kt, params, opt); err != nil {
		logs.Errorf("sync tcloud cert failed, opt: %v, err: %v, rid: %s", params, err, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *certHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveCertDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove cert delete from cloud failed, accountID: %s, region: %s, err: %v, rid: %s",
			hd.request.AccountID, hd.request.Region, err, kt.Rid)
		return err
	}

	return nil
}

// Name get cloud resource type name
func (hd *certHandler) Name() enumor.CloudResourceType {
	return enumor.CertCloudResType
}
