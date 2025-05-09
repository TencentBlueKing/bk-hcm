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

// Package cloud ...
package cloud

import (
	"fmt"
	"net/http"

	"hcm/cmd/data-service/service/audit/cloud"
	"hcm/cmd/data-service/service/capability"
	"hcm/cmd/data-service/service/cloud/cvm"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/audit"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	audittable "hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// InitCloudService initial the cloud service
func InitCloudService(cap *capability.Capability) {
	svc := &cloudSvc{
		dao:   cap.Dao,
		audit: cloud.NewCloudAudit(cap.Dao),
	}

	h := rest.NewHandler()

	h.Add("GetResBasicInfo", http.MethodPost, "/cloud/resources/basics/{type}/id/{id}", svc.GetResourceBasicInfo)
	h.Add("ListResBasicInfo", http.MethodPost, "/cloud/resources/basics/list", svc.ListResourceBasicInfo)
	h.Add("BatchListResBasicInfo", http.MethodPost, "/cloud/resources/basics/batch/list",
		svc.BatchListResourceBasicInfo)
	h.Add("AssignResourceToBiz", http.MethodPost, "/cloud/resources/assign/bizs", svc.AssignResourceToBiz)

	h.Load(cap.WebService)
}

type cloudSvc struct {
	dao   dao.Set
	audit *cloud.Audit
}

// listResUsageBizRel list resource usage biz rel
func (svc cloudSvc) listResUsageBizRel(kt *kit.Kit, resType enumor.CloudResourceType, resIDs []string) (
	map[string][]int64, error) {

	usageBizIDs := make(map[string][]int64, 0)
	for _, batch := range slice.Split(resIDs, int(core.DefaultMaxPageLimit)) {

		listOpt := &types.ListOption{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("res_type", resType),
				tools.RuleIn("res_id", batch),
			),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"res_id", "usage_biz_id"},
		}
		listOpt.Page.Sort = "res_id"

		// 单个资源会对应多条关联关系
		for {
			rst, err := svc.dao.ResUsageBizRel().List(kt, listOpt)
			if err != nil {
				logs.Errorf("failed to list res usage biz rel, err: %v, res_type: %s, res_ids: %v, rid: %s",
					err, resType, resIDs, kt.Rid)
				return nil, err
			}

			for _, item := range rst.Details {
				usageBizIDs[item.ResID] = append(usageBizIDs[item.ResID], item.UsageBizID)
			}

			if len(rst.Details) < int(listOpt.Page.Limit) {
				break
			}

			listOpt.Page.Start += uint32(listOpt.Page.Limit)
		}
	}

	return usageBizIDs, nil
}

// GetResourceBasicInfo get resource basic info.
func (svc cloudSvc) GetResourceBasicInfo(cts *rest.Contexts) (interface{}, error) {
	resourceType := cts.PathParameter("type").String()
	if len(resourceType) == 0 {
		return nil, errf.New(errf.InvalidParameter, "resource type is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "resource id is required")
	}

	req := new(protocloud.GetResourceBasicInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list, err := svc.dao.Cloud().ListResourceBasicInfo(cts.Kit, enumor.CloudResourceType(resourceType), []string{id},
		req.Fields...)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "%s not found resource: %s", resourceType, id)
	}

	if len(list) != 1 {
		logs.Errorf("list resource basic info return count not right, count: %s, resource type: %s, id: %s, rid: %s",
			len(list), resourceType, id, cts.Kit.Rid)
		return nil, fmt.Errorf("list resource basic info return count not right")
	}

	usageBizIDs, err := svc.listResUsageBizRel(cts.Kit, enumor.CloudResourceType(resourceType), []string{id})
	if err != nil {
		logs.Errorf("failed to list res usage biz rel, err: %v, res_type: %s, res_id: %s, rid: %s",
			err, resourceType, id, cts.Kit.Rid)
		return nil, err
	}

	basicInfo := list[0]
	basicInfo.UsageBizIDs = usageBizIDs[id]

	return basicInfo, nil
}

// ListResourceBasicInfo list resource basic info.
func (svc cloudSvc) ListResourceBasicInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ListResourceBasicInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list, err := svc.dao.Cloud().ListResourceBasicInfo(cts.Kit, req.ResourceType, req.IDs, req.Fields...)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "%s not found resource: %v", req.ResourceType, req.IDs)
	}

	usageBizIDs, err := svc.listResUsageBizRel(cts.Kit, req.ResourceType, req.IDs)
	if err != nil {
		logs.Errorf("failed to list res usage biz rel, err: %v, res_type: %s, res_ids: %v, rid: %s",
			err, req.ResourceType, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	result := make(map[string]types.CloudResourceBasicInfo, len(list))
	for _, info := range list {
		info.UsageBizIDs = usageBizIDs[info.ID]
		result[info.ID] = info
	}

	return result, nil
}

// BatchListResourceBasicInfo batch list resource basic info.
func (svc cloudSvc) BatchListResourceBasicInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.BatchListResourceBasicInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result := make(map[string]types.CloudResourceBasicInfo, len(req.Items))
	for _, item := range req.Items {
		list, err := svc.dao.Cloud().ListResourceBasicInfo(cts.Kit, item.ResourceType, item.IDs, item.Fields...)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			return nil, errf.Newf(errf.RecordNotFound, "%s not found resource: %v", item.ResourceType, item.IDs)
		}

		usageBizIDs, err := svc.listResUsageBizRel(cts.Kit, item.ResourceType, item.IDs)
		if err != nil {
			logs.Errorf("failed to list res usage biz rel, err: %v, res_type: %s, res_ids: %v, rid: %s",
				err, item.ResourceType, item.IDs, cts.Kit.Rid)
			return nil, err
		}

		for _, info := range list {
			info.UsageBizIDs = usageBizIDs[info.ID]
			result[info.ID] = info
		}
	}

	return result, nil
}

var assignResAuditTypeMap = map[enumor.CloudResourceType]enumor.AuditResourceType{
	enumor.SecurityGroupCloudResType:    enumor.SecurityGroupAuditResType,
	enumor.VpcCloudResType:              enumor.VpcCloudAuditResType,
	enumor.SubnetCloudResType:           enumor.SubnetAuditResType,
	enumor.EipCloudResType:              enumor.EipAuditResType,
	enumor.CvmCloudResType:              enumor.CvmAuditResType,
	enumor.DiskCloudResType:             enumor.DiskAuditResType,
	enumor.RouteTableCloudResType:       enumor.RouteTableAuditResType,
	enumor.GcpFirewallRuleCloudResType:  enumor.GcpFirewallRuleAuditResType,
	enumor.NetworkInterfaceCloudResType: enumor.NetworkInterfaceAuditResType,
}

// AssignResourceToBiz assign an account's cloud resource to biz, **only for ui**.
func (svc cloudSvc) AssignResourceToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AssignResourceToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		auditOpts := make([]audit.CloudResourceAssignInfo, 0)

		hasCvmAssign := false
		for _, resType := range req.ResTypes {
			if resType == enumor.CvmCloudResType {
				hasCvmAssign = true
			}

			auditType, exists := assignResAuditTypeMap[resType]
			if !exists {
				return nil, errf.Newf(errf.InvalidParameter, "resource type %s cannot be assigned", resType)
			}

			expr := tools.EqualWithOpExpression(filter.And, map[string]interface{}{"account_id": req.AccountID,
				"bk_biz_id": constant.UnassignedBiz})

			ids, err := svc.dao.Cloud().ListResourceIDs(cts.Kit, resType, expr)
			if err != nil {
				return nil, err
			}

			if len(ids) == 0 {
				continue
			}

			assignFilter := tools.ContainersExpression("id", ids)
			err = svc.dao.Cloud().AssignResourceToBiz(cts.Kit, txn, resType, assignFilter, req.BkBizID)
			if err != nil {
				return nil, err
			}

			for _, id := range ids {
				auditOpts = append(auditOpts, audit.CloudResourceAssignInfo{
					ResType:         auditType,
					ResID:           id,
					AssignedResType: enumor.BizAuditAssignedResType,
					AssignedResID:   req.BkBizID,
				})
			}
		}

		// create audit
		if len(auditOpts) == 0 {
			return nil, nil
		}

		if err := svc.createAudit(cts.Kit, txn, auditOpts); err != nil {
			return nil, err
		}

		if hasCvmAssign {
			if err := cvm.SyncCvmToCmdb(cts.Kit, req.AccountID, req.BkBizID); err != nil {
				logs.Errorf("sync cvm to cmdb failed, err: %v, accountID: %s, bkBizID: %d, rid: %s", err,
					req.AccountID, req.BkBizID, cts.Kit.Rid)
				return nil, err
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc cloudSvc) createAudit(kt *kit.Kit, txn *sqlx.Tx, auditOpts []audit.CloudResourceAssignInfo) error {

	auditAssignOpts := slice.Split(auditOpts, constant.BatchOperationMaxLimit)
	allAudits := make([]*audittable.AuditTable, 0, len(auditOpts))
	for _, opts := range auditAssignOpts {
		audits, err := svc.audit.GenCloudResAssignAudit(kt, &audit.CloudResourceAssignAuditReq{Assigns: opts})
		if err != nil {
			return err
		}
		allAudits = append(allAudits, audits...)
	}

	if err := svc.dao.Audit().BatchCreateWithTx(kt, txn, allAudits); err != nil {
		return err
	}

	return nil
}
