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
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	securitygroup "hcm/pkg/adaptor/types/security-group"
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
	hd := &sgHandler{baseHandler: baseHandler{
		resType: enumor.SecurityGroupCloudResType,
		cli:     svc.syncCli,
	}}
	return nil, handler.ResourceSyncV2(cts, hd)
}

// sgHandler sg sync handler.
type sgHandler struct {
	baseHandler
	offset uint64
}

var _ handler.HandlerV2[securitygroup.TCloudSG] = new(sgHandler)

// Next ...
func (hd *sgHandler) Next(kt *kit.Kit) ([]securitygroup.TCloudSG, error) {
	if len(hd.request.CloudIDs) > 0 {
		// 指定id只处理一次
		listOpt := &securitygroup.TCloudListOption{
			Region:   hd.request.Region,
			CloudIDs: hd.request.CloudIDs,
			Page: &typecore.TCloudPage{
				Limit: typecore.TCloudQueryLimit,
			},
			TagFilters: hd.request.TagFilters,
		}
		sgResult, err := hd.syncCli.CloudCli().ListSecurityGroupNew(kt, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list tcloud security group failed, err: %v, opt: %+v, rid: %s",
				err, listOpt, kt.Rid)
			return nil, err
		}
		return sgResult, nil
	}

	listOpt := &securitygroup.TCloudListOption{
		Region: hd.request.Region,
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  typecore.TCloudQueryLimit,
		},
		TagFilters: hd.request.TagFilters,
	}

	sgResult, err := hd.syncCli.CloudCli().ListSecurityGroupNew(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud sg failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(sgResult) == 0 {
		return nil, nil
	}

	hd.offset += typecore.TCloudQueryLimit
	return sgResult, nil
}

// Sync ...
func (hd *sgHandler) Sync(kt *kit.Kit, instances []securitygroup.TCloudSG) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  slice.Map(instances, securitygroup.TCloudSG.GetCloudID),
	}
	if _, err := hd.syncCli.SecurityGroup(kt, params, new(tcloud.SyncSGOption)); err != nil {
		logs.Errorf("sync tcloud sg failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	// 额外从安全组视角拉取并同步关联的 CVM，补齐关联关系
	if err := hd.syncCvmBySecurityGroups(kt, params.CloudIDs); err != nil {
		logs.Errorf("sync cvm by security groups failed, err: %v, sgIDs: %v, rid: %s",
			err, params.CloudIDs, kt.Rid)
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

// RemoveDeletedFromCloud ...
func (hd *sgHandler) RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error {
	params := &tcloud.SyncRemovedParams{
		AccountID:  hd.request.AccountID,
		Region:     hd.request.Region,
		CloudIDs:   hd.request.CloudIDs,
		TagFilters: hd.request.TagFilters,
	}
	err := hd.syncCli.RemoveSGDeleteFromCloudV2(kt, params, allCloudIDMap)
	if err != nil {
		logs.Errorf("remove sg delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// syncCvmBySecurityGroups pulls CVMs associated with the given security groups and syncs their relations.
func (hd *sgHandler) syncCvmBySecurityGroups(kt *kit.Kit, sgCloudIDs []string) error {
	if len(sgCloudIDs) == 0 {
		return nil
	}

	cvmIDSet := make(map[string]struct{})

	// 按批次拆分 SGIDs，避免请求过大
	for _, partSGIDs := range slice.Split(sgCloudIDs, constant.BatchOperationMaxLimit) {
		offset := uint64(0)
		for {
			listOpt := &typecvm.ListCvmWithCountOption{
				Region: hd.request.Region,
				SGIDs:  partSGIDs,
				Page: &typecore.TCloudPage{
					Offset: offset,
					Limit:  typecore.TCloudQueryLimit,
				},
			}

			resp, err := hd.syncCli.CloudCli().ListCvmWithCount(kt, listOpt)
			if err != nil {
				return err
			}

			for _, one := range resp.Cvms {
				cvmIDSet[converter.PtrToVal(one.InstanceId)] = struct{}{}
			}

			// 分页终止条件：当前批次少于 Limit 或已达到总数
			if uint64(len(resp.Cvms)) < typecore.TCloudQueryLimit ||
				offset+typecore.TCloudQueryLimit >= uint64(resp.TotalCount) {
				break
			}
			offset += typecore.TCloudQueryLimit
		}
	}

	if len(cvmIDSet) == 0 {
		return nil
	}

	cloudIDs := make([]string, 0, len(cvmIDSet))
	for id := range cvmIDSet {
		cloudIDs = append(cloudIDs, id)
	}

	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.CvmWithRelRes(kt, params, new(tcloud.SyncCvmWithRelResOption)); err != nil {
		return err
	}
	return nil
}
