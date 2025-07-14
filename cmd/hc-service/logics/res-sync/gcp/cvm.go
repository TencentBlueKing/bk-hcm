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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
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

// Cvm ...
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

// createCvm creates cvm by addSlice
func (cli *client) createCvm(kt *kit.Kit, accountID string, region string, zone string,
	addSlice []typescvm.GcpCvm) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	vpcMap, subnetMap, diskMap, vpcSelfLinks,
		subnetSelfLinks, imageMap, err := cli.getNIAssResMapBySelfLinkFromNI(kt, accountID, region, zone, addSlice)
	if err != nil {
		return err
	}

	lists, err := buildCvmCreateReqList(addSlice, accountID, region, zone, vpcMap, subnetMap, diskMap, vpcSelfLinks,
		subnetSelfLinks, imageMap)
	if err != nil {
		logs.Errorf("[%s] build cvm create req list failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
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

// buildCvmCreateReqList builds cvm create request list
func buildCvmCreateReqList(addSlice []typescvm.GcpCvm, accountID, region, zone string,
	vpcMap map[string]*common.VpcDB, subnetMap map[string]*SubnetDB, diskMap map[string]string,
	vpcSelfLinks []string, subnetSelfLinks []string, imageMap map[string]string) (
	[]protocloud.CvmBatchCreate[corecvm.GcpCvmExtension], error) {

	lists := make([]dataproto.CvmBatchCreate[corecvm.GcpCvmExtension], 0)
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
			return nil, fmt.Errorf("cvm %s can not find vpc", fmt.Sprint(one.Id))
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
				return nil, fmt.Errorf("cvm: %d not found disk: %s in db", one.Id, v.Source)
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
			return nil, fmt.Errorf("conv start time failed, err: %v", err)
		}

		createTime, err := times.ParseToStdTime(time.RFC3339Nano, one.CreationTimestamp)
		if err != nil {
			return nil, fmt.Errorf("conv create time failed, err: %v", err)
		}

		imageID := ""
		if id, exsit := imageMap[one.SourceMachineImage]; exsit {
			imageID = id
		}

		req := buildCvmCreateReq(one, imageID, startTime, createTime, accountID, region, zone,
			vpcMap[inVpcSelfLinks[0]].VpcCloudID, vpcMap[inVpcSelfLinks[0]].VpcID, cloudSubIDs, subnetIDs, vpcSelfLinks,
			subnetSelfLinks, cloudNetWorkInterfaceIDs, disks)
		lists = append(lists, req)
	}

	return lists, nil
}

// buildCvmCreateReq builds cvm create request
func buildCvmCreateReq(one typescvm.GcpCvm, imageID, startTime, createTime, accountID, region, zone,
	vpcCloudID, vpcID string, cloudSubIDs, subnetIDs, vpcSelfLinks, subnetSelfLinks, cloudNetWorkInterfaceIDs []string,
	disks []corecvm.GcpAttachedDisk) protocloud.CvmBatchCreate[corecvm.GcpCvmExtension] {

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
		CloudVpcIDs:    []string{vpcCloudID},
		VpcIDs:         []string{vpcID},
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
	return cvm
}

// updateCvm updates cvm by updateMap
func (cli *client) updateCvm(kt *kit.Kit, accountID string, region string, zone string,
	updateMap map[string]typescvm.GcpCvm) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("cvm updateMap is <= 0, not update")
	}

	updateSlice := make([]typescvm.GcpCvm, 0)
	for _, one := range updateMap {
		updateSlice = append(updateSlice, one)
	}
	vpcMap, subnetMap, diskMap, vpcSelfLinks,
		subnetSelfLinks, imageMap, err := cli.getNIAssResMapBySelfLinkFromNI(kt, accountID, region, zone, updateSlice)
	if err != nil {
		return err
	}

	lists, err := buildCvmUpdateReqList(updateMap, vpcMap, subnetMap, diskMap, vpcSelfLinks,
		subnetSelfLinks, imageMap)
	if err != nil {
		logs.Errorf("[%s] build cvm create req list failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
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

// buildCvmUpdateReqList builds cvm update request list
func buildCvmUpdateReqList(updateMap map[string]typescvm.GcpCvm, vpcMap map[string]*common.VpcDB,
	subnetMap map[string]*SubnetDB, diskMap map[string]string, vpcSelfLinks []string, subnetSelfLinks []string,
	imageMap map[string]string) ([]protocloud.CvmBatchUpdateWithExtension[corecvm.GcpCvmExtension], error) {

	lists := make([]dataproto.CvmBatchUpdateWithExtension[corecvm.GcpCvmExtension], 0)
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
			return nil, fmt.Errorf("cvm %s can not find vpc", fmt.Sprint(one.Id))
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
				return nil, fmt.Errorf("cvm: %d not found disk: %s in db", one.Id, v.Source)
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
			return nil, fmt.Errorf("conv start time failed, err: %v", err)
		}

		imageID := ""
		if id, exsit := imageMap[one.SourceMachineImage]; exsit {
			imageID = id
		}

		req := buildCvmUpdateReq(id, one, vpcMap[inVpcSelfLinks[0]].VpcCloudID, vpcMap[inVpcSelfLinks[0]].VpcID,
			imageID, startTime, cloudSubIDs, subnetIDs, vpcSelfLinks, subnetSelfLinks, cloudNetWorkInterfaceIDs, disks)
		lists = append(lists, req)
	}
	return lists, nil
}

// buildCvmUpdateReq builds cvm update request
func buildCvmUpdateReq(id string, one typescvm.GcpCvm, vpcCloudID, vpcID, imageID, startTime string, cloudSubIDs,
	subnetIDs, vpcSelfLinks, subnetSelfLinks, cloudNetWorkInterfaceIDs []string,
	disks []corecvm.GcpAttachedDisk) protocloud.CvmBatchUpdateWithExtension[corecvm.GcpCvmExtension] {

	priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(one.NetworkInterfaces)
	cvm := dataproto.CvmBatchUpdateWithExtension[corecvm.GcpCvmExtension]{
		CvmBatchUpdate: dataproto.CvmBatchUpdate{
			ID:                   id,
			Name:                 one.Name,
			CloudVpcIDs:          []string{vpcCloudID},
			VpcIDs:               []string{vpcID},
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
		},
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
	return cvm
}

// deleteCvm deletes cvm by delCloudIDs
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

// listCvmFromCloud lists cvm from cloud by params
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

// listCvmFromDB lists cvm from db by params
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

// RemoveCvmDeleteFromCloud ...
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
			if len(cloudIDs) > 0 {
				if err := cli.deleteCvm(kt, accountID, zone, cloudIDs); err != nil {
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
