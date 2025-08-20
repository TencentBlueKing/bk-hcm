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

package gcp

import (
	"hcm/cmd/hc-service/logics/res-sync/common"
	typescvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/tools/slice"
)

func (cli *client) getNIAssResMapBySelfLinkFromNI(kt *kit.Kit, accountID string, region string, zone string,
	cvmSlice []typescvm.GcpCvm) (map[string]*common.VpcDB, map[string]*SubnetDB,
	map[string]string, []string, []string, map[string]string, error) {

	vpcSelfLinks := make([]string, 0)
	subnetSelfLinks := make([]string, 0)
	diskSelfLinks := make([]string, 0)
	imageSelfLinks := make([]string, 0)
	for _, one := range cvmSlice {
		if len(one.NetworkInterfaces) > 0 {
			for _, networkInterface := range one.NetworkInterfaces {
				if networkInterface != nil {
					vpcSelfLinks = append(vpcSelfLinks, networkInterface.Network)
					subnetSelfLinks = append(subnetSelfLinks, networkInterface.Subnetwork)
				}
			}
		}
		for _, one := range one.Disks {
			diskSelfLinks = append(diskSelfLinks, one.Source)
		}
		imageSelfLinks = append(imageSelfLinks, one.SourceMachineImage)
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, vpcSelfLinks)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	subnetMap, err := cli.getSubnetMap(kt, accountID, region, subnetSelfLinks)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	diskMap, err := cli.getDiskMap(kt, accountID, zone, diskSelfLinks)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	imageMap, err := cli.getImageMap(kt, accountID, imageSelfLinks)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	return vpcMap, subnetMap, diskMap, vpcSelfLinks, subnetSelfLinks, imageMap, nil
}

func (cli *client) getImageMap(kt *kit.Kit, accountID string,
	cloudImageIDs []string) (map[string]string, error) {

	imageMap := make(map[string]string)

	elems := slice.Split(cloudImageIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		imageParams := &ListBySelfLinkOption{
			AccountID: accountID,
			SelfLink:  parts,
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

func (cli *client) getVpcMap(kt *kit.Kit, accountID string, selfLink []string) (
	map[string]*common.VpcDB, error) {

	vpcMap := make(map[string]*common.VpcDB)

	elems := slice.Split(selfLink, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		vpcParams := &ListBySelfLinkOption{
			AccountID: accountID,
			SelfLink:  parts,
		}
		vpcFromDB, err := cli.listVpcFromDBBySelfLink(kt, vpcParams)
		if err != nil {
			return vpcMap, err
		}

		for _, vpc := range vpcFromDB {
			vpcMap[vpc.Extension.SelfLink] = &common.VpcDB{
				VpcCloudID: vpc.CloudID,
				VpcID:      vpc.ID,
			}
		}
	}

	return vpcMap, nil
}

func (cli *client) getSubnetMap(kt *kit.Kit, accountID string, region string,
	selfLink []string) (map[string]*SubnetDB, error) {

	subnetMap := make(map[string]*SubnetDB)

	elems := slice.Split(selfLink, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		subnetParams := &ListSubnetBySelfLinkOption{
			AccountID: accountID,
			Region:    region,
			SelfLink:  parts,
		}
		subnetFromDB, err := cli.listSubnetFromDBBySelfLink(kt, subnetParams)
		if err != nil {
			return subnetMap, err
		}

		for _, subnet := range subnetFromDB {
			subnetMap[subnet.Extension.SelfLink] = &SubnetDB{
				SubnetCloudID: subnet.CloudID,
				SubnetID:      subnet.ID,
			}
		}
	}

	return subnetMap, nil
}

func (cli *client) getDiskMap(kt *kit.Kit, accountID string, zone string,
	selfLink []string) (map[string]string, error) {

	diskMap := make(map[string]string)

	elems := slice.Split(selfLink, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		diskParams := &ListDiskBySelfLinkOption{
			AccountID: accountID,
			Zone:      zone,
			SelfLink:  parts,
		}
		diskFromDB, err := cli.listDiskFromDBBySelfLink(kt, diskParams)
		if err != nil {
			return diskMap, err
		}

		for _, disk := range diskFromDB {
			diskMap[disk.Extension.SelfLink] = disk.CloudID
		}
	}

	return diskMap, nil
}
