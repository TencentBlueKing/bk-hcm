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

package tcloud

import (
	logicsaccount "hcm/cmd/cloud-server/logics/account"
	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	adcore "hcm/pkg/adaptor/types/core"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	hcbwpkg "hcm/pkg/api/hc-service/bandwidth-packages"
	"hcm/pkg/criteria/errf"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// CheckReq 检查申请单的数据是否正确
func (a *ApplicationOfCreateTCloudLB) CheckReq() error {
	if err := a.req.Validate(true); err != nil {
		return err
	}

	if err := logicsaccount.IsResourceAccount(a.Cts.Kit, a.Client.DataService(), a.req.AccountID); err != nil {
		return err
	}

	if !a.req.IsExclusive() {
		return nil
	}

	if err := lblogic.CheckExclusiveClusterOwnership(a.Cts.Kit, a.Client.DataService(), a.req); err != nil {
		return err
	}

	if err := a.checkBandwidthPackageEgress(); err != nil {
		return err
	}

	return a.checkVipIdle()
}

// checkVipIdle 指定了VIP时，实时查云校验该VIP在所选四层独占集群中仍然闲置。
func (a *ApplicationOfCreateTCloudLB) checkVipIdle() error {
	vip := cvt.PtrToVal(a.req.Vip)
	if vip == "" {
		return nil
	}

	if len(a.req.CloudClusterIDs) != 1 {
		return errf.New(errf.InvalidParameter, "cloud_cluster_ids must contain exactly one id when vip is specified")
	}

	return lblogic.CheckExclusiveClusterVipIdle(a.Cts.Kit, a.Client.HCService(), a.req.AccountID, a.req.Region,
		a.req.CloudClusterIDs[0], vip)
}

// checkBandwidthPackageEgress 计费方式为共享带宽包时，校验带宽包出口与可能被分配到的独占集群出口是否一致。
func (a *ApplicationOfCreateTCloudLB) checkBandwidthPackageEgress() error {
	if cvt.PtrToVal(a.req.InternetChargeType) != typelb.BandwidthPackage {
		return nil
	}

	egress, err := a.getBandwidthPackageEgress(cvt.PtrToVal(a.req.BandwidthPackageID))
	if err != nil {
		return err
	}

	allowedEgresses, err := lblogic.ComputeExclusiveClusterEgressSet(a.Cts.Kit, a.Client.DataService(), a.req)
	if err != nil {
		return err
	}

	if !slice.IsItemInSlice(allowedEgresses, egress) {
		return errf.Newf(errf.InvalidParameter,
			"bandwidth package egress(%s) is not allowed by the selected exclusive clusters", egress)
	}

	return nil
}

// getBandwidthPackageEgress 实时查云获取带宽包的网络出口，查不到按 InvalidParameter 处理。
func (a *ApplicationOfCreateTCloudLB) getBandwidthPackageEgress(bwPkgID string) (string, error) {
	opt := &hcbwpkg.ListTCloudBwPkgOption{
		AccountID:   a.req.AccountID,
		Region:      a.req.Region,
		Page:        &adcore.TCloudPage{Offset: 0, Limit: 1},
		PkgCloudIds: []string{bwPkgID},
	}
	result, err := a.Client.HCService().TCloud.BandPkg.ListBandwidthPackage(a.Cts.Kit, opt)
	if err != nil {
		return "", err
	}

	if len(result.Packages) == 0 {
		return "", errf.Newf(errf.InvalidParameter, "bandwidth package(%s) not found", bwPkgID)
	}

	return result.Packages[0].Egress, nil
}
