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
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// GetMonitorData get cvm monitor data.
func (svc *cvmSvc) GetMonitorData(cts *rest.Contexts) (interface{}, error) {
	return svc.getMonitorData(cts, handler.ListResourceAuthRes)
}

func (svc *cvmSvc) getMonitorData(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, err
	}

	req := new(cscvm.GetMonitorDataReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listFilter := tools.ContainersExpression("id", req.IDs)

	// 权限校验
	// list authorized instances
	authFilter, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Cvm, Action: meta.Find, Filter: listFilter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return nil, errf.New(errf.InvalidParameter, "no permission")
	}

	// 从数据库查询实例信息，获取region和account_id
	listReq := &core.ListReq{
		Filter: authFilter,
		Page:   core.NewDefaultBasePage(),
	}

	cvms, err := svc.client.DataService().Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list cvm from db failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	if len(cvms.Details) != len(req.IDs) {
		return nil, errf.Newf(errf.RecordNotFound, "some instances not found, need: %d, got: %d",
			len(req.IDs), len(cvms.Details))
	}

	switch vendor {
	case enumor.TCloud:
		return svc.getTCloudMonitorData(cts, req, cvms.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor %s is not supported", vendor)
	}
}

func (svc *cvmSvc) getTCloudMonitorData(cts *rest.Contexts, req *cscvm.GetMonitorDataReq, cvms []cvm.BaseCvm) (
	interface{}, error) {

	// 按account_id + region分组
	// 定义分组key结构
	type groupKey struct {
		AccountID string
		Region    string
	}
	instanceGroups := make(map[groupKey][]cvm.BaseCvm)

	for _, instance := range cvms {
		key := groupKey{
			AccountID: instance.AccountID,
			Region:    instance.Region,
		}
		instanceGroups[key] = append(instanceGroups[key], instance)
	}

	// 合并所有分组的监控数据
	allDataPoints := make([]*cscvm.MonitorDataPointResp, 0)

	// 按account_id + region分组调用hc-service
	for key, instances := range instanceGroups {
		cloudIDToInst := slice.FuncToMap(instances, func(instance cvm.BaseCvm) (string, cvm.BaseCvm) {
			return instance.CloudID, instance
		})
		cloudIDs := maps.Keys(cloudIDToInst)

		hcReq := &protocvm.TCloudMonitorDataReq{
			AccountID:   key.AccountID,
			Region:      key.Region,
			MetricName:  req.MetricName,
			Period:      req.Period,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			InstanceIDs: cloudIDs,
		}

		resp, err := svc.client.HCService().TCloud.Cvm.GetMonitorData(cts.Kit, hcReq)
		if err != nil {
			logs.Errorf("get monitor data failed, err: %v, account_id: %s, region: %s, rid: %s",
				err, key.AccountID, key.Region, cts.Kit.Rid)
			return nil, err
		}

		for _, dataPoint := range resp.DataPoints {
			var cloudID string
			for _, dimension := range dataPoint.Dimensions {
				if dimension.Name == constant.TCloudCvmInstanceIDKey {
					cloudID = dimension.Value
				}
			}
			if len(cloudID) == 0 {
				logs.Warnf("instance_id dimension key not found, account_id: %s, region: %s, rid: %s",
					key.AccountID, key.Region, cts.Kit.Rid)
				continue
			}
			instDetail, ok := cloudIDToInst[cloudID]
			if !ok {
				logs.Warnf("instance not found, account_id: %s, region: %s, cloud_id: %s, rid: %s",
					key.AccountID, key.Region, cloudID, cts.Kit.Rid)
				continue
			}

			allDataPoints = append(allDataPoints, &cscvm.MonitorDataPointResp{
				ID:         instDetail.ID,
				IP:         instDetail.PrivateIPv4Addresses,
				Region:     key.Region,
				InstanceID: cloudID,
				Timestamps: dataPoint.Timestamps,
				Values:     dataPoint.Values,
			})
		}
	}

	return &cscvm.GetMonitorDataResp{
		DataPoints: allDataPoints,
	}, nil
}
