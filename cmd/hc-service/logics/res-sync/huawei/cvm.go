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

package huawei

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/huawei"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	typescvm "hcm/pkg/adaptor/types/cvm"
	networkinterface "hcm/pkg/adaptor/types/network-interface"
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

// Cvm Huawei Cloud Cvm sync
func (cli *client) Cvm(kt *kit.Kit, params *SyncBaseParams, opt *SyncCvmOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmFromCloudOri, err := cli.listCvmFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	cvmFromDB, err := cli.listCvmFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(cvmFromCloudOri) == 0 && len(cvmFromDB) == 0 {
		return new(SyncResult), nil
	}
	cvmFromCloud, err := cli.wrapCvm(kt, params, cvmFromCloudOri)
	if err != nil {
		logs.Errorf("preprocess HuaweiCvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typescvm.HuaWeiCvmWrapper, corecvm.Cvm[cvm.HuaWeiCvmExtension]](
		cvmFromCloud, cvmFromDB, cli.isCvmChange)

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

func (cli *client) wrapCvm(kt *kit.Kit, params *SyncBaseParams, cvmList []typescvm.HuaWeiCvm) (
	[]typescvm.HuaWeiCvmWrapper, error) {

	cvmWrappers := make([]typescvm.HuaWeiCvmWrapper, 0)
	for _, cloudCvm := range cvmList {
		wrapped := typescvm.HuaWeiCvmWrapper{HuaWeiCvm: cloudCvm}
		// 查询对应子网信息
		cloudSubnetIDs, localSubnetIDs, err := cli.getSubnets(kt, params.AccountID, params.Region,
			cloudCvm.GetCloudID(), cloudCvm.GetCloudVpcID())
		if err != nil {
			logs.Errorf("[%s] get subnets failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
			return nil, err
		}
		wrapped.CloudSubnetIDs = cloudSubnetIDs
		wrapped.SubnetIDs = localSubnetIDs

		cvmWrappers = append(cvmWrappers, wrapped)
	}
	return cvmWrappers, nil
}

func (cli *client) updateCvm(kt *kit.Kit, accountID string, region string,
	updateMap map[string]typescvm.HuaWeiCvmWrapper) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("cvm updateMap is <= 0, not update")
	}

	lists := make([]dataproto.CvmBatchUpdate[corecvm.HuaWeiCvmExtension], 0)

	cloudVpcIDs := make([]string, 0)
	cloudImageIDs := make([]string, 0)
	for _, one := range updateMap {
		cloudVpcIDs = append(cloudVpcIDs, one.Metadata["vpc_id"])
		cloudImageIDs = append(cloudImageIDs, one.Image.Id)
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, region, cloudVpcIDs)
	if err != nil {
		return err
	}

	imageMap, err := cli.getImageMap(kt, accountID, region, cloudImageIDs)
	if err != nil {
		return err
	}

	for id, one := range updateMap {
		if _, exist := vpcMap[one.Metadata["vpc_id"]]; !exist {
			return fmt.Errorf("cvm %s can not find vpc", one.Id)
		}

		imageID := ""
		if id, exsit := imageMap[one.Image.Id]; exsit {
			imageID = id
		}

		cvm := dataproto.CvmBatchUpdate[corecvm.HuaWeiCvmExtension]{
			ID:                   id,
			Name:                 one.Name,
			CloudVpcIDs:          []string{one.Metadata["vpc_id"]},
			VpcIDs:               []string{vpcMap[one.Metadata["vpc_id"]].VpcID},
			CloudSubnetIDs:       one.CloudSubnetIDs,
			SubnetIDs:            one.SubnetIDs,
			CloudImageID:         one.Image.Id,
			ImageID:              imageID,
			Memo:                 one.Description,
			Status:               one.Status,
			PrivateIPv4Addresses: one.PrivateIPv4Addresses,
			PrivateIPv6Addresses: one.PrivateIPv6Addresses,
			PublicIPv4Addresses:  one.PublicIPv4Addresses,
			PublicIPv6Addresses:  one.PublicIPv6Addresses,
			CloudLaunchedTime:    one.CloudLaunchedTime,
			CloudExpiredTime:     one.AutoTerminateTime,
			Extension:            convHuaweiCvmExtension(one),
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.HuaWeiCvmExtension]{Cvms: lists}
	if err := cli.dbCli.HuaWei.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateCvm failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to update cvm success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func convHuaweiCvmExtension(one typescvm.HuaWeiCvmWrapper) *corecvm.HuaWeiCvmExtension {

	sgIDs := make([]string, 0)
	for _, v := range one.SecurityGroups {
		sgIDs = append(sgIDs, v.Id)
	}

	ext := &corecvm.HuaWeiCvmExtension{
		AliasName:             one.OSEXTSRVATTRinstanceName,
		HypervisorHostname:    one.OSEXTSRVATTRhypervisorHostname,
		Flavor:                one.Flavor,
		CloudSecurityGroupIDs: sgIDs,
		CloudTenantID:         one.TenantId,
		DiskConfig:            one.OSDCFdiskConfig,
		PowerState:            one.OSEXTSTSpowerState,
		ConfigDrive:           one.ConfigDrive,
		Metadata: &corecvm.HuaWeiMetadata{
			ChargingMode:      one.Metadata["charging_mode"],
			CloudOrderID:      one.Metadata["metering.order_id"],
			CloudProductID:    one.Metadata["metering.product_id"],
			EcmResStatus:      one.Metadata["EcmResStatus"],
			ImageType:         one.Metadata["metering.imagetype"],
			ResourceSpecCode:  one.Metadata["metering.resourcespeccode"],
			ResourceType:      one.Metadata["metering.resourcetype"],
			InstanceExtraInfo: one.Metadata["cascaded.instance_extrainfo"],
			ImageName:         one.Metadata["image_name"],
			AgencyName:        one.Metadata["agency_name"],
			OSBit:             one.Metadata["os_bit"],
			SupportAgentList:  one.Metadata["__support_agent_list"],
		},
		CloudOSVolumeID:          one.CloudOSDiskID,
		CloudDataVolumeIDs:       one.CLoudDataDiskIDs,
		RootDeviceName:           one.OSEXTSRVATTRrootDeviceName,
		CloudEnterpriseProjectID: one.EnterpriseProjectId,
		CpuOptions:               nil,
	}
	if one.CpuOptions != nil {
		ext.CpuOptions = &corecvm.HuaWeiCpuOptions{
			CpuThreads: one.CpuOptions.HwcpuThreads,
		}
	}
	return ext
}

func (cli *client) createCvm(kt *kit.Kit, accountID string, region string, addSlice []typescvm.HuaWeiCvmWrapper) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	lists := make([]dataproto.CvmBatchCreate[corecvm.HuaWeiCvmExtension], 0)

	cloudVpcIDs := make([]string, 0)
	cloudImageIDs := make([]string, 0)
	for _, one := range addSlice {
		cloudVpcIDs = append(cloudVpcIDs, one.Metadata["vpc_id"])
		cloudImageIDs = append(cloudImageIDs, one.Image.Id)
	}

	vpcMap, err := cli.getVpcMap(kt, accountID, region, cloudVpcIDs)
	if err != nil {
		return err
	}

	imageMap, err := cli.getImageMap(kt, accountID, region, cloudImageIDs)
	if err != nil {
		return err
	}

	for _, one := range addSlice {
		if _, exist := vpcMap[one.Metadata["vpc_id"]]; !exist {
			return fmt.Errorf("cvm %s can not find vpc", one.Id)
		}

		imageID := ""
		if id, exsit := imageMap[one.Image.Id]; exsit {
			imageID = id
		}

		cvm := dataproto.CvmBatchCreate[corecvm.HuaWeiCvmExtension]{
			CloudID:              one.Id,
			Name:                 one.Name,
			BkBizID:              constant.UnassignedBiz,
			BkHostID:             constant.UnBindBkHostID,
			BkCloudID:            constant.UnassignedBkCloudID,
			AccountID:            accountID,
			Region:               region,
			Zone:                 one.OSEXTAZavailabilityZone,
			CloudVpcIDs:          []string{one.Metadata["vpc_id"]},
			VpcIDs:               []string{vpcMap[one.Metadata["vpc_id"]].VpcID},
			CloudSubnetIDs:       one.CloudSubnetIDs,
			SubnetIDs:            one.SubnetIDs,
			CloudImageID:         one.Image.Id,
			ImageID:              imageID,
			OsName:               one.Metadata["os_type"],
			Memo:                 one.Description,
			Status:               one.Status,
			PrivateIPv4Addresses: one.PrivateIPv4Addresses,
			PrivateIPv6Addresses: one.PrivateIPv6Addresses,
			PublicIPv4Addresses:  one.PublicIPv4Addresses,
			PublicIPv6Addresses:  one.PublicIPv6Addresses,
			MachineType:          one.Flavor.CloudID,
			CloudCreatedTime:     one.Created,
			CloudLaunchedTime:    one.CloudLaunchedTime,
			CloudExpiredTime:     one.AutoTerminateTime,
			Extension:            convHuaweiCvmExtension(one),
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.HuaWeiCvmExtension]{Cvms: lists}

	_, err = cli.dbCli.HuaWei.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create huawei cvm failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to create cvm success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
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
				VpcID: vpc.ID,
			}
		}
	}

	return vpcMap, nil
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

func (cli *client) getSubnets(kt *kit.Kit, accountID, region, serverID string,
	cloudVpcID string) ([]string, []string, error) {

	opt := &networkinterface.HuaWeiNIListOption{
		Region:   region,
		ServerID: serverID,
	}
	netInterDatas, err := cli.cloudCli.ListNetworkInterface(kt, opt)
	if err != nil {
		logs.Errorf("[%s] request adaptor to list huawei network interface failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return make([]string, 0), make([]string, 0), err
	}

	cloudSubnetIDs := make([]string, 0)
	for _, v := range netInterDatas.Details {
		if v.CloudSubnetID != nil {
			cloudSubnetIDs = append(cloudSubnetIDs, *v.CloudSubnetID)
		}
	}

	subnetParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  cloudSubnetIDs,
	}
	subnetFromDB, err := cli.listSubnetFromDB(kt, subnetParams, cloudVpcID)
	if err != nil {
		return make([]string, 0), make([]string, 0), err
	}

	subnetIDs := make([]string, 0)
	for _, subnet := range subnetFromDB {
		subnetIDs = append(subnetIDs, subnet.ID)
	}

	return cloudSubnetIDs, subnetIDs, nil
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
	tmps, err := cli.listCvmFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	delCvmFromCloud := make([]typescvm.HuaWeiCvm, 0)
	for _, tmp := range tmps {
		if tmp.Status != huawei.CvmDeleteStatus {
			delCvmFromCloud = append(delCvmFromCloud, tmp)
		}
	}

	if len(delCvmFromCloud) > 0 {
		logs.Errorf("[%s] validate cvm not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.HuaWei, checkParams, len(delCvmFromCloud), kt.Rid)
		return fmt.Errorf("validate cvm not exist failed, before delete")
	}

	deleteReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete cvm failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to delete cvm success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listCvmFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typescvm.HuaWeiCvm, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typecvm.HuaWeiListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &typecore.HuaWeiCvmOffsetPage{
			Offset: 1,
			Limit:  int32(constant.CloudResourceSyncMaxLimit),
		},
	}
	result, err := cli.cloudCli.ListCvm(kt, opt)
	if err != nil {
		if strings.Contains(err.Error(), huawei.ErrDataNotFound) {
			return nil, nil
		}

		logs.Errorf("[%s] list cvm from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listCvmFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]corecvm.Cvm[cvm.HuaWeiCvmExtension], error) {

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
	result, err := cli.dbCli.HuaWei.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list cvm from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// RemoveCvmDeleteFromCloud 删除从云上删除已经删除的cvm
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
		resultFromDB, err := cli.dbCli.HuaWei.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
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

		var resultFromCloud []typescvm.HuaWeiCvm
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			tmps, err := cli.listCvmFromCloud(kt, params)
			if err != nil {
				return err
			}

			for _, tmp := range tmps {
				// 过滤掉删除的主机。
				if tmp.Status != huawei.CvmDeleteStatus {
					resultFromCloud = append(resultFromCloud, tmp)
				}
			}
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.Id)
			}

			delCloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(delCloudIDs) > 0 {
				if err = cli.deleteCvm(kt, accountID, region, delCloudIDs); err != nil {
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

func (cli *client) isCvmChange(cloud typescvm.HuaWeiCvmWrapper, db corecvm.Cvm[cvm.HuaWeiCvmExtension]) bool {
	if db.CloudID != cloud.Id {
		return true
	}

	if db.Name != cloud.Name {
		return true
	}

	if db.CloudImageID != cloud.Image.Id {
		return true
	}

	if db.OsName != cloud.Metadata["os_type"] {
		return true
	}

	if db.Status != cloud.Status {
		return true
	}

	if !assert.IsStringSliceEqual(db.CloudSubnetIDs, cloud.CloudSubnetIDs) {
		return true
	}

	if !assert.IsStringSliceEqual(db.CloudVpcIDs, []string{cloud.GetCloudVpcID()}) {
		return true
	}

	if cli.isIPAddressChange(cloud, db) {
		return true
	}

	if db.CloudLaunchedTime != cloud.CloudLaunchedTime {
		return true
	}

	if db.MachineType != cloud.Flavor.CloudID {
		return true
	}

	if db.CloudCreatedTime != cloud.Created {
		return true
	}

	if db.CloudExpiredTime != cloud.AutoTerminateTime {
		return true
	}

	if cli.isCvmExtensionChange(&cloud.HuaWeiCvm, db.Extension) {
		return true
	}

	return false
}

func (cli *client) isIPAddressChange(cloud typescvm.HuaWeiCvmWrapper, db corecvm.Cvm[corecvm.HuaWeiCvmExtension]) bool {

	if !assert.IsStringSliceEqual(cloud.PrivateIPv4Addresses, db.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(cloud.PublicIPv4Addresses, db.PublicIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(cloud.PrivateIPv6Addresses, db.PrivateIPv6Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(cloud.PublicIPv6Addresses, db.PublicIPv6Addresses) {
		return true
	}
	return false
}

func (cli *client) isCvmExtensionChange(cloud *typescvm.HuaWeiCvm, dbExt *corecvm.HuaWeiCvmExtension) bool {

	if dbExt.AliasName != cloud.OSEXTSRVATTRinstanceName {
		return true
	}

	if dbExt.HypervisorHostname != cloud.OSEXTSRVATTRhypervisorHostname {
		return true
	}

	sgIDs := make([]string, 0)
	for _, v := range cloud.SecurityGroups {
		sgIDs = append(sgIDs, v.Id)
	}
	if !assert.IsStringSliceEqual(dbExt.CloudSecurityGroupIDs, sgIDs) {
		return true
	}

	if dbExt.CloudTenantID != cloud.TenantId {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.DiskConfig, cloud.OSDCFdiskConfig) {
		return true
	}

	if dbExt.PowerState != cloud.OSEXTSTSpowerState {
		return true
	}

	if dbExt.ConfigDrive != cloud.ConfigDrive {
		return true
	}

	osDiskId := ""
	dataDiskIds := make([]string, 0)
	for _, v := range cloud.OsExtendedVolumesvolumesAttached {
		if v.BootIndex != nil && *v.BootIndex == "0" {
			osDiskId = v.Id
		} else {
			dataDiskIds = append(dataDiskIds, v.Id)
		}
	}
	if dbExt.CloudOSVolumeID != osDiskId {
		return true
	}
	if !assert.IsStringSliceEqual(dbExt.CloudDataVolumeIDs, dataDiskIds) {
		return true
	}

	if dbExt.RootDeviceName != cloud.OSEXTSRVATTRrootDeviceName {
		return true
	}

	if !assert.IsPtrStringEqual(dbExt.CloudEnterpriseProjectID, cloud.EnterpriseProjectId) {
		return true
	}

	if !assert.IsPtrInt32Equal(dbExt.CpuOptions.CpuThreads, cloud.CpuOptions.HwcpuThreads) {
		return true
	}

	if cli.isExtFlavorChange(cloud.Flavor, dbExt.Flavor) {
		return true
	}

	if cli.isExtMetadataChange(cloud.Metadata, dbExt) {
		return true
	}
	return false
}

func (cli *client) isExtFlavorChange(cloud *corecvm.HuaWeiFlavor, db *corecvm.HuaWeiFlavor) bool {
	if db.CloudID != cloud.CloudID {
		return true
	}

	if db.Name != cloud.Name {
		return true
	}

	if db.Disk != cloud.Disk {
		return true
	}

	if db.VCpus != cloud.VCpus {
		return true
	}

	if db.Ram != cloud.Ram {
		return true
	}
	return false
}

func (cli *client) isExtMetadataChange(metadata map[string]string, dbExt *corecvm.HuaWeiCvmExtension) bool {
	if dbExt.Metadata.ChargingMode != metadata["charging_mode"] {
		return true
	}

	if dbExt.Metadata.CloudOrderID != metadata["metering.order_id"] {
		return true
	}

	if dbExt.Metadata.CloudProductID != metadata["metering.product_id"] {
		return true
	}

	if dbExt.Metadata.EcmResStatus != metadata["EcmResStatus"] {
		return true
	}

	if dbExt.Metadata.ImageType != metadata["metering.imagetype"] {
		return true
	}

	if dbExt.Metadata.ResourceSpecCode != metadata["metering.resourcespeccode"] {
		return true
	}

	if dbExt.Metadata.ResourceType != metadata["metering.resourcetype"] {
		return true
	}

	if dbExt.Metadata.InstanceExtraInfo != metadata["cascaded.instance_extrainfo"] {
		return true
	}

	if dbExt.Metadata.ImageName != metadata["image_name"] {
		return true
	}

	if dbExt.Metadata.AgencyName != metadata["agency_name"] {
		return true
	}

	if dbExt.Metadata.OSBit != metadata["os_bit"] {
		return true
	}

	if dbExt.Metadata.SupportAgentList != metadata["__support_agent_list"] {
		return true
	}
	return false
}
