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
	typescore "hcm/pkg/adaptor/types/core"
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

	if len(cvmFromCloud) == 0 && len(cvmFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typescvm.AzureCvm, corecvm.Cvm[cvm.AzureCvmExtension]](
		cvmFromCloud, cvmFromDB, isCvmChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteCvm(kt, params.AccountID, params.ResourceGroupName, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createCvm(kt, params.AccountID, params.ResourceGroupName, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateCvm(kt, params.AccountID, params.ResourceGroupName, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) listCvmFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typescvm.AzureCvm, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typescore.AzureListByIDOption{
		ResourceGroupName: params.ResourceGroupName,
		CloudIDs:          params.CloudIDs,
	}
	result, err := cli.cloudCli.ListCvmByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list cvm from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	cvms := make([]typescvm.AzureCvm, 0, len(result))
	for _, one := range result {
		cvms = append(cvms, converter.PtrToVal(one))
	}

	return cvms, nil
}

func (cli *client) listCvmFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]corecvm.Cvm[cvm.AzureCvmExtension], error) {

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
					Field: "extension.resource_group_name",
					Op:    filter.JSONEqual.Factory(),
					Value: params.ResourceGroupName,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list cvm from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) createCvm(kt *kit.Kit, accountID string, resGroupName string,
	addSlice []typescvm.AzureCvm) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	lists := make([]dataproto.CvmBatchCreate[corecvm.AzureCvmExtension], 0)

	niMap := make(map[string]string)
	niIDs := make([]string, 0)
	cloudImageIDs := make([]string, 0)
	for _, one := range addSlice {
		niIDs = append(niIDs, one.NetworkInterfaceIDs...)
		for _, niID := range one.NetworkInterfaceIDs {
			niMap[niID] = converter.PtrToVal(one.ID)
		}
		cloudImageIDs = append(cloudImageIDs, converter.PtrToVal(one.CloudImageID))
	}

	cloudMap, cloudVpcIDsMap, cloudSubnetIDsMap, err := cli.getNIAssResMapFromNI(kt, niIDs, resGroupName, niMap)
	if err != nil {
		return err
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, cloudVpcIDsMap)
	if err != nil {
		return err
	}

	subnetMap, err := cli.getSubnetMap(kt, accountID, cloudSubnetIDsMap)
	if err != nil {
		return err
	}

	imageMap, err := cli.getImageMap(kt, accountID, resGroupName, cloudImageIDs)
	if err != nil {
		return err
	}

	for _, one := range addSlice {
		if _, exsit := vpcMap[converter.PtrToVal(one.ID)]; !exsit {
			return fmt.Errorf("cvm: %s not found vpc", converter.PtrToVal(one.ID))
		}

		if _, exsit := subnetMap[converter.PtrToVal(one.ID)]; !exsit {
			return fmt.Errorf("cvm: %s not found subnets", converter.PtrToVal(one.ID))
		}

		imageID := ""
		if id, exsit := imageMap[converter.PtrToVal(one.CloudImageID)]; exsit {
			imageID = id
		}

		cvm := dataproto.CvmBatchCreate[corecvm.AzureCvmExtension]{
			CloudID:   converter.PtrToVal(one.ID),
			Name:      converter.PtrToVal(one.Name),
			BkBizID:   constant.UnassignedBiz,
			BkHostID:  constant.UnBindBkHostID,
			BkCloudID: constant.UnassignedBkCloudID,
			AccountID: accountID,
			Region:    converter.PtrToVal(one.Location),
			// 云上不支持该字段，azure可用区非地域概念
			Zone:           "",
			CloudVpcIDs:    []string{vpcMap[converter.PtrToVal(one.ID)].VpcCloudID},
			VpcIDs:         []string{vpcMap[converter.PtrToVal(one.ID)].VpcID},
			CloudSubnetIDs: cloudMap[converter.PtrToVal(one.ID)].CloudSubnetIDs,
			SubnetIDs:      subnetMap[converter.PtrToVal(one.ID)],
			CloudImageID:   converter.PtrToVal(one.CloudImageID),
			ImageID:        imageID,
			OsName:         converter.PtrToVal(one.ComputerName),
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(one.Status),
			PrivateIPv4Addresses: cloudMap[converter.PtrToVal(one.ID)].PrivateIPv4Addresses,
			PrivateIPv6Addresses: cloudMap[converter.PtrToVal(one.ID)].PrivateIPv6Addresses,
			PublicIPv4Addresses:  cloudMap[converter.PtrToVal(one.ID)].PublicIPv4Addresses,
			PublicIPv6Addresses:  cloudMap[converter.PtrToVal(one.ID)].PublicIPv6Addresses,
			MachineType:          string(converter.PtrToVal(one.VMSize)),
			CloudCreatedTime:     times.ConvStdTimeFormat(converter.PtrToVal(one.TimeCreated)),
			// 云上不支持该字段
			CloudLaunchedTime: "",
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AzureCvmExtension{
				ResourceGroupName: resGroupName,
				AdditionalCapabilities: &corecvm.AzureAdditionalCapabilities{
					HibernationEnabled: one.HibernationEnabled,
					UltraSSDEnabled:    one.UltraSSDEnabled,
				},
				BillingProfile: &corecvm.AzureBillingProfile{
					MaxPrice: one.MaxPrice,
				},
				EvictionPolicy: (*string)(one.EvictionPolicy),
				HardwareProfile: &corecvm.AzureHardwareProfile{
					VmSize: (*string)(one.VMSize),
					VmSizeProperties: &corecvm.AzureVmSizeProperties{
						VCPUsAvailable: one.VCPUsAvailable,
						VCPUsPerCore:   one.VCPUsPerCore,
					},
				},
				LicenseType:              one.LicenseType,
				CloudNetworkInterfaceIDs: one.NetworkInterfaceIDs,
				Priority:                 (*string)(one.Priority),
				StorageProfile: &corecvm.AzureStorageProfile{
					CloudDataDiskIDs: one.CloudDataDiskIDs,
					CloudOsDiskID:    one.CloudOsDiskID,
				},
				Zones: converter.PtrToSlice(one.Zones),
			},
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.AzureCvmExtension]{
		Cvms: lists,
	}
	_, err = cli.dbCli.Azure.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create azure cvm failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to create cvm success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateCvm(kt *kit.Kit, accountID string, resGroupName string,
	updateMap map[string]typescvm.AzureCvm) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("cvm updateMap is <= 0, not update")
	}

	lists := make([]dataproto.CvmBatchUpdate[corecvm.AzureCvmExtension], 0)

	niMap := make(map[string]string)
	niIDs := make([]string, 0)
	cloudImageIDs := make([]string, 0)
	for _, one := range updateMap {
		niIDs = append(niIDs, one.NetworkInterfaceIDs...)
		for _, niID := range one.NetworkInterfaceIDs {
			niMap[niID] = converter.PtrToVal(one.ID)
		}
		cloudImageIDs = append(cloudImageIDs, converter.PtrToVal(one.CloudImageID))
	}

	cloudMap, cloudVpcIDsMap, cloudSubnetIDsMap, err := cli.getNIAssResMapFromNI(kt, niIDs, resGroupName, niMap)
	if err != nil {
		return err
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, cloudVpcIDsMap)
	if err != nil {
		return err
	}

	subnetMap, err := cli.getSubnetMap(kt, accountID, cloudSubnetIDsMap)
	if err != nil {
		return err
	}

	imageMap, err := cli.getImageMap(kt, accountID, resGroupName, cloudImageIDs)
	if err != nil {
		return err
	}

	for id, one := range updateMap {
		if _, exsit := vpcMap[converter.PtrToVal(one.ID)]; !exsit {
			return fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.ID))
		}

		if _, exsit := subnetMap[converter.PtrToVal(one.ID)]; !exsit {
			return fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.ID))
		}

		imageID := ""
		if id, exsit := imageMap[converter.PtrToVal(one.CloudImageID)]; exsit {
			imageID = id
		}

		cvm := dataproto.CvmBatchUpdate[corecvm.AzureCvmExtension]{
			ID:             id,
			Name:           converter.PtrToVal(one.Name),
			CloudVpcIDs:    []string{vpcMap[converter.PtrToVal(one.ID)].VpcCloudID},
			VpcIDs:         []string{vpcMap[converter.PtrToVal(one.ID)].VpcID},
			CloudSubnetIDs: cloudMap[converter.PtrToVal(one.ID)].CloudSubnetIDs,
			SubnetIDs:      subnetMap[converter.PtrToVal(one.ID)],
			CloudImageID:   converter.PtrToVal(one.CloudImageID),
			ImageID:        imageID,
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(one.Status),
			PrivateIPv4Addresses: cloudMap[converter.PtrToVal(one.ID)].PrivateIPv4Addresses,
			PrivateIPv6Addresses: cloudMap[converter.PtrToVal(one.ID)].PrivateIPv6Addresses,
			PublicIPv4Addresses:  cloudMap[converter.PtrToVal(one.ID)].PublicIPv4Addresses,
			PublicIPv6Addresses:  cloudMap[converter.PtrToVal(one.ID)].PublicIPv6Addresses,
			// 云上不支持该字段
			CloudLaunchedTime: "",
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AzureCvmExtension{
				ResourceGroupName: resGroupName,
				AdditionalCapabilities: &corecvm.AzureAdditionalCapabilities{
					HibernationEnabled: one.HibernationEnabled,
					UltraSSDEnabled:    one.UltraSSDEnabled,
				},
				BillingProfile: &corecvm.AzureBillingProfile{
					MaxPrice: one.MaxPrice,
				},
				EvictionPolicy: (*string)(one.EvictionPolicy),
				HardwareProfile: &corecvm.AzureHardwareProfile{
					VmSize: (*string)(one.VMSize),
					VmSizeProperties: &corecvm.AzureVmSizeProperties{
						VCPUsAvailable: one.VCPUsAvailable,
						VCPUsPerCore:   one.VCPUsPerCore,
					},
				},
				LicenseType:              one.LicenseType,
				CloudNetworkInterfaceIDs: one.NetworkInterfaceIDs,
				Priority:                 (*string)(one.Priority),
				StorageProfile: &corecvm.AzureStorageProfile{
					CloudDataDiskIDs: one.CloudDataDiskIDs,
					CloudOsDiskID:    one.CloudOsDiskID,
				},
				Zones: converter.PtrToSlice(one.Zones),
			},
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.AzureCvmExtension]{
		Cvms: lists,
	}
	if err := cli.dbCli.Azure.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateCvm failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to update cvm success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

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

func (cli *client) getNIAssResMapFromNI(kt *kit.Kit, niIDs []string, resGroupName string,
	niMap map[string]string) (map[string]*CloudData, map[string]string, map[string]string, error) {

	netInterOpt := &typescore.AzureListByIDOption{
		ResourceGroupName: resGroupName,
		CloudIDs:          niIDs,
	}
	netInterDatas, err := cli.cloudCli.ListNetworkInterfaceByID(kt, netInterOpt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list azure net interface failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return nil, nil, nil, err
	}

	cloudMap := make(map[string]*CloudData)
	cloudVpcIDsMap := make(map[string]string, 0)
	cloudSubnetIDsMap := make(map[string]string, 0)
	for _, niData := range netInterDatas.Details {
		if cvmID, exsit := niMap[converter.PtrToVal(niData.CloudID)]; exsit {
			cloudMap[cvmID] = new(CloudData)
			if niData.CloudSubnetID != nil {
				cloudMap[cvmID].CloudSubnetIDs = make([]string, 0)
				cloudMap[cvmID].CloudSubnetIDs = append(cloudMap[cvmID].CloudSubnetIDs,
					converter.PtrToVal(niData.CloudSubnetID))
				cloudSubnetIDsMap[cvmID] = converter.PtrToVal(niData.CloudSubnetID)
			}
			if niData.CloudVpcID != nil {
				cloudVpcIDsMap[cvmID] = converter.PtrToVal(niData.CloudVpcID)
			}
			cloudMap[cvmID].PrivateIPv4Addresses = make([]string, 0)
			cloudMap[cvmID].PrivateIPv4Addresses = append(cloudMap[cvmID].PrivateIPv4Addresses, niData.PrivateIPv4...)
			cloudMap[cvmID].PrivateIPv6Addresses = make([]string, 0)
			cloudMap[cvmID].PrivateIPv6Addresses = append(cloudMap[cvmID].PrivateIPv6Addresses, niData.PrivateIPv6...)
			cloudMap[cvmID].PublicIPv4Addresses = make([]string, 0)
			cloudMap[cvmID].PublicIPv4Addresses = append(cloudMap[cvmID].PublicIPv4Addresses, niData.PublicIPv4...)
			cloudMap[cvmID].PublicIPv6Addresses = make([]string, 0)
			cloudMap[cvmID].PublicIPv6Addresses = append(cloudMap[cvmID].PublicIPv6Addresses, niData.PublicIPv6...)
		}
	}

	return cloudMap, cloudVpcIDsMap, cloudSubnetIDsMap, nil
}

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

func (cli *client) deleteCvm(kt *kit.Kit, accountID string, resGroupName string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("cvm delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          delCloudIDs,
	}
	delCvmFromCloud, err := cli.listCvmFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delCvmFromCloud) > 0 {
		logs.Errorf("[%s] validate cvm not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delCvmFromCloud), kt.Rid)
		return fmt.Errorf("validate cvm not exist failed, before delete")
	}

	deleteReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete cvm failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to delete cvm success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) RemoveCvmDeleteFromCloud(kt *kit.Kit, accountID string, resGroupName string) error {
	req := &protocloud.CvmListReq{
		Field: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(),
					Value: resGroupName},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Azure.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list cvm failed, err: %v, req: %v, rid: %s", enumor.Azure,
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
			AccountID:         accountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		resultFromCloud, err := cli.listCvmFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.ID))
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteCvm(kt, accountID, resGroupName, cloudIDs); err != nil {
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

func isCvmChange(cloud typescvm.AzureCvm, db corecvm.Cvm[cvm.AzureCvmExtension]) bool {

	if db.CloudID != converter.PtrToVal(cloud.ID) {
		return true
	}

	if db.Name != converter.PtrToVal(cloud.Name) {
		return true
	}

	if db.CloudImageID != converter.PtrToVal(cloud.CloudImageID) {
		return true
	}

	if db.OsName != converter.PtrToVal(cloud.ComputerName) {
		return true
	}

	if db.Status != converter.PtrToVal(cloud.Status) {
		return true
	}

	if db.MachineType != string(converter.PtrToVal(cloud.VMSize)) {
		return true
	}

	if db.Extension.AdditionalCapabilities != nil {
		if !assert.IsPtrBoolEqual(db.Extension.AdditionalCapabilities.HibernationEnabled,
			cloud.HibernationEnabled) {
			return true
		}

		if !assert.IsPtrBoolEqual(db.Extension.AdditionalCapabilities.UltraSSDEnabled,
			cloud.UltraSSDEnabled) {
			return true
		}
	}

	if db.Extension.BillingProfile != nil {
		if !assert.IsPtrFloat64Equal(db.Extension.BillingProfile.MaxPrice,
			cloud.MaxPrice) {
			return true
		}
	}

	if !assert.IsPtrStringEqual(db.Extension.EvictionPolicy,
		(*string)(cloud.EvictionPolicy)) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.LicenseType, cloud.LicenseType) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.Priority,
		(*string)(cloud.Priority)) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.Zones, converter.PtrToSlice(cloud.Zones)) {
		return true
	}

	if db.Extension.StorageProfile.CloudOsDiskID != cloud.CloudOsDiskID {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.StorageProfile.CloudDataDiskIDs, cloud.CloudDataDiskIDs) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.HardwareProfile.VmSize,
		(*string)(cloud.VMSize)) {
		return true
	}

	if db.Extension.HardwareProfile.VmSizeProperties != nil {
		if !assert.IsPtrInt32Equal(db.Extension.HardwareProfile.VmSizeProperties.VCPUsAvailable,
			cloud.VCPUsAvailable) {
			return true
		}

		if !assert.IsPtrInt32Equal(db.Extension.HardwareProfile.VmSizeProperties.VCPUsPerCore,
			cloud.VCPUsPerCore) {
			return true
		}
	}

	return false
}
