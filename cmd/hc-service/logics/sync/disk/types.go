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

package disk

import (
	typescore "hcm/pkg/adaptor/types/core"
	typedisk "hcm/pkg/adaptor/types/disk"
	"hcm/pkg/api/core"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	"google.golang.org/api/compute/v1"
)

// TCloudDiskSyncDiff diff tcloud disk
type TCloudDiskSyncDiff struct {
	Disk *cbs.Disk
}

// HuaWeiDiskSyncDiff diff huawei disk
type HuaWeiDiskSyncDiff struct {
	Disk model.VolumeDetail
}

// GcpDiskSyncDiff diff gcp disk
type GcpDiskSyncDiff struct {
	Disk *compute.Disk
}

// AzureDiskSyncDiff diff azure disk struct
type AzureDiskSyncDiff struct {
	Disk *typedisk.AzureDisk
}

// AwsDiskSyncDiff aws disk diff struct
type AwsDiskSyncDiff struct {
	Disk *ec2.Volume
}

// DiskSyncDS disk data-service
type DiskSyncDS struct {
	IsUpdated bool
	HcDisk    *dataproto.DiskResult
}

// TCloudDiskSyncDS disk data-service
type TCloudDiskSyncDS struct {
	IsUpdated bool
	HcDisk    *dataproto.DiskExtResult[dataproto.TCloudDiskExtensionResult]
}

// HuaWeiDiskSyncDS disk data-service
type HuaWeiDiskSyncDS struct {
	IsUpdated bool
	HcDisk    *dataproto.DiskExtResult[dataproto.HuaWeiDiskExtensionResult]
}

// GcpDiskSyncDS disk data-service
type GcpDiskSyncDS struct {
	IsUpdated bool
	HcDisk    *dataproto.DiskExtResult[dataproto.GcpDiskExtensionResult]
}

// AzureDiskSyncDS disk data-service
type AzureDiskSyncDS struct {
	IsUpdated bool
	HcDisk    *dataproto.DiskExtResult[dataproto.AzureDiskExtensionResult]
}

// AwsDiskSyncDS disk data-service
type AwsDiskSyncDS struct {
	IsUpdated bool
	HcDisk    *dataproto.DiskExtResult[dataproto.AwsDiskExtensionResult]
}

// GetDatasFromDSForCvmDiskSync ...
func GetDatasFromDSForCvmDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
	dataCli *dataservice.Client) (map[string]*DiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskResult, 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		if len(req.SelfLinks) > 0 {
			filter := filter.AtomRule{Field: "extension.self_link", Op: filter.JSONIn.Factory(),
				Value: req.SelfLinks}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Global.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	dsMap := make(map[string]*DiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(DiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

// GetDatasFromDSForDiskSync ...
func GetDatasFromDSForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
	dataCli *dataservice.Client) (map[string]*DiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskResult, 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Global.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	dsMap := make(map[string]*DiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(DiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

// GetDatasFromDSForGcpDiskSync ...
func GetDatasFromDSForGcpDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
	dataCli *dataservice.Client) (map[string]*DiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskResult, 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "zone", Op: filter.Equal.Factory(), Value: req.Zone},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Global.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	dsMap := make(map[string]*DiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(DiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

// GetSelfLinkMapFromDSForDiskSync ...
func GetSelfLinkMapFromDSForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
	dataCli *dataservice.Client) (map[string]*DiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskExtResult[datadisk.GcpDiskExtensionResult], 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		if len(req.SelfLinks) > 0 {
			filter := filter.AtomRule{Field: "extension.self_link", Op: filter.JSONIn.Factory(),
				Value: req.SelfLinks}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Gcp.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	dsMap := make(map[string]*DiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(DiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = &datadisk.DiskResult{
			ID:        result.ID,
			Vendor:    result.Vendor,
			AccountID: result.AccountID,
			Name:      result.Name,
			BkBizID:   result.BkBizID,
			CloudID:   result.CloudID,
			Region:    result.Region,
			Zone:      result.Zone,
			DiskSize:  result.DiskSize,
			DiskType:  result.DiskType,
			Status:    result.Status,
			Memo:      result.Memo,
			Creator:   result.Creator,
			Reviser:   result.Reviser,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		}
		dsMap[result.Extension.SelfLink] = sg
	}

	return dsMap, nil
}

// DiffDiskSyncDelete ...
func DiffDiskSyncDelete(kt *kit.Kit, deleteCloudIDs []string,
	dataCli *dataservice.Client) error {

	elems := slice.Split(deleteCloudIDs, typescore.TCloudQueryLimit)

	for _, partDeleteCloudIDs := range elems {
		batchDeleteReq := &datadisk.DiskDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", partDeleteCloudIDs),
		}
		if _, err := dataCli.Global.DeleteDisk(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
			logs.Errorf("request dataservice delete tcloud disk failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
