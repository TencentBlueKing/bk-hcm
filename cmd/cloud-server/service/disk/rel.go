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
 */

package disk

import (
	"fmt"

	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	datarelproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListDiskExtByCvmID list disk with extension by cvm_id.
func (svc *diskSvc) ListDiskExtByCvmID(cts *rest.Contexts) (interface{}, error) {
	return svc.listDiskExtByCvmID(cts, handler.ResOperateAuth)
}

// ListBizDiskExtByCvmID list biz disk with extension by cvm_id.
func (svc *diskSvc) ListBizDiskExtByCvmID(cts *rest.Contexts) (interface{}, error) {
	return svc.listDiskExtByCvmID(cts, handler.BizOperateAuth)
}

func (svc *diskSvc) listDiskExtByCvmID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error,
) {
	CvmID := cts.Request.PathParameter("cvm_id")
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.CvmCloudResType, CvmID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if basicInfo.Vendor != vendor {
		return nil, errf.NewFromErr(
			errf.InvalidParameter,
			fmt.Errorf(
				"the vendor(%s) of the cvm does not match the vendor(%s) in url path",
				basicInfo.Vendor,
				vendor,
			),
		)
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer,
		ResType: meta.Disk, Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	reqData := &datarelproto.DiskCvmRelWithDiskListReq{CvmIDs: []string{CvmID}}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.ListDiskCvmRelWithDisk(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	case enumor.Aws:
		return svc.client.DataService().Aws.ListDiskCvmRelWithDisk(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.ListDiskCvmRelWithDisk(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.ListDiskCvmRelWithDisk(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	case enumor.Azure:
		return svc.client.DataService().Azure.ListDiskCvmRelWithDisk(cts.Kit.Ctx, cts.Kit.Header(), reqData)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}

// ListDiskCvmRel ...
func (svc *diskSvc) ListDiskCvmRel(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.client.DataService().Global.ListDiskCvmRel(cts.Kit, req)
}

// ListRelWithCvm list disk with extension by cvm_id.
func (svc *diskSvc) ListRelWithCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listRelWithCvm(cts, handler.ListResourceAuthRes)
}

// ListBizRelWithCvm list biz disk with extension by cvm_id.
func (svc *diskSvc) ListBizRelWithCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listRelWithCvm(cts, handler.ListBizAuthRes)
}

func (svc *diskSvc) listRelWithCvm(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.ListWithCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Cvm, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}
	req.Filter = expr

	listReq := &datarelproto.ListWithCvmReq{
		Fields:         req.Fields,
		Filter:         req.Filter,
		Page:           req.Page,
		NotEqualDiskID: req.NotEqualDiskID,
	}
	return svc.client.DataService().Global.ListDiskCvmRelWithCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
}

// ListRelDiskWithoutCvm list disk not bind cvm
func (svc *diskSvc) ListRelDiskWithoutCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listRelDiskWithoutCvm(cts, handler.ListResourceAuthRes)
}

// ListBizRelDiskWithoutCvm list biz disk not bind cvm
func (svc *diskSvc) ListBizRelDiskWithoutCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listRelDiskWithoutCvm(cts, handler.ListBizAuthRes)
}

func (svc *diskSvc) listRelDiskWithoutCvm(cts *rest.Contexts,
	authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.ListDiskWithoutCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Disk, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}
	req.Filter = expr

	listReq := &datarelproto.ListDiskWithoutCvmReq{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.ListDiskWithoutCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
}
