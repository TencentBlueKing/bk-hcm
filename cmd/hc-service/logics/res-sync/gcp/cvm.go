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
	"time"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/gcp"
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
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

// SyncCvmOption ...
type SyncCvmOption struct {
	Region string `json:"region" validate:"required"`
	Zone   string `json:"zone" validate:"required"`
}

// Validate ...
func (opt SyncCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}

func (cli *client) Cvm(kt *kit.Kit, params *SyncBaseParams, opt *SyncCvmOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmFromCloud, err := cli.listCvmFromCloud(kt, params, opt)
	if err != nil {
		return nil, err
	}

	cvmFromDB, err := cli.listCvmFromDB(kt, params, opt.Zone)
	if err != nil {
		return nil, err
	}

	if len(cvmFromCloud) == 0 && len(cvmFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typescvm.GcpCvm, corecvm.Cvm[cvm.GcpCvmExtension]](
		cvmFromCloud, cvmFromDB, isCvmChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteCvm(kt, params.AccountID, opt.Zone, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createCvm(kt, params.AccountID, opt.Region, opt.Zone, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateCvm(kt, params.AccountID, opt.Region, opt.Zone, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createCvm(kt *kit.Kit, accountID string, region string, zone string,
	addSlice []typescvm.GcpCvm) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	lists := make([]dataproto.CvmBatchCreate[corecvm.GcpCvmExtension], 0)

	vpcMap, subnetMap, diskMap, vpcSelfLinks,
		subnetSelfLinks, imageMap, err := cli.getNIAssResMapBySelfLinkFromNI(kt, accountID, region, zone, addSlice)
	if err != nil {
		return err
	}

	for _, one := range addSlice {
		inVpcSelfLinks := make([]string, 0)
		inSubnetSelfLinks := make([]string, 0)
		cloudNetWorkInterfaceIDs := make([]string, 0)
		if len(one.NetworkInterfaces) > 0 {
			for _, networkInterface := range one.NetworkInterfaces {
				if networkInterface != nil {
					cloudNetInterfaceID := fmt.Sprintf("%d", one.Id) + "_" + networkInterface.Name
					cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
					inVpcSelfLinks = append(inVpcSelfLinks, networkInterface.Network)
					inSubnetSelfLinks = append(inSubnetSelfLinks, networkInterface.Subnetwork)
				}
			}
		}

		if _, exsit := vpcMap[inVpcSelfLinks[0]]; !exsit {
			return fmt.Errorf("cvm %s can not find vpc", fmt.Sprint(one.Id))
		}

		subnetIDs := make([]string, 0)
		cloudSubIDs := make([]string, 0)
		for _, one := range inSubnetSelfLinks {
			if _, exsit := subnetMap[one]; exsit {
				subnetIDs = append(subnetIDs, subnetMap[one].SubnetID)
				cloudSubIDs = append(cloudSubIDs, subnetMap[one].SubnetCloudID)
			}
		}

		disks := make([]corecvm.GcpAttachedDisk, 0)
		for _, v := range one.Disks {
			cloudID, exist := diskMap[v.Source]
			if !exist {
				return fmt.Errorf("cvm: %d not found disk: %s in db", one.Id, v.Source)
			}

			tmp := corecvm.GcpAttachedDisk{
				SelfLink:   v.Source,
				Boot:       v.Boot,
				Index:      v.Index,
				CloudID:    cloudID,
				DeviceName: v.DeviceName,
			}
			disks = append(disks, tmp)
		}

		startTime, err := times.ParseToStdTime(time.RFC3339Nano, one.LastStartTimestamp)
		if err != nil {
			return fmt.Errorf("conv start time failed, err: %v", err)
		}

		createTime, err := times.ParseToStdTime(time.RFC3339Nano, one.CreationTimestamp)
		if err != nil {
			return fmt.Errorf("conv create time failed, err: %v", err)
		}

		imageID := ""
		if id, exsit := imageMap[one.SourceMachineImage]; exsit {
			imageID = id
		}

		priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(one.NetworkInterfaces)
		cvm := dataproto.CvmBatchCreate[corecvm.GcpCvmExtension]{
			CloudID:        fmt.Sprintf("%d", one.Id),
			Name:           one.Name,
			BkBizID:        constant.UnassignedBiz,
			BkHostID:       constant.UnBindBkHostID,
			BkCloudID:      constant.UnassignedBkCloudID,
			AccountID:      accountID,
			Region:         region,
			Zone:           zone,
			CloudVpcIDs:    []string{vpcMap[inVpcSelfLinks[0]].VpcCloudID},
			VpcIDs:         []string{vpcMap[inVpcSelfLinks[0]].VpcID},
			CloudSubnetIDs: cloudSubIDs,
			SubnetIDs:      subnetIDs,
			CloudImageID:   one.SourceMachineImage,
			ImageID:        imageID,
			// gcp镜像是与硬盘绑定的
			OsName:               "",
			Memo:                 converter.ValToPtr(one.Description),
			Status:               one.Status,
			PrivateIPv4Addresses: priIPv4,
			PrivateIPv6Addresses: priIPv6,
			PublicIPv4Addresses:  pubIPv4,
			PublicIPv6Addresses:  pubIPv6,
			MachineType:          gcp.GetMachineType(one.MachineType),
			CloudCreatedTime:     createTime,
			CloudLaunchedTime:    startTime,
			CloudExpiredTime:     "",
			Extension: &corecvm.GcpCvmExtension{
				VpcSelfLinks:             vpcSelfLinks,
				SubnetSelfLinks:          subnetSelfLinks,
				DeletionProtection:       one.DeletionProtection,
				CanIpForward:             one.CanIpForward,
				CloudNetworkInterfaceIDs: cloudNetWorkInterfaceIDs,
				Disks:                    disks,
				SelfLink:                 one.SelfLink,
				CpuPlatform:              one.CpuPlatform,
				Labels:                   one.Labels,
				MinCpuPlatform:           one.MinCpuPlatform,
				StartRestricted:          one.StartRestricted,
				ResourcePolicies:         one.ResourcePolicies,
				ReservationAffinity:      nil,
				Fingerprint:              one.Fingerprint,
				AdvancedMachineFeatures:  nil,
			},
		}

		if one.ReservationAffinity != nil {
			cvm.Extension.ReservationAffinity = &corecvm.GcpReservationAffinity{
				ConsumeReservationType: one.ReservationAffinity.ConsumeReservationType,
				Key:                    one.ReservationAffinity.Key,
				Values:                 one.ReservationAffinity.Values,
			}
		}

		if one.AdvancedMachineFeatures != nil {
			cvm.Extension.AdvancedMachineFeatures = &corecvm.GcpAdvancedMachineFeatures{
				EnableNestedVirtualization: one.AdvancedMachineFeatures.EnableNestedVirtualization,
				EnableUefiNetworking:       one.AdvancedMachineFeatures.EnableUefiNetworking,
				ThreadsPerCore:             one.AdvancedMachineFeatures.ThreadsPerCore,
				VisibleCoreCount:           one.AdvancedMachineFeatures.VisibleCoreCount,
			}
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.GcpCvmExtension]{
		Cvms: lists,
	}

	_, err = cli.dbCli.Gcp.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create gcp cvm failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to create cvm success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateCvm(kt *kit.Kit, accountID string, region string, zone string,
	updateMap map[string]typescvm.GcpCvm) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("cvm updateMap is <= 0, not update")
	}

	lists := make([]dataproto.CvmBatchUpdate[corecvm.GcpCvmExtension], 0)

	updateSlice := make([]typescvm.GcpCvm, 0)
	for _, one := range updateMap {
		updateSlice = append(updateSlice, one)
	}

	vpcMap, subnetMap, diskMap, vpcSelfLinks,
		subnetSelfLinks, imageMap, err := cli.getNIAssResMapBySelfLinkFromNI(kt, accountID, region, zone, updateSlice)
	if err != nil {
		return err
	}

	for id, one := range updateMap {
		inVpcSelfLinks := make([]string, 0)
		inSubnetSelfLinks := make([]string, 0)
		cloudNetWorkInterfaceIDs := make([]string, 0)
		if len(one.NetworkInterfaces) > 0 {
			for _, networkInterface := range one.NetworkInterfaces {
				if networkInterface != nil {
					cloudNetInterfaceID := fmt.Sprintf("%d", one.Id) + "_" + networkInterface.Name
					cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
					inVpcSelfLinks = append(inVpcSelfLinks, networkInterface.Network)
					inSubnetSelfLinks = append(inSubnetSelfLinks, networkInterface.Subnetwork)
				}
			}
		}

		if _, exsit := vpcMap[inVpcSelfLinks[0]]; !exsit {
			return fmt.Errorf("cvm %s can not find vpc", fmt.Sprint(one.Id))
		}

		subnetIDs := make([]string, 0)
		cloudSubIDs := make([]string, 0)
		for _, one := range inSubnetSelfLinks {
			if _, exsit := subnetMap[one]; exsit {
				subnetIDs = append(subnetIDs, subnetMap[one].SubnetID)
				cloudSubIDs = append(cloudSubIDs, subnetMap[one].SubnetCloudID)
			}
		}

		disks := make([]corecvm.GcpAttachedDisk, 0)
		for _, v := range one.Disks {
			cloudID, exist := diskMap[v.Source]
			if !exist {
				return fmt.Errorf("cvm: %d not found disk: %s in db", one.Id, v.Source)
			}

			tmp := corecvm.GcpAttachedDisk{
				SelfLink:   v.Source,
				Boot:       v.Boot,
				Index:      v.Index,
				CloudID:    cloudID,
				DeviceName: v.DeviceName,
			}
			disks = append(disks, tmp)
		}

		startTime, err := times.ParseToStdTime(time.RFC3339Nano, one.LastStartTimestamp)
		if err != nil {
			return fmt.Errorf("conv start time failed, err: %v", err)
		}

		imageID := ""
		if id, exsit := imageMap[one.SourceMachineImage]; exsit {
			imageID = id
		}

		priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(one.NetworkInterfaces)
		cvm := dataproto.CvmBatchUpdate[corecvm.GcpCvmExtension]{
			ID:                   id,
			Name:                 one.Name,
			CloudVpcIDs:          []string{vpcMap[inVpcSelfLinks[0]].VpcCloudID},
			VpcIDs:               []string{vpcMap[inVpcSelfLinks[0]].VpcID},
			CloudSubnetIDs:       cloudSubIDs,
			SubnetIDs:            subnetIDs,
			Memo:                 converter.ValToPtr(one.Description),
			Status:               one.Status,
			PrivateIPv4Addresses: priIPv4,
			PrivateIPv6Addresses: priIPv6,
			PublicIPv4Addresses:  pubIPv4,
			PublicIPv6Addresses:  pubIPv6,
			CloudLaunchedTime:    startTime,
			CloudExpiredTime:     "",
			CloudImageID:         one.SourceMachineImage,
			ImageID:              imageID,
			Extension: &corecvm.GcpCvmExtension{
				VpcSelfLinks:             vpcSelfLinks,
				SubnetSelfLinks:          subnetSelfLinks,
				DeletionProtection:       one.DeletionProtection,
				CanIpForward:             one.CanIpForward,
				CloudNetworkInterfaceIDs: cloudNetWorkInterfaceIDs,
				Disks:                    disks,
				SelfLink:                 one.SelfLink,
				CpuPlatform:              one.CpuPlatform,
				Labels:                   one.Labels,
				MinCpuPlatform:           one.MinCpuPlatform,
				StartRestricted:          one.StartRestricted,
				ResourcePolicies:         one.ResourcePolicies,
				ReservationAffinity:      nil,
				Fingerprint:              one.Fingerprint,
				AdvancedMachineFeatures:  nil,
			},
		}

		if one.ReservationAffinity != nil {
			cvm.Extension.ReservationAffinity = &corecvm.GcpReservationAffinity{
				ConsumeReservationType: one.ReservationAffinity.ConsumeReservationType,
				Key:                    one.ReservationAffinity.Key,
				Values:                 one.ReservationAffinity.Values,
			}
		}

		if one.AdvancedMachineFeatures != nil {
			cvm.Extension.AdvancedMachineFeatures = &corecvm.GcpAdvancedMachineFeatures{
				EnableNestedVirtualization: one.AdvancedMachineFeatures.EnableNestedVirtualization,
				EnableUefiNetworking:       one.AdvancedMachineFeatures.EnableUefiNetworking,
				ThreadsPerCore:             one.AdvancedMachineFeatures.ThreadsPerCore,
				VisibleCoreCount:           one.AdvancedMachineFeatures.VisibleCoreCount,
			}
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.GcpCvmExtension]{
		Cvms: lists,
	}

	if err := cli.dbCli.Gcp.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateCvm failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}
	logs.Infof("[%s] sync cvm to update cvm success, count: %d, ids: %v, rid: %s",
		enumor.Gcp, len(updateMap), maps.Keys(updateMap), kt.Rid)

	return nil
}

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

func (cli *client) deleteCvm(kt *kit.Kit, accountID string, zone string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("cvm delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  delCloudIDs,
	}
	delCvmFromCloud, err := cli.listCvmFromCloud(kt, checkParams, &SyncCvmOption{Zone: zone})
	if err != nil {
		return err
	}

	if len(delCvmFromCloud) > 0 {
		logs.Errorf("[%s] validate cvm not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delCvmFromCloud), kt.Rid)
		return fmt.Errorf("validate cvm not exist failed, before delete")
	}

	deleteReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete cvm failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cvm to delete cvm success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listCvmFromCloud(kt *kit.Kit, params *SyncBaseParams, option *SyncCvmOption) ([]typescvm.GcpCvm,
	error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typescvm.GcpListOption{
		Zone:     option.Zone,
		CloudIDs: params.CloudIDs,
		Page: &adcore.GcpPage{
			PageSize: adcore.GcpQueryLimit,
		},
	}
	result, _, err := cli.cloudCli.ListCvm(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list cvm from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listCvmFromDB(kt *kit.Kit, params *SyncBaseParams, zone string) (
	[]corecvm.Cvm[cvm.GcpCvmExtension], error) {

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
					Field: "zone",
					Op:    filter.Equal.Factory(),
					Value: zone,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list cvm from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) RemoveCvmDeleteFromCloud(kt *kit.Kit, accountID string, zone string) error {
	req := &protocloud.CvmListReq{
		Field: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "zone", Op: filter.Equal.Factory(), Value: zone},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Gcp.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list cvm failed, err: %v, req: %v, rid: %s", enumor.Gcp,
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
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listCvmFromCloud(kt, params, &SyncCvmOption{Zone: zone})
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, fmt.Sprint(one.Id))
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if err := cli.deleteCvm(kt, accountID, zone, cloudIDs); err != nil {
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

func isCvmChange(cloud typescvm.GcpCvm, db corecvm.Cvm[cvm.GcpCvmExtension]) bool {

	if db.CloudID != fmt.Sprintf("%d", cloud.Id) {
		return true
	}

	if db.Name != cloud.Name {
		return true
	}

	vpcSelfLinks := make([]string, 0)
	subnetSelfLinks := make([]string, 0)
	cloudNetWorkInterfaceIDs := make([]string, 0)
	if len(cloud.NetworkInterfaces) > 0 {
		for _, networkInterface := range cloud.NetworkInterfaces {
			if networkInterface != nil {
				cloudNetInterfaceID := fmt.Sprintf("%d", cloud.Id) + "_" + networkInterface.Name
				cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
				vpcSelfLinks = append(vpcSelfLinks, networkInterface.Network)
				subnetSelfLinks = append(subnetSelfLinks, networkInterface.Subnetwork)
			}
		}
	}

	if len(db.Extension.VpcSelfLinks) == 0 || len(vpcSelfLinks) == 0 ||
		(db.Extension.VpcSelfLinks[0] != vpcSelfLinks[0]) {
		return true
	}

	if len(db.Extension.SubnetSelfLinks) == 0 || len(subnetSelfLinks) == 0 ||
		!assert.IsStringSliceEqual(db.Extension.SubnetSelfLinks, subnetSelfLinks) {
		return true
	}

	if db.CloudImageID != cloud.SourceMachineImage {
		return true
	}

	if db.Status != cloud.Status {
		return true
	}

	priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(cloud.NetworkInterfaces)

	if !assert.IsStringSliceEqual(db.PrivateIPv4Addresses, priIPv4) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PublicIPv4Addresses, pubIPv4) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PrivateIPv6Addresses, priIPv6) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PublicIPv6Addresses, pubIPv6) {
		return true
	}

	if db.MachineType != gcp.GetMachineType(cloud.MachineType) {
		return true
	}

	createTime, err := times.ParseToStdTime(time.RFC3339Nano, cloud.CreationTimestamp)
	if err != nil {
		logs.Errorf("[%s] conv CreationTimestamp to std time failed, err: %v", enumor.Gcp, err)
		return true
	}

	if db.CloudCreatedTime != createTime {
		return true
	}

	startTime, err := times.ParseToStdTime(time.RFC3339Nano, cloud.LastStartTimestamp)
	if err != nil {
		logs.Errorf("[%s] conv LastStartTimestamp to std time failed, err: %v", enumor.Gcp, err)
		return true
	}

	if db.CloudLaunchedTime != startTime {
		return true
	}

	if db.Extension.DeletionProtection != cloud.DeletionProtection {
		return true
	}

	if db.Extension.CpuPlatform != cloud.CpuPlatform {
		return true
	}

	if db.Extension.CanIpForward != cloud.CanIpForward {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.CloudNetworkInterfaceIDs, cloudNetWorkInterfaceIDs) {
		return true
	}

	if db.Extension.SelfLink != cloud.SelfLink {
		return true
	}

	if db.Extension.MinCpuPlatform != cloud.MinCpuPlatform {
		return true
	}

	if db.Extension.StartRestricted != cloud.StartRestricted {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.ResourcePolicies, cloud.ResourcePolicies) {
		return true
	}

	if db.Extension.Fingerprint != cloud.Fingerprint {
		return true
	}

	if (db.Extension.ReservationAffinity == nil && cloud.ReservationAffinity != nil) ||
		(db.Extension.ReservationAffinity != nil && cloud.ReservationAffinity == nil) {
		return true
	}

	if db.Extension.ReservationAffinity != nil && cloud.ReservationAffinity != nil {
		if db.Extension.ReservationAffinity.ConsumeReservationType != cloud.ReservationAffinity.ConsumeReservationType {
			return true
		}

		if db.Extension.ReservationAffinity.Key != cloud.ReservationAffinity.Key {
			return true
		}

		if !assert.IsStringSliceEqual(db.Extension.ReservationAffinity.Values, cloud.ReservationAffinity.Values) {
			return true
		}
	}

	if (db.Extension.AdvancedMachineFeatures != nil && cloud.AdvancedMachineFeatures == nil) ||
		(db.Extension.AdvancedMachineFeatures == nil && cloud.AdvancedMachineFeatures != nil) {
		return true
	}

	if db.Extension.AdvancedMachineFeatures != nil && cloud.AdvancedMachineFeatures != nil {
		if db.Extension.AdvancedMachineFeatures.EnableNestedVirtualization != cloud.AdvancedMachineFeatures.EnableNestedVirtualization {
			return true
		}

		if db.Extension.AdvancedMachineFeatures.EnableUefiNetworking != cloud.AdvancedMachineFeatures.EnableUefiNetworking {
			return true
		}

		if db.Extension.AdvancedMachineFeatures.ThreadsPerCore != cloud.AdvancedMachineFeatures.ThreadsPerCore {
			return true
		}

		if db.Extension.AdvancedMachineFeatures.VisibleCoreCount != cloud.AdvancedMachineFeatures.ThreadsPerCore {
			return true
		}
	}

	if !assert.IsStringMapEqual(db.Extension.Labels, cloud.Labels) {
		return true
	}

	for _, dbValue := range db.Extension.Disks {
		isEqual := false
		for _, cloudValue := range cloud.Disks {
			if dbValue.Boot == cloudValue.Boot && dbValue.Index == cloudValue.Index &&
				dbValue.SelfLink == cloudValue.Source && dbValue.DeviceName == cloudValue.DeviceName {
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
