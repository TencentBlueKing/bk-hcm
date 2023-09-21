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

package cvm

import (
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
)

// BatchStopCvm batch stop cvm.
func (c *cvm) BatchStopCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	result *core.BatchOperateAllResult, err error) {

	result = &core.BatchOperateAllResult{}
	if len(basicInfoMap) == 0 {
		return result, nil
	}

	ids := make([]string, 0, len(basicInfoMap))
	for id := range basicInfoMap {
		ids = append(ids, id)
	}

	if err := c.audit.ResBaseOperationAudit(kt, enumor.CvmAuditResType, protoaudit.Stop, ids); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, kt.Rid)
		for _, cvmId := range ids {
			result.Failed = append(result.Failed, core.FailedInfo{ID: cvmId, Error: err})
		}
		return result, err
	}

	cvmVendorMap := classifier.ClassifyBasicInfoByVendor(basicInfoMap)
	for vendor, infos := range cvmVendorMap {
		switch vendor {
		case enumor.TCloud, enumor.Aws, enumor.HuaWei:
			// 支持batch
			batchStopRes := c.batchStopCvm(kt, vendor, infos)
			result.Succeeded = append(result.Succeeded, batchStopRes.Succeeded...)
			result.Failed = append(result.Failed, batchStopRes.Failed...)

		case enumor.Gcp:
			for _, cvmInfo := range infos {
				if err := c.client.HCService().Gcp.Cvm.StopCvm(kt.Ctx, kt.Header(), cvmInfo.ID); err != nil {
					result.Failed = append(result.Failed, core.FailedInfo{ID: cvmInfo.ID, Error: err})
				} else {
					result.Succeeded = append(result.Succeeded, cvmInfo.ID)
				}
			}
		case enumor.Azure:
			req := &hcprotocvm.AzureStopReq{SkipShutdown: false}
			for _, cvmInfo := range infos {
				if err := c.client.HCService().Azure.Cvm.StopCvm(kt.Ctx, kt.Header(), cvmInfo.ID, req); err != nil {
					result.Failed = append(result.Failed, core.FailedInfo{ID: cvmInfo.ID, Error: err})
				} else {
					result.Succeeded = append(result.Succeeded, cvmInfo.ID)
				}
			}
		default:
			err := errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
			for _, cvmInfo := range infos {
				result.Failed = append(result.Failed, core.FailedInfo{ID: cvmInfo.ID, Error: err})
			}
		}
	}
	return result, err
}

func (c *cvm) stopCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, string, error) {

	successIDs := make([]string, 0)
	for _, one := range infoMap {
		switch vendor {
		case enumor.Gcp:
			if err := c.client.HCService().Gcp.Cvm.StopCvm(kt.Ctx, kt.Header(), one.ID); err != nil {
				return successIDs, one.ID, err
			}

		case enumor.Azure:
			req := &hcprotocvm.AzureStopReq{
				SkipShutdown: false,
			}
			if err := c.client.HCService().Azure.Cvm.StopCvm(kt.Ctx, kt.Header(), one.ID, req); err != nil {
				return successIDs, one.ID, err
			}

		default:
			return successIDs, one.ID, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}
	}

	return successIDs, "", nil
}

// batchStopCvm stop cvm.
func (c *cvm) batchStopCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	result *core.BatchOperateAllResult) {

	result = &core.BatchOperateAllResult{}
	cvmMap := classifier.ClassifyBasicInfoByAccount(infoMap)
	markFail := func(err error, ids ...string) {
		for _, id := range ids {
			result.Failed = append(result.Failed, core.FailedInfo{ID: id, Error: err})
		}
	}
	for accountID, reginMap := range cvmMap {
		for region, ids := range reginMap {
			switch vendor {
			case enumor.TCloud:
				req := &hcprotocvm.TCloudBatchStopReq{
					AccountID:   accountID,
					Region:      region,
					IDs:         ids,
					StopType:    typecvm.SoftFirst,
					StoppedMode: typecvm.KeepCharging,
				}
				if err := c.client.HCService().TCloud.Cvm.BatchStopCvm(kt.Ctx, kt.Header(), req); err != nil {
					markFail(err, ids...)
					continue
				}

			case enumor.Aws:
				req := &hcprotocvm.AwsBatchStopReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
					Force:     true,
					Hibernate: false,
				}
				if err := c.client.HCService().Aws.Cvm.BatchStopCvm(kt.Ctx, kt.Header(), req); err != nil {
					markFail(err, ids...)
					continue
				}

			case enumor.HuaWei:
				req := &hcprotocvm.HuaWeiBatchStopReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
					Force:     true,
				}
				if err := c.client.HCService().HuaWei.Cvm.BatchStopCvm(kt.Ctx, kt.Header(), req); err != nil {
					markFail(err, ids...)
					continue
				}

			default:
				e := errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
				for _, id := range ids {
					result.Failed = append(result.Failed, core.FailedInfo{ID: id, Error: e})
				}
			}

			result.Succeeded = append(result.Succeeded, ids...)
		}
	}

	return result
}

// CheckAndStopCvm 检查在非停止状态的主机并尝试关机
func (c *cvm) checkAndStopCvm(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo) (
	result *core.BatchOperateAllResult) {

	result = &core.BatchOperateAllResult{}
	if len(infoMap) == 0 {
		return result
	}
	cvmIds := converter.MapKeyToSlice(infoMap)
	// filter out not stopped cvm
	notStoppedRule := filter.AtomRule{Field: "status", Op: filter.NotIn.Factory(), Value: []string{"STOPPING",
		"STOPPED", "stopping", "stopped", "SUSPENDING", "SUSPENDED", "PowerState/stopped", "SHUTOFF"}}
	notStoppedFilter, err := tools.And(tools.ContainersExpression("id", cvmIds), notStoppedRule)
	if err != nil {
		for _, cvmId := range cvmIds {
			result.Failed = append(result.Failed, core.FailedInfo{ID: cvmId, Error: err})
		}
		return result
	}
	notStoppedReq := &cloud.CvmListReq{Field: []string{"id"}, Filter: notStoppedFilter, Page: core.NewDefaultBasePage()}

	notStoppedCvmRes, err := c.client.DataService().Global.Cvm.ListCvm(kt.Ctx, kt.Header(), notStoppedReq)
	if err != nil {
		for _, cvmId := range cvmIds {
			result.Failed = append(result.Failed, core.FailedInfo{ID: cvmId, Error: err})
		}
		return result
	}

	notStoppedMap := make(map[string]types.CloudResourceBasicInfo)
	for _, cvm := range notStoppedCvmRes.Details {
		notStoppedMap[cvm.ID] = infoMap[cvm.ID]
	}

	// stop cvm
	stopRes, err := c.BatchStopCvm(kt, notStoppedMap)
	result.Succeeded = append(result.Succeeded, stopRes.Succeeded...)
	result.Failed = append(result.Failed, stopRes.Failed...)
	if err != nil {
		logs.Errorf("stop cvm failed, err: %v, resp: %+v, infos: %+v, rid: %s", err, stopRes, notStoppedMap, kt.Rid)
	}
	return result
}
