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

package aws

import (
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/aws"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncSecurityGroup ....
func (svc *service) SyncSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &sgHandler{cli: svc.syncCli})
}

// sgHandler sg sync handler.
type sgHandler struct {
	cli ressync.Interface

	// Perpare 构建参数
	request   *sync.AwsSyncReq
	syncCli   aws.Interface
	nextToken *string
}

var _ handler.Handler = new(sgHandler)

// Prepare ...
func (hd *sgHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *sgHandler) Next(kt *kit.Kit) ([]string, error) {
	if len(hd.request.CloudIDs) > 0 {
		// 指定id只处理一次
		listOpt := &securitygroup.AwsListOption{
			Region:   hd.request.Region,
			CloudIDs: hd.request.CloudIDs,
		}
		sgResult, _, err := hd.syncCli.CloudCli().ListSecurityGroup(kt, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list aws sg failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
			return nil, err
		}
		return slice.Map(sgResult, func(one securitygroup.AwsSG) string { return converter.PtrToVal(one.GroupId) }), nil
	}
	listOpt := &securitygroup.AwsListOption{
		Region: hd.request.Region,
		Page: &typecore.AwsPage{
			NextToken:  hd.nextToken,
			MaxResults: converter.ValToPtr(int64(constant.CloudResourceSyncMaxLimit)),
		},
	}

	sgResult, resp, err := hd.syncCli.CloudCli().ListSecurityGroup(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list aws sg failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(sgResult) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(sgResult))
	for _, one := range sgResult {
		cloudIDs = append(cloudIDs, converter.PtrToVal(one.GroupId))
	}

	hd.nextToken = resp.NextToken
	return cloudIDs, nil
}

// Sync ...
func (hd *sgHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &aws.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.SecurityGroup(kt, params, new(aws.SyncSGOption)); err != nil {
		logs.Errorf("sync aws sg failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *sgHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveSecurityGroupDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove sg delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *sgHandler) Name() enumor.CloudResourceType {
	return enumor.SecurityGroupCloudResType
}
