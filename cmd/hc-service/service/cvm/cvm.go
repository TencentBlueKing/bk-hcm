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

// Package cvm ...
package cvm

import (
	"fmt"

	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// InitCvmService initial the cvm service.
func InitCvmService(cap *capability.Capability) {
	svc := &cvmSvc{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	svc.initTCloudCvmService(cap)
	svc.initAwsCvmService(cap)
	svc.initAzureCvmService(cap)
	svc.initGcpCvmService(cap)
	svc.initHuaWeiCvmService(cap)
}

type cvmSvc struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
	client  *client.ClientSet
}

func (svc *cvmSvc) listCvms(kt *kit.Kit, cvmIDs ...string) ([]corecvm.BaseCvm, error) {
	if len(cvmIDs) == 0 {
		return nil, nil
	}

	result := make([]corecvm.BaseCvm, 0, len(cvmIDs))
	for _, ids := range slice.Split(cvmIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.dataCli.Global.Cvm.ListCvm(kt, req)
		if err != nil {
			logs.Errorf("get cvms failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		result = append(result, resp.Details...)
	}

	return result, nil
}

func (svc *cvmSvc) listSecurityGroupMap(kt *kit.Kit, sgIDs ...string) (map[string]cloud.BaseSecurityGroup, error) {
	if len(sgIDs) == 0 {
		return nil, fmt.Errorf("security group ids is empty")
	}

	result := make(map[string]cloud.BaseSecurityGroup)
	for _, ids := range slice.Split(sgIDs, int(core.DefaultMaxPageLimit)) {
		req := &protocloud.SecurityGroupListReq{
			Filter: tools.ContainersExpression("id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list security groups failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, one := range resp.Details {
			result[one.ID] = one
		}
	}

	return result, nil
}

func (svc *cvmSvc) getSecurityGroupMapByCloudIDs(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string) (
	map[string]string, error) {

	cloudIDs = slice.Unique(cloudIDs)
	m := make(map[string]string)
	for _, ids := range slice.Split(cloudIDs, int(core.DefaultMaxPageLimit)) {
		req := &protocloud.SecurityGroupListReq{
			Field: []string{"id", "cloud_id"},
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", ids),
				tools.RuleEqual("vendor", vendor),
			),
			Page: core.NewDefaultBasePage(),
		}
		resp, err := svc.dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("request dataservice list security group failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, one := range resp.Details {
			m[one.CloudID] = one.ID
		}
	}
	return m, nil
}

// createSGCommonRels 先删除cvmID关联的安全组关系，再创建新的安全组关系
func (svc *cvmSvc) createSGCommonRels(kt *kit.Kit, vendor enumor.Vendor, resType enumor.CloudResourceType, cvmID string, sgIDs []string) error {

	createReq := &protocloud.SGCommonRelBatchUpsertReq{
		DeleteReq: &dataproto.BatchDeleteReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("res_id", cvmID),
				tools.RuleEqual("res_type", resType),
			),
		},
	}

	for i, sgID := range sgIDs {
		createReq.Rels = append(createReq.Rels, protocloud.SGCommonRelCreate{
			SecurityGroupID: sgID,
			ResVendor:       vendor,
			ResID:           cvmID,
			ResType:         resType,
			Priority:        int64(i) + 1,
		})
	}
	if err := svc.dataCli.Global.SGCommonRel.BatchUpsertSgCommonRels(kt, createReq); err != nil {
		logs.Errorf("request dataservice create security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, createReq, kt.Rid)
		return err
	}
	return nil
}

func (svc *cvmSvc) getCvms(kt *kit.Kit, vendor enumor.Vendor, region string, cvmIDs []string) ([]corecvm.BaseCvm,
	error) {
	if len(cvmIDs) == 0 {
		return nil, nil
	}

	result := make([]corecvm.BaseCvm, 0, len(cvmIDs))
	for _, ids := range slice.Split(cvmIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleEqual("region", region),
				tools.RuleIn("id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		listResp, err := svc.dataCli.Global.Cvm.ListCvm(kt, listReq)
		if err != nil {
			logs.Errorf("request dataservice list cvm failed, err: %v, ids: %v, rid: %s", err, cvmIDs, kt.Rid)
			return nil, err
		}
		result = append(result, listResp.Details...)
	}

	return result, nil
}
