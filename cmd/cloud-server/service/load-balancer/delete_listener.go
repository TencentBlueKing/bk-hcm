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

package loadbalancer

import (
	"fmt"

	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// DeleteBizListener delete biz listener.
func (svc *lbSvc) DeleteBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteListener(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) deleteListener(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ListenerCloudResType,
		IDs:          req.IDs,
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list listener basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}
	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.ListenerAuditResType, basicInfoReq.IDs); err != nil {
		logs.Errorf("create operation audit listener failed, ids: %v, err: %v, rid: %s",
			basicInfoReq.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	if err = svc.validateListenerTargetWeight(cts.Kit, req.IDs); err != nil {
		logs.Errorf("validate listener target weight failed, ids: %s, err: %v, rid: %s", req.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	infoByVendor := classifier.ClassifyMap(basicInfoMap, func(item types.CloudResourceBasicInfo) enumor.Vendor {
		return item.Vendor
	})
	for vendor, infos := range infoByVendor {
		ids := slice.Map(infos, func(item types.CloudResourceBasicInfo) string {
			return item.ID
		})
		deleteReq := &core.BatchDeleteReq{
			IDs: ids,
		}
		switch vendor {
		case enumor.TCloud:
			err = svc.client.HCService().TCloud.Clb.DeleteListener(cts.Kit, deleteReq)
			if err != nil {
				logs.Errorf("[%s] request hcservice to delete listener failed, ids: %s, err: %v, rid: %s",
					enumor.TCloud, req.IDs, err, cts.Kit.Rid)
				return nil, err
			}
		default:
			logs.Errorf("delete listener not support vendor: %s, ids: %v, rid: %s", vendor, req.IDs, cts.Kit.Rid)
			return nil, fmt.Errorf("delete listener not support vendor: %s", vendor)
		}
	}

	return nil, nil
}

// validateListenerTargetWeight 校验监听器绑定的所有rs权重是否为0
func (svc *lbSvc) validateListenerTargetWeight(kt *kit.Kit, ids []string) error {
	for _, id := range ids {
		stat, err := svc.getListenerTargetWeightStat(kt, id)
		if err != nil {
			return err
		}
		if stat.NonZeroWeightCount > 0 {
			return fmt.Errorf("listener %s has targets with non-zero weight", id)
		}
	}
	return nil
}

// getTGIDsByListenerID 七层监听器会对应多个目标组，四层监听器只有一个目标组
func (svc *lbSvc) getTGIDsByListenerID(kt *kit.Kit, listenerID string) ([]string, error) {
	targetGroupIDs := make([]string, 0)
	listTGReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lbl_id", listenerID),
		),
		Page: core.NewDefaultBasePage(),
	}
	for {
		rels, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, listTGReq)
		if err != nil {
			logs.Errorf("list target group listener rel failed, req: %+v, err: %v, rid: %s", listTGReq, err, kt.Rid)
			return nil, err
		}
		for _, detail := range rels.Details {
			targetGroupIDs = append(targetGroupIDs, detail.TargetGroupID)
		}
		if len(rels.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		listTGReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return targetGroupIDs, nil
}

func (svc *lbSvc) getListenerTargetWeightStat(kt *kit.Kit, listenerID string) (*cslb.ListenerTargetsStat, error) {

	targetGroupIDs, err := svc.getTGIDsByListenerID(kt, listenerID)
	if err != nil {
		logs.Errorf("get target group ids by listener id failed, listenerID: %s, err: %v, rid: %s",
			listenerID, err, kt.Rid)
		return nil, err
	}
	result := &cslb.ListenerTargetsStat{}
	for _, batch := range slice.Split(targetGroupIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("target_group_id", batch),
			),
			Page: core.NewDefaultBasePage(),
		}
		for {
			targets, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listReq)
			if err != nil {
				logs.Errorf("list target failed, req: %+v, err: %v, rid: %s", listReq, err, kt.Rid)
				return nil, err
			}
			for _, detail := range targets.Details {
				if converter.PtrToVal(detail.Weight) == 0 {
					result.ZeroWeightCount++
				}
			}
			result.TotalCount += len(targets.Details)
			if len(targets.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	result.NonZeroWeightCount = result.TotalCount - result.ZeroWeightCount
	return result, nil
}

// ListBizListenerTargetWeightStat list biz listener rs weight stat.
func (svc *lbSvc) ListBizListenerTargetWeightStat(cts *rest.Contexts) (interface{}, error) {
	req := new(cslb.ListListenerTargetsStatReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ListenerCloudResType,
		IDs:          req.IDs,
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list listener basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}
	// validate biz and authorize
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Find, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	result := make(map[string]*cslb.ListenerTargetsStat)
	for _, id := range req.IDs {
		stat, err := svc.getListenerTargetWeightStat(cts.Kit, id)
		if err != nil {
			logs.Errorf("get listener target weight stat failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
			return nil, err
		}
		result[id] = stat
	}

	return result, nil
}
