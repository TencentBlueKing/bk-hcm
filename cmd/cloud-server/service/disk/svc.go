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
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"hcm/cmd/cloud-server/logics/audit"
	disklgc "hcm/cmd/cloud-server/logics/disk"
	cloudproto "hcm/pkg/api/cloud-server/disk"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	datarelproto "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

type diskSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	diskLgc    disklgc.Interface
}

// ListDisk list disk.
func (svc *diskSvc) ListDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.listDisk(cts, handler.ListResourceAuthRes)
}

// ListBizDisk list biz disk.
func (svc *diskSvc) ListBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.listDisk(cts, handler.ListBizAuthRes)
}

func (svc *diskSvc) listDisk(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(cloudproto.DiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.Disk, Action: meta.Find, Filter: req.Filter,
	})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &cloudproto.DiskListResult{Details: make([]*cloudproto.DiskResult, 0)}, nil
	}

	// filter out disk in recycle bin
	req.Filter, err = tools.And(expr, &filter.AtomRule{
		Field: "recycle_status", Op: filter.NotEqual.Factory(),
		Value: enumor.RecycleStatus,
	})
	if err != nil {
		return nil, err
	}

	listReq := &dataproto.DiskListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	resp, err := svc.client.DataService().Global.ListDisk(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}

	if len(resp.Details) == 0 {
		return &cloudproto.DiskListResult{Details: make([]*cloudproto.DiskResult, 0), Count: resp.Count}, nil
	}

	diskIDs := make([]string, len(resp.Details))
	for idx, diskData := range resp.Details {
		diskIDs[idx] = diskData.ID
	}

	listRelReq := &datarelproto.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("disk_id", diskIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rels, err := svc.client.DataService().Global.ListDiskCvmRel(cts.Kit.Ctx, cts.Kit.Header(), listRelReq)
	if err != nil {
		return nil, err
	}

	diskIDToCvmID := make(map[string]string)
	for _, relData := range rels.Details {
		diskIDToCvmID[relData.DiskID] = relData.CvmID
	}

	details := make([]*cloudproto.DiskResult, len(resp.Details))
	for idx, diskData := range resp.Details {
		// Gcp disk type 截取类型展示
		if diskData.Vendor == string(enumor.Gcp) {
			diskData.DiskType = extractGcpDiskType(diskData.DiskType)
		}

		details[idx] = &cloudproto.DiskResult{
			InstanceID:   diskIDToCvmID[diskData.ID],
			InstanceType: "cvm",
			DiskResult:   diskData,
		}
	}

	return &cloudproto.DiskListResult{Details: details, Count: resp.Count}, nil
}

// DeleteDisk 删除云盘.
func (svc *diskSvc) DeleteDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteDisk(cts, handler.ResValidWithAuth)
}

// DeleteBizDisk 删除业务下的云盘.
func (svc *diskSvc) DeleteBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteDisk(cts, handler.BizValidWithAuth)
}

func (svc *diskSvc) deleteDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	diskID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.CloudResourceType(table.DiskTable),
		diskID,
		append(types.CommonBasicInfoFields, "recycle_status")...,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Delete, BasicInfo: basicInfo,
	})
	if err != nil {
		return nil, err
	}

	err = svc.diskLgc.DeleteDisk(cts.Kit, basicInfo.Vendor, basicInfo.ID, basicInfo.AccountID)
	if err != nil {
		return nil, err
	}

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

// RetrieveDisk 查询云盘详情.
func (svc *diskSvc) RetrieveDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveDisk(cts, handler.ResValidWithAuth)
}

// RetrieveBizDisk 查询业务下的云盘详情.
func (svc *diskSvc) RetrieveBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveDisk(cts, handler.BizValidWithAuth)
}

// RetrieveRecycledDisk get recycled disk.
func (svc *diskSvc) RetrieveRecycledDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveDisk(cts, handler.RecycleValidWithAuth)
}

func (svc *diskSvc) retrieveDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	diskID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.CloudResourceType(table.DiskTable),
		diskID,
		append(types.CommonBasicInfoFields, "recycle_status")...,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Find, BasicInfo: basicInfo,
	})
	if err != nil {
		return nil, err
	}

	rels, err := svc.client.DataService().Global.ListDiskCvmRel(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&datarelproto.DiskCvmRelListReq{
			Filter: tools.EqualExpression("disk_id", diskID),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		return nil, err
	}

	var instanceID, instanceName string
	if len(rels.Details) > 0 {
		instanceID = rels.Details[0].CvmID
		instanceName, err = getCvmName(cts, svc.client.DataService(), instanceID)
		if err != nil {
			return nil, err
		}
	}

	return svc.retrieveDiskByVendor(cts, basicInfo.Vendor, diskID, instanceID, instanceName)
}

func (svc *diskSvc) retrieveDiskByVendor(
	cts *rest.Contexts,
	vendor enumor.Vendor,
	diskID string,
	instID string,
	instName string,
) (interface{}, error) {
	switch vendor {
	case enumor.TCloud:
		resp, err := svc.client.DataService().TCloud.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.TCloudDiskExtResult{
			DiskExtResult: resp,
			InstanceType:  string(enumor.DiskBindCvm),
			InstanceID:    instID,
			InstanceName:  instName,
		}, nil
	case enumor.Aws:
		resp, err := svc.client.DataService().Aws.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.AwsDiskExtResult{
			DiskExtResult: resp,
			InstanceType:  string(enumor.DiskBindCvm),
			InstanceID:    instID,
			InstanceName:  instName,
		}, nil
	case enumor.HuaWei:
		resp, err := svc.client.DataService().HuaWei.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.HuaWeiDiskExtResult{
			DiskExtResult: resp,
			InstanceType:  string(enumor.DiskBindCvm),
			InstanceID:    instID,
			InstanceName:  instName,
		}, nil
	case enumor.Gcp:
		resp, err := svc.client.DataService().Gcp.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}

		resp.DiskType = extractGcpDiskType(resp.DiskType)
		return cloudproto.GcpDiskExtResult{
			DiskExtResult: resp,
			InstanceType:  string(enumor.DiskBindCvm),
			InstanceID:    instID,
			InstanceName:  instName,
		}, nil
	case enumor.Azure:
		resp, err := svc.client.DataService().Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.AzureDiskExtResult{
			DiskExtResult: resp,
			InstanceType:  string(enumor.DiskBindCvm),
			InstanceID:    instID,
			InstanceName:  instName,
		}, nil
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
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

// AttachDisk attach disk.
func (svc *diskSvc) AttachDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.attachDisk(cts, handler.ResValidWithAuth)
}

// AttachBizDisk  attach biz disk.
func (svc *diskSvc) AttachBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.attachDisk(cts, handler.BizValidWithAuth)
}

func (svc *diskSvc) attachDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	diskID, err := extractDiskID(cts)
	if err != nil {
		return nil, err
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.DiskCloudResType,
		diskID,
		append(types.CommonBasicInfoFields, "recycle_status")...,
	)
	if err != nil {
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.tcloudAttachDisk(cts, basicInfo, validHandler)
	case enumor.Aws:
		return svc.awsAttachDisk(cts, basicInfo, validHandler)
	case enumor.HuaWei:
		return svc.huaweiAttachDisk(cts, basicInfo, validHandler)
	case enumor.Gcp:
		return svc.gcpAttachDisk(cts, basicInfo, validHandler)
	case enumor.Azure:
		return svc.azureAttachDisk(cts, basicInfo, validHandler)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}

// DetachDisk detach disk.
func (svc *diskSvc) DetachDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.detachDisk(cts, handler.ResValidWithAuth)
}

// DetachBizDisk  detach biz disk.
func (svc *diskSvc) DetachBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.detachDisk(cts, handler.BizValidWithAuth)
}

func (svc *diskSvc) detachDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(cloudproto.DiskDetachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.DiskCloudResType,
		req.DiskID,
		append(types.CommonBasicInfoFields, "recycle_status")...,
	)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Disassociate, BasicInfo: basicInfo,
	})
	if err != nil {
		return nil, err
	}

	// get cvm id
	rels, err := svc.client.DataService().Global.ListDiskCvmRel(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&datarelproto.DiskCvmRelListReq{
			Filter: tools.EqualExpression("disk_id", req.DiskID),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if len(rels.Details) == 0 {
		return nil, fmt.Errorf("disk(%s) not attached", req.DiskID)
	}

	cvmID := rels.Details[0].CvmID

	err = svc.diskLgc.DetachDisk(cts.Kit, basicInfo.Vendor, cvmID, req.DiskID, basicInfo.AccountID)
	if err != nil {
		logs.Errorf("detach disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

func (svc *diskSvc) authorizeDiskAssignOp(kt *kit.Kit, ids []string) error {
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          ids,
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
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
		}, BizID: info.BkBizID})
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

func extractDiskID(cts *rest.Contexts) (string, error) {
	req := new(cloudproto.DiskReq)
	reqData, err := ioutil.ReadAll(cts.Request.Request.Body)
	if err != nil {
		logs.Errorf("read request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", err
	}

	cts.Request.Request.Body = ioutil.NopCloser(bytes.NewReader(reqData))
	if err := cts.DecodeInto(req); err != nil {
		return "", err
	}

	if err := req.Validate(); err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	cts.Request.Request.Body = ioutil.NopCloser(bytes.NewReader(reqData))

	return req.DiskID, nil
}

func getCvmName(cts *rest.Contexts, cli *dataservice.Client, cvmID string) (string, error) {
	cvms, err := cli.Global.Cvm.ListCvm(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&cloud.CvmListReq{
			Filter: tools.ContainersExpression("id", []string{cvmID}),
			Page: &core.BasePage{
				Limit: core.DefaultMaxPageLimit,
			},
		},
	)
	if err != nil {
		return "", err
	}

	if len(cvms.Details) > 0 {
		return cvms.Details[0].Name, nil
	}
	return "", fmt.Errorf("cvm(%s) does not exist", cvmID)
}

func extractGcpDiskType(rawDiskType string) string {
	lastIdx := strings.LastIndex(rawDiskType, "/")
	return rawDiskType[lastIdx+1:]
}
