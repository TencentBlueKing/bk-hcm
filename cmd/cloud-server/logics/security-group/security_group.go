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
	"hcm/pkg/tools/slice"
)

// Interface define disk interface.
type Interface interface {
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

func (s *securityGroup) listRelBizsWithCVM(kt *kit.Kit, currentBizID int64, cvmIDs []string) ([]int64, error) {
	relBizMap := make(map[int64]interface{})
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
			relBizMap[item.BkBizID] = struct{}{}
		}
	}

	if _, ok := relBizMap[currentBizID]; ok {
		delete(relBizMap, currentBizID)
	}

	// 当前业务必须在列表的第一个
	relBizs := make([]int64, 0, len(relBizMap)+1)
	if currentBizID != 0 {
		relBizs[0] = currentBizID
	}
	for bizID := range relBizMap {
		relBizs = append(relBizs, bizID)
	}

	return relBizs, nil
}

func (s *securityGroup) listRelBizsWithLB(kt *kit.Kit, currentBizID int64, lbIDs []string) ([]int64, error) {
	relBizMap := make(map[int64]interface{})
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
			relBizMap[item.BkBizID] = struct{}{}
		}
	}

	if _, ok := relBizMap[currentBizID]; ok {
		delete(relBizMap, currentBizID)
	}

	// 当前业务必须在列表的第一个
	relBizs := make([]int64, 0, len(relBizMap)+1)
	if currentBizID != 0 {
		relBizs[0] = currentBizID
	}
	for bizID := range relBizMap {
		relBizs = append(relBizs, bizID)
	}

	return relBizs, nil
}
