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
	adcore "hcm/pkg/adaptor/types/core"
	typesni "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud/network-interface"
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
)

// SyncNIOption ...
type SyncNIOption struct {
}

// Validate ...
func (opt SyncNIOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// NetworkInterface ...
func (cli *client) NetworkInterface(kt *kit.Kit, params *SyncBaseParams, opt *SyncNIOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	niFromCloud, err := cli.listNetworkInterfaceFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	niFromDB, err := cli.listNetworkInterfaceFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(niFromCloud) == 0 && len(niFromDB) == 0 {
		return new(SyncResult), nil
	}

	addNetworkInterface, updateMap, delCloudIDs := common.Diff[typesni.AzureNI,
		coreni.NetworkInterface[coreni.AzureNIExtension]](niFromCloud, niFromDB, isNIChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteNetworkInterface(kt, params.AccountID, params.ResourceGroupName, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addNetworkInterface) > 0 {
		if err = cli.createNetworkInterface(kt, params.AccountID, params.ResourceGroupName,
			addNetworkInterface); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateNetworkInterface(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// RemoveNetworkInterfaceDeleteFromCloud ...
func (cli *client) RemoveNetworkInterfaceDeleteFromCloud(kt *kit.Kit, accountID string, resGroupName string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
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
		resultFromDB, err := cli.dbCli.Global.NetworkInterface.List(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list ni failed, err: %v, req: %v, rid: %s", enumor.Azure,
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

		var resultFromCloud []typesni.AzureNI
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID:         accountID,
				ResourceGroupName: resGroupName,
				CloudIDs:          cloudIDs,
			}
			resultFromCloud, err = cli.listNetworkInterfaceFromCloud(kt, params)
			if err != nil {
				return err
			}
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, *one.CloudID)
			}

			delCloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if err = cli.deleteNetworkInterface(kt, accountID, resGroupName, delCloudIDs); err != nil {
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

// deleteNetworkInterface deletes network interfaces from the database
func (cli *client) deleteNetworkInterface(kt *kit.Kit, accountID string, resGroupName string, delCloudIDs []string) error {

	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete network interface, network interfaces is required")
	}

	checkParams := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          delCloudIDs,
	}
	delNetworkInterfaceFromCloud, err := cli.listNetworkInterfaceFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delNetworkInterfaceFromCloud) > 0 {
		logs.Errorf("[%s] validate ni not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delNetworkInterfaceFromCloud), kt.Rid)
		return fmt.Errorf("validate ni not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.NetworkInterface.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete ni failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync ni to delete ni success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// updateNetworkInterface updates network interfaces in the database
func (cli *client) updateNetworkInterface(kt *kit.Kit, accountID string, updateMap map[string]typesni.AzureNI) error {

	if len(updateMap) == 0 {
		return fmt.Errorf("update network interface, network interfaces is required")
	}

	nis := make([]dataproto.NetworkInterfaceUpdateReq[dataproto.AzureNICreateExt], 0)
	for id, item := range updateMap {
		tmpRes := dataproto.NetworkInterfaceUpdateReq[dataproto.AzureNICreateExt]{
			ID:            id,
			AccountID:     accountID,
			Name:          converter.PtrToVal(item.Name),
			Region:        converter.PtrToVal(item.Region),
			Zone:          converter.PtrToVal(item.Zone),
			CloudID:       converter.PtrToVal(item.CloudID),
			VpcID:         converter.PtrToVal(item.VpcID),
			CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
			SubnetID:      converter.PtrToVal(item.SubnetID),
			CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
			PrivateIPv4:   item.PrivateIPv4,
			PrivateIPv6:   item.PrivateIPv6,
			PublicIPv4:    item.PublicIPv4,
			PublicIPv6:    item.PublicIPv6,
			InstanceID:    converter.PtrToVal(item.InstanceID),
		}
		if item.Extension != nil {
			tmpRes.Extension = &dataproto.AzureNICreateExt{
				ResourceGroupName: item.Extension.ResourceGroupName,
				MacAddress:        converter.PtrToVal(item.Extension.MacAddress),
				// EnableAcceleratedNetworking 是否加速网络
				EnableAcceleratedNetworking: item.Extension.EnableAcceleratedNetworking,
				// EnableIPForwarding 是否允许IP转发
				EnableIPForwarding: item.Extension.EnableIPForwarding,
				// DNSSettings DNS设置
				DNSSettings: item.Extension.DNSSettings,
				// GatewayLoadBalancerID 网关负载均衡器ID
				CloudGatewayLoadBalancerID: item.Extension.CloudGatewayLoadBalancerID,
				// CloudSecurityGroupID 网络安全组ID
				CloudSecurityGroupID: item.Extension.CloudSecurityGroupID,
				SecurityGroupID:      item.Extension.SecurityGroupID,
			}
			// IPConfigurations IP配置列表
			var tmpIPConfigs []*coreni.InterfaceIPConfiguration
			for _, cidrItem := range item.Extension.IPConfigurations {
				tmpIPConfigs = append(tmpIPConfigs, cidrItem)
			}
			tmpRes.Extension.IPConfigurations = tmpIPConfigs
		}

		nis = append(nis, tmpRes)
	}

	updateReq := &dataproto.NetworkInterfaceBatchUpdateReq[dataproto.AzureNICreateExt]{
		NetworkInterfaces: nis,
	}
	if err := cli.dbCli.Azure.NetworkInterface.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db ni failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync ni to update ni success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// createNetworkInterface creates network interfaces in the database
func (cli *client) createNetworkInterface(kt *kit.Kit, accountID, resGroupName string, adds []typesni.AzureNI) error {

	if len(adds) == 0 {
		return fmt.Errorf("create network interface, network interfaces is required")
	}

	cvmMap, err := cli.getCvmMapFromDB(kt, accountID, resGroupName, adds)
	if err != nil {
		return err
	}

	nis := make([]dataproto.NetworkInterfaceReq[dataproto.AzureNICreateExt], 0, len(adds))
	for _, item := range adds {
		bizID := int64(constant.UnassignedBiz)
		if item.InstanceID != nil {
			if one, exist := cvmMap[converter.PtrToVal(item.InstanceID)]; exist {
				bizID = one.BkBizID
			}
		}

		tmpRes := dataproto.NetworkInterfaceReq[dataproto.AzureNICreateExt]{
			AccountID:     accountID,
			Vendor:        string(enumor.Azure),
			Name:          converter.PtrToVal(item.Name),
			Region:        converter.PtrToVal(item.Region),
			Zone:          converter.PtrToVal(item.Zone),
			CloudID:       converter.PtrToVal(item.CloudID),
			VpcID:         converter.PtrToVal(item.VpcID),
			CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
			SubnetID:      converter.PtrToVal(item.SubnetID),
			CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
			PrivateIPv4:   item.PrivateIPv4,
			PrivateIPv6:   item.PrivateIPv6,
			PublicIPv4:    item.PublicIPv4,
			PublicIPv6:    item.PublicIPv6,
			InstanceID:    converter.PtrToVal(item.InstanceID),
			BkBizID:       bizID,
		}
		if item.Extension != nil {
			tmpRes.Extension = &dataproto.AzureNICreateExt{
				ResourceGroupName: item.Extension.ResourceGroupName,
				MacAddress:        converter.PtrToVal(item.Extension.MacAddress),
				// EnableAcceleratedNetworking 是否加速网络
				EnableAcceleratedNetworking: item.Extension.EnableAcceleratedNetworking,
				// EnableIPForwarding 是否允许IP转发
				EnableIPForwarding: item.Extension.EnableIPForwarding,
				// DNSSettings DNS设置
				DNSSettings: item.Extension.DNSSettings,
				// GatewayLoadBalancerID 网关负载均衡器ID
				CloudGatewayLoadBalancerID: item.Extension.CloudGatewayLoadBalancerID,
				// CloudSecurityGroupID 网络安全组ID
				CloudSecurityGroupID: item.Extension.CloudSecurityGroupID,
				SecurityGroupID:      item.Extension.SecurityGroupID,
			}
			// IPConfigurations IP配置列表
			var tmpIPConfigs []*coreni.InterfaceIPConfiguration
			for _, cidrItem := range item.Extension.IPConfigurations {
				tmpIPConfigs = append(tmpIPConfigs, cidrItem)
			}
			tmpRes.Extension.IPConfigurations = tmpIPConfigs
		}

		nis = append(nis, tmpRes)
	}

	createReq := &dataproto.NetworkInterfaceBatchCreateReq[dataproto.AzureNICreateExt]{
		NetworkInterfaces: nis,
	}
	if _, err := cli.dbCli.Azure.NetworkInterface.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create ni failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync ni to create ni success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(adds), kt.Rid)

	return nil
}

// getCvmMapFromDB retrieves a map of CVMs from the database
func (cli *client) getCvmMapFromDB(kt *kit.Kit, accountID string, resGroupName string, adds []typesni.AzureNI) (
	map[string]corecvm.Cvm[cvm.AzureCvmExtension], error) {

	instanceIDs := make([]string, 0)
	for _, one := range adds {
		if one.InstanceID != nil {
			instanceIDs = append(instanceIDs, *one.InstanceID)
		}
	}

	if len(instanceIDs) == 0 {
		return make(map[string]corecvm.Cvm[cvm.AzureCvmExtension]), nil
	}

	params := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          instanceIDs,
	}
	cvms, err := cli.listCvmFromDB(kt, params)
	if err != nil {
		logs.Errorf("[%s] list cvm from db failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return nil, err
	}

	cvmMap := make(map[string]corecvm.Cvm[cvm.AzureCvmExtension])
	for _, one := range cvms {
		cvmMap[one.CloudID] = one
	}

	return cvmMap, nil
}

// listNetworkInterfaceFromCloud retrieves network interfaces from the cloud based on the provided parameters.
func (cli *client) listNetworkInterfaceFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typesni.AzureNI, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adcore.AzureListByIDOption{
		ResourceGroupName: params.ResourceGroupName,
		CloudIDs:          params.CloudIDs,
	}
	result, err := cli.cloudCli.ListNetworkInterfaceByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list ni from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// listNetworkInterfaceFromDB retrieves network interfaces from the database based on the provided parameters.
func (cli *client) listNetworkInterfaceFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]coreni.NetworkInterface[coreni.AzureNIExtension], error) {

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
					Field: "extension.resource_group_name",
					Op:    filter.JSONEqual.Factory(),
					Value: params.ResourceGroupName,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.NetworkInterface.ListNetworkInterfaceExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list ni from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// isNIChange checks if the Azure network interface has changed compared to the database information.
func isNIChange(item typesni.AzureNI, dbInfo coreni.NetworkInterface[coreni.AzureNIExtension]) bool {

	if dbInfo.Name != converter.PtrToVal(item.Name) || dbInfo.Region != converter.PtrToVal(item.Region) ||
		dbInfo.Zone != converter.PtrToVal(item.Zone) || dbInfo.CloudID != converter.PtrToVal(item.CloudID) ||
		dbInfo.CloudVpcID != converter.PtrToVal(item.CloudVpcID) ||
		dbInfo.CloudSubnetID != converter.PtrToVal(item.CloudSubnetID) ||
		dbInfo.InstanceID != converter.PtrToVal(item.InstanceID) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PrivateIPv4, dbInfo.PrivateIPv4) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PrivateIPv6, dbInfo.PrivateIPv6) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv4, dbInfo.PublicIPv4) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv6, dbInfo.PublicIPv6) {
		return true
	}
	if dbInfo.Extension == nil {
		return false
	}

	extRet := checkAzureExt(item, dbInfo)
	if extRet {
		return true
	}

	ipExtRet := checkAzureIPConfigIsUpdate(item, dbInfo)
	if ipExtRet {
		return true
	}

	return false
}

// checkAzureExt checks if the Azure network interface extension has been updated.
func checkAzureExt(item typesni.AzureNI, dbInfo coreni.NetworkInterface[coreni.AzureNIExtension]) bool {
	if item.Extension.ResourceGroupName != dbInfo.Extension.ResourceGroupName {
		return true
	}
	if item.Extension.MacAddress != nil &&
		!assert.IsPtrStringEqual(item.Extension.MacAddress, dbInfo.Extension.MacAddress) {
		return true
	}
	if item.Extension.EnableAcceleratedNetworking != nil &&
		!assert.IsPtrBoolEqual(item.Extension.EnableAcceleratedNetworking,
			dbInfo.Extension.EnableAcceleratedNetworking) {
		return true
	}
	if item.Extension.EnableIPForwarding != nil &&
		!assert.IsPtrBoolEqual(item.Extension.EnableIPForwarding, dbInfo.Extension.EnableIPForwarding) {
		return true
	}
	if item.Extension.DNSSettings != nil {
		if !assert.IsPtrStringSliceEqual(item.Extension.DNSSettings.DNSServers,
			dbInfo.Extension.DNSSettings.DNSServers) {
			return true
		}
	}
	if item.Extension.CloudGatewayLoadBalancerID != nil &&
		!assert.IsPtrStringEqual(item.Extension.CloudGatewayLoadBalancerID,
			dbInfo.Extension.CloudGatewayLoadBalancerID) {
		return true
	}
	if item.Extension.CloudSecurityGroupID != nil &&
		!assert.IsPtrStringEqual(item.Extension.CloudSecurityGroupID, dbInfo.Extension.CloudSecurityGroupID) {
		return true
	}
	return false
}

// checkAzureIPConfigIsUpdate checks if the IP configurations of the Azure network interface have been updated.
func checkAzureIPConfigIsUpdate(item typesni.AzureNI,
	dbInfo coreni.NetworkInterface[coreni.AzureNIExtension]) bool {

	for index, remote := range item.Extension.IPConfigurations {
		if len(dbInfo.Extension.IPConfigurations) > index {
			dbIpInfo := dbInfo.Extension.IPConfigurations[index]
			if !assert.IsPtrStringEqual(remote.CloudID, dbIpInfo.CloudID) {
				return true
			}
			if !assert.IsPtrStringEqual(remote.Name, dbIpInfo.Name) {
				return true
			}
			if !assert.IsPtrStringEqual(remote.Type, dbIpInfo.Type) {
				return true
			}
			if dbIpInfo.Properties != nil {
				if !assert.IsPtrBoolEqual(remote.Properties.Primary, dbIpInfo.Properties.Primary) {
					return true
				}
				if !assert.IsPtrStringEqual(remote.Properties.PrivateIPAddress, dbIpInfo.Properties.PrivateIPAddress) {
					return true
				}
				if !assert.IsPtrStringEqual((*string)(remote.Properties.PrivateIPAddressVersion),
					(*string)(dbIpInfo.Properties.PrivateIPAddressVersion)) {
					return true
				}
				if !assert.IsPtrStringEqual((*string)(remote.Properties.PrivateIPAllocationMethod),
					(*string)(dbIpInfo.Properties.PrivateIPAllocationMethod)) {
					return true
				}
				if !assert.IsPtrStringEqual(remote.Properties.CloudSubnetID, dbIpInfo.Properties.CloudSubnetID) {
					return true
				}
				if dbIpInfo.Properties.PublicIPAddress != nil && dbIpInfo.Properties.PublicIPAddress.Properties != nil {
					if !assert.IsPtrStringEqual(remote.Properties.PublicIPAddress.Properties.IPAddress,
						dbIpInfo.Properties.PublicIPAddress.Properties.IPAddress) {
						return true
					}
					if !assert.IsPtrStringEqual(
						(*string)(remote.Properties.PublicIPAddress.Properties.PublicIPAddressVersion),
						(*string)(dbIpInfo.Properties.PublicIPAddress.Properties.PublicIPAddressVersion)) {
						return true
					}
					if !assert.IsPtrStringEqual(
						(*string)(remote.Properties.PublicIPAddress.Properties.PublicIPAllocationMethod),
						(*string)(dbIpInfo.Properties.PublicIPAddress.Properties.PublicIPAllocationMethod)) {
						return true
					}
				}
			}
		}
	}
	return false
}
