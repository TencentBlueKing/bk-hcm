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

package cvm

import (
	"fmt"

	logicaudit "hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/disk"
	"hcm/cmd/cloud-server/logics/eip"
	logicsni "hcm/cmd/cloud-server/logics/network-interface"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// Assign 分配主机及关联资源到业务下
func Assign(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error {

	if len(ids) == 0 {
		return fmt.Errorf("ids is required")
	}

	// 校验主机信息
	if err := ValidateBeforeAssign(kt, cli, ids); err != nil {
		return err
	}

	// 获取主机关联资源
	eipIDs, diskIDs, niIDs, err := GetCvmRelResIDs(kt, cli, ids)
	if err != nil {
		return err
	}

	// 校验主机关联资源信息
	if err := ValidateCvmRelResBeforeAssign(kt, cli, bizID, eipIDs, diskIDs, niIDs); err != nil {
		return err
	}

	// create assign audit
	audit := logicaudit.NewAudit(cli)
	if err := audit.ResBizAssignAudit(kt, enumor.CvmAuditResType, ids, bizID); err != nil {
		logs.Errorf("create assign cvm audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 分配主机关联资源
	if err := AssignCvmRelRes(kt, cli, eipIDs, diskIDs, niIDs, bizID); err != nil {
		return err
	}

	// 分配主机
	update := &dataproto.CvmCommonInfoBatchUpdateReq{IDs: ids, BkBizID: bizID}
	if err := cli.Global.Cvm.BatchUpdateCvmCommonInfo(kt, update); err != nil {
		logs.Errorf("batch update cvm common info failed, err: %v, req: %v, rid: %s", err, update, kt.Rid)
		return err
	}

	return nil
}

// AssignCvmRelRes 分配主机关联资源
func AssignCvmRelRes(kt *kit.Kit, cli *dataservice.Client, eipIDs []string,
	diskIDs []string, niIDs []string, bizID int64) error {

	if len(eipIDs) != 0 {
		if err := eip.Assign(kt, cli, eipIDs, uint64(bizID), true); err != nil {
			return err
		}
	}

	if len(diskIDs) != 0 {
		if err := disk.Assign(kt, cli, diskIDs, uint64(bizID), true); err != nil {
			return err
		}
	}

	if len(niIDs) != 0 {
		if err := logicsni.Assign(kt, cli, niIDs, bizID, true); err != nil {
			return err
		}
	}

	return nil
}

// GetCvmRelResIDs 获取主机关联资源ID列表
func GetCvmRelResIDs(kt *kit.Kit, cli *dataservice.Client, ids []string) (
	eipIDs []string, diskIDs []string, niIDs []string, err error) {

	listRelReq := &core.ListReq{
		Filter: tools.ContainersExpression("cvm_id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	diskResp, err := cli.Global.ListDiskCvmRel(kt, listRelReq)
	if err != nil {
		logs.Errorf("list disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	listRelReq = &core.ListReq{
		Filter: tools.ContainersExpression("cvm_id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	eipResp, err := cli.Global.ListEipCvmRel(kt, listRelReq)
	if err != nil {
		logs.Errorf("list eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	listRelReq = &core.ListReq{
		Filter: tools.ContainersExpression("cvm_id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	niResp, err := cli.Global.NetworkInterfaceCvmRel.List(kt, listRelReq)
	if err != nil {
		logs.Errorf("list network_interface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	eipIDs = slice.Map(eipResp.Details, func(rel *dataproto.EipCvmRelResult) string {
		return rel.EipID
	})

	diskIDs = slice.Map(diskResp.Details, func(rel *dataproto.DiskCvmRelResult) string {
		return rel.DiskID
	})

	niIDs = slice.Map(niResp.Details, func(rel *dataproto.NetworkInterfaceCvmRelResult) string {
		return rel.NetworkInterfaceID
	})

	return
}

// ValidateCvmRelResBeforeAssign 校验主机关联资源在分配前
func ValidateCvmRelResBeforeAssign(kt *kit.Kit, cli *dataservice.Client, targetBizId int64, eipIDs []string,
	diskIDs []string, niIDs []string) error {

	if len(eipIDs) != 0 {
		if err := eip.ValidateBeforeAssign(kt, cli, targetBizId, eipIDs, true); err != nil {
			return err
		}
	}

	if len(diskIDs) != 0 {
		if err := disk.ValidateBeforeAssign(kt, cli, targetBizId, diskIDs, true); err != nil {
			return err
		}
	}

	if len(niIDs) != 0 {
		if err := logicsni.ValidateBeforeAssign(kt, cli, targetBizId, niIDs, true); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBeforeAssign 分配主机前校验主机信息
func ValidateBeforeAssign(kt *kit.Kit, cli *dataservice.Client, ids []string) error {
	listReq := &core.ListReq{
		Fields: []string{"id", "bk_biz_id", "bk_cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: ids},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.Global.Cvm.ListCvm(kt, listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	accountIDMap := make(map[string]struct{}, 0)
	assignedIDs := make([]string, 0)
	unBindCloudIDs := make([]string, 0)
	for _, one := range result.Details {
		accountIDMap[one.AccountID] = struct{}{}

		if one.BkBizID != constant.UnassignedBiz {
			assignedIDs = append(assignedIDs, one.ID)
		}

		if one.BkCloudID == constant.UnbindBkCloudID {
			unBindCloudIDs = append(unBindCloudIDs, one.ID)
		}
	}

	if len(assignedIDs) != 0 {
		return fmt.Errorf("cvm(ids=%v) already assigned", assignedIDs)
	}

	if len(unBindCloudIDs) != 0 {
		return fmt.Errorf("cvm(ids=%v) not bind cloud area", unBindCloudIDs)
	}

	return nil
}
