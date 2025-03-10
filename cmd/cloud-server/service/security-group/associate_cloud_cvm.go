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

package securitygroup

import (
	protoaudit "hcm/pkg/api/data-service/audit"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchAssociateCvm 关联安全组，安全组本地id，cvm 云id
func (svc *securityGroupSvc) BatchAssociateCvm(cts *rest.Contexts) (any, error) {
	return svc.batchAssociateCvms(cts, handler.ResOperateAuth)
}

// BatchAssociateBizCvm 业务下关联安全组，安全组本地id，cvm 云id
func (svc *securityGroupSvc) BatchAssociateBizCvm(cts *rest.Contexts) (any, error) {
	return svc.batchAssociateCvms(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) batchAssociateCvms(cts *rest.Contexts,
	validHandler handler.ValidWithAuthHandler) (any, error) {

	req, sgInfo, err := svc.decodeAssociateReq(cts, validHandler, meta.Associate)
	if err != nil {
		return nil, err
	}

	if err = svc.createBatchAssociateCvmAudit(cts.Kit, req.SecurityGroupID, req.CvmIDs); err != nil {
		logs.Errorf("create associate cvm audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch sgInfo.Vendor {
	case enumor.TCloud:
		return svc.batchAssociateTCloudCvms(cts, req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support for batch associate cvm", sgInfo.Vendor)
	}

}

func (svc *securityGroupSvc) createBatchAssociateCvmAudit(kt *kit.Kit, sgID string, cvmIDs []string) error {
	// create operation audit.
	audits := make([]protoaudit.CloudResourceOperationInfo, 0, len(cvmIDs))
	for _, cvmID := range cvmIDs {
		audits = append(audits, protoaudit.CloudResourceOperationInfo{
			ResType:           enumor.SecurityGroupAuditResType,
			ResID:             sgID,
			Action:            protoaudit.Associate,
			AssociatedResType: enumor.CvmAuditResType,
			AssociatedResID:   cvmID,
		})
	}

	if err := svc.audit.BatchResOperationAudit(kt, audits); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (svc *securityGroupSvc) batchAssociateTCloudCvms(cts *rest.Contexts,
	req *hcproto.SecurityGroupBatchAssociateCvmReq) (any, error) {

	err := svc.client.HCService().TCloud.SecurityGroup.BatchAssociateCvm(cts.Kit, req.SecurityGroupID,
		req.CvmIDs)
	if err != nil {
		logs.Errorf("fail to call hc service associate cloud cvm, err: %v, sg_id: %s, cloud_cvm_ids: %v, rid:%s",
			err, req.SecurityGroupID, req.CvmIDs, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// BatchDisassociateCvm disassociate cvm.
func (svc *securityGroupSvc) BatchDisassociateCvm(cts *rest.Contexts) (any, error) {
	return svc.batchDisassociateCvm(cts, handler.ResOperateAuth)
}

// BatchDisassociateBizCvm disassociate biz cvm.
func (svc *securityGroupSvc) BatchDisassociateBizCvm(cts *rest.Contexts) (any, error) {
	return svc.batchDisassociateCvm(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) batchDisassociateCvm(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	any, error) {

	req, sgInfo, err := svc.decodeAssociateReq(cts, validHandler, meta.Disassociate)
	if err != nil {
		return nil, err
	}

	switch sgInfo.Vendor {

	case enumor.TCloud:
		// create operation audit.
		if err = svc.createTCloudDisassociateCvmAudit(cts, req); err != nil {
			return nil, err
		}

		err := svc.client.HCService().TCloud.SecurityGroup.BatchDisassociateCvm(cts.Kit,
			req.SecurityGroupID, req.CvmIDs)
		if err != nil {
			logs.Errorf("fail to call hc service dissociate cloud cvm, err: %v, sg_id: %s, cloud_cvm_ids: %v, rid:%s",
				err, req.SecurityGroupID, req.CvmIDs, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", sgInfo.Vendor)
	}

}

func (svc *securityGroupSvc) createTCloudDisassociateCvmAudit(cts *rest.Contexts,
	req *hcproto.SecurityGroupBatchAssociateCvmReq) error {

	audits := make([]protoaudit.CloudResourceOperationInfo, 0, len(req.CvmIDs))
	for _, cvmID := range req.CvmIDs {
		audits = append(audits, protoaudit.CloudResourceOperationInfo{
			ResType:           enumor.SecurityGroupRuleAuditResType,
			ResID:             req.SecurityGroupID,
			Action:            protoaudit.Disassociate,
			AssociatedResType: enumor.CvmAuditResType,
			AssociatedResID:   cvmID,
		})
	}

	if err := svc.audit.BatchResOperationAudit(cts.Kit, audits); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}
	return nil
}

func (svc *securityGroupSvc) decodeAssociateReq(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler,
	action meta.Action) (*hcproto.SecurityGroupBatchAssociateCvmReq, *types.CloudResourceBasicInfo, error) {

	req := new(hcproto.SecurityGroupBatchAssociateCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, req.SecurityGroupID)
	if err != nil {
		logs.Errorf("get security group  resource basic info failed, err: %v, sg_id: %s, rid: %s",
			err, req.SecurityGroupID, cts.Kit.Rid)
		return nil, nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: action, BasicInfo: sgInfo})
	if err != nil {
		return nil, nil, err
	}
	return req, sgInfo, nil
}
