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

	cvmrelmgr "hcm/cmd/hc-service/logics/res-sync/cvm-rel-manager"
	typecvm "hcm/pkg/adaptor/types/cvm"
	typeseip "hcm/pkg/adaptor/types/eip"
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
		step8: sync cvm
		step9: sync cvm_sg_rel
		step10: sync cvm_disk_rel
		step11: sync cvm_eip_rel
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
	diskBootMap, mgr, err := cli.buildCvmRelManger(kt, params.Region, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] build cvm rel manager failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step3: sync vpc
	if err = mgr.Sync(kt, enumor.VpcCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Vpc(kt, assResParams, new(SyncVpcOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate vpc failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step4: sync subnet
	if err = mgr.Sync(kt, enumor.SubnetCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Subnet(kt, assResParams, new(SyncSubnetOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate subnet failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step5: sync security group
	if err = mgr.Sync(kt, enumor.SecurityGroupCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.SecurityGroup(kt, assResParams, new(SyncSGOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate disk failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step6: sync disk
	if err = mgr.Sync(kt, enumor.DiskCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudIDs:  cloudIDs,
		}
		syncDiskOpt := &SyncDiskOption{
			BootMap: diskBootMap,
		}
		if _, err := cli.Disk(kt, assResParams, syncDiskOpt); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate disk failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step7: sync eip
	if err = mgr.Sync(kt, enumor.EipCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Eip(kt, assResParams, new(SyncEipOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm associate eip failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step8: sync cvm
	if err = mgr.Sync(kt, enumor.CvmCloudResType, func(kt *kit.Kit, cloudIDs []string) error {
		assResParams := &SyncBaseParams{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := cli.Cvm(kt, assResParams, new(SyncCvmOption)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("[%s] sync cvm failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	syncRelOpt := &cvmrelmgr.SyncRelOption{
		Vendor: enumor.Aws,
	}

	// step9: sync cvm_sg_rel
	syncRelOpt.ResType = enumor.SecurityGroupCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_securityGroup_rel failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step10: sync cvm_disk_rel
	syncRelOpt.ResType = enumor.DiskCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_disk_rel failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// step11: sync cvm_eip_rel
	syncRelOpt.ResType = enumor.EipCloudResType
	if err = mgr.SyncRel(kt, syncRelOpt); err != nil {
		logs.Errorf("[%s] sync cvm_eip_rel failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

// getEipMapFromCloudCvm 查询主机所对应的Eip信息。
func (cli *client) getEipMapFromCloudCvm(kt *kit.Kit, region string, cvmFromCloud []typecvm.AwsCvm) (
	map[string]string, error) {

	ipMap := make(map[string]struct{}, 0)
	for _, one := range cvmFromCloud {
		if one.PublicIpAddress != nil && len(*one.PublicIpAddress) != 0 {
			ipMap[*one.PublicIpAddress] = struct{}{}
		}
	}

	if len(ipMap) == 0 {
		return make(map[string]string), nil
	}

	ips := converter.MapKeyToStringSlice(ipMap)

	result := make(map[string]string, 0)
	split := slice.Split(ips, constant.CloudResourceSyncMaxLimit)
	for _, partIPs := range split {
		opt := &typeseip.AwsEipListOption{
			Region: region,
			Ips:    partIPs,
		}
		resp, err := cli.cloudCli.ListEip(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list eip by ip from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
				enumor.Aws, err, opt, kt.Rid)
			return nil, err
		}

		for _, one := range resp.Details {
			result[*one.PublicIp] = one.CloudID
		}
	}

	return result, nil
}

func (cli *client) buildCvmRelManger(kt *kit.Kit, region string, cvmFromCloud []typecvm.AwsCvm) (
	map[string]struct{}, *cvmrelmgr.CvmRelManger, error) {

	if len(cvmFromCloud) == 0 {
		return nil, nil, fmt.Errorf("cvms that from cloud is required")
	}

	eipMap, err := cli.getEipMapFromCloudCvm(kt, region, cvmFromCloud)
	if err != nil {
		logs.Errorf("[%s] get eip map failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return nil, nil, err
	}

	rootDeviceMap := make(map[string]struct{}, 0)
	mgr := cvmrelmgr.NewCvmRelManager(cli.dbCli)
	for _, cvm := range cvmFromCloud {
		// SecurityGroup
		for _, sg := range cvm.SecurityGroups {
			if sg == nil {
				continue
			}

			mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.SecurityGroupCloudResType, converter.PtrToVal(sg.GroupId))
		}

		// Vpc
		if cvm.VpcId != nil {
			mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.VpcCloudResType, *cvm.VpcId)
		}

		// Subnet
		if cvm.SubnetId != nil {
			mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.SubnetCloudResType, *cvm.SubnetId)
		}

		// Disk
		var rootDiskName string
		if converter.PtrToVal(cvm.RootDeviceType) == "ebs" && cvm.RootDeviceName != nil && *cvm.RootDeviceName != "" {
			rootDiskName = *cvm.RootDeviceName
		}

		for _, disk := range cvm.BlockDeviceMappings {
			if disk.Ebs == nil {
				continue
			}

			mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.DiskCloudResType, *disk.Ebs.VolumeId)

			if disk.DeviceName != nil && *disk.DeviceName != "" && *disk.DeviceName == rootDiskName {
				rootDeviceMap[*disk.Ebs.VolumeId] = struct{}{}
			}
		}

		// Eip
		if cvm.PublicIpAddress != nil && len(*cvm.PublicIpAddress) != 0 {
			eipCloudID, exist := eipMap[*cvm.PublicIpAddress]
			if exist {
				mgr.CvmAppendAssResCloudID(cvm.GetCloudID(), enumor.EipCloudResType, eipCloudID)
			}
		}
	}

	return rootDeviceMap, mgr, nil
}
