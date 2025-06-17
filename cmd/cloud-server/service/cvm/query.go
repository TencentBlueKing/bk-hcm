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
	"fmt"

	proto "hcm/pkg/api/cloud-server"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// ListCvm list cvm.
func (svc *cvmSvc) ListCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listCvm(cts, handler.ListResourceAuthRes)
}

// ListBizCvm list biz cvm.
func (svc *cvmSvc) ListBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listCvm(cts, handler.ListBizAuthRes)
}

func (svc *cvmSvc) listCvm(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.ListReq)
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

	listReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.Cvm.ListCvm(cts.Kit, listReq)
}

// GetCvm get cvm.
func (svc *cvmSvc) GetCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.getCvm(cts, handler.ListResourceAuthRes)
}

// GetBizCvm get biz cvm.
func (svc *cvmSvc) GetBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.getCvm(cts, handler.ListBizAuthRes)
}

// GetRecyclingCvm get recycled cvm.
// Deprecated: use GetCvm with recycle_status='recycling' instead
func (svc *cvmSvc) GetRecyclingCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.getCvm(cts, handler.GetRecyclingAuth)
}

// GetBizRecyclingCvm get recycled cvm that is previously in biz.
// Deprecated: use GetBizCvm with recycle_status='recycling' instead
func (svc *cvmSvc) GetBizRecyclingCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.getCvm(cts, handler.BizRecyclingAuth)
}

func (svc *cvmSvc) getCvm(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.CvmCloudResType, id, append(types.CommonBasicInfoFields, "recycle_status")...)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.Cvm, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get cvm")
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Aws:
		return svc.client.DataService().Aws.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Gcp:
		return svc.client.DataService().Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Azure:
		return svc.client.DataService().Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	case enumor.Other:
		return svc.client.DataService().Other.Cvm.GetCvm(cts.Kit, id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}

// CheckCvmsInBiz check if cvms are in the specified biz.
func CheckCvmsInBiz(kt *kit.Kit, client *client.ClientSet, rule filter.RuleFactory, bizID int64) error {
	req := &core.ListReq{
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
	result, err := client.DataService().Global.Cvm.ListCvm(kt, req)
	if err != nil {
		logs.Errorf("count cvms that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d cvms are already assigned", result.Count)
	}

	return nil
}

// QueryCvmRelatedRes ...
func (svc *cvmSvc) QueryCvmRelatedRes(cts *rest.Contexts) (interface{}, error) {
	return svc.queryCvmRelatedRes(cts, handler.ResOperateAuth)
}

// QueryBizCvmRelatedRes ...
func (svc *cvmSvc) QueryBizCvmRelatedRes(cts *rest.Contexts) (interface{}, error) {
	return svc.queryCvmRelatedRes(cts, handler.BizOperateAuth)
}

// QueryCvmRelatedRes 统计cvm 关联资源数量，目前包含disk和eip，主要用于回收时展示关联资源数量
func (svc *cvmSvc) queryCvmRelatedRes(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(cscvm.BatchQueryCvmRelatedReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          req.IDs,
	}
	basicInfo, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}
	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Find, BasicInfos: basicInfo})
	if err != nil {
		return nil, err
	}

	// 查询磁盘信息
	diskReq := &core.ListReq{
		Filter: tools.ContainersExpression("cvm_id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	diskRel, err := svc.client.DataService().Global.ListDiskCvmRel(cts.Kit, diskReq)
	if err != nil {
		logs.Errorf("fail to list disk cvm relation, err: %v, cvmIds: %v , rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}
	diskCount := make(map[string]int, len(req.IDs))
	for _, rel := range diskRel.Details {
		diskCount[rel.CvmID]++
	}

	eipList := make(map[string][]string, len(req.IDs))
	eipReq := &dataproto.EipCvmRelWithEipListReq{CvmIDs: req.IDs}
	cvmEipRel, err := svc.client.DataService().Global.ListEipCvmRelWithEip(cts.Kit, eipReq)
	if err != nil {
		logs.Errorf("fail to list cvm related eip, err: %v, cvm_ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}
	for _, eipRel := range cvmEipRel {
		eipList[eipRel.CvmID] = append(eipList[eipRel.CvmID], eipRel.PublicIp)
	}

	relatedInfos := slice.Map(req.IDs, func(cvmId string) cscvm.CvmRelatedInfo {
		return cscvm.CvmRelatedInfo{
			DiskCount: diskCount[cvmId],
			EipCount:  len(eipList[cvmId]),
			Eip:       eipList[cvmId],
		}
	})
	return relatedInfos, nil
}
