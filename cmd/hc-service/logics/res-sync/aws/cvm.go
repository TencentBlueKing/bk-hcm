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
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/aws"
	typescvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

// SyncCvmOption ...
type SyncCvmOption struct {
}

// Validate ...
func (opt SyncCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}

func (cli *client) Cvm(kt *kit.Kit, params *SyncBaseParams, opt *SyncCvmOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmFromCloud, err := cli.listCvmFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	cvmFromDB, err := cli.listCvmFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(cvmFromCloud) == 0 && len(cvmFromCloud) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typescvm.AwsCvm, corecvm.Cvm[cvm.AwsCvmExtension]](
		cvmFromCloud, cvmFromDB, isCvmChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteCvm(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createCvm(kt, params.AccountID, params.Region, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateCvm(kt, params.AccountID, params.Region, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createCvm(kt *kit.Kit, accountID string, region string,
	addSlice []typescvm.AwsCvm) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	lists := make([]dataproto.CvmBatchCreate[corecvm.AwsCvmExtension], 0)

	vpcMap, subnetMap, imageMap, err := cli.getCvmRelResMaps(kt, accountID, region, addSlice)
	if err != nil {
		return err
	}

	for _, one := range addSlice {
		if _, exsit := vpcMap[converter.PtrToVal(one.VpcId)]; !exsit {
			return fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.InstanceId))
		}

		if _, exsit := subnetMap[converter.PtrToVal(one.SubnetId)]; !exsit {
			return fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.InstanceId))
		}

		privateIPv4Addresses := make([]string, 0)
		if one.PrivateIpAddress != nil {
			privateIPv4Addresses = append(privateIPv4Addresses, converter.PtrToVal(one.PrivateIpAddress))
		}
		publicIPv4Addresses := make([]string, 0)
		if one.PublicIpAddress != nil {
			publicIPv4Addresses = append(publicIPv4Addresses, converter.PtrToVal(one.PublicIpAddress))
		}
		publicIPv6Addresses := make([]string, 0)
		if one.Ipv6Address != nil {
			publicIPv6Addresses = append(publicIPv6Addresses, converter.PtrToVal(one.Ipv6Address))
		}

		sgIDs := make([]string, 0)
		if len(one.SecurityGroups) > 0 {
			for _, sg := range one.SecurityGroups {
				if sg.GroupId != nil {
					sgIDs = append(sgIDs, converter.PtrToVal(sg.GroupId))
				}
			}
		}

		awsBlockDeviceMapping := make([]corecvm.AwsBlockDeviceMapping, 0)
		if len(one.BlockDeviceMappings) > 0 {
			for _, v := range one.BlockDeviceMappings {
				if v != nil {
					tmp := corecvm.AwsBlockDeviceMapping{
						Status:        v.Ebs.Status,
						CloudVolumeID: v.Ebs.VolumeId,
					}
					awsBlockDeviceMapping = append(awsBlockDeviceMapping, tmp)
				}
			}
		}

		imageID := ""
		if id, exsit := imageMap[converter.PtrToVal(one.ImageId)]; exsit {
			imageID = id
		}

		cvm := dataproto.CvmBatchCreate[corecvm.AwsCvmExtension]{
			CloudID:        converter.PtrToVal(one.InstanceId),
			Name:           converter.PtrToVal(aws.GetCvmNameFromTags(one.Tags)),
			BkBizID:        constant.UnassignedBiz,
			BkCloudID:      vpcMap[converter.PtrToVal(one.VpcId)].BkCloudID,
			AccountID:      accountID,
			Region:         region,
			Zone:           converter.PtrToVal(one.Placement.AvailabilityZone),
			CloudVpcIDs:    []string{converter.PtrToVal(one.VpcId)},
			VpcIDs:         []string{vpcMap[converter.PtrToVal(one.VpcId)].VpcID},
			CloudSubnetIDs: []string{converter.PtrToVal(one.SubnetId)},
			SubnetIDs:      []string{subnetMap[converter.PtrToVal(one.SubnetId)]},
			CloudImageID:   converter.PtrToVal(one.ImageId),
			ImageID:        imageID,
			OsName:         converter.PtrToVal(one.PlatformDetails),
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(one.State.Name),
			PrivateIPv4Addresses: privateIPv4Addresses,
			// 云上不支持该字段
			PrivateIPv6Addresses: nil,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			MachineType:          converter.PtrToVal(one.InstanceType),
			// 云上不支持该字段
			CloudCreatedTime:  "",
			CloudLaunchedTime: times.ConvStdTimeFormat(converter.PtrToVal(one.LaunchTime)),
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AwsCvmExtension{
				CpuOptions: &corecvm.AwsCpuOptions{
					CoreCount:      one.CpuOptions.CoreCount,
					ThreadsPerCore: one.CpuOptions.ThreadsPerCore,
				},
				Platform:              one.Platform,
				DnsName:               one.PublicDnsName,
				EbsOptimized:          one.EbsOptimized,
				CloudSecurityGroupIDs: sgIDs,
				PrivateDnsName:        one.PrivateDnsName,
				PrivateDnsNameOptions: nil,
				CloudRamDiskID:        one.RamdiskId,
				RootDeviceName:        one.RootDeviceName,
				RootDeviceType:        one.RootDeviceType,
				SourceDestCheck:       one.SourceDestCheck,
				SriovNetSupport:       one.SriovNetSupport,
				VirtualizationType:    one.VirtualizationType,
				BlockDeviceMapping:    awsBlockDeviceMapping,
			},
		}

		if one.PrivateDnsNameOptions != nil {
			cvm.Extension.PrivateDnsNameOptions = &corecvm.AwsPrivateDnsNameOptions{
				EnableResourceNameDnsAAAARecord: one.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
				EnableResourceNameDnsARecord:    one.PrivateDnsNameOptions.EnableResourceNameDnsARecord,
				HostnameType:                    one.PrivateDnsNameOptions.HostnameType,
			}
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.AwsCvmExtension]{
		Cvms: lists,
	}

	_, err = cli.dbCli.Aws.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create aws cvm failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to create cvm success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateCvm(kt *kit.Kit, accountID string, region string,
	updateMap map[string]typescvm.AwsCvm) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("cvm updateMap is <= 0, not update")
	}

	lists := make([]dataproto.CvmBatchUpdate[corecvm.AwsCvmExtension], 0)

	cloudDataSlice := make([]typescvm.AwsCvm, 0, len(updateMap))
	for _, one := range updateMap {
		cloudDataSlice = append(cloudDataSlice, one)
	}
	vpcMap, subnetMap, imageMap, err := cli.getCvmRelResMaps(kt, accountID, region, cloudDataSlice)
	if err != nil {
		return err
	}

	for id, one := range updateMap {
		if _, exsit := vpcMap[converter.PtrToVal(one.VpcId)]; !exsit {
			return fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.InstanceId))
		}

		if _, exsit := subnetMap[converter.PtrToVal(one.SubnetId)]; !exsit {
			return fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.InstanceId))
		}

		privateIPv4Addresses := make([]string, 0)
		if one.PrivateIpAddress != nil {
			privateIPv4Addresses = append(privateIPv4Addresses, converter.PtrToVal(one.PrivateIpAddress))
		}
		publicIPv4Addresses := make([]string, 0)
		if one.PublicIpAddress != nil {
			publicIPv4Addresses = append(publicIPv4Addresses, converter.PtrToVal(one.PublicIpAddress))
		}
		publicIPv6Addresses := make([]string, 0)
		if one.Ipv6Address != nil {
			publicIPv6Addresses = append(publicIPv6Addresses, converter.PtrToVal(one.Ipv6Address))
		}

		sgIDs := make([]string, 0)
		if len(one.SecurityGroups) > 0 {
			for _, sg := range one.SecurityGroups {
				if sg.GroupId != nil {
					sgIDs = append(sgIDs, converter.PtrToVal(sg.GroupId))
				}
			}
		}

		awsBlockDeviceMapping := make([]corecvm.AwsBlockDeviceMapping, 0)
		if len(one.BlockDeviceMappings) > 0 {
			for _, v := range one.BlockDeviceMappings {
				if v != nil {
					tmp := corecvm.AwsBlockDeviceMapping{
						Status:        v.Ebs.Status,
						CloudVolumeID: v.Ebs.VolumeId,
					}
					awsBlockDeviceMapping = append(awsBlockDeviceMapping, tmp)
				}
			}
		}

		imageID := ""
		if id, exsit := imageMap[converter.PtrToVal(one.ImageId)]; exsit {
			imageID = id
		}

		cvm := dataproto.CvmBatchUpdate[corecvm.AwsCvmExtension]{
			ID:             id,
			Name:           converter.PtrToVal(aws.GetCvmNameFromTags(one.Tags)),
			BkCloudID:      vpcMap[converter.PtrToVal(one.VpcId)].BkCloudID,
			CloudVpcIDs:    []string{converter.PtrToVal(one.VpcId)},
			VpcIDs:         []string{vpcMap[converter.PtrToVal(one.VpcId)].VpcID},
			CloudSubnetIDs: []string{converter.PtrToVal(one.SubnetId)},
			SubnetIDs:      []string{subnetMap[converter.PtrToVal(one.SubnetId)]},
			CloudImageID:   converter.PtrToVal(one.ImageId),
			ImageID:        imageID,
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(one.State.Name),
			PrivateIPv4Addresses: privateIPv4Addresses,
			// 云上不支持该字段
			PrivateIPv6Addresses: nil,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			CloudLaunchedTime:    times.ConvStdTimeFormat(converter.PtrToVal(one.LaunchTime)),
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AwsCvmExtension{
				CpuOptions: &corecvm.AwsCpuOptions{
					CoreCount:      one.CpuOptions.CoreCount,
					ThreadsPerCore: one.CpuOptions.ThreadsPerCore,
				},
				Platform:              one.Platform,
				DnsName:               one.PublicDnsName,
				EbsOptimized:          one.EbsOptimized,
				CloudSecurityGroupIDs: sgIDs,
				PrivateDnsName:        one.PrivateDnsName,
				PrivateDnsNameOptions: nil,
				CloudRamDiskID:        one.RamdiskId,
				RootDeviceName:        one.RootDeviceName,
				RootDeviceType:        one.RootDeviceType,
				SourceDestCheck:       one.SourceDestCheck,
				SriovNetSupport:       one.SriovNetSupport,
				VirtualizationType:    one.VirtualizationType,
				BlockDeviceMapping:    awsBlockDeviceMapping,
			},
		}

		if one.PrivateDnsNameOptions != nil {
			cvm.Extension.PrivateDnsNameOptions = &corecvm.AwsPrivateDnsNameOptions{
				EnableResourceNameDnsAAAARecord: one.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
				EnableResourceNameDnsARecord:    one.PrivateDnsNameOptions.EnableResourceNameDnsARecord,
				HostnameType:                    one.PrivateDnsNameOptions.HostnameType,
			}
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.AwsCvmExtension]{
		Cvms: lists,
	}

	if err := cli.dbCli.Aws.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateCvm failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to update cvm success, count: %d, ids: %v, rid: %s", enumor.Aws,
		len(updateMap), updateMap, kt.Rid)

	return nil
}

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
				BkCloudID:  vpc.BkCloudID,
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

func (cli *client) deleteCvm(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("cvm delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delCvmFromCloud, err := cli.listCvmFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delCvmFromCloud) > 0 {
		logs.Errorf("[%s] validate cvm not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Aws, checkParams, len(delCvmFromCloud), kt.Rid)
		return fmt.Errorf("validate cvm not exist failed, before delete")
	}

	deleteReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete cvm failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to delete cvm success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listCvmFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typescvm.AwsCvm, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typescvm.AwsListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
	}
	result, _, err := cli.cloudCli.ListCvm(kt, opt)
	if err != nil {
		if strings.Contains(err.Error(), aws.ErrCvmNotFound) {
			return make([]typescvm.AwsCvm, 0), nil
		}

		logs.Errorf("[%s] list cvm from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listCvmFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]corecvm.Cvm[cvm.AwsCvmExtension], error) {

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

func (cli *client) RemoveCvmDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
	req := &protocloud.CvmListReq{
		Field: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Aws.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list cvm failed, err: %v, req: %v, rid: %s", enumor.Aws,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		var delCloudIDs []string
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			delCloudIDs, err = cli.listRemoveCvmID(kt, params)
			if err != nil {
				return err
			}
		}

		if len(delCloudIDs) != 0 {
			if err = cli.deleteCvm(kt, accountID, region, delCloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func (cli *client) listRemoveCvmID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delCloudIDs := make([]string, 0)
	for _, one := range params.CloudIDs {
		opt := &typescvm.AwsListOption{
			Region:   params.Region,
			CloudIDs: []string{one},
		}
		_, _, err := cli.cloudCli.ListCvm(kt, opt)
		if err != nil {
			if strings.Contains(err.Error(), aws.ErrCvmNotFound) {
				delCloudIDs = append(delCloudIDs, one)
			}
		}
	}

	return delCloudIDs, nil
}

func isCvmChange(cloud typescvm.AwsCvm, db corecvm.Cvm[cvm.AwsCvmExtension]) bool {

	if db.CloudID != converter.PtrToVal(cloud.InstanceId) {
		return true
	}

	if db.Name != converter.PtrToVal(aws.GetCvmNameFromTags(cloud.Tags)) {
		return true
	}

	if db.CloudImageID != converter.PtrToVal(cloud.ImageId) {
		return true
	}

	if db.OsName != converter.PtrToVal(cloud.PlatformDetails) {
		return true
	}

	if db.Status != converter.PtrToVal(cloud.State.Name) {
		return true
	}

	if len(db.CloudVpcIDs) == 0 || (db.CloudVpcIDs[0] != converter.PtrToVal(cloud.VpcId)) {
		return true
	}

	if len(db.CloudSubnetIDs) == 0 || (db.CloudSubnetIDs[0] != converter.PtrToVal(cloud.SubnetId)) {
		return true
	}

	privateIPv4Addresses := make([]string, 0)
	if cloud.PrivateIpAddress != nil {
		privateIPv4Addresses = append(privateIPv4Addresses, converter.PtrToVal(cloud.PrivateIpAddress))
	}
	publicIPv4Addresses := make([]string, 0)
	if cloud.PublicIpAddress != nil {
		publicIPv4Addresses = append(publicIPv4Addresses, converter.PtrToVal(cloud.PublicIpAddress))
	}
	publicIPv6Addresses := make([]string, 0)
	if cloud.Ipv6Address != nil {
		publicIPv6Addresses = append(publicIPv6Addresses, converter.PtrToVal(cloud.Ipv6Address))
	}

	if !assert.IsStringSliceEqual(privateIPv4Addresses, db.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv4Addresses, db.PublicIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv6Addresses, db.PublicIPv6Addresses) {
		return true
	}

	if db.MachineType != converter.PtrToVal(cloud.InstanceType) {
		return true
	}

	if db.CloudLaunchedTime != times.ConvStdTimeFormat(converter.PtrToVal(cloud.LaunchTime)) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.Platform, cloud.Platform) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.DnsName, cloud.PublicDnsName) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.EbsOptimized, cloud.EbsOptimized) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.PrivateDnsName, cloud.PrivateDnsName) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.CloudRamDiskID, cloud.RamdiskId) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.RootDeviceName, cloud.RootDeviceName) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.PrivateDnsName, cloud.PrivateDnsName) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.RootDeviceType, cloud.RootDeviceType) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.SourceDestCheck, cloud.SourceDestCheck) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.SriovNetSupport, cloud.SriovNetSupport) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.VirtualizationType, cloud.VirtualizationType) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Extension.CpuOptions.CoreCount, cloud.CpuOptions.CoreCount) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Extension.CpuOptions.ThreadsPerCore, cloud.CpuOptions.ThreadsPerCore) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
		cloud.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.PrivateDnsNameOptions.EnableResourceNameDnsARecord, cloud.PrivateDnsNameOptions.EnableResourceNameDnsARecord) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.PrivateDnsNameOptions.HostnameType, cloud.PrivateDnsNameOptions.HostnameType) {
		return true
	}

	sgIDs := make([]string, 0)
	if len(cloud.SecurityGroups) > 0 {
		for _, sg := range cloud.SecurityGroups {
			if sg.GroupId != nil {
				sgIDs = append(sgIDs, converter.PtrToVal(sg.GroupId))
			}
		}
	}
	if !assert.IsStringSliceEqual(db.Extension.CloudSecurityGroupIDs, sgIDs) {
		return true
	}

	for _, dbValue := range db.Extension.BlockDeviceMapping {
		isEqual := false
		for _, cloudValue := range cloud.BlockDeviceMappings {
			if dbValue.CloudVolumeID == cloudValue.Ebs.VolumeId && dbValue.Status == cloudValue.Ebs.Status {
				isEqual = true
				break
			}
		}
		if !isEqual {
			return true
		}
	}

	return false
}
