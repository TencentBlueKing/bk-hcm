/*
 *
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

package lblogic

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// RSInfo RS info
type RSInfo struct {
	InstType enumor.InstType `json:"inst_type"`
	IP       string          `json:"rsip"`
	Port     int             `json:"rsport"`
	EndPort  int             `json:"-"`
	Weight   int             `json:"weight"`
}

// CheckTarget 校验RSInfo是否存在
func (r *RSInfo) CheckTarget(kt *kit.Kit, vendor enumor.Vendor, bkBizID int64, client *dataservice.Client) error {

	switch vendor {
	case enumor.TCloud:
		return r.checkRSInCvm(kt, client, vendor, bkBizID)
	default:
		return fmt.Errorf("unsupported vendor %s", vendor)
	}
}

func (r *RSInfo) checkRSInCvm(kt *kit.Kit, client *dataservice.Client, vendor enumor.Vendor, bkBizID int64) error {
	// ENI 不进行检查
	if r.InstType == enumor.EniInstType {
		return nil
	}

	_, err := r.getCvmInfo(kt, client, vendor, bkBizID)
	if err != nil {
		logs.Errorf("get cvm info failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (r *RSInfo) getCvmInfo(kt *kit.Kit, client *dataservice.Client, vendor enumor.Vendor, bkBizID int64) (*corecvm.BaseCvm, error) {
	expr, err := tools.And(
		tools.ExpressionOr(
			tools.RuleJSONContains("private_ipv4_addresses", r.IP),
			tools.RuleJSONContains("private_ipv6_addresses", r.IP),
			tools.RuleJSONContains("public_ipv4_addresses", r.IP),
			tools.RuleJSONContains("public_ipv6_addresses", r.IP),
		),
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("bk_biz_id", bkBizID),
	)
	if err != nil {
		return nil, err
	}
	listReq := &core.ListReq{
		Filter: expr,
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	cvms, err := client.Global.Cvm.ListCvm(kt, listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(cvms.Details) == 0 {
		logs.Errorf("the RS %s not found, rid: %s", r.IP, kt.Rid)
		return nil, fmt.Errorf("the RS %s not found", r.IP)
	}
	return &cvms.Details[0], nil
}

// GetTargetReq 获取主机信息, TCloud load from table cvm
func (r *RSInfo) GetTargetReq(kt *kit.Kit, vendor enumor.Vendor, bkBizID int64, tgID, accountID string,
	client *dataservice.Client, cvmLgc cvm.Interface) (*cloud.TargetBaseReq, error) {

	switch vendor {
	case enumor.TCloud:
		return r.getTCloudTargetReq(kt, client, vendor, tgID, accountID, bkBizID)
	default:
		return nil, fmt.Errorf("unsupported vendor %s", vendor)
	}
}

func (r *RSInfo) getTCloudTargetReq(kt *kit.Kit, client *dataservice.Client, vendor enumor.Vendor,
	targetGroupID, accountID string, bkBizID int64) (*cloud.TargetBaseReq, error) {

	cvmInfo, err := r.getCvmInfo(kt, client, vendor, bkBizID)
	if err != nil {
		return nil, err
	}

	privateIPs := append(cvmInfo.PrivateIPv4Addresses, cvmInfo.PrivateIPv6Addresses...)
	publicIPs := append(cvmInfo.PublicIPv4Addresses, cvmInfo.PublicIPv6Addresses...)

	return &cloud.TargetBaseReq{
		IP:               r.IP,
		InstType:         r.InstType,
		Port:             int64(r.Port),
		Weight:           cvt.ValToPtr(int64(r.Weight)),
		AccountID:        accountID,
		TargetGroupID:    targetGroupID,
		CloudInstID:      cvmInfo.CloudID,
		InstName:         cvmInfo.Name,
		PrivateIPAddress: privateIPs,
		PublicIPAddress:  publicIPs,
		CloudVpcIDs:      cvmInfo.CloudVpcIDs,
		Zone:             cvmInfo.Zone,
	}, nil
}

func (r *RSInfo) checkTargetAlreadyExist(kt *kit.Kit, client *dataservice.Client, tgID string) error {
	expression, err := tools.And(
		tools.RuleEqual("target_group_id", tgID),
		tools.RuleEqual("port", r.Port),
		tools.ExpressionOr(
			tools.RuleJSONContains("private_ip_address", r.IP),
			tools.RuleJSONContains("public_ip_address", r.IP),
		),
	)
	if err != nil {
		return err
	}
	listReq := &core.ListReq{
		Filter: expression,
		Page:   core.NewDefaultBasePage(),
	}
	target, err := client.Global.LoadBalancer.ListTarget(kt, listReq)
	if err != nil {
		return err
	}

	if len(target.Details) > 0 {
		return fmt.Errorf("target %s:%d already exists", r.IP, r.Port)
	}

	return nil
}

// GetKey get key
func (r *RSInfo) GetKey() string {
	return fmt.Sprintf("%s:%d", r.IP, r.Port)
}
