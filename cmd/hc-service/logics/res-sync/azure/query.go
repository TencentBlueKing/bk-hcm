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

package azure

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// getVpcMap retrieves the VPC mapping from the database based on the provided cloud VPC IDs.
func (cli *client) getVpcMap(kt *kit.Kit, accountID string, cloudVpcIDsMap map[string]string) (
	map[string]*common.VpcDB, error) {

	vpcMap := make(map[string]*common.VpcDB)

	cloudVpcIDs := make([]string, 0)
	for _, cloudID := range cloudVpcIDsMap {
		cloudVpcIDs = append(cloudVpcIDs, cloudID)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: cloudVpcIDs},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.Vpc.ListVpcExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list vpc from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			accountID, req, kt.Rid)
		return nil, err
	}
	vpcFromDB := result.Details

	if len(vpcFromDB) <= 0 {
		return vpcMap, fmt.Errorf("can not find vpc form db")
	}

	if err != nil {
		return vpcMap, err
	}

	for _, vpc := range vpcFromDB {
		for cvmID, vpcID := range cloudVpcIDsMap {
			if vpcID == vpc.CloudID {
				vpcMap[cvmID] = &common.VpcDB{
					VpcCloudID: vpc.CloudID,
					VpcID:      vpc.ID,
				}
			}
		}
	}

	return vpcMap, nil
}

// getSubnetMap retrieves the subnet mapping from the database based on the provided cloud subnet IDs.
func (cli *client) getSubnetMap(kt *kit.Kit, accountID string, cloudSubnetsIDsMap map[string]string) (
	map[string][]string, error) {

	subnetMap := make(map[string][]string)

	cloudSubnetsIDs := make([]string, 0)
	for _, cloudID := range cloudSubnetsIDsMap {
		cloudSubnetsIDs = append(cloudSubnetsIDs, cloudID)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: cloudSubnetsIDs},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list subnet from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			accountID, req, kt.Rid)
		return nil, err
	}

	subnetFromDB := result.Details

	for _, subnet := range subnetFromDB {
		for cvmID, subnetID := range cloudSubnetsIDsMap {
			if subnet.CloudID == subnetID {
				subnetMap[cvmID] = append(subnetMap[cvmID], subnet.ID)
			}
		}
	}

	return subnetMap, nil
}

// getImageMap retrieves the image mapping from the database based on the provided cloud image IDs.
func (cli *client) getImageMap(kt *kit.Kit, accountID string, resGroupName string,
	cloudImageIDs []string) (map[string]string, error) {

	imageMap := make(map[string]string)

	elems := slice.Split(cloudImageIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		imageParams := &SyncBaseParams{
			AccountID:         accountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          parts,
		}
		imageFromDB, err := cli.listImageFromDBForCvm(kt, imageParams)
		if err != nil {
			return imageMap, err
		}

		for _, image := range imageFromDB {
			imageMap[image.CloudID] = image.ID
		}
	}

	return imageMap, nil
}
