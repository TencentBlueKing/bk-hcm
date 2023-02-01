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
	"fmt"
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	cloudproto "hcm/pkg/api/cloud-server/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitDiskService initialize the disk service.
func InitDiskService(c *capability.Capability) {
	svc := &diskSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("ListDisk", http.MethodPost, "/disks/list", svc.ListDisk)
	h.Add("RetrieveDisk", http.MethodGet, "/disks/{id}", svc.RetrieveDisk)
	h.Add("AssignDisk", http.MethodPost, "/disks/assign/bizs", svc.AssignDisk)

	h.Load(c.WebService)
}

type diskSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// ListDisk ...
func (dSvc *diskSvc) ListDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudproto.DiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	filterExp := req.Filter
	// 如果查询条件为空, 表示全量. 这里通过判断一个不存在的值, 表示全量返回
	if req.Filter == nil {
		filterExp = tools.AllExpression()
	}
	return dSvc.client.DataService().Global.ListDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.DiskListReq{
			Filter: filterExp,
			Page:   req.Page,
		},
	)
}

// RetrieveDisk 查询云盘详情
func (dSvc *diskSvc) RetrieveDisk(cts *rest.Contexts) (interface{}, error) {
	diskID := cts.PathParameter("id").String()

	baseInfo, err := dSvc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.CloudResourceType(disk.TableName),
		diskID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return dSvc.client.DataService().TCloud.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.Aws:
		return dSvc.client.DataService().Aws.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.HuaWei:
		return dSvc.client.DataService().HuaWei.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.Gcp:
		return dSvc.client.DataService().Gcp.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.Azure:
		return dSvc.client.DataService().Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}

// AssignDisk 将云盘分配给指定业务
func (dSvc *diskSvc) AssignDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudproto.DiskAssignReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return dSvc.client.DataService().Global.BatchUpdateDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.DiskBatchUpdateReq{IDs: req.IDs, BkBizID: req.BkBizID},
	)
}
