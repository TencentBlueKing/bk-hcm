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

	"hcm/cmd/cloud-server/logics/audit"
	cloudproto "hcm/pkg/api/cloud-server/disk"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

type diskSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// ListDisk ...
func (svc *diskSvc) ListDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudproto.DiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	authOpt := &meta.ListAuthResInput{Type: meta.Disk, Action: meta.Find}
	expr, noPermFlag, err := svc.authorizer.ListAuthInstWithFilter(cts.Kit, authOpt, req.Filter, "account_id")
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &dataproto.DiskListResult{Details: make([]*dataproto.DiskResult, 0)}, nil
	}

	return svc.client.DataService().Global.ListDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.DiskListReq{
			Filter: expr,
			Page:   req.Page,
		},
	)
}

// DeleteDisk 删除云盘
func (svc *diskSvc) DeleteDisk(cts *rest.Contexts) (interface{}, error) {
	diskID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.CloudResourceType(disk.TableName),
		diskID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// TODO 增加鉴权和审计
	// TODO 判断云盘是否可删除

	deleteReq := &hcproto.DiskDeleteReq{DiskID: diskID, AccountID: basicInfo.AccountID}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Disk.DeleteDisk(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.Aws:
		return nil, svc.client.HCService().Aws.Disk.DeleteDisk(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.HuaWei:
		return nil, svc.client.HCService().HuaWei.Disk.DeleteDisk(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.Gcp:
		return nil, svc.client.HCService().Gcp.Disk.DeleteDisk(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.Azure:
		return nil, svc.client.HCService().Azure.Disk.DeleteDisk(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}

// RetrieveDisk 查询云盘详情
func (svc *diskSvc) RetrieveDisk(cts *rest.Contexts) (interface{}, error) {
	diskID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.CloudResourceType(disk.TableName),
		diskID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{
		Type: meta.Disk, Action: meta.Find,
		ResourceID: basicInfo.AccountID,
	}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.Aws:
		return svc.client.DataService().Aws.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	case enumor.Azure:
		return svc.client.DataService().Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}

// AssignDisk 将云盘分配给指定业务
func (svc *diskSvc) AssignDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudproto.DiskAssignReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authorizeDiskAssignOp(cts.Kit, req.IDs); err != nil {
		return nil, err
	}

	// check if all disks are not assigned to biz, right now assigning resource twice is not allowed
	diskFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs}
	err := svc.checkDisksInBiz(cts.Kit, diskFilter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create assign audit
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.EipAuditResType, req.IDs, int64(req.BkBizID))
	if err != nil {
		logs.Errorf("create assign disk audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.client.DataService().Global.BatchUpdateDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.DiskBatchUpdateReq{IDs: req.IDs, BkBizID: req.BkBizID},
	)
}

// AttachDisk ...
func (svc *diskSvc) AttachDisk(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.tcloudAttachDisk(cts)
	case enumor.Aws:
		return svc.awsAttachDisk(cts)
	case enumor.HuaWei:
		return svc.huaweiAttachDisk(cts)
	case enumor.Gcp:
		return svc.gcpAttachDisk(cts)
	case enumor.Azure:
		return svc.azureAttachDisk(cts)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// DetachDisk ...
func (svc *diskSvc) DetachDisk(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloudproto.DiskDetachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// TODO 增加鉴权和审计

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.DiskCloudResType,
		req.DiskID,
	)
	if err != nil {
		return nil, err
	}

	detachReq := &hcproto.DiskDetachReq{
		AccountID: basicInfo.AccountID,
		CvmID:     req.CvmID,
		DiskID:    req.DiskID,
	}

	switch vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Disk.DetachDisk(cts.Kit.Ctx, cts.Kit.Header(), detachReq)
	case enumor.Aws:
		return nil, svc.client.HCService().Aws.Disk.DetachDisk(cts.Kit.Ctx, cts.Kit.Header(), detachReq)
	case enumor.HuaWei:
		return nil, svc.client.HCService().HuaWei.Disk.DetachDisk(cts.Kit.Ctx, cts.Kit.Header(), detachReq)
	case enumor.Gcp:
		return nil, svc.client.HCService().Gcp.Disk.DetachDisk(cts.Kit.Ctx, cts.Kit.Header(), detachReq)
	case enumor.Azure:
		return nil, svc.client.HCService().Gcp.Disk.DetachDisk(cts.Kit.Ctx, cts.Kit.Header(), detachReq)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (svc *diskSvc) authorizeDiskAssignOp(kt *kit.Kit, ids []string) error {
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(
		kt.Ctx,
		kt.Header(),
		basicInfoReq,
	)
	if err != nil {
		return err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{
			Type: meta.Disk, Action: meta.Assign,
			ResourceID: info.AccountID,
		}})
	}
	err = svc.authorizer.AuthorizeWithPerm(kt, authRes...)
	if err != nil {
		return err
	}

	return nil
}

// checkDisksInBiz check if disks are in the specified biz.
func (svc *diskSvc) checkDisksInBiz(kt *kit.Kit, rule filter.RuleFactory, bizID int64) error {
	req := &dataproto.DiskListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: bizID}, rule,
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := svc.client.DataService().Global.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count disks that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != nil && *result.Count != 0 {
		return fmt.Errorf("%d disks are already assigned", result.Count)
	}

	return nil
}
