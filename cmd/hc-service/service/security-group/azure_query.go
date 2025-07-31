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
	adazure "hcm/pkg/adaptor/azure"
	typescore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

func (g *securityGroup) listAzureSecurityGroupFromCloud(kt *kit.Kit, client *adazure.Azure, resourceGroupName string,
	cloudIDs []string) ([]*armnetwork.SecurityGroup, error) {

	opt := &typescore.AzureListByIDOption{
		ResourceGroupName: resourceGroupName,
		CloudIDs:          cloudIDs,
	}
	resp, err := client.ListRawSecurityGroupByID(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list azure security group by id failed, err: %v, opt: %v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	return resp, nil
}

func (g *securityGroup) getAzureSecurityGroupMap(kt *kit.Kit, sgIDs []string) (
	map[string]corecloud.SecurityGroup[corecloud.AzureSecurityGroupExtension], error) {

	sgReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", sgIDs),
		Page:   core.NewDefaultBasePage(),
	}
	sgResult, err := g.dataCli.Azure.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), sgReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud security group failed, err: %v, ids: %v, rid: %s",
			err, sgIDs, kt.Rid)
		return nil, err
	}

	sgMap := make(map[string]corecloud.SecurityGroup[corecloud.AzureSecurityGroupExtension], len(sgResult.Details))
	for _, sg := range sgResult.Details {
		sgMap[sg.ID] = sg
	}

	return sgMap, nil
}
