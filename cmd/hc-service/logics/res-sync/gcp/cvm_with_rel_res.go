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
	"fmt"

	cvmrelmgr "hcm/cmd/hc-service/logics/res-sync/cvm-rel-manager"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	typesdisk "hcm/pkg/adaptor/types/disk"
	typeseip "hcm/pkg/adaptor/types/eip"
	adtysubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncCvmWithRelResOption ...
type SyncCvmWithRelResOption struct {
	Region string `json:"region" validate:"required"`
	Zone   string `json:"zone" validate:"required"`
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
		step5: sync disk
		step6: sync eip
		step7: sync cvm
		step8: sync network interface
		step9: sync cvm_disk_rel
		step10: sync cvm_eip_rel
		step11: sync cvm_network_interface_rel
*/
func (cli *client) CvmWithRelRes(kt *kit.Kit, params *SyncBaseParams, opt *SyncCvmWithRelResOption) (
	*SyncResult, error) {

	syncCvmOption := &SyncCvmOption{
		Region: opt.Region,
		Zone:   opt.Zone,
	}
	cvmFromCloud, err := cli.listCvmFromCloud(kt, params, syncCvmOption)
	if err != nil {
		return nil, err
	}

	// step1: 如果cvm全部不存在，仅同步主机即可，有可能主机被从云上删除
	if len(cvmFromCloud) == 0 {
		if _, err = cli.Cvm(kt, params, syncCvmOption); err != nil {
			return nil, err
		}

		return new(SyncResult), nil
	}

	// step2: 获取cvm和关联资源的关联关系
	diskBootMap, mgr, err := cli.buildCvmRelManger(kt, cvmFromCloud, opt)
	if err != nil {
		logs.Errorf("[%s] build cvm rel manager failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step3: sync vpc
	if err = mgr.Sync(kt, enumor.VpcCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Vpc(kt, assResParams, new(SyncVpcOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate vpc failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step4: sync subnet
	if err = mgr.Sync(kt, enumor.SubnetCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Subnet(kt, assResParams, &SyncSubnetOption{Region: opt.Region}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate subnet failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step5: sync disk
	if err = mgr.Sync(kt, enumor.DiskCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Disk(kt, assResParams, &SyncDiskOption{Zone: opt.Zone, BootMap: diskBootMap}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate disk failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step6: sync eip
	if err = mgr.Sync(kt, enumor.EipCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Eip(kt, assResParams, &SyncEipOption{Region: opt.Region}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate eip failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step7: sync cvm
	if err = mgr.Sync(kt, enumor.CvmCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Cvm(kt, assResParams, syncCvmOption); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step8: sync network interface 网络接口同步是用的主机ID，因为网络接口不能单独存在
	if err = mgr.Sync(kt, enumor.CvmCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			CloudIDs:  cloudIDs,
		}
		syncNIOpt := &SyncNIOption{
			Region: opt.Region,
			Zone:   opt.Zone,
		}
		if _, err := cli.NetworkInterface(kt, assResParams, syncNIOpt); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync network interface failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	syncRelOpt := &cvmrelmgr.SyncRelOption{
		Vendor: enumor.Gcp,
	}

	// step9: sync cvm_disk_rel
	syncRelOpt.ResType = enumor.DiskCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_disk_rel failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step10: sync cvm_eip_rel
	syncRelOpt.ResType = enumor.EipCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_eip_rel failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	// step11: sync cvm_eip_rel
	syncRelOpt.ResType = enumor.NetworkInterfaceCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_network_interface_rel failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) getVpcMapFromCloud(kt *kit.Kit, cvmFromCloud []typecvm.GcpCvm) (map[string]string, error) {
	selfLinkMap := make(map[string]struct{}, 0)
	for _, cvm := range cvmFromCloud {
		for _, networkInterface := range cvm.NetworkInterfaces {
			if networkInterface != nil {
				selfLinkMap[networkInterface.Network] = struct{}{}
			}
		}
	}

	if len(selfLinkMap) == 0 {
		return make(map[string]string), nil
	}

	selfLinks := converter.MapKeyToStringSlice(selfLinkMap)

	result := make(map[string]string, 0)
	split := slice.Split(selfLinks, core.GcpSelfLinkMaxQueryLimit)
	for _, part := range split {
		opt := &types.GcpListOption{
			SelfLinks: part,
			Page: &core.GcpPage{
				PageSize: core.GcpSelfLinkMaxQueryLimit,
			},
		}
		vpcResult, err := cli.cloudCli.ListVpc(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list vpc by self link from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
				enumor.Gcp, err, opt, kt.Rid)
			return nil, err
		}

		for _, one := range vpcResult.Details {
			result[one.Extension.SelfLink] = one.CloudID
		}
	}

	return result, nil
}

func (cli *client) getDiskMapFromCloud(kt *kit.Kit, zone string, cvmFromCloud []typecvm.GcpCvm) (
	map[string]string, error) {

	selfLinkMap := make(map[string]struct{}, 0)
	for _, cvm := range cvmFromCloud {
		for _, disk := range cvm.Disks {
			if disk != nil {
				selfLinkMap[disk.Source] = struct{}{}
			}
		}
	}

	if len(selfLinkMap) == 0 {
		return make(map[string]string), nil
	}

	selfLinks := converter.MapKeyToStringSlice(selfLinkMap)

	result := make(map[string]string, 0)
	split := slice.Split(selfLinks, core.GcpSelfLinkMaxQueryLimit)
	for _, part := range split {
		opt := &typesdisk.GcpDiskListOption{
			Zone:      zone,
			SelfLinks: part,
			Page: &core.GcpPage{
				PageSize: core.GcpSelfLinkMaxQueryLimit,
			},
		}
		disks, _, err := cli.cloudCli.ListDisk(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list disk by self link from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
				enumor.Gcp, err, opt, kt.Rid)
			return nil, err
		}

		for _, one := range disks {
			result[one.SelfLink] = fmt.Sprintf("%d", one.Id)
		}
	}

	return result, nil
}

func (cli *client) getSubnetMapFromCloud(kt *kit.Kit, region string, cvmFromCloud []typecvm.GcpCvm) (
	map[string]string, error) {

	selfLinkMap := make(map[string]struct{}, 0)
	for _, cvm := range cvmFromCloud {
		for _, networkInterface := range cvm.NetworkInterfaces {
			if networkInterface != nil {
				selfLinkMap[networkInterface.Subnetwork] = struct{}{}
			}
		}
	}

	if len(selfLinkMap) == 0 {
		return make(map[string]string), nil
	}

	selfLinks := converter.MapKeyToStringSlice(selfLinkMap)

	result := make(map[string]string, 0)
	split := slice.Split(selfLinks, core.GcpSelfLinkMaxQueryLimit)
	for _, part := range split {
		opt := &adtysubnet.GcpSubnetListOption{
			GcpListOption: core.GcpListOption{
				Page: &core.GcpPage{
					PageSize: core.GcpSelfLinkMaxQueryLimit,
				},
				SelfLinks: part,
			},
			Region: region,
		}
		subnetResult, err := cli.cloudCli.ListSubnet(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list subnet by self link from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
				enumor.Gcp, err, opt, kt.Rid)
			return nil, err
		}

		for _, one := range subnetResult.Details {
			result[one.Extension.SelfLink] = one.CloudID
		}
	}

	return result, nil
}

func (cli *client) buildCvmRelManger(kt *kit.Kit, cvmFromCloud []typecvm.GcpCvm, opt *SyncCvmWithRelResOption) (
	map[string]struct{}, *cvmrelmgr.CvmRelManger, error) {

	if len(cvmFromCloud) == 0 {
		return nil, nil, fmt.Errorf("cvms that from cloud is required")
	}

	diskBootMap := make(map[string]struct{})

	eipMap, err := cli.getEipMapFromCloud(kt, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] get eip map failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, nil, err
	}

	vpcMap, err := cli.getVpcMapFromCloud(kt, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] get vpc map failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, nil, err
	}

	subnetMap, err := cli.getSubnetMapFromCloud(kt, opt.Region, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] get subnet map failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, nil, err
	}

	diskMap, err := cli.getDiskMapFromCloud(kt, opt.Zone, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] get disk map failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, nil, err
	}

	mgr := cvmrelmgr.NewCvmRelManager(cli.dbCli)
	for _, cvm := range cvmFromCloud {
		// Disk
		for _, disk := range cvm.Disks {
			if disk != nil {
				if disk.Boot {
					diskBootMap[disk.Source] = struct{}{}
				}
				diskID, exist := diskMap[disk.Source]
				if !exist {
					return nil, nil, fmt.Errorf("disk: %s not found", disk.Source)
				}
				mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.DiskCloudResType, diskID)
			}
		}

		for _, networkInterface := range cvm.NetworkInterfaces {
			mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.NetworkInterfaceCloudResType,
				fmt.Sprintf("%d", cvm.Id)+"_"+networkInterface.Name)
			if networkInterface != nil {
				// NetworkInterface Gcp网络接口的同步是按主机维度进行的，所以这里传入的是主机ID

				// Vpc
				vpcID, exist := vpcMap[networkInterface.Network]
				if !exist {
					return nil, nil, fmt.Errorf("vpc: %s not found", networkInterface.Network)
				}
				mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.VpcCloudResType, vpcID)

				// Subnet
				subnetID, exist := subnetMap[networkInterface.Subnetwork]
				if !exist {
					return nil, nil, fmt.Errorf("subnet: %s not found", networkInterface.Subnetwork)
				}
				mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.SubnetCloudResType, subnetID)

				// Eip
				for _, config := range networkInterface.AccessConfigs {
					if eipCloudID, exist := eipMap[config.NatIP]; exist {
						mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.EipCloudResType, eipCloudID)
					}
				}
			}
		}
	}
	return diskBootMap, mgr, nil
}

// getEipMapFromCloud 通过弹性IP的IP地址获取IP和EipID的映射关系，内置ips分页查询，对ips数量没有限制。
func (cli *client) getEipMapFromCloud(kt *kit.Kit, cvmFromCloud []typecvm.GcpCvm) (map[string]string, error) {

	ipMap := make(map[string]struct{}, 0)
	for _, one := range cvmFromCloud {
		for _, networkInterface := range one.NetworkInterfaces {
			for _, config := range networkInterface.AccessConfigs {
				ipMap[config.NatIP] = struct{}{}
			}
		}
	}

	if len(ipMap) == 0 {
		return make(map[string]string), nil
	}

	ips := converter.MapKeyToStringSlice(ipMap)

	result := make(map[string]string, 0)
	split := slice.Split(ips, constant.CloudResourceSyncMaxLimit)
	for _, partIPs := range split {
		opt := &typeseip.GcpEipAggregatedListOption{
			IPAddresses: partIPs,
		}
		eips, err := cli.cloudCli.ListAggregatedEip(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list eip by ip from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
				enumor.Gcp, err, opt, kt.Rid)
			return nil, err
		}

		for _, one := range eips {
			result[one.Address] = fmt.Sprintf("%d", one.Id)
		}
	}

	return result, nil
}
