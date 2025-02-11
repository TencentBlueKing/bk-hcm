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
	"errors"
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// Interface define security group interface.
type Interface interface {
	ListSGRelCVM(kt *kit.Kit, sgID string, resBizID int64, listReq *core.ListReq) (
		*dataproto.SGCommonRelWithCVMListResp, error)
	ListSGRelLoadBalancer(kt *kit.Kit, sgID string, resBizID int64, oriReq *core.ListReq) (
		*dataproto.SGCommonRelWithLBListResp, error)
	ListSGRelBusiness(kt *kit.Kit, currentBizID int64, sgID string) (*proto.ListSGRelBusinessResp, error)
	UpdateSGMgmtAttr(kt *kit.Kit, mgmtAttr *proto.SecurityGroupUpdateMgmtAttrReq, sgID string) error
	BatchUpdateSGMgmtAttr(kt *kit.Kit, mgmtAttrs []proto.BatchUpdateSGMgmtAttrItem) error
}

type securityGroup struct {
	client *client.ClientSet
	audit  audit.Interface
}

// NewSecurityGroup new security group.
func NewSecurityGroup(client *client.ClientSet, audit audit.Interface) Interface {
	return &securityGroup{
		client: client,
		audit:  audit,
	}
}

func (s *securityGroup) listBaseSecurityGroups(kt *kit.Kit, sgIDs []string) ([]cloud.BaseSecurityGroup, error) {
	baseSG := make([]cloud.BaseSecurityGroup, 0)
	for _, ids := range slice.Split(sgIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &dataproto.SecurityGroupListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", ids)),
			Page:   core.NewDefaultBasePage(),
		}

		listRst, err := s.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list security group failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}

		baseSG = append(baseSG, listRst.Details...)
	}

	return baseSG, nil
}

// ListSGRelCVM list security group rel cvm.
// only summary information will be return to avoid the risk of exceeding authority.
func (s *securityGroup) ListSGRelCVM(kt *kit.Kit, sgID string, resBizID int64, oriReq *core.ListReq) (
	*dataproto.SGCommonRelWithCVMListResp, error) {

	listFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			oriReq.Filter,
			tools.RuleEqual("bk_biz_id", resBizID),
		},
	}

	listReq := &dataproto.SGCommonRelListReq{
		SGIDs: []string{sgID},
		ListReq: core.ListReq{
			Filter: listFilter,
			Page:   oriReq.Page,
			Fields: oriReq.Fields,
		},
	}

	return s.client.DataService().Global.SGCommonRel.ListWithCVMSummary(kt, listReq)
}

// ListSGRelLoadBalancer list security group rel load balancer.
// only summary information will be return to avoid the risk of exceeding authority.
func (s *securityGroup) ListSGRelLoadBalancer(kt *kit.Kit, sgID string, resBizID int64, oriReq *core.ListReq) (
	*dataproto.SGCommonRelWithLBListResp, error) {

	listFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			oriReq.Filter,
			tools.RuleEqual("bk_biz_id", resBizID),
		},
	}

	listReq := &dataproto.SGCommonRelListReq{
		SGIDs: []string{sgID},
		ListReq: core.ListReq{
			Filter: listFilter,
			Page:   oriReq.Page,
			Fields: oriReq.Fields,
		},
	}

	return s.client.DataService().Global.SGCommonRel.ListWithLBSummary(kt, listReq)
}

func (s *securityGroup) listAllSGRel(kt *kit.Kit, listFilter *filter.Expression, listFields []string) (
	[]cloud.SecurityGroupCommonRel, error) {

	listReq := &core.ListReq{
		Fields: listFields,
		Filter: listFilter,
		Page:   core.NewDefaultBasePage(),
	}

	sgRels := make([]cloud.SecurityGroupCommonRel, 0)
	for {
		rst, err := s.client.DataService().Global.SGCommonRel.ListSgCommonRels(kt, listReq)
		if err != nil {
			logs.Errorf("list security group rel failed, filter: %+v, err: %v, rid: %s", listFilter, err, kt.Rid)
			return nil, err
		}

		sgRels = append(sgRels, rst.Details...)
		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return sgRels, nil
}

// ListSGRelBusiness List the biz IDs that have resources associated with the security group. Group by resource type.
// Use sg_ID to query res_IDs in the rel table, then use res_IDs to query res bizs.
func (s *securityGroup) ListSGRelBusiness(kt *kit.Kit, currentBizID int64, sgID string) (
	*proto.ListSGRelBusinessResp, error) {

	relListFilter := tools.EqualExpression("security_group_id", sgID)
	relListFields := []string{"res_id", "res_type"}

	// list security group rel resources
	sgRelResources, err := s.listAllSGRel(kt, relListFilter, relListFields)
	if err != nil {
		logs.Errorf("list security group rel resources failed, id: %s, err: %v, rid: %s", sgID, err, kt.Rid)
		return nil, err
	}

	if sgRelResources == nil {
		logs.Errorf("security group rel resources is empty, id: %s, rid: %s", sgID, kt.Rid)
		return nil, errors.New("security group rel resources is empty")
	}

	cvmIDs := make([]string, 0)
	lbIDs := make([]string, 0)

	for _, relRes := range sgRelResources {
		switch relRes.ResType {
		case enumor.CvmCloudResType:
			cvmIDs = append(cvmIDs, relRes.ResID)
		case enumor.LoadBalancerCloudResType:
			lbIDs = append(lbIDs, relRes.ResID)
		default:
			logs.Errorf("unsupported res type: %s, sg id: %s, res id: %s, rid: %s", relRes.ResType, sgID,
				relRes.ID, kt.Rid)
			return nil, fmt.Errorf("unsupported res type: %s", relRes.ResType)
		}
	}

	// list business ids associated with CVM and load balancer resources.
	cvmRelBizs, err := s.listRelBizsWithCVM(kt, currentBizID, cvmIDs)
	if err != nil {
		logs.Errorf("list security group rel cvm bizs failed, err: %v, id: %s, rid: %s", err, sgID, kt.Rid)
		return nil, err
	}

	lbRelBizs, err := s.listRelBizsWithLB(kt, currentBizID, lbIDs)
	if err != nil {
		logs.Errorf("list security group rel lb bizs failed, err: %v, id: %s, rid: %s", err, sgID, kt.Rid)
		return nil, err
	}

	return &proto.ListSGRelBusinessResp{
		CVM:          cvmRelBizs,
		LoadBalancer: lbRelBizs,
	}, nil
}

func (s *securityGroup) listRelBizsWithCVM(kt *kit.Kit, currentBizID int64, cvmIDs []string) (
	[]proto.ListSGRelBusinessItem, error) {

	relBizMap := make(map[int64]int64)
	for _, batch := range slice.Split(cvmIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Fields: []string{"bk_biz_id"},
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}

		res, err := s.client.DataService().Global.Cvm.ListCvm(kt, req)
		if err != nil {
			logs.Errorf("list security group rel cvm failed, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, kt.Rid)
			return nil, err
		}

		for _, item := range res.Details {
			relBizMap[item.BkBizID] += 1
		}
	}

	return tidySGRelBusiness(currentBizID, relBizMap), nil
}

func (s *securityGroup) listRelBizsWithLB(kt *kit.Kit, currentBizID int64, lbIDs []string) (
	[]proto.ListSGRelBusinessItem, error) {

	relBizMap := make(map[int64]int64)
	for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Fields: []string{"bk_biz_id"},
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}

		res, err := s.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, req)
		if err != nil {
			logs.Errorf("list security group rel load balancer failed, err: %v, lbIDs: %v, rid: %s", err, lbIDs,
				kt.Rid)
			return nil, err
		}

		for _, item := range res.Details {
			relBizMap[item.BkBizID] += 1
		}
	}

	return tidySGRelBusiness(currentBizID, relBizMap), nil
}

func tidySGRelBusiness(currentBizID int64, relBizMap map[int64]int64) []proto.ListSGRelBusinessItem {
	var currentBizResC int64
	if resCount, ok := relBizMap[currentBizID]; ok {
		currentBizResC = resCount
		delete(relBizMap, currentBizID)
	}

	// 当前业务必须在列表的第一个
	relBizs := make([]proto.ListSGRelBusinessItem, 0, len(relBizMap)+1)
	if currentBizID != constant.UnassignedBiz {
		relBizs[0] = proto.ListSGRelBusinessItem{
			BkBizID:  currentBizID,
			ResCount: currentBizResC,
		}
	}
	for bizID, count := range relBizMap {
		relBizs = append(relBizs, proto.ListSGRelBusinessItem{
			BkBizID:  bizID,
			ResCount: count,
		})
	}

	return relBizs
}

// UpdateSGMgmtAttr update security group management attributes
func (s *securityGroup) UpdateSGMgmtAttr(kt *kit.Kit, mgmtAttr *proto.SecurityGroupUpdateMgmtAttrReq,
	sgID string) error {

	sgs, err := s.listBaseSecurityGroups(kt, []string{sgID})
	if err != nil {
		logs.Errorf("list base security group failed, err: %v, sg_id: %s, rid: %s", err, sgID, kt.Rid)
		return err
	}

	if len(sgs) != 1 {
		logs.Errorf("list base security group failed, len: %d, sg_id: %s, rid: %s", len(sgs), sgID, kt.Rid)
		return errors.New("security group not found")
	}

	sg := sgs[0]
	// 管理类型已确定的不能修改
	if sg.MgmtType.Validate() == nil && sg.MgmtType != mgmtAttr.MgmtType {
		logs.Errorf("security group mgmt type cannot be modified, sg_id: %s, old_type: %s, new_type: %s, rid: %s",
			sg.ID, sg.MgmtType, mgmtAttr.MgmtType, kt.Rid)
		return fmt.Errorf("security group: %s mgmt type cannot be modified", sg.ID)
	}

	// 平台管理不可修改管理业务
	if sg.MgmtType == enumor.MgmtTypePlatform {
		if mgmtAttr.MgmtBizID != constant.UnassignedBiz && mgmtAttr.MgmtBizID != 0 {
			logs.Errorf("security group: %s cannot be assigned to a business, rid: %s", sg.ID, kt.Rid)
			return fmt.Errorf("security group: %s cannot be assigned to a business", sg.ID)
		}
	}

	// 更新管理属性
	updateItems := make([]dataproto.BatchUpdateSGMgmtAttrItem, 0)
	updateItems = append(updateItems, dataproto.BatchUpdateSGMgmtAttrItem{
		ID:         sg.ID,
		MgmtType:   mgmtAttr.MgmtType,
		MgmtBizID:  mgmtAttr.MgmtBizID,
		Manager:    mgmtAttr.Manager,
		BakManager: mgmtAttr.BakManager,
	})

	updateReq := &dataproto.BatchUpdateSecurityGroupMgmtAttrReq{
		SecurityGroups: updateItems,
	}

	if err := s.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupMgmtAttr(kt.Ctx, kt.Header(),
		updateReq); err != nil {
		logs.Errorf("batch update security group management attributes failed, err: %v, rid: %s", err,
			kt.Rid)
		return err
	}

	// 更新使用业务列表
	setRelReq := &dataproto.ResUsageBizRelUpdateReq{
		UsageBizIDs: mgmtAttr.UsageBizIDs,
		ResCloudID:  sg.CloudID,
		ResVendor:   sg.Vendor,
	}
	err = s.client.DataService().Global.ResUsageBizRel.SetBizRels(kt, enumor.SecurityGroupCloudResType, sgID, setRelReq)
	if err != nil {
		logs.Errorf("set security group usage biz rel failed, err: %v, sg_id: %s, rid: %s", err, sgID, kt.Rid)
		return err
	}

	return nil
}

// BatchUpdateSGMgmtAttr batch update security group management attributes
func (s *securityGroup) BatchUpdateSGMgmtAttr(kt *kit.Kit, mgmtAttrs []proto.BatchUpdateSGMgmtAttrItem) error {
	// 获取变更安全组当前的基本信息
	sgIDs := make([]string, len(mgmtAttrs))
	for i, sgAttr := range mgmtAttrs {
		sgIDs[i] = sgAttr.ID
	}

	sgs, err := s.listBaseSecurityGroups(kt, sgIDs)
	if err != nil {
		logs.Errorf("list base security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 仅当所有管理属性均不存在时才允许批量编辑，平台管理不可批量编辑
	for _, sg := range sgs {
		if sg.MgmtType == enumor.MgmtTypePlatform {
			logs.Errorf("platform security group cannot be batch updated, sg_id: %s, rid: %s", sg.ID, kt.Rid)
			return fmt.Errorf("platform security group cannot be batch updated, id: %s", sg.ID)
		}

		if sg.MgmtBizID != constant.UnassignedBiz || sg.Manager != "" || sg.BakManager != "" {
			logs.Errorf("security group management attributes already exist, sg_id: %s, rid: %s", sg.ID, kt.Rid)
			return fmt.Errorf("security group: %s management attributes already exist", sg.ID)
		}
	}

	for _, batch := range slice.Split(mgmtAttrs, constant.BatchOperationMaxLimit) {
		updateItems := make([]dataproto.BatchUpdateSGMgmtAttrItem, len(batch))
		for i, attr := range batch {
			updateItems[i] = dataproto.BatchUpdateSGMgmtAttrItem{
				ID: attr.ID,
				// 批量更新默认更新为业务管理
				MgmtType:   enumor.MgmtTypeBiz,
				MgmtBizID:  attr.MgmtBizID,
				Manager:    attr.Manager,
				BakManager: attr.BakManager,
			}
		}

		updateReq := &dataproto.BatchUpdateSecurityGroupMgmtAttrReq{
			SecurityGroups: updateItems,
		}

		if err := s.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupMgmtAttr(kt.Ctx, kt.Header(),
			updateReq); err != nil {
			logs.Errorf("batch update security group management attributes failed, err: %v, rid: %s", err,
				kt.Rid)
			return err
		}
	}

	return nil
}
