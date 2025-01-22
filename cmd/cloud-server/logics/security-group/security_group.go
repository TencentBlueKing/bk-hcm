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
	cssgproto "hcm/pkg/api/cloud-server/security-group"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// Interface define disk interface.
type Interface interface {
	ListSGRelCVM(kt *kit.Kit, sgID string, resBizID int64, listReq *core.ListReq) (
		*dataproto.SGCommonRelWithCVMListResp, error)
	ListSGRelLoadBalancer(kt *kit.Kit, sgID string, resBizID int64, oriReq *core.ListReq) (
		*dataproto.SGCommonRelWithLBListResp, error)
	ListSGRelBusiness(kt *kit.Kit, currentBizID int64, sgID string) (*cssgproto.ListSGRelBusinessResp, error)
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

// ListSGRelBusiness List the biz IDs that have resources associated with the security group. Group by resource type.
// Use sg_ID to query res_IDs in the rel table, then use res_IDs to query res bizs.
func (s *securityGroup) ListSGRelBusiness(kt *kit.Kit, currentBizID int64, sgID string) (
	*cssgproto.ListSGRelBusinessResp, error) {

	relListReq := &dataproto.SGCommonRelWithSecurityGroupListReq{
		SGIDs: []string{sgID},
	}

	// list security group rel resources
	sgRelResources, err := s.client.DataService().Global.SGCommonRel.ListWithSecurityGroup(kt, relListReq)
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

	for _, relRes := range *sgRelResources {
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

	return &cssgproto.ListSGRelBusinessResp{
		CVM:          cvmRelBizs,
		LoadBalancer: lbRelBizs,
	}, nil
}

func (s *securityGroup) listRelBizsWithCVM(kt *kit.Kit, currentBizID int64, cvmIDs []string) (
	[]cssgproto.ListSGRelBusinessItem, error) {

	relBizMap := make(map[int64]int64)
	for _, batch := range slice.Split(cvmIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
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
	[]cssgproto.ListSGRelBusinessItem, error) {

	relBizMap := make(map[int64]int64)
	for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
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

func tidySGRelBusiness(currentBizID int64, relBizMap map[int64]int64) []cssgproto.ListSGRelBusinessItem {
	var currentBizResC int64
	if resCount, ok := relBizMap[currentBizID]; ok {
		currentBizResC = resCount
		delete(relBizMap, currentBizID)
	}

	// 当前业务必须在列表的第一个
	relBizs := make([]cssgproto.ListSGRelBusinessItem, 0, len(relBizMap)+1)
	if currentBizID != 0 {
		relBizs[0] = cssgproto.ListSGRelBusinessItem{
			BkBizID:  currentBizID,
			ResCount: currentBizResC,
		}
	}
	for bizID, count := range relBizMap {
		relBizs = append(relBizs, cssgproto.ListSGRelBusinessItem{
			BkBizID:  bizID,
			ResCount: count,
		})
	}

	return relBizs
}
