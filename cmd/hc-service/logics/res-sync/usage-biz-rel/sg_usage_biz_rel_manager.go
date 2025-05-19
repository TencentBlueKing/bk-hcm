/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package usagebizrelmgr

import (
	"slices"
	"sort"

	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// UsageBizRelManager 安全组使用业务关联管理
type UsageBizRelManager struct {
	dbCli *dataservice.Client
}

// NewUsageBizRelManager ...
func NewUsageBizRelManager(dbCli *dataservice.Client) *UsageBizRelManager {
	return &UsageBizRelManager{
		dbCli: dbCli,
	}
}

// SyncSecurityGroupUsageBiz 同步安全组使用业务
func (mgr *UsageBizRelManager) SyncSecurityGroupUsageBiz(kt *kit.Kit, sg *cloudcore.BaseSecurityGroup) error {

	// upgrade old version data
	if sg.BkBizID > 0 && sg.MgmtBizID <= 0 {
		req := &protocloud.BatchUpdateSecurityGroupMgmtAttrReq{
			SecurityGroups: []protocloud.BatchUpdateSGMgmtAttrItem{{
				ID:        sg.ID,
				MgmtType:  enumor.MgmtTypeBiz,
				MgmtBizID: sg.BkBizID,
				Vendor:    sg.Vendor,
			}},
		}
		err := mgr.dbCli.Global.SecurityGroup.BatchUpdateSecurityGroupMgmtAttr(kt, req)
		if err != nil {
			logs.Errorf("update security group mgmt biz and type failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if slice.IsItemInSlice(sg.UsageBizIDs, constant.AttachedAllBiz) {
		// already max biz range
		if len(sg.UsageBizIDs) == 1 {
			return nil
		}
		// if UsageBizIDs contains -1 and other bizs,
		// reset UsageBizIDs to []int64{constant.AttachedAllBiz} as required
		req := &protocloud.ResUsageBizRelUpdateReq{
			UsageBizIDs: []int64{constant.AttachedAllBiz},
			ResCloudID:  sg.CloudID,
			ResVendor:   sg.Vendor,
		}
		err := mgr.dbCli.Global.ResUsageBizRel.SetBizRels(kt, enumor.SecurityGroupCloudResType, sg.ID, req)
		if err != nil {
			logs.Errorf("reset sg(%s/%s) res usage biz to -1 failed, err: %v, rid: %s", sg.Vendor, sg.ID, err, kt.Rid)
			return err
		}
		return nil
	}
	bizIDList, err := mgr.querySGIDMap(kt, sg)
	if err != nil {
		return err
	}
	// 要求顺序一致
	if slices.Compare(sg.UsageBizIDs, bizIDList) == 0 {
		// no change
		return nil
	}
	req := &protocloud.ResUsageBizRelUpdateReq{
		UsageBizIDs: bizIDList,
		ResCloudID:  sg.CloudID,
		ResVendor:   sg.Vendor,
	}
	err = mgr.dbCli.Global.ResUsageBizRel.SetBizRels(kt, enumor.SecurityGroupCloudResType, sg.ID, req)
	if err != nil {
		logs.Errorf("update res usage biz rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (mgr *UsageBizRelManager) querySGIDMap(kt *kit.Kit, sg *cloudcore.BaseSecurityGroup) ([]int64, error) {

	var usageBizResCountMap = make(map[int64]int, len(sg.UsageBizIDs))
	for i := range sg.UsageBizIDs {
		if sg.UsageBizIDs[i] == sg.MgmtBizID {
			// 跳过管理业务
			continue
		}
		usageBizResCountMap[sg.UsageBizIDs[i]] = 0
	}

	// 1. 查询当前实际的安全组关联使用业务id
	relReq := &core.ListReq{
		Filter: tools.EqualExpression("security_group_id", sg.ID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"res_type", "res_id"},
	}
	resTypeIDMap := make(map[enumor.CloudResourceType][]string)
	for {
		resRelResp, err := mgr.dbCli.Global.SGCommonRel.ListSgCommonRels(kt, relReq)
		// 2. 查询当前实际的业务
		if err != nil {
			logs.Errorf("list sg common rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for i := range resRelResp.Details {
			detail := resRelResp.Details[i]
			resTypeIDMap[detail.ResType] = append(resTypeIDMap[detail.ResType], detail.ResID)
		}
		if len(resRelResp.Details) < int(relReq.Page.Limit) {
			break
		}
		relReq.Page.Start += uint32(relReq.Page.Limit)
	}
	for resType := range resTypeIDMap {
		resIDs := resTypeIDMap[resType]
		for _, resIDBatch := range slice.Split(resIDs, constant.BatchOperationMaxLimit) {
			// 查询资源的业务id
			basicInfoReq := protocloud.ListResourceBasicInfoReq{
				ResourceType: resType,
				IDs:          resIDBatch,
				Fields:       []string{"id", "bk_biz_id"},
			}
			basicInfo, err := mgr.dbCli.Global.Cloud.ListResBasicInfo(kt, basicInfoReq)
			if err != nil {
				logs.Errorf("list res basic info failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			for resID := range basicInfo {
				if basicInfo[resID].BkBizID <= 0 {
					// 跳过无效业务id
					continue
				}
				if basicInfo[resID].BkBizID == sg.MgmtBizID {
					// 跳过管理业务id
					continue
				}
				usageBizResCountMap[basicInfo[resID].BkBizID] += 1
			}
		}
	}

	bizList := cvt.MapKeyToSlice(usageBizResCountMap)
	// 按资源数量排序
	sort.Slice(bizList, func(i, j int) bool {
		return usageBizResCountMap[bizList[i]] > usageBizResCountMap[bizList[j]]
	})
	if sg.MgmtBizID > 0 {
		// 保证管理业务id在第一位
		newBizList := make([]int64, len(usageBizResCountMap)+1)
		newBizList[0] = sg.MgmtBizID
		copy(newBizList[1:], bizList)
		bizList = newBizList
	}
	return bizList, nil
}
