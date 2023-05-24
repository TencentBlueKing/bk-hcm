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
	typeseip "hcm/pkg/adaptor/types/eip"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncEip ....
func (svc *service) SyncEip(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &eipHandler{cli: svc.syncCli})
}

// eipHandler eip sync handler.
type eipHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.HuaWeiSyncReq
	syncCli huawei.Interface
	// marker 取值为上一页数据的最后一条记录的id，为空时为查询第一页
	marker *string
}

var _ handler.Handler = new(eipHandler)

// Prepare ...
func (hd *eipHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *eipHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &typeseip.HuaWeiEipListOption{
		Limit:  converter.ValToPtr(int32(constant.CloudResourceSyncMaxLimit)),
		Region: hd.request.Region,
		Marker: hd.marker,
	}

	eipResult, err := hd.syncCli.CloudCli().ListEip(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list huawei eip failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(eipResult.Details) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(eipResult.Details))
	for _, one := range eipResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	hd.marker = converter.ValToPtr(eipResult.Details[len(eipResult.Details)-1].CloudID)
	return cloudIDs, nil
}

// Sync ...
func (hd *eipHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &huawei.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.Eip(kt, params, new(huawei.SyncEipOption)); err != nil {
		logs.Errorf("sync huawei eip failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *eipHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveEipDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove eip delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *eipHandler) Name() enumor.CloudResourceType {
	return enumor.EipCloudResType
}
