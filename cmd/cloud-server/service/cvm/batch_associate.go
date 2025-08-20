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
	"errors"
	"fmt"

	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	proto "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchAssociateSecurityGroups 主机批量关联安全组
func (svc *cvmSvc) BatchAssociateSecurityGroups(cts *rest.Contexts) (interface{}, error) {
	return svc.batchAssociateSecurityGroups(cts, handler.ResOperateAuth)
}

// BizBatchAssociateSecurityGroups 主机批量关联安全组
func (svc *cvmSvc) BizBatchAssociateSecurityGroups(cts *rest.Contexts) (interface{}, error) {
	return svc.batchAssociateSecurityGroups(cts, handler.BizOperateAuth)
}

func (svc *cvmSvc) batchAssociateSecurityGroups(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	cvmID, sgIDs, err := svc.decodeAndValidateAssociateSGsReq(cts, meta.Associate, validHandler)
	if err != nil {
		logs.Errorf("decode and validate batch associate security groups req failed, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, err
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", cvmID)),
		Page:   core.NewDefaultBasePage(),
	}
	cvms, err := svc.client.DataService().Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, req: %+v, rid: %s", err, listReq, cts.Kit.Rid)
		return nil, err
	}
	if len(cvms.Details) == 0 {
		logs.Errorf("cvm not found, cvm_id: %s, rid: %s", cvmID, cts.Kit.Rid)
		return nil, fmt.Errorf("cvm not found, cvm_id: %s", cvmID)
	}
	curCvm := cvms.Details[0]

	switch curCvm.Vendor {
	case enumor.TCloud:
		req := &protocvm.TCloudCvmBatchAssociateSecurityGroupReq{
			SecurityGroupIDs: sgIDs,
			CvmID:            cvmID,
			Region:           curCvm.Region,
			AccountID:        curCvm.AccountID,
		}
		err = svc.client.HCService().TCloud.Cvm.BatchAssociateSecurityGroup(cts.Kit, req)
		if err != nil {
			logs.Errorf("batch associate security groups failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, err
		}
	case enumor.Aws:
		req := &protocvm.AwsCvmBatchAssociateSecurityGroupReq{
			SecurityGroupIDs: sgIDs,
			CvmID:            cvmID,
			Region:           curCvm.Region,
			AccountID:        curCvm.AccountID,
		}
		err = svc.client.HCService().Aws.Cvm.BatchAssociateSecurityGroup(cts.Kit, req)
		if err != nil {
			logs.Errorf("batch associate security groups failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, err
		}
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support for batch associate security groups", curCvm.Vendor)
	}
	return nil, nil
}

func (svc *cvmSvc) deleteSGAndCvmRelationship(kt *kit.Kit, cvmID string, sgIDs []string) error {
	batchDeleteReq := &proto.BatchDeleteReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", cvmID),
			tools.RuleEqual("res_type", enumor.CvmCloudResType),
			tools.RuleIn("security_group_id", sgIDs),
		),
	}
	err := svc.client.DataService().Global.SGCommonRel.BatchDeleteSgCommonRels(kt, batchDeleteReq)
	if err != nil {
		logs.Errorf("delete security group and cvm relationship failed, err: %v, req: %+v, rid: %s",
			err, batchDeleteReq, kt.Rid)
		return err
	}
	return nil
}

func (svc *cvmSvc) decodeAndValidateAssociateSGsReq(cts *rest.Contexts, action meta.Action,
	validHandler handler.ValidWithAuthHandler) (cvmID string, sgIDs []string, err error) {

	cvmID = cts.PathParameter("cvm_id").String()
	if cvmID == "" {
		return "", nil, errf.NewFromErr(errf.InvalidParameter, errors.New("cvm_id is required"))
	}

	req := new(cscvm.BatchAssociateSecurityGroupsReq)
	if err := cts.DecodeInto(req); err != nil {
		return "", nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return "", nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.SecurityGroupCloudResType, IDs: req.SecurityGroupIDs,
				Fields: types.CommonBasicInfoFields},
			{ResourceType: enumor.CvmCloudResType, IDs: []string{cvmID}, Fields: types.ResWithRecycleBasicFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return "", nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: action, BasicInfos: basicInfos})
	if err != nil {
		return "", nil, err
	}

	return cvmID, req.SecurityGroupIDs, nil
}
