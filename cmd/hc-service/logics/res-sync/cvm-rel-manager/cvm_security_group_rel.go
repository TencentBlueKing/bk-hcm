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

package cvmrelmgr

import (
	"sort"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service"
	datacloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// syncCvmSGRel sync cvm securityGroup rel.
// getCvmIDWithAssResIDMap CvmAppendAssResCloudID
// 根据上面两个方法可以得知 获取到的sg列表是有序的，按照这个顺序作为优先级写入关联关系表即可
func (mgr *CvmRelManger) syncCvmSGRel(kt *kit.Kit, cvmMap map[string]string, opt *SyncRelOption) error {

	if err := opt.Validate(); err != nil {
		return err
	}

	securityGroupMap, err := mgr.getSGMap(kt)
	if err != nil {
		logs.Errorf("get securityGroup map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmIDs, cvmIDToSgIDMapFromCloud, err := mgr.getCvmIDWithAssResIDMap(enumor.SecurityGroupCloudResType, cvmMap,
		securityGroupMap)
	if err != nil {
		logs.Errorf("get cvm id with ass res id map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmIDToRelsFromDB, err := mgr.listCvmSGRelsFromDB(kt, cvmIDs)
	if err != nil {
		logs.Errorf("get cvm_sg_rel map from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	err = mgr.compareCvmSGRel(kt, cvmIDToSgIDMapFromCloud, cvmIDToRelsFromDB, opt.Vendor)
	if err != nil {
		logs.Errorf("compare cvm sg rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (mgr *CvmRelManger) compareCvmSGRel(kt *kit.Kit, cvmIDToSgIDMapFromCloud map[string][]string,
	cvmIDToSGRelsMapFromDB map[string][]cloud.SGCommonRelWithBaseSecurityGroup, vendor enumor.Vendor) error {

	for cvmID, sgIDs := range cvmIDToSgIDMapFromCloud {
		localSGRels := cvmIDToSGRelsMapFromDB[cvmID]
		// 按优先级从小到大排序
		sort.Slice(localSGRels, func(i, j int) bool {
			return localSGRels[i].Priority < localSGRels[j].Priority
		})
		localLen := len(localSGRels)
		cloudLen := len(sgIDs)
		// 找到所有相等的列表
		var idx int
		var sgID string
		var stayLocalIDs []string
		for ; idx < cloudLen; idx++ {
			sgID = sgIDs[idx]
			if idx >= localLen || localSGRels[idx].ID != sgID || localSGRels[idx].Priority != int64(idx+1) {
				// 剩下的全部加入新增列表里
				break
			}
			// 加入可以保留的安全组id列表中
			stayLocalIDs = append(stayLocalIDs, sgID)
		}
		err := mgr.upsertSgRelForCvm(kt, cvmID, idx, stayLocalIDs, sgIDs[idx:], vendor)
		if err != nil {
			logs.Errorf("fail to upsert cvm(%s) security group rel, err: %v, rid: %s", cvmID, err, kt.Rid)
			return err
		}
	}

	return nil
}

func (mgr *CvmRelManger) upsertSgRelForCvm(kt *kit.Kit, cvmID string, startIdx int, stayLocalIDs []string,
	sgIDs []string, vendor enumor.Vendor) error {

	createDel := &datacloud.SGCommonRelBatchUpsertReq{Rels: make([]datacloud.SGCommonRelCreate, 0)}
	// 删除所有不在给定id中的安全组，防止误删
	createDel.DeleteReq = &dataproto.BatchDeleteReq{Filter: tools.ExpressionAnd(
		tools.RuleEqual("res_type", enumor.CvmCloudResType),
		tools.RuleEqual("res_id", cvmID),
	)}
	for i, sgID := range sgIDs {
		createDel.Rels = append(createDel.Rels, datacloud.SGCommonRelCreate{
			SecurityGroupID: sgID,
			ResVendor:       vendor,
			ResID:           cvmID,
			ResType:         enumor.CvmCloudResType,
			Priority:        int64(i + startIdx + 1),
		})

	}
	if len(stayLocalIDs) > 0 {
		createDel.DeleteReq.Filter.Rules = append(createDel.DeleteReq.Filter.Rules,
			tools.RuleNotIn("security_group_id", stayLocalIDs))
	}
	if len(createDel.Rels) > 0 {
		// 同时需要删除和创建
		err := mgr.dataCli.Global.SGCommonRel.BatchUpsertSgCommonRels(kt, createDel)
		if err != nil {
			logs.Errorf("fail to upsert cvm(%s) security group rel, err: %v, req: %+v, rid: %s",
				cvmID, err, createDel, kt.Rid)
			return err
		}
		return nil
	}

	// 只需要尝试删除多余关联关系即可
	err := mgr.dataCli.Global.SGCommonRel.BatchDeleteSgCommonRels(kt, createDel.DeleteReq)
	if err != nil {
		logs.Errorf("fail to delete cvm(%s) security group rel, err: %v, req: %+v, rid: %s",
			cvmID, err, createDel.DeleteReq, kt.Rid)
		return err
	}

	return nil
}

func (mgr *CvmRelManger) listCvmSGRelsFromDB(kt *kit.Kit, cvmIDs []string) (
	map[string][]cloud.SGCommonRelWithBaseSecurityGroup, error) {

	result := make(map[string][]cloud.SGCommonRelWithBaseSecurityGroup)
	for _, ids := range slice.Split(cvmIDs, constant.BatchOperationMaxLimit) {
		listReq := &datacloud.SGCommonRelWithSecurityGroupListReq{
			ResIDs:  ids,
			ResType: enumor.CvmCloudResType,
		}
		respResult, err := mgr.dataCli.Global.SGCommonRel.ListWithSecurityGroup(kt, listReq)
		if err != nil {
			logs.Errorf("list securityGroup cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, rel := range *respResult {
			if _, exist := result[rel.ResID]; !exist {
				result[rel.ResID] = make([]cloud.SGCommonRelWithBaseSecurityGroup, 0)
			}
			result[rel.ResID] = append(result[rel.ResID], rel)
		}
	}

	return result, nil
}

func (mgr *CvmRelManger) getSGMap(kt *kit.Kit) (map[string]string, error) {
	cloudIDs := mgr.getAllCvmAssResCloudIDs(enumor.SecurityGroupCloudResType)

	sgMap := make(map[string]string)
	split := slice.Split(cloudIDs, int(core.DefaultMaxPageLimit))
	for _, partCloudIDs := range split {
		req := &datacloud.SecurityGroupListReq{
			Field:  []string{"id", "cloud_id"},
			Filter: tools.ContainersExpression("cloud_id", partCloudIDs),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := mgr.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list securityGroup failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			sgMap[one.CloudID] = one.ID
		}
	}

	return sgMap, nil
}
