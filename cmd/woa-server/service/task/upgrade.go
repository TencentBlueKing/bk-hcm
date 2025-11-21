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

package task

import (
	"fmt"
	"strings"

	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// CreateBizUpgradeCRPOrder create biz upgrade crp order
func (s *service) CreateBizUpgradeCRPOrder(cts *rest.Contexts) (any, error) {
	input := new(types.CreateUpgradeCrpOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create upgrade crp order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}
	input.BkBizID = bkBizID
	input.User = cts.Kit.User

	err = input.Validate()
	if err != nil {
		logs.Errorf("failed to create upgrade crp order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// 业务-IAAS资源创建权限
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Create}, BizID: input.BkBizID,
	})
	if err != nil {
		logs.Errorf("no permission to create upgrade crp order, bizID: %d, err: %v, rid: %s", input.BkBizID,
			err, cts.Kit.Rid)
		return nil, err
	}

	upgradeCVMList, err := s.getInstanceDetails(cts.Kit, input.BkBizID, input.User, input.UpgradeCvmList)
	if err != nil {
		logs.Errorf("failed to get instance details, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 统一使用 applyReq 实现交付逻辑
	applyReq := &types.ApplyReq{
		BkBizId: input.BkBizID,
		User:    input.User,
		// 升降配默认常规项目
		RequireType: enumor.RequireTypeRegular,
		Remark:      input.Remark,
		Suborders: []*types.Suborder{
			{
				ResourceType:   types.ResourceTypeUpgradeCvm,
				Replicas:       uint(len(upgradeCVMList)),
				AppliedCore:    0,
				Remark:         input.Remark,
				UpgradeCVMList: upgradeCVMList,
			},
		},
	}

	return s.createUpgradeCRPOrder(cts.Kit, applyReq)
}

func (s *service) getInstanceDetails(kt *kit.Kit, bkBizID int64, user string, upgradeList []types.UpgradeCvmItem) (
	[]*types.UpgradeCVMSpec, error) {

	// 根据 bkHostID 获取实例详情（机型、地域）、总核心数
	bkHostIDsMap := slice.FuncToMap(upgradeList, func(i types.UpgradeCvmItem) (int64, interface{}) {
		return i.BkHostID, nil
	})
	bkHostIDs := maps.Keys(bkHostIDsMap)

	// 根据 instanceID 获取实例详情（机型、地域）、总核心数
	instanceIDsMap := slice.FuncToMap(upgradeList, func(i types.UpgradeCvmItem) (string, interface{}) {
		return i.InstanceID, nil
	})
	instanceIDs := maps.Keys(instanceIDsMap)

	bkHostIDsDetails, instanceIDsDetails, err := s.listCVMByInstanceIDANDBkHostID(kt, bkHostIDs, instanceIDs)
	if err != nil {
		logs.Errorf("failed to list cvm by instance id and bk host id, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// generate upgrade cvm list
	rst := make([]*types.UpgradeCVMSpec, 0, len(upgradeList))
	for _, item := range upgradeList {
		var cvmDetail cvm.Cvm[cvm.TCloudZiyanHostExtension]
		var exists bool
		if item.InstanceID != "" {
			cvmDetail, exists = instanceIDsDetails[item.InstanceID]
			if !exists {
				logs.Errorf("cannot found instance detail by instance_id: %s, rid: %s", item.InstanceID, kt.Rid)
				return nil, fmt.Errorf("cannot found instance detail by instance_id: %s", item.InstanceID)
			}
		} else {
			cvmDetail, exists = bkHostIDsDetails[item.BkHostID]
			if !exists {
				logs.Errorf("cannot found instance detail by bk_host_id: %d, rid: %s", item.BkHostID, kt.Rid)
				return nil, fmt.Errorf("cannot found instance detail by bk_host_id: %d", item.BkHostID)
			}
		}

		// 业务校验
		if cvmDetail.BkBizID != bkBizID {
			logs.Errorf("instance %s is not belong to biz %d, rid: %s", item.InstanceID, bkBizID, kt.Rid)
			return nil, fmt.Errorf("instance %s is not belong to biz %d", item.InstanceID, bkBizID)
		}
		// 校验主备负责人
		if !strings.Contains(cvmDetail.Extension.Operator, user) &&
			!strings.Contains(cvmDetail.Extension.BkBakOperator, user) {
			logs.Errorf("user %s is not the operator of instance %s, rid: %s", user, item.InstanceID, kt.Rid)
			return nil, fmt.Errorf("user %s is not the operator of instance %s", user, item.InstanceID)
		}
		rst = append(rst, &types.UpgradeCVMSpec{
			InstanceID:           cvmDetail.CloudID,
			PrivateIPv4Addresses: cvmDetail.PrivateIPv4Addresses,
			PrivateIPv6Addresses: cvmDetail.PrivateIPv6Addresses,
			BkAssetID:            cvmDetail.BkAssetID,
			DeviceType:           cvmDetail.MachineType,
			RegionID:             cvmDetail.Region,
			ZoneID:               cvmDetail.Zone,
			TargetInstanceType:   item.TargetInstanceType,
		})
	}

	return rst, nil
}

func (s *service) listCVMByInstanceIDANDBkHostID(kt *kit.Kit, bkHostIDs []int64, instanceIDs []string) (
	map[int64]cvm.Cvm[cvm.TCloudZiyanHostExtension],
	map[string]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {

	listField := []string{"id", "cloud_id", "bk_biz_id", "bk_host_id", "machine_type", "region", "zone",
		"private_ipv4_addresses", "private_ipv6_addresses", "extension"}

	bkHostIDDetails := make(map[int64]cvm.Cvm[cvm.TCloudZiyanHostExtension])
	for _, batch := range slice.Split(bkHostIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.CvmListReq{
			Filter: tools.ContainersExpression("bk_host_id", batch),
			Page:   core.NewDefaultBasePage(),
			Field:  listField,
		}
		cvms, err := s.client.DataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("failed to get instance details by bk_host_id, err: %v, bk_host_id: %v, rid: %s", err,
				bkHostIDs, kt.Rid)
			return nil, nil, err
		}
		for _, item := range cvms.Details {
			bkHostIDDetails[item.BkHostID] = item
		}
	}

	instanceIDsDetails := make(map[string]cvm.Cvm[cvm.TCloudZiyanHostExtension])
	for _, batch := range slice.Split(instanceIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.CvmListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", enumor.TCloudZiyan),
				tools.RuleIn("cloud_id", batch),
			),
			Page:  core.NewDefaultBasePage(),
			Field: listField,
		}
		cvms, err := s.client.DataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("failed to get instance details by cloud_id, err: %v, cloud_ids: %v, rid: %s", err,
				instanceIDs, kt.Rid)
			return nil, nil, err
		}
		for _, item := range cvms.Details {
			instanceIDsDetails[item.CloudID] = item
		}
	}

	return bkHostIDDetails, instanceIDsDetails, nil
}

// createUpgradeCRPOrder creates upgrade crp order
func (s *service) createUpgradeCRPOrder(kt *kit.Kit, input *types.ApplyReq) (any, error) {
	// 校验预测
	if err := s.verifyResPlanDemand(kt, input); err != nil {
		logs.Errorf("failed to verify res plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 创建 ticket + suborder
	rst, err := s.logics.Scheduler().CreateUpgradeTicketANDOrder(kt, input)
	if err != nil {
		logs.Errorf("failed to create apply order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rst, nil
}
