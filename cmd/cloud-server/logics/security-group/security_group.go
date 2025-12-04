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
	"hcm/pkg/ziyan"
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
	FindYuntiDefaultSG(kt *kit.Kit, sgIDs []string) (yuntiSGCloudID string, err error)
}

type securityGroup struct {
	// ziyan 特殊逻辑
	sgOp ziyan.SgIf

	client *client.ClientSet
	audit  audit.Interface
}

// NewSecurityGroup new security group.
func NewSecurityGroup(client *client.ClientSet, audit audit.Interface) Interface {
	return &securityGroup{
		sgOp:   ziyan.NewSgOp(client),
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

	// list business ids associated with CVM and load balancer resources.
	cvmRelBizs, err := s.listRelBizs(kt, currentBizID, sgID, enumor.CvmCloudResType)
	if err != nil {
		logs.Errorf("list security group rel cvm bizs failed, err: %v, id: %s, rid: %s", err, sgID, kt.Rid)
		return nil, err
	}

	lbRelBizs, err := s.listRelBizs(kt, currentBizID, sgID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("list security group rel lb bizs failed, err: %v, id: %s, rid: %s", err, sgID, kt.Rid)
		return nil, err
	}

	return &proto.ListSGRelBusinessResp{
		CVM:          cvmRelBizs,
		LoadBalancer: lbRelBizs,
	}, nil
}

func (s *securityGroup) listRelBizs(kt *kit.Kit, currentBizID int64, sgID string,
	resourceType enumor.CloudResourceType) ([]proto.ListSGRelBusinessItem, error) {

	req := &dataproto.SGCommonRelCountBizInfoReq{
		SgID:    sgID,
		ResType: resourceType,
	}
	resBizInfo, err := s.client.DataService().Global.SGCommonRel.CountSGRelatedResBizInfo(kt, req)
	if err != nil {
		logs.Errorf("count security group related res biz info failed, err: %v, sg_id: %s, rid: %s", err, sgID, kt.Rid)
		return nil, err
	}
	return tidySGRelBusiness(currentBizID, resBizInfo.Items), nil
}

func tidySGRelBusiness(currentBizID int64,
	relBizList []dataproto.ListSGRelBusinessItem) []proto.ListSGRelBusinessItem {

	relBizs := make([]proto.ListSGRelBusinessItem, 0, len(relBizList)+1)
	currentDetail := proto.ListSGRelBusinessItem{
		BkBizID:  currentBizID,
		ResCount: 0,
	}

	for _, item := range relBizList {
		if item.BkBizID == currentBizID {
			currentDetail.ResCount = item.ResCount
			continue
		}

		relBizs = append(relBizs, proto.ListSGRelBusinessItem{
			BkBizID:  item.BkBizID,
			ResCount: item.ResCount,
		})
	}

	// 当前业务必须在列表的第一个
	// 如果关联资源中没有当前业务，需默认添加到列表开头
	if currentBizID != constant.UnassignedBiz {
		relBizs = append([]proto.ListSGRelBusinessItem{currentDetail}, relBizs...)
	}

	return relBizs
}

func isSGMgmtModifiable(kt *kit.Kit, sg cloud.BaseSecurityGroup, mgmtType enumor.MgmtType, mgmtBizID int64) error {
	// 管理类型已确定的不能修改
	if sg.MgmtType != "" {
		if mgmtType != "" && sg.MgmtType != mgmtType {
			logs.Errorf("security group mgmt type cannot be modified, sg_id: %s, old_type: %s, new_type: %s, rid: %s",
				sg.ID, sg.MgmtType, mgmtType, kt.Rid)
			return fmt.Errorf("security group: %s mgmt type cannot be modified", sg.ID)
		}
	}

	// 已分配的安全组，不可修改管理业务
	if sg.BkBizID != constant.UnassignedBiz && mgmtBizID != 0 {
		logs.Errorf("security group: %s is assigned, cannot modify the assigned business, rid: %s", sg.ID, kt.Rid)
		return fmt.Errorf("security group: %s is assigned, cannot modify the assigned business", sg.ID)
	}

	// 非业务管理的安全组，不可修改管理业务
	if sg.MgmtType != enumor.MgmtTypeBiz && mgmtType != enumor.MgmtTypeBiz && mgmtBizID != 0 {
		logs.Errorf("security group: %s mgmt_type is unconfirmed, cannot modify the assigned business, rid: %s",
			sg.ID, kt.Rid)
		return fmt.Errorf("security group: %s mgmt_type is unconfirmed, cannot modify the assigned business",
			sg.ID)
	}

	return nil
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
	err = isSGMgmtModifiable(kt, sg, mgmtAttr.MgmtType, mgmtAttr.MgmtBizID)
	if err != nil {
		return err
	}

	// 管理业务和使用业务必须在帐号下的业务列表中
	belongTo, err := s.isBizsBelongToAccount(kt, sg.AccountID, mgmtAttr.MgmtBizID, mgmtAttr.UsageBizIDs)
	if err != nil {
		logs.Errorf("check bizs belong to account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if !belongTo {
		logs.Errorf("bizs: %d, %v are not belong to account: %s, sec_group: %s, rid: %s", mgmtAttr.MgmtBizID,
			mgmtAttr.UsageBizIDs, sg.AccountID, sg.ID, kt.Rid)
		return fmt.Errorf("bizs: %d, %v are not belong to account: %s", mgmtAttr.MgmtBizID, mgmtAttr.UsageBizIDs,
			sg.AccountID)
	}

	// 更新管理属性到云上（如果有需要的话）
	if err := s.updateSingleCloudMgmtAttr(kt, mgmtAttr, sg.ID); err != nil {
		logs.Errorf("update cloud security group management attributes failed, err: %v, sg_id: %s, rid: %s", err,
			sg.ID, kt.Rid)
		return err
	}

	// 更新管理属性
	updateItems := make([]dataproto.BatchUpdateSGMgmtAttrItem, 0)
	updateItems = append(updateItems, dataproto.BatchUpdateSGMgmtAttrItem{
		ID:         sg.ID,
		MgmtType:   mgmtAttr.MgmtType,
		MgmtBizID:  mgmtAttr.MgmtBizID,
		Manager:    mgmtAttr.Manager,
		BakManager: mgmtAttr.BakManager,
		Vendor:     sg.Vendor,
		CloudID:    sg.CloudID,
	})
	updateReq := &dataproto.BatchUpdateSecurityGroupMgmtAttrReq{
		SecurityGroups: updateItems,
	}
	if err := s.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupMgmtAttr(kt, updateReq); err != nil {
		logs.Errorf("batch update security group management attributes failed, err: %v, rid: %s", err,
			kt.Rid)
		return err
	}

	// 更新使用业务列表
	if len(mgmtAttr.UsageBizIDs) <= 0 {
		return nil
	}
	// 使用业务非全部时，将管理业务加入使用业务
	if mgmtAttr.MgmtBizID != 0 && mgmtAttr.UsageBizIDs[0] != constant.AttachedAllBiz {
		mgmtAttr.UsageBizIDs = append(mgmtAttr.UsageBizIDs, mgmtAttr.MgmtBizID)
	}
	setRelReq := &dataproto.ResUsageBizRelUpdateReq{
		UsageBizIDs: slice.Unique(mgmtAttr.UsageBizIDs),
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

// updateSingleCloudMgmtAttr 更新单个安全组的云上管理属性，以本地数据为准
// 因先同步再更新本地数据，需要将本次期望的更新属性合并到本地数据中
func (s *securityGroup) updateSingleCloudMgmtAttr(kt *kit.Kit, mgmtAttr *proto.SecurityGroupUpdateMgmtAttrReq,
	sgID string) error {

	listReq := &dataproto.SecurityGroupListReq{
		Field:  []string{},
		Filter: tools.EqualExpression("id", sgID),
		Page:   core.NewDefaultBasePage(),
	}
	sgInfos, err := s.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, sg_id: %s, rid: %s", err, sgID, kt.Rid)
		return err
	}

	if len(sgInfos.Details) != 1 {
		logs.Errorf("list security group failed, len: %d, sg_id: %s, rid: %s", len(sgInfos.Details), sgID, kt.Rid)
		return fmt.Errorf("security group: %s not found", sgID)
	}

	// 平台管理不更新
	if sgInfos.Details[0].MgmtType == enumor.MgmtTypePlatform {
		logs.Infof("platform security group cannot be updated cloud mgmt attr, sg_id: %s, rid: %s", sgID, kt.Rid)
		return nil
	}

	updateAttrItem := proto.BatchUpdateSGMgmtAttrItem{
		ID:         sgID,
		Manager:    sgInfos.Details[0].Manager,
		BakManager: sgInfos.Details[0].BakManager,
		MgmtBizID:  sgInfos.Details[0].MgmtBizID,
	}
	if mgmtAttr.MgmtBizID != 0 {
		updateAttrItem.MgmtBizID = mgmtAttr.MgmtBizID
	}
	if mgmtAttr.Manager != "" {
		updateAttrItem.Manager = mgmtAttr.Manager
	}
	if mgmtAttr.BakManager != "" {
		updateAttrItem.BakManager = mgmtAttr.BakManager
	}

	sgMap := make(map[string]cloud.BaseSecurityGroup)
	sgMap[sgID] = sgInfos.Details[0]
	if err := s.updateCloudMgmtAttr(kt, []proto.BatchUpdateSGMgmtAttrItem{updateAttrItem}, sgMap); err != nil {
		logs.Errorf("update cloud security group management attributes failed, err: %v, rid: %s", err, kt.Rid)
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

	if len(sgs) != len(mgmtAttrs) {
		logs.Errorf("security group count not match, sgIDs count: %d, mgmtAttrs count: %d, rid: %s",
			len(sgIDs), len(mgmtAttrs), kt.Rid)
		return errors.New("security group count not match")
	}

	sgInfos := make(map[string]cloud.BaseSecurityGroup)
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

		sgInfos[sg.ID] = sg
	}

	// 管理业务和使用业务必须在帐号下的业务列表中
	for _, attr := range mgmtAttrs {
		sg := sgInfos[attr.ID]
		belongTo, err := s.isBizsBelongToAccount(kt, sg.AccountID, attr.MgmtBizID, []int64{})
		if err != nil {
			logs.Errorf("check bizs belong to account failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		if !belongTo {
			logs.Errorf("biz: %d are not belong to account: %s, sec_group: %s, rid: %s", attr.MgmtBizID,
				sg.AccountID, sg.ID, kt.Rid)
			return fmt.Errorf("biz: %d are not belong to account: %s", attr.MgmtBizID, sg.AccountID)
		}
	}

	// 更新管理属性到云上（如果有需要的话）
	if err := s.updateCloudMgmtAttr(kt, mgmtAttrs, sgInfos); err != nil {
		logs.Errorf("update cloud security group management attributes failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := s.batchUpdateSecurityGroupMgmtAttr(kt, mgmtAttrs, sgInfos); err != nil {
		logs.Errorf("batch update security group management attributes failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (s *securityGroup) batchUpdateSecurityGroupMgmtAttr(kt *kit.Kit, mgmtAttrs []proto.BatchUpdateSGMgmtAttrItem,
	sgInfos map[string]cloud.BaseSecurityGroup) error {

	for _, batch := range slice.Split(mgmtAttrs, constant.BatchOperationMaxLimit) {
		updateItems := make([]dataproto.BatchUpdateSGMgmtAttrItem, len(batch))
		for i, attr := range batch {
			updateItem := dataproto.BatchUpdateSGMgmtAttrItem{
				ID: attr.ID,
				// 批量更新默认更新为业务管理
				MgmtType:   enumor.MgmtTypeBiz,
				MgmtBizID:  attr.MgmtBizID,
				Manager:    attr.Manager,
				BakManager: attr.BakManager,
			}

			if _, ok := sgInfos[attr.ID]; ok {
				updateItem.Vendor = sgInfos[attr.ID].Vendor
				updateItem.CloudID = sgInfos[attr.ID].CloudID
			}

			updateItems[i] = updateItem
		}

		updateReq := &dataproto.BatchUpdateSecurityGroupMgmtAttrReq{
			SecurityGroups: updateItems,
		}

		if err := s.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupMgmtAttr(kt,
			updateReq); err != nil {
			logs.Errorf("batch update security group management attributes failed, err: %v, rid: %s", err,
				kt.Rid)
			return err
		}
	}

	return nil
}

func (s *securityGroup) isBizsBelongToAccount(kt *kit.Kit, accountID string, mgmtBiz int64, usageBizs []int64) (
	bool, error) {

	// 管理业务不可修改为未分配
	if mgmtBiz == constant.UnassignedBiz {
		return false, errors.New("cannot update security group management business to unassigned")
	}

	accountBizs := make([]int64, 0)
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("account_id", accountID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"account_id", "bk_biz_id"},
	}
	for {
		rst, err := s.client.DataService().Global.Account.ListAccountBizRel(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list account biz rel failed, err: %v, rid: %s", err, kt.Rid)
			return false, err
		}

		for _, bizID := range rst.Details {
			// 帐号全业务可见时，直接返回true
			if bizID.BkBizID == constant.AttachedAllBiz {
				return true, nil
			}
			accountBizs = append(accountBizs, bizID.BkBizID)
		}

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}

		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	if mgmtBiz != 0 {
		if !slice.IsItemInSlice(accountBizs, mgmtBiz) {
			return false, nil
		}
	}

	for _, bizID := range usageBizs {
		if !slice.IsItemInSlice(accountBizs, bizID) {
			return false, nil
		}
	}

	return true, nil
}

// updateCloudMgmtAttr 更新安全组在云上的管理属性
func (s *securityGroup) updateCloudMgmtAttr(kt *kit.Kit, mgmtAttrs []proto.BatchUpdateSGMgmtAttrItem,
	sgInfos map[string]cloud.BaseSecurityGroup) error {

	groupByVendor := make(map[enumor.Vendor][]proto.BatchUpdateSGMgmtAttrItem)
	for _, attr := range mgmtAttrs {
		if _, ok := sgInfos[attr.ID]; !ok {
			logs.Warnf("failed to update cloud mgmt attr, security group: %s not found, rid: %s", attr.ID, kt.Rid)
			continue
		}

		groupByVendor[sgInfos[attr.ID].Vendor] = append(groupByVendor[sgInfos[attr.ID].Vendor], attr)
	}

	for vendor, attrs := range groupByVendor {
		switch vendor {
		case enumor.TCloudZiyan:
			if err := s.updateTCloudZiyanMgmtAttr(kt, attrs, sgInfos); err != nil {
				logs.Errorf("failed to update cloud mgmt attr, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
				return err
			}
		case enumor.TCloud, enumor.Aws, enumor.Gcp, enumor.Azure, enumor.HuaWei, enumor.Zenlayer, enumor.Kaopu:
			// 暂不需要更新云上属性
			continue
		default:
			logs.Warnf("update cloud mgmt attr, unsupported vendor: %s, rid: %s", vendor, kt.Rid)
			continue
		}
	}

	return nil
}
