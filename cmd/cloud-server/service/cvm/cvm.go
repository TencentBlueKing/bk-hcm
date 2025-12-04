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

// Package cvm ...
package cvm

import (
	"fmt"
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/cvm"
	"hcm/cmd/cloud-server/logics/disk"
	"hcm/cmd/cloud-server/logics/eip"
	moalogic "hcm/cmd/cloud-server/logics/moa"
	securitygroup "hcm/cmd/cloud-server/logics/security-group"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// InitCvmService initialize the cvm service.
func InitCvmService(c *capability.Capability) {

	svc := &cvmSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		diskLgc:    c.Logics.Disk,
		cvmLgc:     c.Logics.Cvm,
		eipLgc:     c.Logics.Eip,
		cmdbCli:    c.CmdbCli,

		moaLogic: c.Logics.Moa,
		sgLogic:  c.Logics.SecurityGroup,
	}

	h := rest.NewHandler()

	h.Add("GetCvm", http.MethodGet, "/cvms/{id}", svc.GetCvm)
	h.Add("ListCvmExt", http.MethodPost, "/cvms/list", svc.ListCvm)
	h.Add("CreateCvm", http.MethodPost, "/cvms/create", svc.CreateCvm)
	h.Add("InquiryPriceCvm", http.MethodPost, "/cvms/prices/inquiry", svc.InquiryPriceCvm)
	h.Add("BatchDeleteCvm", http.MethodDelete, "/cvms/batch", svc.BatchDeleteCvm)
	h.Add("AssignCvmToBiz", http.MethodPost, "/cvms/assign/bizs", svc.AssignCvmToBiz)
	h.Add("AssignCvmToBizPreview", http.MethodPost, "/cvms/assign/bizs/preview", svc.AssignCvmToBizPreview)
	h.Add("ListAssignedCvmMatchHost", http.MethodPost, "/cvms/assign/hosts/match/list", svc.ListAssignedCvmMatchHost)
	h.Add("BatchStartCvm", http.MethodPost, "/cvms/batch/start", svc.BatchStartCvm)
	h.Add("BatchStopCvm", http.MethodPost, "/cvms/batch/stop", svc.BatchStopCvm)
	h.Add("BatchRebootCvm", http.MethodPost, "/cvms/batch/reboot", svc.BatchRebootCvm)
	h.Add("QueryCvmRelatedRes", http.MethodPost, "/cvms/rel_res/batch", svc.QueryCvmRelatedRes)
	h.Add("ListInstanceConfig", http.MethodPost,
		"/vendors/{vendor}/instances/config/query_from_cloud", svc.ListInstanceConfig)
	h.Add("ListBizInstanceConfig", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/{vendor}/instances/config/query_from_cloud", svc.ListBizInstanceConfig)

	// 资源下回收相关接口
	h.Add("RecycleCvm", http.MethodPost, "/cvms/recycle", svc.RecycleCvm)
	h.Add("RecoverCvm", http.MethodPost, "/cvms/recover", svc.RecoverCvm)
	h.Add("GetRecycledCvm", http.MethodGet, "/recycled/cvms/{id}", svc.GetRecyclingCvm)
	h.Add("BatchDeleteRecycledCvm", http.MethodDelete, "/recycled/cvms/batch", svc.BatchDeleteRecycledCvm)

	// cvm apis in biz
	h.Add("GetBizCvm", http.MethodGet, "/bizs/{bk_biz_id}/cvms/{id}", svc.GetBizCvm)
	h.Add("ListBizCvmExt", http.MethodPost, "/bizs/{bk_biz_id}/cvms/list", svc.ListBizCvm)
	h.Add("BatchDeleteBizCvm", http.MethodDelete, "/bizs/{bk_biz_id}/cvms/batch", svc.BatchDeleteBizCvm)
	h.Add("BatchStartBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/batch/start", svc.BatchStartBizCvm)
	h.Add("BatchStopBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/batch/stop", svc.BatchStopBizCvm)
	h.Add("BatchRebootBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/batch/reboot", svc.BatchRebootBizCvm)
	h.Add("QueryBizCvmRelatedRes", http.MethodPost, "/bizs/{bk_biz_id}/cvms/rel_res/batch", svc.QueryBizCvmRelatedRes)
	h.Add("ListCvmSecurityGroupRules", http.MethodPost,
		"/bizs/{bk_biz_id}/cvms/{cvm_id}/security_groups/{security_group_id}/rules/list", svc.ListCvmSecurityGroupRules)
	h.Add("InquiryBizPriceCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/prices/inquiry", svc.InquiryBizPriceCvm)
	h.Add("BatchResetAsyncBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/batch/reset_async",
		svc.BatchResetAsyncBizCvm)
	h.Add("BatchSopsAsyncResetBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/sops/batch/reset_async",
		svc.BatchSopsAsyncResetBizCvm)
	h.Add("BatchUnVerifyAsyncResetBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/unverify/batch/reset_async",
		svc.BatchUnVerifyAsyncResetBizCvm)

	h.Add("ListBizCvmOperateStatus", http.MethodPost, "/bizs/{bk_biz_id}/cvms/list/operate/status",
		svc.ListBizCvmOperateStatus)
	h.Add("BatchAsyncStartCvm", http.MethodPost, "/cvms/batch/start_async", svc.BatchAsyncStartCvm)
	h.Add("BatchAsyncStopCvm", http.MethodPost, "/cvms/batch/stop_async", svc.BatchAsyncStopCvm)
	h.Add("BatchAsyncRebootCvm", http.MethodPost, "/cvms/batch/reboot_async", svc.BatchAsyncRebootCvm)
	h.Add("BatchAsyncStartBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/batch/start_async",
		svc.BatchAsyncStartBizCvm)
	h.Add("BatchAsyncStopBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/batch/stop_async", svc.BatchAsyncStopBizCvm)
	h.Add("BatchSopsAsyncStopBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/sops/batch/stop_async",
		svc.BatchSopsAsyncStopBizCvm)
	h.Add("BatchAsyncRebootBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/batch/reboot_async",
		svc.BatchAsyncRebootBizCvm)
	h.Add("BatchSopsAsyncRebootBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/sops/batch/reboot_async",
		svc.BatchSopsAsyncRebootBizCvm)

	// 业务下回收接口
	h.Add("RecycleBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/recycle", svc.RecycleBizCvm)
	h.Add("RecoverBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/cvms/recover", svc.RecoverBizCvm)
	h.Add("GetBizRecycledCvm", http.MethodGet, "/bizs/{bk_biz_id}/recycled/cvms/{id}", svc.GetBizRecyclingCvm)
	h.Add("BatchDeleteBizRecycledCvm", http.MethodDelete, "/bizs/{bk_biz_id}/recycled/cvms/batch",
		svc.BatchDeleteBizRecycledCvm)

	h.Add("BatchAssociateSecurityGroups", http.MethodPost,
		"/cvms/{cvm_id}/security_groups/batch_associate", svc.BatchAssociateSecurityGroups)
	h.Add("BizBatchAssociateSecurityGroups", http.MethodPost,
		"/bizs/{bk_biz_id}/cvms/{cvm_id}/security_groups/batch_associate", svc.BizBatchAssociateSecurityGroups)

	h.Add("BatchIdleCheckBizCvm", http.MethodPost,
		"/bizs/{bk_biz_id}/cvms/batch/idle_check", svc.BatchIdleCheckBizCvm)
	h.Add("GetIdleCheckCvmResult", http.MethodPost,
		"/bizs/{bk_biz_id}/cvms/idle_check/result/{suborder_id}", svc.GetIdleCheckCvmResult)

	initCvmServiceHooks(svc, h)

	h.Load(c.WebService)
}

type cvmSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	diskLgc    disk.Interface
	cvmLgc     cvm.Interface
	eipLgc     eip.Interface
	cmdbCli    cmdb.Client

	moaLogic moalogic.Interface
	sgLogic  securitygroup.Interface
}

// batchListCvmByIDs 批量获取CVM列表
func (svc *cvmSvc) batchListCvmByIDs(kt *kit.Kit, ids []string) ([]corecvm.BaseCvm, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("ids is empty")
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", ids),
		),
		Page: core.NewDefaultBasePage(),
	}
	list := make([]corecvm.BaseCvm, 0)
	for {
		cvmResp, err := svc.client.DataService().Global.Cvm.ListCvm(kt, listReq)
		if err != nil {
			return nil, err
		}

		list = append(list, cvmResp.Details...)
		if len(cvmResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return list, nil
}
