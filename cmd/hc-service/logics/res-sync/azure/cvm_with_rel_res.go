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

	cvmrelmgr "hcm/cmd/hc-service/logics/res-sync/cvm-rel-manager"
	typescore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	typesni "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// SyncCvmWithRelResOption ...
type SyncCvmWithRelResOption struct {
}

// Validate ...
func (opt SyncCvmWithRelResOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// CvmWithRelRes ...
/*
	同步流程：
		step1: 如果cvm全部不存在，仅同步主机即可，有可能主机被从云上删除
		step2: 获取cvm和关联资源的关联关系
		step3: sync vpc
		step4: sync subnet
		step5: sync security group
		step6: sync disk
		step7: sync eip
		step8: sync network interface
		step9: sync cvm
		step10: sync cvm_sg_rel
		step11: sync cvm_disk_rel
		step12: sync cvm_eip_rel
		step13: sync cvm_network_interface_rel
*/
func (cli *client) CvmWithRelRes(kt *kit.Kit, params *SyncBaseParams, opt *SyncCvmWithRelResOption) (
	*SyncResult, error) {

	cvmFromCloud, err := cli.listCvmFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	// step1: 如果cvm全部不存在，仅同步主机即可，有可能主机被从云上删除
	if len(cvmFromCloud) == 0 {
		if _, err = cli.Cvm(kt, params, new(SyncCvmOption)); err != nil {
			return nil, err
		}

		return new(SyncResult), nil
	}

	// step2: 获取cvm和关联资源的关联关系
	mgr, err := cli.buildCvmRelManger(kt, params.ResourceGroupName, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] build cvm rel manager failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step3: sync vpc
	if err = mgr.SyncForAzure(kt, enumor.VpcCloudResType, func(kt *kit.Kit, resGroupName string,
		cloudIDs []string) error {

		assResParams := &SyncBaseParams{
			AccountID:         params.AccountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		if _, err := cli.Vpc(kt, assResParams, new(SyncVpcOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate vpc failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step4: sync subnet
	if err = mgr.SyncDependParentResForAzure(kt, enumor.SubnetCloudResType, func(kt *kit.Kit, resGroupName string,
		cloudVpcID string, cloudIDs []string) error {

		assResParams := &SyncBaseParams{
			AccountID:         params.AccountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		syncSubnetOpt := &SyncSubnetOption{
			CloudVpcID: cloudVpcID,
		}
		if _, err := cli.Subnet(kt, assResParams, syncSubnetOpt); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate subnet failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step5: sync security group
	if err = mgr.SyncForAzure(kt, enumor.SecurityGroupCloudResType, func(kt *kit.Kit, resGroupName string,
		cloudIDs []string) error {

		assResParams := &SyncBaseParams{
			AccountID:         params.AccountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		if _, err := cli.SecurityGroup(kt, assResParams, new(SyncSGOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate disk failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step6: sync disk
	if err = mgr.SyncForAzure(kt, enumor.DiskCloudResType, func(kt *kit.Kit, resGroupName string,
		cloudIDs []string) error {

		assResParams := &SyncBaseParams{
			AccountID:         params.AccountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		if _, err := cli.Disk(kt, assResParams, new(SyncDiskOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate disk failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step7: sync eip
	if err = mgr.SyncForAzure(kt, enumor.EipCloudResType, func(kt *kit.Kit, resGroupName string,
		cloudIDs []string) error {

		assResParams := &SyncBaseParams{
			AccountID:         params.AccountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		if _, err := cli.Eip(kt, assResParams, new(SyncEipOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate eip failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step8: sync network interface
	if err = mgr.SyncForAzure(kt, enumor.NetworkInterfaceCloudResType, func(kt *kit.Kit, resGroupName string,
		cloudIDs []string) error {

		assResParams := &SyncBaseParams{
			AccountID:         params.AccountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		if _, err := cli.NetworkInterface(kt, assResParams, new(SyncNIOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync network interface failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step9: sync cvm
	if err = mgr.SyncForAzure(kt, enumor.CvmCloudResType, func(kt *kit.Kit, resGroupName string,
		cloudIDs []string) error {

		assResParams := &SyncBaseParams{
			AccountID:         params.AccountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		if _, err := cli.Cvm(kt, assResParams, new(SyncCvmOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	syncRelOpt := &cvmrelmgr.SyncRelOption{
		Vendor: enumor.Azure,
	}

	// step10: sync cvm_sg_rel
	syncRelOpt.ResType = enumor.SecurityGroupCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_securityGroup_rel failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step11: sync cvm_disk_rel
	syncRelOpt.ResType = enumor.DiskCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_disk_rel failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step12: sync cvm_eip_rel
	syncRelOpt.ResType = enumor.EipCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_eip_rel failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	// step13: sync cvm_network_interface_rel
	syncRelOpt.ResType = enumor.NetworkInterfaceCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_network_interface_rel failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) buildCvmRelManger(kt *kit.Kit, resGroupName string, cvmFromCloud []typecvm.AzureCvm) (
	*cvmrelmgr.CvmRelManger, error) {

	if len(cvmFromCloud) == 0 {
		return nil, fmt.Errorf("cvms that from cloud is required")
	}

	vpcSubnetMap := make(map[string]map[string]struct{})

	niMap, err := cli.getEipMapFromCloudByIP(kt, resGroupName, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] get eip map failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return nil, err
	}

	mgr := cvmrelmgr.NewCvmRelManager(cli.dbCli)
	for _, cvm := range cvmFromCloud {
		for _, niCloudID := range cvm.NetworkInterfaceIDs {
			// Network interface
			mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.NetworkInterfaceCloudResType, niCloudID)

			ni, exist := niMap[niCloudID]
			if !exist {
				return nil, fmt.Errorf("network interface: %s not found", niCloudID)
			}

			// VPC
			if ni.CloudVpcID != nil {
				mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.VpcCloudResType, *ni.CloudVpcID)
			}

			// Subnet 同步依赖与CloudVpcID
			if ni.CloudSubnetID != nil {
				if _, exist = vpcSubnetMap[*ni.CloudVpcID]; !exist {
					vpcSubnetMap[*ni.CloudVpcID] = make(map[string]struct{}, 0)
				}

				vpcSubnetMap[*ni.CloudVpcID][*ni.CloudSubnetID] = struct{}{}
			}

			// Eip
			for _, ip := range ni.Extension.IPConfigurations {
				if ip.Properties.PublicIPAddress != nil {
					mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.EipCloudResType,
						*ip.Properties.PublicIPAddress.CloudID)
				}
			}
		}

		// Disk
		mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.DiskCloudResType, cvm.CloudOsDiskID)

		for _, diskID := range cvm.CloudDataDiskIDs {
			mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.DiskCloudResType, diskID)
		}
	}

	vpcSubnetListMap := make(map[string][]string)
	for cloudVpcID, subnetMap := range vpcSubnetMap {
		vpcSubnetListMap[cloudVpcID] = converter.MapKeyToStringSlice(subnetMap)
	}

	mgr.AddAssParentWithChildRes(enumor.SubnetCloudResType, vpcSubnetListMap)

	return mgr, nil
}

// getEipMapFromCloudByIP 通过弹性IP的IP地址获取IP和EipID的映射关系，内置ips分页查询，对ips数量没有限制。
func (cli *client) getEipMapFromCloudByIP(kt *kit.Kit, resGroupName string, cvmFromCloud []typecvm.AzureCvm) (
	map[string]typesni.AzureNI, error) {

	nis := make(map[string]struct{}, 0)
	for _, cvm := range cvmFromCloud {
		for _, one := range cvm.NetworkInterfaceIDs {
			nis[one] = struct{}{}
		}
	}

	if len(nis) == 0 {
		return make(map[string]typesni.AzureNI), nil
	}

	opt := &typescore.AzureListByIDOption{
		ResourceGroupName: resGroupName,
		CloudIDs:          converter.MapKeyToStringSlice(nis),
	}
	resp, err := cli.cloudCli.ListNetworkInterfaceByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list eip by ip from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
			enumor.Azure, err, opt, kt.Rid)
		return nil, err
	}

	result := make(map[string]typesni.AzureNI)
	for _, one := range resp.Details {
		result[*one.CloudID] = one
	}

	return result, nil
}
