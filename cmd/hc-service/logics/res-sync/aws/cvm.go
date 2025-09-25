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

	"github.com/aws/aws-sdk-go/service/ec2"
)

// SyncCvmOption ...
type SyncCvmOption struct {
}

// Validate ...
func (opt SyncCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Cvm ...
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

	if len(cvmFromCloud) == 0 && len(cvmFromDB) == 0 {
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

// createCvm ...
func (cli *client) createCvm(kt *kit.Kit, accountID string, region string,
	addSlice []typescvm.AwsCvm) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	vpcMap, subnetMap, imageMap, err := cli.getCvmRelResMaps(kt, accountID, region, addSlice)
	if err != nil {
		return err
	}

	lists, err := buildCvmBatchCreateList(addSlice, accountID, region, vpcMap, subnetMap, imageMap)
	if err != nil {
		logs.Errorf("[%s] build cvm batch create list failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
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

// buildCvmBatchCreateList build cvm batch create list
func buildCvmBatchCreateList(addSlice []typescvm.AwsCvm, accountID, region string, vpcMap map[string]*common.VpcDB,
	subnetMap map[string]string, imageMap map[string]string) (
	[]protocloud.CvmBatchCreate[corecvm.AwsCvmExtension], error) {

	lists := make([]dataproto.CvmBatchCreate[corecvm.AwsCvmExtension], 0)
	for _, one := range addSlice {
		if _, exsit := vpcMap[converter.PtrToVal(one.VpcId)]; !exsit {
			return nil, fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.InstanceId))
		}

		if _, exsit := subnetMap[converter.PtrToVal(one.SubnetId)]; !exsit {
			return nil, fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.InstanceId))
		}

		privateIPv4Addresses, publicIPv4Addresses, publicIPv6Addresses := getIPAddress(one)

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

		req := buildAwsCvmCreateReq(one, accountID, region, vpcMap[converter.PtrToVal(one.VpcId)].VpcID,
			subnetMap[converter.PtrToVal(one.SubnetId)], imageID, privateIPv4Addresses, publicIPv4Addresses,
			publicIPv6Addresses, sgIDs, awsBlockDeviceMapping)
		lists = append(lists, req)
	}
	return lists, nil
}

// buildAwsCvmCreateReq builds a cvm create request for AWS
func buildAwsCvmCreateReq(one typescvm.AwsCvm, accountID, region, vpcID, subnetID, imageID string,
	privateIPv4Addresses, publicIPv4Addresses, publicIPv6Addresses, sgIDs []string,
	awsBlockDeviceMapping []corecvm.AwsBlockDeviceMapping) protocloud.CvmBatchCreate[corecvm.AwsCvmExtension] {

	cvm := dataproto.CvmBatchCreate[corecvm.AwsCvmExtension]{
		CloudID:        converter.PtrToVal(one.InstanceId),
		Name:           converter.PtrToVal(aws.GetCvmNameFromTags(one.Tags)),
		BkBizID:        constant.UnassignedBiz,
		BkHostID:       constant.UnBindBkHostID,
		BkCloudID:      constant.UnassignedBkCloudID,
		AccountID:      accountID,
		Region:         region,
		Zone:           converter.PtrToVal(one.Placement.AvailabilityZone),
		CloudVpcIDs:    []string{converter.PtrToVal(one.VpcId)},
		VpcIDs:         []string{vpcID},
		CloudSubnetIDs: []string{converter.PtrToVal(one.SubnetId)},
		SubnetIDs:      []string{subnetID},
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
		Extension:        convAwsExtension(one, sgIDs, awsBlockDeviceMapping),
	}

	return cvm
}

// updateCvm ...
func (cli *client) updateCvm(kt *kit.Kit, accountID string, region string,
	updateMap map[string]typescvm.AwsCvm) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("cvm updateMap is <= 0, not update")
	}

	lists := make([]dataproto.CvmBatchUpdateWithExtension[corecvm.AwsCvmExtension], 0)

	cloudDataSlice := make([]typescvm.AwsCvm, 0, len(updateMap))
	for _, one := range updateMap {
		cloudDataSlice = append(cloudDataSlice, one)
	}
	vpcMap, subnetMap, imageMap, err := cli.getCvmRelResMaps(kt, accountID, region, cloudDataSlice)
	if err != nil {
		return err
	}

	for id, one := range updateMap {
		if _, exists := vpcMap[converter.PtrToVal(one.VpcId)]; !exists {
			return fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.InstanceId))
		}

		if _, exists := subnetMap[converter.PtrToVal(one.SubnetId)]; !exists {
			return fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.InstanceId))
		}

		req := buildCvmUpdateReqWithAwsExtension(id, one, vpcMap, subnetMap, imageMap)
		lists = append(lists, req)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.AwsCvmExtension]{
		Cvms: lists,
	}

	if err := cli.dbCli.Aws.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateCvm failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to update cvm success, count: %d, rid: %s", enumor.Aws, len(updateMap), kt.Rid)

	return nil
}

func buildCvmUpdateReqWithAwsExtension(id string, one typescvm.AwsCvm, vpcMap map[string]*common.VpcDB,
	subnetMap map[string]string,
	imageMap map[string]string) protocloud.CvmBatchUpdateWithExtension[corecvm.AwsCvmExtension] {

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
	if id, exists := imageMap[converter.PtrToVal(one.ImageId)]; exists {
		imageID = id
	}
	privateIPv4Addresses, publicIPv4Addresses, publicIPv6Addresses := parseIPInfo(one)

	req := dataproto.CvmBatchUpdateWithExtension[corecvm.AwsCvmExtension]{
		CvmBatchUpdate: dataproto.CvmBatchUpdate{
			ID:             id,
			Name:           converter.PtrToVal(aws.GetCvmNameFromTags(one.Tags)),
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
		},
		Extension: convAwsExtension(one, sgIDs, awsBlockDeviceMapping),
	}
	return req
}

func parseIPInfo(one typescvm.AwsCvm) (privateIPv4Addresses, publicIPv4Addresses, publicIPv6Addresses []string) {
	if one.PrivateIpAddress != nil {
		privateIPv4Addresses = append(privateIPv4Addresses, converter.PtrToVal(one.PrivateIpAddress))
	}
	if one.PublicIpAddress != nil {
		publicIPv4Addresses = append(publicIPv4Addresses, converter.PtrToVal(one.PublicIpAddress))
	}
	if one.Ipv6Address != nil {
		publicIPv6Addresses = append(publicIPv6Addresses, converter.PtrToVal(one.Ipv6Address))
	}
	return
}

// deleteCvm ...
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

// listCvmFromCloud ...
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

// RemoveCvmDeleteFromCloud ...
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

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listCvmFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.InstanceId))
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteCvm(kt, accountID, region, cloudIDs); err != nil {
					return err
				}
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

// listRemoveCvmID ...
func (cli *client) listRemoveCvmID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delCloudIDs := make([]string, 0)
	cloudIDs := params.CloudIDs
	for {
		opt := &typescvm.AwsListOption{
			Region:   params.Region,
			CloudIDs: cloudIDs,
		}
		cvms, _, err := cli.cloudCli.ListCvm(kt, opt)
		if err != nil {
			if strings.Contains(err.Error(), aws.ErrCvmNotFound) {
				var delCloudID string
				cloudIDs, delCloudID = removeNotFoundCloudID(cloudIDs, err)
				delCloudIDs = append(delCloudIDs, delCloudID)

				if len(cloudIDs) <= 0 {
					break
				}

				continue
			}

			logs.Errorf("[%s] list cvm from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
				params.AccountID, opt, kt.Rid)
			return nil, err
		}

		fmt.Println(len(cvms))

		break
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

	privateIPv4Addresses, publicIPv4Addresses, publicIPv6Addresses := getIPAddress(cloud)

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

	if isCvmExtensionChange(cloud, db.Extension) {
		return false
	}

	return false
}

func isCvmExtensionChange(cloud typescvm.AwsCvm, dbExt *corecvm.AwsCvmExtension) bool {
	if !assert.IsPtrStringEqual(dbExt.Platform, cloud.Platform) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.DnsName, cloud.PublicDnsName) {
		return true
	}

	if !assert.IsPtrBoolEqual(dbExt.EbsOptimized, cloud.EbsOptimized) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.PrivateDnsName, cloud.PrivateDnsName) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.CloudRamDiskID, cloud.RamdiskId) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.RootDeviceName, cloud.RootDeviceName) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.PrivateDnsName, cloud.PrivateDnsName) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.RootDeviceType, cloud.RootDeviceType) {
		return true
	}

	if !assert.IsPtrBoolEqual(dbExt.SourceDestCheck, cloud.SourceDestCheck) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.SriovNetSupport, cloud.SriovNetSupport) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.VirtualizationType, cloud.VirtualizationType) {
		return true
	}

	if !assert.IsPtrInt64Equal(dbExt.CpuOptions.CoreCount, cloud.CpuOptions.CoreCount) {
		return true
	}

	if !assert.IsPtrInt64Equal(dbExt.CpuOptions.ThreadsPerCore, cloud.CpuOptions.ThreadsPerCore) {
		return true
	}

	if !assert.IsPtrBoolEqual(dbExt.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
		cloud.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord) {
		return true
	}

	if !assert.IsPtrBoolEqual(dbExt.PrivateDnsNameOptions.EnableResourceNameDnsARecord,
		cloud.PrivateDnsNameOptions.EnableResourceNameDnsARecord) {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.PrivateDnsNameOptions.HostnameType, cloud.PrivateDnsNameOptions.HostnameType) {
		return true
	}

	if isCvmSecurityGroupChange(cloud, dbExt) {
		return true
	}

	if isAwsVolumeChange(cloud.BlockDeviceMappings, dbExt.BlockDeviceMapping) {
		return true
	}

	return false
}

func isAwsVolumeChange(cloud []*ec2.InstanceBlockDeviceMapping, db []corecvm.AwsBlockDeviceMapping) bool {

	dbVolumeIDs := slice.Map(db, func(one corecvm.AwsBlockDeviceMapping) *string {
		return one.CloudVolumeID
	})
	cloudVolumeIDs := slice.Map(cloud, func(one *ec2.InstanceBlockDeviceMapping) *string {
		return one.Ebs.VolumeId
	})
	if !assert.IsPtrStringSliceEqual(dbVolumeIDs, cloudVolumeIDs) {
		return true
	}
	return false
}
func isCvmSecurityGroupChange(cloud typescvm.AwsCvm, dbExt *corecvm.AwsCvmExtension) bool {
	sgIDs := make([]string, 0)
	if len(cloud.SecurityGroups) > 0 {
		for _, sg := range cloud.SecurityGroups {
			if sg.GroupId != nil {
				sgIDs = append(sgIDs, converter.PtrToVal(sg.GroupId))
			}
		}
	}
	if !assert.IsStringSliceEqual(dbExt.CloudSecurityGroupIDs, sgIDs) {
		return true
	}
	return false
}

func convAwsExtension(one typescvm.AwsCvm, sgIDs []string,
	awsBlockDeviceMapping []corecvm.AwsBlockDeviceMapping) *corecvm.AwsCvmExtension {

	extension := &corecvm.AwsCvmExtension{
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
	}

	if one.PrivateDnsNameOptions != nil {
		extension.PrivateDnsNameOptions = &corecvm.AwsPrivateDnsNameOptions{
			EnableResourceNameDnsAAAARecord: one.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
			EnableResourceNameDnsARecord:    one.PrivateDnsNameOptions.EnableResourceNameDnsARecord,
			HostnameType:                    one.PrivateDnsNameOptions.HostnameType,
		}
	}
	return extension
}

func getIPAddress(one typescvm.AwsCvm) (privateIPv4Addresses, publicIPv4Addresses, publicIPv6Addresses []string) {
	privateIPv4Addresses = make([]string, 0)
	if one.PrivateIpAddress != nil {
		privateIPv4Addresses = append(privateIPv4Addresses, converter.PtrToVal(one.PrivateIpAddress))
	}
	publicIPv4Addresses = make([]string, 0)
	if one.PublicIpAddress != nil {
		publicIPv4Addresses = append(publicIPv4Addresses, converter.PtrToVal(one.PublicIpAddress))
	}
	publicIPv6Addresses = make([]string, 0)
	if one.Ipv6Address != nil {
		publicIPv6Addresses = append(publicIPv6Addresses, converter.PtrToVal(one.Ipv6Address))
	}
	return privateIPv4Addresses, publicIPv4Addresses, publicIPv6Addresses
}
