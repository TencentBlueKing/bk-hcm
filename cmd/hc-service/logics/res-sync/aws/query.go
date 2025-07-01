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

package aws

import (
	"hcm/cmd/hc-service/logics/res-sync/common"
	typescvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coredisk "hcm/pkg/api/core/cloud/disk"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

func (cli *client) getCvmRelResMaps(kt *kit.Kit, accountID string, region string,
	cloudDataSlice []typescvm.AwsCvm) (map[string]*common.VpcDB, map[string]string, map[string]string, error) {

	cloudVpcIDs := make([]string, 0)
	cloudSubnetIDs := make([]string, 0)
	cloudImageIDs := make([]string, 0)
	for _, one := range cloudDataSlice {
		cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(one.VpcId))
		cloudSubnetIDs = append(cloudSubnetIDs, converter.PtrToVal(one.SubnetId))
		cloudImageIDs = append(cloudImageIDs, converter.PtrToVal(one.ImageId))
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, region, cloudVpcIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	subnetMap, err := cli.getSubnetMap(kt, accountID, region, cloudSubnetIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	imageMap, err := cli.getImageMap(kt, accountID, region, cloudImageIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return vpcMap, subnetMap, imageMap, nil
}

func (cli *client) getVpcMap(kt *kit.Kit, accountID string, region string,
	cloudVpcIDs []string) (map[string]*common.VpcDB, error) {

	vpcMap := make(map[string]*common.VpcDB)

	elems := slice.Split(cloudVpcIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		vpcParams := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  parts,
		}
		vpcFromDB, err := cli.listVpcFromDB(kt, vpcParams)
		if err != nil {
			return vpcMap, err
		}

		for _, vpc := range vpcFromDB {
			vpcMap[vpc.CloudID] = &common.VpcDB{
				VpcCloudID: vpc.CloudID,
				VpcID:      vpc.ID,
			}
		}
	}

	return vpcMap, nil
}

func (cli *client) getSubnetMap(kt *kit.Kit, accountID string, region string,
	cloudSubnetsIDs []string) (map[string]string, error) {

	subnetMap := make(map[string]string)

	elems := slice.Split(cloudSubnetsIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		subnetParams := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  parts,
		}
		subnetFromDB, err := cli.listSubnetFromDB(kt, subnetParams)
		if err != nil {
			return subnetMap, err
		}

		for _, subnet := range subnetFromDB {
			subnetMap[subnet.CloudID] = subnet.ID
		}
	}

	return subnetMap, nil
}

// getImageMap get image map from database based on the cloud image IDs.
func (cli *client) getImageMap(kt *kit.Kit, accountID string, region string,
	cloudImageIDs []string) (map[string]string, error) {

	imageMap := make(map[string]string)

	elems := slice.Split(cloudImageIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		imageParams := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  parts,
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

// listCvmFromDB lists CVM from the database based on the provided parameters.
func (cli *client) listCvmFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]corecvm.Cvm[corecvm.AwsCvmExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &protocloud.CvmListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: params.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Aws.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list cvm from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// listDiskFromDB list disk from db.
func (cli *client) listDiskFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*coredisk.Disk[coredisk.AwsExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: params.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Aws.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}
