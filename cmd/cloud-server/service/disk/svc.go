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
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
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
		Authorizer: svc.authorizer, ResType: meta.Disk, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &cloudproto.DiskListResult{Details: make([]*cloudproto.DiskResult, 0)}, nil
	}

	resp, err := svc.client.DataService().Global.ListDisk(cts.Kit,
		&core.ListReq{Filter: expr, Page: req.Page})
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

	rels, err := svc.client.DataService().Global.ListDiskCvmRel(cts.Kit,
		&core.ListReq{
			Filter: tools.ContainersExpression("disk_id", diskIDs),
			Page:   core.NewDefaultBasePage(),
		})
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

		details[idx] = &cloudproto.DiskResult{InstanceID: diskIDToCvmID[diskData.ID], InstanceType: "cvm",
			BaseDisk: diskData,
		}
	}

	return &cloudproto.DiskListResult{Details: details, Count: resp.Count}, nil
}

// DeleteDisk 删除云盘.
func (svc *diskSvc) DeleteDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteDisk(cts, handler.ResOperateAuth)
}

// DeleteBizDisk 删除业务下的云盘.
func (svc *diskSvc) DeleteBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteDisk(cts, handler.BizOperateAuth)
}

func (svc *diskSvc) deleteDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	diskID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.CloudResourceType(table.DiskTable), diskID, append(types.CommonBasicInfoFields, "recycle_status")...)
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

	err = svc.diskLgc.DeleteDisk(cts.Kit, basicInfo.Vendor, basicInfo.ID)
	if err != nil {
		return nil, err
	}
	return nil, err
}

// GetDisk 查询云盘详情.
func (svc *diskSvc) GetDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveDisk(cts, handler.ListResourceAuthRes)
}

// GetBizDisk 查询业务下的云盘详情.
func (svc *diskSvc) GetBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveDisk(cts, handler.ListBizAuthRes)
}

// GetBizRecycledDisk get biz recycled disk.
// Deprecated: use GetDisk with recycle_status='recycling' instead
func (svc *diskSvc) GetBizRecycledDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveDisk(cts, handler.GetRecyclingAuth)
}

// GetRecycledDisk get recycled disk.
// Deprecated: use GetBizDisk with recycle_status='recycling' instead
func (svc *diskSvc) GetRecycledDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveDisk(cts, handler.BizRecyclingAuth)
}

func (svc *diskSvc) retrieveDisk(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (interface{}, error) {
	diskID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.CloudResourceType(table.DiskTable), diskID, append(types.CommonBasicInfoFields, "recycle_status")...)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authOpt := &handler.ListAuthResOption{
		Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Find,
	}
	// validate biz and authorize
	_, noPerm, err := validHandler(cts, authOpt)
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get disk")
	}

	rels, err := svc.client.DataService().Global.ListDiskCvmRel(cts.Kit,
		&core.ListReq{
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

func (svc *diskSvc) retrieveDiskByVendor(cts *rest.Contexts, vendor enumor.Vendor,
	diskID string, instID string, instName string) (interface{}, error) {

	switch vendor {
	case enumor.TCloud:
		resp, err := svc.client.DataService().TCloud.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.TCloudDiskExtResult{
			Disk:         resp,
			InstanceType: string(enumor.DiskBindCvm),
			InstanceID:   instID,
			InstanceName: instName,
		}, nil
	case enumor.Aws:
		resp, err := svc.client.DataService().Aws.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.AwsDiskExtResult{
			Disk:         resp,
			InstanceType: string(enumor.DiskBindCvm),
			InstanceID:   instID,
			InstanceName: instName,
		}, nil
	case enumor.HuaWei:
		resp, err := svc.client.DataService().HuaWei.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.HuaWeiDiskExtResult{
			Disk:         resp,
			InstanceType: string(enumor.DiskBindCvm),
			InstanceID:   instID,
			InstanceName: instName,
		}, nil
	case enumor.Gcp:
		resp, err := svc.client.DataService().Gcp.RetrieveDisk(cts.Kit, diskID)
		if err != nil {
			return nil, err
		}

		resp.DiskType = extractGcpDiskType(resp.DiskType)
		return cloudproto.GcpDiskExtResult{
			Disk:         resp,
			InstanceType: string(enumor.DiskBindCvm),
			InstanceID:   instID,
			InstanceName: instName,
		}, nil
	case enumor.Azure:
		resp, err := svc.client.DataService().Azure.RetrieveDisk(cts.Kit.Ctx, cts.Kit.Header(), diskID)
		if err != nil {
			return nil, err
		}
		return cloudproto.AzureDiskExtResult{
			Disk:         resp,
			InstanceType: string(enumor.DiskBindCvm),
			InstanceID:   instID,
			InstanceName: instName,
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

	return nil, disklgc.Assign(cts.Kit, svc.client.DataService(), req.IDs, req.BkBizID, false)
}

func (svc *diskSvc) authorizeDiskAssignOp(kt *kit.Kit, ids []string) error {
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          ids,
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(kt, basicInfoReq)
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
		cts.Kit,
		&core.ListReq{
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
