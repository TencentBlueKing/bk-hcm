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
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
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

	if len(cvmFromCloud) == 0 && len(cvmFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typescvm.TCloudCvm, corecvm.Cvm[cvm.TCloudCvmExtension]](
		cvmFromCloud, cvmFromDB, isCvmChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteCvm(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
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

func (cli *client) updateCvm(kt *kit.Kit, accountID string, region string,
	updateMap map[string]typescvm.TCloudCvm) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("cvm updateMap is <= 0, not update")
	}

	lists := make([]dataproto.CvmBatchUpdate[corecvm.TCloudCvmExtension], 0, len(updateMap))

	cloudVpcIDs := make([]string, 0)
	cloudSubnetIDs := make([]string, 0)
	cloudImageIDs := make([]string, 0)
	for _, one := range updateMap {
		cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(one.VirtualPrivateCloud.VpcId))
		cloudSubnetIDs = append(cloudSubnetIDs, converter.PtrToVal(one.VirtualPrivateCloud.SubnetId))
		cloudImageIDs = append(cloudImageIDs, converter.PtrToVal(one.ImageId))
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, region, cloudVpcIDs)
	if err != nil {
		return err
	}

	subnetMap, err := cli.getSubnetMap(kt, accountID, region, cloudSubnetIDs)
	if err != nil {
		return err
	}

	imageMap, err := cli.getImageMap(kt, accountID, region, cloudImageIDs)
	if err != nil {
		return err
	}

	for id, one := range updateMap {
		if _, exsit := vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)]; !exsit {
			return fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.InstanceId))
		}

		if _, exsit := subnetMap[converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)]; !exsit {
			return fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.InstanceId))
		}

		imageID := ""
		if id, exsit := imageMap[converter.PtrToVal(one.ImageId)]; exsit {
			imageID = id
		}

		extension := BuildCVMExtension(one)
		updateOne := dataproto.CvmBatchUpdate[corecvm.TCloudCvmExtension]{
			ID:             id,
			Name:           converter.PtrToVal(one.InstanceName),
			BkCloudID:      vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)].BkCloudID,
			CloudVpcIDs:    []string{converter.PtrToVal(one.VirtualPrivateCloud.VpcId)},
			VpcIDs:         []string{vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)].VpcID},
			CloudSubnetIDs: []string{converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)},
			SubnetIDs:      []string{subnetMap[converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)]},
			CloudImageID:   converter.PtrToVal(one.ImageId),
			ImageID:        imageID,
			// 备注字段云上没有，仅限hcm内部使用
			Memo:                 nil,
			Status:               converter.PtrToVal(one.InstanceState),
			PrivateIPv4Addresses: converter.PtrToSlice(one.PrivateIpAddresses),
			PublicIPv4Addresses:  converter.PtrToSlice(one.PublicIpAddresses),
			// 云上该字段没有
			CloudLaunchedTime: "",
			CloudExpiredTime:  converter.PtrToVal(one.ExpiredTime),
			Extension:         extension,
		}

		if len(one.IPv6Addresses) != 0 {
			one.PublicIpAddresses, one.PrivateIpAddresses, err = cli.cloudCli.DetermineIPv6Type(kt,
				region, one.IPv6Addresses)
			if err != nil {
				return err
			}
		}

		lists = append(lists, updateOne)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.TCloudCvmExtension]{
		Cvms: lists,
	}
	if err := cli.dbCli.TCloud.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request tcloud dataservice BatchUpdateCvm failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to update cvm success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createCvm(kt *kit.Kit, accountID string, region string,
	addSlice []typescvm.TCloudCvm) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	lists := make([]dataproto.CvmBatchCreate[corecvm.TCloudCvmExtension], 0)

	cloudVpcIDs := make([]string, 0)
	cloudSubnetIDs := make([]string, 0)
	cloudImageIDs := make([]string, 0)
	for _, one := range addSlice {
		cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(one.VirtualPrivateCloud.VpcId))
		cloudSubnetIDs = append(cloudSubnetIDs, converter.PtrToVal(one.VirtualPrivateCloud.SubnetId))
		cloudImageIDs = append(cloudImageIDs, converter.PtrToVal(one.ImageId))
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, region, cloudVpcIDs)
	if err != nil {
		return err
	}

	subnetMap, err := cli.getSubnetMap(kt, accountID, region, cloudSubnetIDs)
	if err != nil {
		return err
	}

	imageMap, err := cli.getImageMap(kt, accountID, region, cloudImageIDs)
	if err != nil {
		return err
	}

	for _, one := range addSlice {
		if _, exsit := vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)]; !exsit {
			return fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.InstanceId))
		}

		if _, exsit := subnetMap[converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)]; !exsit {
			return fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.InstanceId))
		}

		imageID := ""
		if id, exsit := imageMap[converter.PtrToVal(one.ImageId)]; exsit {
			imageID = id
		}

		extension := BuildCVMExtension(one)
		addOne := dataproto.CvmBatchCreate[corecvm.TCloudCvmExtension]{
			CloudID:        converter.PtrToVal(one.InstanceId),
			Name:           converter.PtrToVal(one.InstanceName),
			BkBizID:        constant.UnassignedBiz,
			BkCloudID:      vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)].BkCloudID,
			AccountID:      accountID,
			Region:         region,
			Zone:           converter.PtrToVal(one.Placement.Zone),
			CloudVpcIDs:    []string{converter.PtrToVal(one.VirtualPrivateCloud.VpcId)},
			VpcIDs:         []string{vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)].VpcID},
			CloudSubnetIDs: []string{converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)},
			SubnetIDs:      []string{subnetMap[converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)]},
			CloudImageID:   converter.PtrToVal(one.ImageId),
			ImageID:        imageID,
			OsName:         converter.PtrToVal(one.OsName),
			// 备注字段云上没有，仅限hcm内部使用
			Memo:                 nil,
			Status:               converter.PtrToVal(one.InstanceState),
			PrivateIPv4Addresses: converter.PtrToSlice(one.PrivateIpAddresses),
			PublicIPv4Addresses:  converter.PtrToSlice(one.PublicIpAddresses),
			MachineType:          *one.InstanceType,
			CloudCreatedTime:     *one.CreatedTime,
			// 该字段云上没有
			CloudLaunchedTime: "",
			CloudExpiredTime:  converter.PtrToVal(one.ExpiredTime),
			Extension:         extension,
		}

		if len(one.IPv6Addresses) != 0 {
			one.PublicIpAddresses, one.PrivateIpAddresses, err = cli.cloudCli.DetermineIPv6Type(kt,
				region, one.IPv6Addresses)
			if err != nil {
				return err
			}
		}

		lists = append(lists, addOne)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.TCloudCvmExtension]{
		Cvms: lists,
	}

	_, err = cli.dbCli.TCloud.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud cvm failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to create cvm success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(addSlice), kt.Rid)

	return nil
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
				VpcID:     vpc.ID,
				BkCloudID: vpc.BkCloudID,
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
			enumor.TCloud, checkParams, len(delCvmFromCloud), kt.Rid)
		return fmt.Errorf("validate cvm not exist failed, before delete")
	}

	deleteReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete cvm failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to delete cvm success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listCvmFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typescvm.TCloudCvm, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typescvm.TCloudListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListCvm(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list cvm from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listCvmFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]corecvm.Cvm[cvm.TCloudCvmExtension], error) {

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
	result, err := cli.dbCli.TCloud.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list cvm from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.TCloud,
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
		resultFromDB, err := cli.dbCli.TCloud.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list cvm failed, err: %v, req: %v, rid: %s", enumor.TCloud,
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
			if err := cli.deleteCvm(kt, accountID, region, cloudIDs); err != nil {
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

// BuildCVMExtension build extension
func BuildCVMExtension(one typescvm.TCloudCvm) *corecvm.TCloudCvmExtension {
	dataDiskIDs := make([]string, 0)
	for _, disk := range one.DataDisks {
		dataDiskIDs = append(dataDiskIDs, *disk.DiskId)
	}

	extension := &corecvm.TCloudCvmExtension{
		Placement: &corecvm.TCloudPlacement{
			CloudProjectID: one.Placement.ProjectId,
		},
		InstanceChargeType:    one.InstanceChargeType,
		Cpu:                   one.CPU,
		Memory:                one.Memory,
		CloudDataDiskIDs:      dataDiskIDs,
		RenewFlag:             one.RenewFlag,
		CloudSecurityGroupIDs: converter.PtrToSlice(one.SecurityGroupIds),
		StopChargingMode:      one.StopChargingMode,
		UUID:                  one.Uuid,
		IsolatedSource:        one.IsolatedSource,
		DisableApiTermination: one.DisableApiTermination,
	}

	if one.SystemDisk != nil {
		extension.CloudSystemDiskID = one.SystemDisk.DiskId
	}

	if one.InternetAccessible != nil {
		extension.InternetAccessible = &corecvm.TCloudInternetAccessible{
			InternetChargeType:      one.InternetAccessible.InternetChargeType,
			InternetMaxBandwidthOut: one.InternetAccessible.InternetMaxBandwidthOut,
			PublicIPAssigned:        one.InternetAccessible.PublicIpAssigned,
			CloudBandwidthPackageID: one.InternetAccessible.BandwidthPackageId,
		}
	}

	if one.VirtualPrivateCloud != nil {
		extension.VirtualPrivateCloud = &corecvm.TCloudVirtualPrivateCloud{
			AsVpcGateway: one.VirtualPrivateCloud.AsVpcGateway,
		}
	}

	return extension
}

func isCvmChange(cloud typescvm.TCloudCvm, db corecvm.Cvm[cvm.TCloudCvmExtension]) bool {

	if db.CloudID != converter.PtrToVal(cloud.InstanceId) {
		return true
	}

	if db.Name != converter.PtrToVal(cloud.InstanceName) {
		return true
	}

	if len(db.CloudVpcIDs) == 0 || (db.CloudVpcIDs[0] !=
		converter.PtrToVal(cloud.VirtualPrivateCloud.VpcId)) {
		return true
	}

	if len(db.CloudSubnetIDs) == 0 || (db.CloudSubnetIDs[0] != converter.PtrToVal(cloud.VirtualPrivateCloud.SubnetId)) {
		return true
	}

	if db.CloudImageID != converter.PtrToVal(cloud.ImageId) {
		return true
	}

	if db.OsName != converter.PtrToVal(cloud.OsName) {
		return true
	}

	if db.Status != converter.PtrToVal(cloud.InstanceState) {
		return true
	}

	if len(cloud.IPv6Addresses) == 0 && (len(db.PublicIPv6Addresses) != 0 || len(db.PrivateIPv6Addresses) != 0) {
		return true
	}

	if len(cloud.IPv6Addresses) != 0 {
		if len(db.PublicIPv6Addresses) == 0 && len(db.PrivateIPv6Addresses) == 0 {
			return true
		}

		tmpMap := converter.StringSliceToMap(converter.PtrToSlice(cloud.IPv6Addresses))
		for _, address := range db.PublicIPv6Addresses {
			delete(tmpMap, address)
		}

		for _, address := range db.PrivateIPv6Addresses {
			delete(tmpMap, address)
		}

		if len(tmpMap) != 0 {
			return true
		}
	}

	if !assert.IsStringSliceEqual(converter.PtrToSlice(cloud.PrivateIpAddresses), db.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(converter.PtrToSlice(cloud.PublicIpAddresses), db.PublicIPv4Addresses) {
		return true
	}

	if db.MachineType != converter.PtrToVal(cloud.InstanceType) {
		return true
	}

	if db.CloudCreatedTime != converter.PtrToVal(cloud.CreatedTime) {
		return true
	}

	if db.CloudExpiredTime != converter.PtrToVal(cloud.ExpiredTime) {
		return true
	}

	cloudExt := BuildCVMExtension(cloud)
	return IsCvmExtensionChange(cloudExt, db.Extension)
}

// IsCvmExtensionChange check if the extension of cloudCvm is changed
func IsCvmExtensionChange(cloud *corecvm.TCloudCvmExtension, db *corecvm.TCloudCvmExtension) bool {
	if cloud == nil || db == nil {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Placement.CloudProjectID, cloud.Placement.CloudProjectID) {
		return true
	}

	if !assert.IsPtrStringEqual(db.InstanceChargeType, cloud.InstanceChargeType) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Cpu, cloud.Cpu) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Memory, cloud.Memory) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudSystemDiskID, db.CloudSystemDiskID) {
		return true
	}

	if !assert.IsStringSliceEqual(cloud.CloudDataDiskIDs, db.CloudDataDiskIDs) {
		return true
	}

	if changed := isInternetInternetAccessibleChanged(cloud.InternetAccessible, db.InternetAccessible); changed {
		return true
	}

	if !assert.IsPtrBoolEqual(db.VirtualPrivateCloud.AsVpcGateway, cloud.VirtualPrivateCloud.AsVpcGateway) {
		return true
	}

	if !assert.IsPtrStringEqual(db.RenewFlag, cloud.RenewFlag) {
		return true
	}

	if !assert.IsStringSliceEqual(db.CloudSecurityGroupIDs, cloud.CloudSecurityGroupIDs) {
		return true
	}

	if !assert.IsPtrStringEqual(db.StopChargingMode, cloud.StopChargingMode) {
		return true
	}

	if !assert.IsPtrStringEqual(db.UUID, cloud.UUID) {
		return true
	}

	if !assert.IsPtrStringEqual(db.IsolatedSource, cloud.IsolatedSource) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.DisableApiTermination, cloud.DisableApiTermination) {
		return true
	}

	return false
}

func isInternetInternetAccessibleChanged(cloud *corecvm.TCloudInternetAccessible,
	db *corecvm.TCloudInternetAccessible) bool {

	if db == nil || cloud == nil {
		return true
	}

	if !assert.IsPtrStringEqual(db.InternetChargeType, cloud.InternetChargeType) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.InternetMaxBandwidthOut, cloud.InternetMaxBandwidthOut) {
		return true
	}

	if !assert.IsPtrStringEqual(db.CloudBandwidthPackageID, cloud.CloudBandwidthPackageID) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.PublicIPAssigned, cloud.PublicIPAssigned) {
		return true
	}

	// CVM查询接口 不返回带宽包id，这里无法比较
	// if !assert.IsPtrStringEqual(db.CloudBandwidthPackageID, cloud.BandwidthPackageId) {
	// 	return true
	// }
	return false
}
